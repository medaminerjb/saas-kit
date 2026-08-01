// Package provider implements the OIDC Provider (OpenID Connect server)
// using the zitadel/oidc/v3 library. SaaSKit acts as an OIDC identity provider.
//
// This package contains:
//   - models.go:    Internal OIDC models (AuthRequest, Token, RefreshToken, Client adapter)
//   - storage.go:   op.Storage implementation backed by PostgreSQL
//   - provider.go:  OIDC provider setup and HTTP handler mounting
//   - login.go:     Login/consent UI handlers
package provider
