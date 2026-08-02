// Package main is the entry point for the SaaSKit server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/saaskit/saaskit/internal/audit"
	"github.com/saaskit/saaskit/internal/config"
	idcrypto "github.com/saaskit/saaskit/internal/identity/crypto"
	"github.com/saaskit/saaskit/internal/identity/handler"
	"github.com/saaskit/saaskit/internal/identity/repository"
	"github.com/saaskit/saaskit/internal/identity/service"
	oidcprovider "github.com/saaskit/saaskit/internal/oidc/provider"
	"github.com/saaskit/saaskit/internal/oidc/relyingparty"
	platformcrypto "github.com/saaskit/saaskit/internal/platform/crypto"
	"github.com/saaskit/saaskit/internal/platform/database"
	"github.com/saaskit/saaskit/internal/platform/events"
	"github.com/saaskit/saaskit/internal/platform/jobs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ─── Configuration ────────────────────────────────
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ─── Logger ───────────────────────────────────────
	logger := setupLogger(cfg.Log)
	logger.Info("starting SaaSKit",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Port),
	)

	// ─── Database ─────────────────────────────────────
	pool, err := database.Connect(ctx, database.Config{
		URL:      cfg.Database.DSN(),
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()
	logger.Info("database connected")

	// ─── Run Migrations ───────────────────────────────
	if cfg.IsDevelopment() {
		logger.Info("running migrations (dev mode)")
		if err := runMigrations(cfg.Database.DSN()); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
	}

	// ─── Signing Keys ─────────────────────────────────
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(
		cfg.JWT.KeyPath,
		cfg.JWT.Algorithm,
		cfg.IsDevelopment(),
	)
	if err != nil {
		return fmt.Errorf("loading signing keys: %w", err)
	}
	logger.Info("signing keys loaded", slog.String("algorithm", keyPair.Algorithm))

	// ─── Crypto ───────────────────────────────────────
	hasher := idcrypto.NewHasher(idcrypto.Argon2Params{
		Memory:      cfg.Argon2.Memory,
		Iterations:  cfg.Argon2.Iterations,
		Parallelism: cfg.Argon2.Parallelism,
		SaltLength:  cfg.Argon2.SaltLength,
		KeyLength:   cfg.Argon2.KeyLength,
	})

	tokenHasher := idcrypto.NewTokenHasher(cfg.ServerSecret)

	// ─── Envelope Encryption ─────────────────────────
	var envelope *platformcrypto.Envelope
	if cfg.EncryptionMasterKey != "" {
		envelope, err = platformcrypto.NewEnvelope(cfg.EncryptionMasterKey)
		if err != nil {
			return fmt.Errorf("initializing envelope encryption: %w", err)
		}
		logger.Info("envelope encryption initialized")
	}

	// ─── Event Publisher ──────────────────────────────
	logPublisher := events.NewLogPublisher(logger)
	auditPublisher := audit.NewAuditPublisher(logger, pool)
	publisher := events.NewMultiPublisher(logPublisher, auditPublisher)

	// ─── Repositories ─────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	tokenRepo := repository.NewTokenRepo(pool)

	// ─── Services ─────────────────────────────────────
	tokenService := service.NewTokenService(service.TokenServiceConfig{
		PrivateKey: keyPair.PrivateKey,
		PublicKey:  keyPair.PublicKey,
		Algorithm:  keyPair.Algorithm,
		KeyID:      keyPair.KeyID,
		Issuer:     cfg.JWT.Issuer,
		AccessTTL:  cfg.JWT.AccessTTL,
	}, logger)

	authService := service.NewAuthService(service.AuthServiceConfig{
		Users:        userRepo,
		Sessions:     sessionRepo,
		Tokens:       tokenRepo,
		Hasher:       hasher,
		TokenHasher:  tokenHasher,
		TokenService: tokenService,
		Publisher:    publisher,
		RefreshTTL:   cfg.JWT.RefreshTTL,
	}, logger)

	userService := service.NewUserService(userRepo, publisher, logger)

	identityManager := service.NewIdentityManager(authService, userService, tokenService, logger)

	// ─── OIDC Provider ────────────────────────────────
	oidcStorage := oidcprovider.NewStorage(oidcprovider.StorageConfig{
		Pool:     pool,
		UserRepo: userRepo,
		Hasher:   hasher,
		Envelope: envelope,
		KeyPair:  keyPair,
		Logger:   logger,
	})

	oidcHandler, err := oidcprovider.SetupProvider(cfg.BaseURL, oidcStorage, logger)
	if err != nil {
		return fmt.Errorf("setting up OIDC provider: %w", err)
	}

	// ─── Social Login ─────────────────────────────────
	socialRouter := chi.NewRouter()
	socialHandler := relyingparty.NewHandler(relyingparty.Config{
		GoogleClientID:     cfg.OAuth.Google.ClientID,
		GoogleClientSecret: cfg.OAuth.Google.ClientSecret,
		GitHubClientID:     cfg.OAuth.GitHub.ClientID,
		GitHubClientSecret: cfg.OAuth.GitHub.ClientSecret,
		BaseURL:            cfg.BaseURL,
		UserRepo:           userRepo,
		IdentityManager:    identityManager,
		Publisher:          publisher,
		Logger:             logger,
	})
	socialHandler.Routes(socialRouter)

	// ─── HTTP Server ──────────────────────────────────
	router := handler.NewRouter(handler.RouterConfig{
		Identity:     identityManager,
		Pool:         pool,
		Logger:       logger,
		OIDCProvider: oidcHandler,
		SocialLogin:  socialRouter,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Background Jobs ──────────────────────────────
	scheduler := jobs.NewScheduler(logger)
	scheduler.Register(&cleanupJob{
		name:     "cleanup_sessions",
		interval: 1 * time.Hour,
		fn:       func(ctx context.Context) error { _, err := sessionRepo.DeleteExpired(ctx); return err },
	})
	scheduler.Register(&cleanupJob{
		name:     "cleanup_tokens",
		interval: 1 * time.Hour,
		fn:       func(ctx context.Context) error { _, err := tokenRepo.DeleteExpired(ctx); return err },
	})
	scheduler.Start(ctx)

	// ─── Start Server ─────────────────────────────────
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// ─── Graceful Shutdown ────────────────────────────
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		scheduler.Wait()
		logger.Info("server stopped")
	}

	return nil
}

// cleanupJob is a simple background job adapter.
type cleanupJob struct {
	name     string
	interval time.Duration
	fn       func(ctx context.Context) error
}

func (j *cleanupJob) Name() string            { return j.name }
func (j *cleanupJob) Interval() time.Duration { return j.interval }
func (j *cleanupJob) Run(ctx context.Context) error { return j.fn(ctx) }

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var handler slog.Handler
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runMigrations(databaseURL string) error {
	// In development, we use goose programmatically.
	// For production, use `make migrate-up` or the goose CLI.
	// This is a placeholder — will wire goose.Up() when dependency is added.
	_ = databaseURL
	return nil
}
