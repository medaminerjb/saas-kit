package provider

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// SetupProvider creates and configures the OpenID Connect Provider.
func SetupProvider(issuer string, storage *Storage, logger *slog.Logger) (http.Handler, error) {
	config := &op.Config{
		CodeMethodS256:        true,
		AuthMethodPost:        true,
		GrantTypeRefreshToken: true,
	}

	provider, err := op.NewProvider(
		config,
		storage,
		op.StaticIssuer(issuer),
		op.WithAllowInsecure(), // Allow HTTP in development
	)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	// Mount the standard OIDC endpoints from zitadel/oidc
	r.Mount("/", provider)

	// Mount the login/consent UI
	r.Get("/login", LoginHandler(storage))
	r.Post("/login", LoginSubmitHandler(storage, provider))

	logger.Info("OIDC provider initialized", slog.String("issuer", issuer))

	return r, nil
}
