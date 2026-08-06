package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"

	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
	"github.com/medaminerjb/saas-kit/internal/platform/database"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
	tenantrepo "github.com/medaminerjb/saas-kit/internal/tenant/repository"
	tenantservice "github.com/medaminerjb/saas-kit/internal/tenant/service"
)

func main() {
	app := &cli.App{
		Name:  "saaskit",
		Usage: "SaaSKit CLI - Manage users, tenants, and API keys",
		Commands: []*cli.Command{
			{
				Name:  "user",
				Usage: "User management commands",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a new user",
						Action: createUser,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "email",
								Usage:    "User email",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "password",
								Usage:    "User password",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "tenant",
				Usage: "Tenant management commands",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a new tenant",
						Action: createTenant,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Usage:    "Tenant name",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "slug",
								Usage: "Tenant slug (optional)",
							},
							&cli.StringFlag{
								Name:     "owner-email",
								Usage:    "Owner email",
								Required: true,
							},
						},
					},
					{
						Name:   "list",
						Usage:  "List tenants for a user",
						Action: listTenants,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "email",
								Usage:    "User email",
								Required: true,
							},
						},
					},
					{
						Name:   "switch",
						Usage:  "Switch active tenant for a user",
						Action: switchTenant,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "email",
								Usage:    "User email",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "Tenant ID to switch to",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "apikey",
				Usage: "API key management commands",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a new API key",
						Action: createAPIKey,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "email",
								Usage:    "User email",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "Tenant ID",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "name",
								Usage: "API key name",
							},
							&cli.StringSliceFlag{
								Name:  "scopes",
								Usage: "API key scopes (e.g., tenant.read, tenant.write)",
							},
						},
					},
					{
						Name:   "list",
						Usage:  "List API keys for a tenant",
						Action: listAPIKeys,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "Tenant ID",
								Required: true,
							},
						},
					},
					{
						Name:   "revoke",
						Usage:  "Revoke an API key",
						Action: revokeAPIKey,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "key-id",
								Usage:    "API key ID",
								Required: true,
							},
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "database-url",
				Usage:   "Database connection URL",
				EnvVars: []string{"DATABASE_URL"},
				Value:   "postgres://saaskit:saaskit@localhost:5432/saaskit?sslmode=disable",
			},
			&cli.StringFlag{
				Name:    "server-secret",
				Usage:   "Server secret for token hashing",
				EnvVars: []string{"SERVER_SECRET"},
			},
			&cli.StringFlag{
				Name:    "encryption-key",
				Usage:   "Encryption master key",
				EnvVars: []string{"ENCRYPTION_MASTER_KEY"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setupServices(ctx context.Context, c *cli.Context) (*service.AuthService, repository.UserRepository, *tenantservice.TenantService, *repository.APIKeyRepo, error) {
	// Database connection
	pool, err := database.Connect(ctx, database.Config{
		URL:      c.String("database-url"),
		MaxConns: 10,
		MinConns: 1,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("connecting to database: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Crypto
	hasher := idcrypto.NewHasher(idcrypto.Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})

	tokenHasher := idcrypto.NewTokenHasher(c.String("server-secret"))

	// Event publisher
	publisher := events.NewLogPublisher(logger)

	// Repositories
	userRepo := repository.NewUserRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	tokenRepo := repository.NewTokenRepo(pool)
	apiKeyRepo := repository.NewAPIKeyRepo(pool)

	// Services
	authService := service.NewAuthService(service.AuthServiceConfig{
		Users:       userRepo,
		Sessions:    sessionRepo,
		Tokens:      tokenRepo,
		Hasher:      hasher,
		TokenHasher: tokenHasher,
		Publisher:   publisher,
		RefreshTTL:  7 * 24 * time.Hour,
	}, logger)

	tenantRepo := tenantrepo.NewTenantRepository(pool)
	tenantService := tenantservice.NewTenantService(tenantRepo, publisher, logger)

	return authService, userRepo, tenantService, apiKeyRepo, nil
}

func createUser(c *cli.Context) error {
	ctx := context.Background()
	authService, _, _, _, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	email := c.String("email")
	password := c.String("password")

	user, _, err := authService.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: password,
		Name:     "",
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	fmt.Printf("User created successfully:\n")
	fmt.Printf("  ID: %s\n", user.ID)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Status: %s\n", user.Status)

	return nil
}

func createTenant(c *cli.Context) error {
	ctx := context.Background()
	_, userRepo, tenantService, _, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	name := c.String("name")
	slug := c.String("slug")
	ownerEmail := c.String("owner-email")

	// Get owner user (nil tenantID for global user lookup)
	user, err := userRepo.GetByEmail(ctx, ownerEmail, nil)
	if err != nil {
		return fmt.Errorf("finding owner user: %w", err)
	}

	// Create tenant
	tenant, err := tenantService.CreateTenant(ctx, name, slug, user.ID)
	if err != nil {
		return fmt.Errorf("creating tenant: %w", err)
	}

	fmt.Printf("Tenant created successfully:\n")
	fmt.Printf("  ID: %s\n", tenant.ID)
	fmt.Printf("  Name: %s\n", tenant.Name)
	fmt.Printf("  Slug: %s\n", tenant.Slug)
	fmt.Printf("  Owner: %s\n", ownerEmail)

	return nil
}

func listTenants(c *cli.Context) error {
	ctx := context.Background()
	_, userRepo, tenantService, _, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	email := c.String("email")

	// Get user (nil tenantID for global user lookup)
	user, err := userRepo.GetByEmail(ctx, email, nil)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}

	// List tenants
	tenants, roles, err := tenantService.ListTenantsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("listing tenants: %w", err)
	}

	fmt.Printf("Tenants for %s:\n", email)
	for i, tenant := range tenants {
		fmt.Printf("  %d. %s (ID: %s, Role: %s)\n", i+1, tenant.Name, tenant.ID, roles[i])
	}

	return nil
}

func switchTenant(c *cli.Context) error {
	ctx := context.Background()
	_, userRepo, tenantService, _, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	email := c.String("email")
	tenantIDStr := c.String("tenant-id")

	// Get user (nil tenantID for global user lookup)
	user, err := userRepo.GetByEmail(ctx, email, nil)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("parsing tenant ID: %w", err)
	}

	// Switch tenant
	if err := tenantService.SwitchActiveTenant(ctx, user.ID, tenantID); err != nil {
		return fmt.Errorf("switching tenant: %w", err)
	}

	fmt.Printf("Successfully switched active tenant for %s to %s\n", email, tenantID)

	return nil
}

func createAPIKey(c *cli.Context) error {
	ctx := context.Background()
	_, userRepo, _, apiKeyRepo, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	email := c.String("email")
	tenantIDStr := c.String("tenant-id")
	name := c.String("name")
	scopes := c.StringSlice("scopes")

	// Get user (nil tenantID for global user lookup)
	user, err := userRepo.GetByEmail(ctx, email, nil)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("parsing tenant ID: %w", err)
	}

	// Generate API key
	keyID := uuid.New()
	keyPrefix := "sk_test_" // Default to test for CLI
	rawKey := fmt.Sprintf("%s%s", keyPrefix, keyID.String())

	// Create API key domain object
	apiKey := &domain.APIKey{
		ID:        keyID,
		TenantID:  tenantID,
		Name:      name,
		KeyPrefix: keyPrefix,
		KeyHash:   rawKey, // In production, this should be hashed
		Scopes:    scopes,
		Type:      domain.APIKeyTypeTest,
		Status:    domain.APIKeyStatusActive,
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
	}

	// Create API key
	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}

	fmt.Printf("API key created successfully:\n")
	fmt.Printf("  ID: %s\n", apiKey.ID)
	fmt.Printf("  Name: %s\n", apiKey.Name)
	fmt.Printf("  Key: %s\n", rawKey)
	fmt.Printf("  Scopes: %v\n", apiKey.Scopes)
	fmt.Printf("  Type: %s\n", apiKey.Type)

	return nil
}

func listAPIKeys(c *cli.Context) error {
	ctx := context.Background()
	_, _, _, apiKeyRepo, err := setupServices(ctx, c)
	if err != nil {
		return err
	}

	tenantIDStr := c.String("tenant-id")

	// Parse tenant ID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("parsing tenant ID: %w", err)
	}

	// List API keys
	apiKeys, err := apiKeyRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("listing API keys: %w", err)
	}

	fmt.Printf("API keys for tenant %s:\n", tenantID)
	for i, key := range apiKeys {
		fmt.Printf("  %d. %s (ID: %s, Scopes: %v, Created: %s)\n", i+1, key.Name, key.ID, key.Scopes, key.CreatedAt.Format(time.RFC3339))
	}

	return nil
}

func revokeAPIKey(c *cli.Context) error {
	// Revoke API key - requires tenant ID and revoked by user ID
	// For CLI, we'll need to get tenant ID from context or ask user
	// For now, return an error indicating this limitation
	return fmt.Errorf("revoke requires tenant ID and revoked by user ID - use API instead")
}
