package provider

import (
	"time"

	"golang.org/x/text/language"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// ───────────────────────────────────────────────────────
// AuthRequest — implements op.AuthRequest
// ───────────────────────────────────────────────────────

// AuthRequest is the internal representation of an OIDC authorization request.
type AuthRequest struct {
	ID            string
	CreationDate  time.Time
	ClientID      string
	RedirectURI   string
	State         string
	Nonce         string
	ResponseType  oidc.ResponseType
	ResponseMode  oidc.ResponseMode
	Scopes        []string
	Prompt        []string
	UILocales     []language.Tag
	LoginHint     string
	MaxAuthAge    *time.Duration
	UserID        string
	CodeChallenge *CodeChallenge

	IsDone   bool
	AuthTime time.Time
}

func (a *AuthRequest) GetID() string                          { return a.ID }
func (a *AuthRequest) GetACR() string                         { return "" }
func (a *AuthRequest) GetAudience() []string                  { return []string{a.ClientID} }
func (a *AuthRequest) GetAuthTime() time.Time                 { return a.AuthTime }
func (a *AuthRequest) GetClientID() string                    { return a.ClientID }
func (a *AuthRequest) GetNonce() string                       { return a.Nonce }
func (a *AuthRequest) GetRedirectURI() string                 { return a.RedirectURI }
func (a *AuthRequest) GetResponseType() oidc.ResponseType     { return a.ResponseType }
func (a *AuthRequest) GetResponseMode() oidc.ResponseMode     { return a.ResponseMode }
func (a *AuthRequest) GetScopes() []string                    { return a.Scopes }
func (a *AuthRequest) GetState() string                       { return a.State }
func (a *AuthRequest) GetSubject() string                     { return a.UserID }
func (a *AuthRequest) Done() bool                             { return a.IsDone }

func (a *AuthRequest) GetAMR() []string {
	if a.IsDone {
		return []string{"pwd"}
	}
	return nil
}

func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	if a.CodeChallenge == nil {
		return nil
	}
	method := oidc.CodeChallengeMethodPlain
	if a.CodeChallenge.Method == "S256" {
		method = oidc.CodeChallengeMethodS256
	}
	return &oidc.CodeChallenge{
		Challenge: a.CodeChallenge.Challenge,
		Method:    method,
	}
}

// CodeChallenge holds PKCE challenge data.
type CodeChallenge struct {
	Challenge string
	Method    string
}

// ───────────────────────────────────────────────────────
// OIDCToken — opaque access token record
// ───────────────────────────────────────────────────────

// OIDCToken represents a stored access token.
type OIDCToken struct {
	ID             string
	ApplicationID  string
	RefreshTokenID string
	Subject        string
	Audience       []string
	Scopes         []string
	Expiration     time.Time
}

// ───────────────────────────────────────────────────────
// OIDCRefreshToken — refresh token record
// ───────────────────────────────────────────────────────

// OIDCRefreshToken represents a stored refresh token.
type OIDCRefreshToken struct {
	ID            string
	Token         string
	AuthTime      time.Time
	AMR           []string
	ApplicationID string
	UserID        string
	Audience      []string
	Scopes        []string
	Expiration    time.Time
	AccessToken   string
}

// RefreshTokenRequest wraps an OIDCRefreshToken to implement op.RefreshTokenRequest.
type RefreshTokenRequest struct {
	*OIDCRefreshToken
}

func (r *RefreshTokenRequest) GetAMR() []string           { return r.AMR }
func (r *RefreshTokenRequest) GetAudience() []string      { return r.Audience }
func (r *RefreshTokenRequest) GetAuthTime() time.Time     { return r.AuthTime }
func (r *RefreshTokenRequest) GetClientID() string        { return r.ApplicationID }
func (r *RefreshTokenRequest) GetScopes() []string        { return r.Scopes }
func (r *RefreshTokenRequest) GetSubject() string         { return r.UserID }
func (r *RefreshTokenRequest) SetCurrentScopes(s []string) { r.Scopes = s }

// ───────────────────────────────────────────────────────
// ClientAdapter — wraps DB client to implement op.Client
// ───────────────────────────────────────────────────────

// ClientAdapter adapts our database OIDC client to the op.Client interface.
type ClientAdapter struct {
	ID                  string
	Secret              string
	RedirectURIs_       []string
	AppType             op.ApplicationType
	AuthMethod_         oidc.AuthMethod
	ResponseTypes_      []oidc.ResponseType
	GrantTypes_         []oidc.GrantType
	AccessTokenType_    op.AccessTokenType
	LoginURL_           func(string) string
	DevMode_            bool
	IDTokenClaimAssert  bool
	ClockSkew_          time.Duration
}

func (c *ClientAdapter) GetID() string                         { return c.ID }
func (c *ClientAdapter) RedirectURIs() []string                { return c.RedirectURIs_ }
func (c *ClientAdapter) PostLogoutRedirectURIs() []string      { return nil }
func (c *ClientAdapter) ApplicationType() op.ApplicationType   { return c.AppType }
func (c *ClientAdapter) AuthMethod() oidc.AuthMethod           { return c.AuthMethod_ }
func (c *ClientAdapter) ResponseTypes() []oidc.ResponseType    { return c.ResponseTypes_ }
func (c *ClientAdapter) GrantTypes() []oidc.GrantType          { return c.GrantTypes_ }
func (c *ClientAdapter) LoginURL(id string) string             { return c.LoginURL_(id) }
func (c *ClientAdapter) AccessTokenType() op.AccessTokenType   { return c.AccessTokenType_ }
func (c *ClientAdapter) IDTokenLifetime() time.Duration        { return 1 * time.Hour }
func (c *ClientAdapter) DevMode() bool                         { return c.DevMode_ }
func (c *ClientAdapter) IDTokenUserinfoClaimsAssertion() bool  { return c.IDTokenClaimAssert }
func (c *ClientAdapter) ClockSkew() time.Duration              { return c.ClockSkew_ }

func (c *ClientAdapter) RestrictAdditionalIdTokenScopes() func([]string) []string {
	return func(scopes []string) []string { return scopes }
}

func (c *ClientAdapter) RestrictAdditionalAccessTokenScopes() func([]string) []string {
	return func(scopes []string) []string { return scopes }
}

func (c *ClientAdapter) IsScopeAllowed(scope string) bool {
	return false
}

// ───────────────────────────────────────────────────────
// Helpers
// ───────────────────────────────────────────────────────

func promptToInternal(oidcPrompt oidc.SpaceDelimitedArray) []string {
	prompts := make([]string, 0, len(oidcPrompt))
	for _, p := range oidcPrompt {
		switch p {
		case oidc.PromptNone, oidc.PromptLogin, oidc.PromptConsent, oidc.PromptSelectAccount:
			prompts = append(prompts, p)
		}
	}
	return prompts
}

func maxAgeToInternal(maxAge *uint) *time.Duration {
	if maxAge == nil {
		return nil
	}
	dur := time.Duration(*maxAge) * time.Second
	return &dur
}

func authRequestFromOIDC(req *oidc.AuthRequest, userID string) *AuthRequest {
	var challenge *CodeChallenge
	if req.CodeChallenge != "" {
		challenge = &CodeChallenge{
			Challenge: req.CodeChallenge,
			Method:    string(req.CodeChallengeMethod),
		}
	}
	return &AuthRequest{
		CreationDate:  time.Now(),
		ClientID:      req.ClientID,
		RedirectURI:   req.RedirectURI,
		State:         req.State,
		Nonce:         req.Nonce,
		ResponseType:  req.ResponseType,
		ResponseMode:  req.ResponseMode,
		Scopes:        req.Scopes,
		Prompt:        promptToInternal(req.Prompt),
		UILocales:     req.UILocales,
		LoginHint:     req.LoginHint,
		MaxAuthAge:    maxAgeToInternal(req.MaxAge),
		UserID:        userID,
		CodeChallenge: challenge,
	}
}
