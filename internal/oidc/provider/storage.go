package provider

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"sync"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"

	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	platformcrypto "github.com/medaminerjb/saas-kit/internal/platform/crypto"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Verify interface compliance at compile time.
var _ op.Storage = (*Storage)(nil)

// Storage implements op.Storage, bridging zitadel/oidc to SaaSKit's PostgreSQL backend.
type Storage struct {
	mu            sync.Mutex
	authRequests  map[string]*AuthRequest
	codes         map[string]string
	tokens        map[string]*OIDCToken
	refreshTokens map[string]*OIDCRefreshToken

	pool         *pgxpool.Pool
	userRepo     repository.UserRepository
	hasher       *idcrypto.Hasher
	envelope     *platformcrypto.Envelope
	keyPair      *idcrypto.KeyPair
	sigKey       *storageSigningKey
	logger       *slog.Logger
}

// StorageConfig holds Storage dependencies.
type StorageConfig struct {
	Pool     *pgxpool.Pool
	UserRepo repository.UserRepository
	Hasher   *idcrypto.Hasher
	Envelope *platformcrypto.Envelope
	KeyPair  *idcrypto.KeyPair
	Logger   *slog.Logger
}

// NewStorage creates a new OIDC storage adapter.
func NewStorage(cfg StorageConfig) *Storage {
	return &Storage{
		authRequests:  make(map[string]*AuthRequest),
		codes:         make(map[string]string),
		tokens:        make(map[string]*OIDCToken),
		refreshTokens: make(map[string]*OIDCRefreshToken),
		pool:          cfg.Pool,
		userRepo:      cfg.UserRepo,
		hasher:        cfg.Hasher,
		envelope:      cfg.Envelope,
		keyPair:       cfg.KeyPair,
		sigKey:        newStorageSigningKey(cfg.KeyPair),
		logger:        cfg.Logger,
	}
}

// ───────────────────────────────────────────────────────
// Auth Request lifecycle
// ───────────────────────────────────────────────────────

func (s *Storage) CreateAuthRequest(ctx context.Context, authReq *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(authReq.Prompt) == 1 && authReq.Prompt[0] == "none" {
		return nil, oidc.ErrLoginRequired()
	}

	request := authRequestFromOIDC(authReq, userID)
	request.ID = uuid.NewString()
	s.authRequests[request.ID] = request
	return request, nil
}

func (s *Storage) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, ok := s.authRequests[id]
	if !ok {
		return nil, fmt.Errorf("auth request not found")
	}
	return req, nil
}

func (s *Storage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	requestID, ok := func() (string, bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		id, ok := s.codes[code]
		return id, ok
	}()
	if !ok {
		return nil, fmt.Errorf("code invalid or expired")
	}
	return s.AuthRequestByID(ctx, requestID)
}

func (s *Storage) SaveAuthCode(ctx context.Context, id string, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code] = id
	return nil
}

func (s *Storage) DeleteAuthRequest(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.authRequests, id)
	for code, reqID := range s.codes {
		if reqID == id {
			delete(s.codes, code)
			break
		}
	}
	return nil
}

// ───────────────────────────────────────────────────────
// Tokens
// ───────────────────────────────────────────────────────

func (s *Storage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (string, time.Time, error) {
	var appID string
	switch req := request.(type) {
	case *AuthRequest:
		appID = req.ClientID
	case op.TokenExchangeRequest:
		appID = req.GetClientID()
	}

	token := s.createOIDCToken(appID, "", request.GetSubject(), request.GetAudience(), request.GetScopes())
	return token.ID, token.Expiration, nil
}

func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (string, string, time.Time, error) {
	appID, authTime, amr := getInfoFromRequest(request)

	if currentRefreshToken == "" {
		refreshID := uuid.NewString()
		token := s.createOIDCToken(appID, refreshID, request.GetSubject(), request.GetAudience(), request.GetScopes())
		refresh := s.createOIDCRefreshToken(token, amr, authTime)
		return token.ID, refresh, token.Expiration, nil
	}

	newRefreshID := uuid.NewString()
	token := s.createOIDCToken(appID, newRefreshID, request.GetSubject(), request.GetAudience(), request.GetScopes())

	if err := s.rotateRefreshToken(currentRefreshToken, newRefreshID, token.ID); err != nil {
		return "", "", time.Time{}, err
	}

	return token.ID, newRefreshID, token.Expiration, nil
}

func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.refreshTokens[refreshToken]
	if !ok {
		return nil, fmt.Errorf("invalid refresh_token")
	}
	return &RefreshTokenRequest{token}, nil
}

func (s *Storage) TerminateSession(ctx context.Context, userID string, clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, token := range s.tokens {
		if token.ApplicationID == clientID && token.Subject == userID {
			delete(s.tokens, token.ID)
			delete(s.refreshTokens, token.RefreshTokenID)
		}
	}
	return nil
}

func (s *Storage) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, ok := s.refreshTokens[token]
	if !ok {
		return "", "", op.ErrInvalidRefreshToken
	}
	return rt.UserID, rt.ID, nil
}

func (s *Storage) RevokeToken(ctx context.Context, tokenIDOrToken string, userID string, clientID string) *oidc.Error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if at, ok := s.tokens[tokenIDOrToken]; ok {
		if at.ApplicationID != clientID {
			return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
		}
		delete(s.tokens, at.ID)
		return nil
	}

	if rt, ok := s.refreshTokens[tokenIDOrToken]; ok {
		if rt.ApplicationID != clientID {
			return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
		}
		delete(s.refreshTokens, rt.ID)
		delete(s.tokens, rt.AccessToken)
		return nil
	}

	return nil
}

// ───────────────────────────────────────────────────────
// Client
// ───────────────────────────────────────────────────────

func (s *Storage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, client_name, client_secret_hash, redirect_uris, response_types, grant_types,
		       application_type, token_endpoint_auth_method, pkce_required, access_token_type, is_active
		FROM oidc_clients WHERE id = $1 AND is_active = true`, clientID)

	var (
		id, name       string
		secretHash     *string
		redirectURIs   []string
		responseTypes  []string
		grantTypes     []string
		appType        string
		authMethod     string
		pkceRequired   bool
		tokenType      string
		isActive       bool
	)

	if err := row.Scan(&id, &name, &secretHash, &redirectURIs, &responseTypes, &grantTypes,
		&appType, &authMethod, &pkceRequired, &tokenType, &isActive); err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	return &ClientAdapter{
		ID:               id,
		Secret:           stringOrEmpty(secretHash),
		RedirectURIs_:    redirectURIs,
		AppType:          toAppType(appType),
		AuthMethod_:      toAuthMethod(authMethod),
		ResponseTypes_:   toResponseTypes(responseTypes),
		GrantTypes_:      toGrantTypes(grantTypes),
		AccessTokenType_: toAccessTokenType(tokenType),
		LoginURL_:        func(reqID string) string { return "/login?authRequestID=" + reqID },
		DevMode_:         false,
	}, nil
}

func (s *Storage) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
	row := s.pool.QueryRow(ctx, `SELECT client_secret_hash FROM oidc_clients WHERE id = $1 AND is_active = true`, clientID)

	var secretHash *string
	if err := row.Scan(&secretHash); err != nil {
		return fmt.Errorf("client not found")
	}
	if secretHash == nil || *secretHash == "" {
		return fmt.Errorf("client has no secret configured")
	}

	ok, err := s.hasher.Verify(clientSecret, *secretHash)
	if err != nil || !ok {
		return fmt.Errorf("invalid client secret")
	}
	return nil
}

// ───────────────────────────────────────────────────────
// Userinfo
// ───────────────────────────────────────────────────────

func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	return nil
}

func (s *Storage) SetUserinfoFromRequest(ctx context.Context, userinfo *oidc.UserInfo, request op.IDTokenRequest, scopes []string) error {
	return s.setUserinfo(ctx, userinfo, request.GetSubject(), request.GetClientID(), scopes)
}

func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, subject, origin string) error {
	s.mu.Lock()
	token, ok := s.tokens[tokenID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("token is invalid or has expired")
	}
	if token.Expiration.Before(time.Now()) {
		return fmt.Errorf("token is expired")
	}
	return s.setUserinfo(ctx, userinfo, token.Subject, token.ApplicationID, token.Scopes)
}

func (s *Storage) SetIntrospectionFromToken(ctx context.Context, introspection *oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
	s.mu.Lock()
	token, ok := s.tokens[tokenID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("token is invalid")
	}

	introspection.Expiration = oidc.FromTime(token.Expiration)
	if token.Expiration.Before(time.Now()) {
		return fmt.Errorf("token is expired")
	}

	for _, aud := range token.Audience {
		if aud == clientID {
			userInfo := new(oidc.UserInfo)
			if err := s.setUserinfo(ctx, userInfo, subject, clientID, token.Scopes); err != nil {
				return err
			}
			introspection.SetUserInfo(userInfo)
			introspection.Scope = token.Scopes
			introspection.ClientID = token.ApplicationID
			return nil
		}
	}
	return fmt.Errorf("token is not valid for this client")
}

func (s *Storage) GetPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (map[string]any, error) {
	return nil, nil
}

// ───────────────────────────────────────────────────────
// Keys
// ───────────────────────────────────────────────────────

func (s *Storage) SigningKey(ctx context.Context) (op.SigningKey, error) {
	return s.sigKey, nil
}

func (s *Storage) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{s.sigKey.algorithm}, nil
}

func (s *Storage) KeySet(ctx context.Context) ([]op.Key, error) {
	return []op.Key{&storagePublicKey{*s.sigKey}}, nil
}

func (s *Storage) GetKeyByIDAndClientID(ctx context.Context, keyID, clientID string) (*jose.JSONWebKey, error) {
	return nil, fmt.Errorf("JWT Profile not supported")
}

func (s *Storage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	allowed := make([]string, 0)
	for _, scope := range scopes {
		if scope == oidc.ScopeOpenID {
			allowed = append(allowed, scope)
		}
	}
	return allowed, nil
}

// ───────────────────────────────────────────────────────
// Device Authorization (stubs for interface compliance)
// ───────────────────────────────────────────────────────

func (s *Storage) StoreDeviceAuthorization(ctx context.Context, clientID, deviceCode, userCode string, expires time.Time, scopes []string) error {
	return fmt.Errorf("device flow not yet implemented")
}

func (s *Storage) GetDeviceAuthorizatonState(ctx context.Context, clientID, deviceCode string) (*op.DeviceAuthorizationState, error) {
	return nil, fmt.Errorf("device flow not yet implemented")
}

func (s *Storage) GetDeviceAuthorizationByUserCode(ctx context.Context, userCode string) (*op.DeviceAuthorizationState, error) {
	return nil, fmt.Errorf("device flow not yet implemented")
}

func (s *Storage) CompleteDeviceAuthorization(ctx context.Context, userCode, subject string) error {
	return fmt.Errorf("device flow not yet implemented")
}

func (s *Storage) DenyDeviceAuthorization(ctx context.Context, userCode string) error {
	return fmt.Errorf("device flow not yet implemented")
}

// Health implements op.Storage.
func (s *Storage) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// ───────────────────────────────────────────────────────
// Authentication — called by login handler
// ───────────────────────────────────────────────────────

// CheckUsernamePassword authenticates a user during the OIDC authorization code flow.
func (s *Storage) CheckUsernamePassword(ctx context.Context, username, password, authRequestID string) error {
	s.mu.Lock()
	req, ok := s.authRequests[authRequestID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("auth request not found")
	}

	user, err := s.userRepo.GetByEmail(ctx, username, nil)
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}

	if !user.CanLogin() {
		return fmt.Errorf("account is not active")
	}

	if !user.HasPassword() {
		return fmt.Errorf("invalid credentials")
	}

	ok2, err := s.hasher.Verify(password, *user.PasswordHash)
	if err != nil || !ok2 {
		return fmt.Errorf("invalid credentials")
	}

	s.mu.Lock()
	req.UserID = user.ID.String()
	req.IsDone = true
	req.AuthTime = time.Now()
	s.mu.Unlock()

	return nil
}

// ───────────────────────────────────────────────────────
// Internal helpers
// ───────────────────────────────────────────────────────

func (s *Storage) setUserinfo(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userinfo.Subject = user.ID.String()
		case oidc.ScopeEmail:
			userinfo.Email = user.Email
			userinfo.EmailVerified = oidc.Bool(user.EmailVerified)
		case oidc.ScopeProfile:
			userinfo.Name = user.Name
			userinfo.PreferredUsername = user.Email
			if user.AvatarURL != nil {
				userinfo.Picture = *user.AvatarURL
			}
		}
	}
	return nil
}

func (s *Storage) createOIDCToken(appID, refreshTokenID, subject string, audience, scopes []string) *OIDCToken {
	s.mu.Lock()
	defer s.mu.Unlock()
	token := &OIDCToken{
		ID:             uuid.NewString(),
		ApplicationID:  appID,
		RefreshTokenID: refreshTokenID,
		Subject:        subject,
		Audience:       audience,
		Scopes:         scopes,
		Expiration:     time.Now().Add(5 * time.Minute),
	}
	s.tokens[token.ID] = token
	return token
}

func (s *Storage) createOIDCRefreshToken(accessToken *OIDCToken, amr []string, authTime time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt := &OIDCRefreshToken{
		ID:            accessToken.RefreshTokenID,
		Token:         accessToken.RefreshTokenID,
		AuthTime:      authTime,
		AMR:           amr,
		ApplicationID: accessToken.ApplicationID,
		UserID:        accessToken.Subject,
		Audience:      accessToken.Audience,
		Scopes:        accessToken.Scopes,
		Expiration:    time.Now().Add(5 * time.Hour),
		AccessToken:   accessToken.ID,
	}
	s.refreshTokens[rt.ID] = rt
	return rt.Token
}

func (s *Storage) rotateRefreshToken(currentToken, newToken, newAccessTokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, ok := s.refreshTokens[currentToken]
	if !ok {
		return fmt.Errorf("invalid refresh token")
	}
	delete(s.refreshTokens, currentToken)
	delete(s.tokens, rt.AccessToken)

	if rt.Expiration.Before(time.Now()) {
		return fmt.Errorf("expired refresh token")
	}

	rt.Token = newToken
	rt.ID = newToken
	rt.Expiration = time.Now().Add(5 * time.Hour)
	rt.AccessToken = newAccessTokenID
	s.refreshTokens[newToken] = rt
	return nil
}

func getInfoFromRequest(req op.TokenRequest) (string, time.Time, []string) {
	if ar, ok := req.(*AuthRequest); ok {
		return ar.ClientID, ar.AuthTime, ar.GetAMR()
	}
	if rr, ok := req.(*RefreshTokenRequest); ok {
		return rr.ApplicationID, rr.AuthTime, rr.AMR
	}
	return "", time.Time{}, nil
}

// ───────────────────────────────────────────────────────
// Signing key types (bridge crypto.Signer → op.SigningKey)
// ───────────────────────────────────────────────────────

type storageSigningKey struct {
	id        string
	algorithm jose.SignatureAlgorithm
	key       crypto.Signer
}

func (k *storageSigningKey) SignatureAlgorithm() jose.SignatureAlgorithm { return k.algorithm }
func (k *storageSigningKey) Key() any                                    { return k.key }
func (k *storageSigningKey) ID() string                                  { return k.id }

type storagePublicKey struct {
	storageSigningKey
}

func (k *storagePublicKey) ID() string                          { return k.id }
func (k *storagePublicKey) Algorithm() jose.SignatureAlgorithm  { return k.algorithm }
func (k *storagePublicKey) Use() string                         { return "sig" }
func (k *storagePublicKey) Key() any {
	switch key := k.key.(type) {
	case *rsa.PrivateKey:
		return &key.PublicKey
	case *ecdsa.PrivateKey:
		return &key.PublicKey
	case ed25519.PrivateKey:
		return key.Public()
	default:
		return k.key.Public()
	}
}

func newStorageSigningKey(kp *idcrypto.KeyPair) *storageSigningKey {
	var alg jose.SignatureAlgorithm
	switch kp.Algorithm {
	case "ES256":
		alg = jose.ES256
	case "EdDSA":
		alg = jose.EdDSA
	default:
		alg = jose.RS256
	}
	return &storageSigningKey{
		id:        kp.KeyID,
		algorithm: alg,
		key:       kp.PrivateKey,
	}
}

// ───────────────────────────────────────────────────────
// Type conversion helpers
// ───────────────────────────────────────────────────────

func toAppType(s string) op.ApplicationType {
	switch s {
	case "native":
		return op.ApplicationTypeNative
	case "user_agent":
		return op.ApplicationTypeUserAgent
	default:
		return op.ApplicationTypeWeb
	}
}

func toAuthMethod(s string) oidc.AuthMethod {
	switch s {
	case "client_secret_post":
		return oidc.AuthMethodPost
	case "none":
		return oidc.AuthMethodNone
	case "private_key_jwt":
		return oidc.AuthMethodPrivateKeyJWT
	default:
		return oidc.AuthMethodBasic
	}
}

func toResponseTypes(ss []string) []oidc.ResponseType {
	out := make([]oidc.ResponseType, len(ss))
	for i, s := range ss {
		out[i] = oidc.ResponseType(s)
	}
	return out
}

func toGrantTypes(ss []string) []oidc.GrantType {
	out := make([]oidc.GrantType, len(ss))
	for i, s := range ss {
		out[i] = oidc.GrantType(s)
	}
	return out
}

func toAccessTokenType(s string) op.AccessTokenType {
	if s == "jwt" {
		return op.AccessTokenTypeJWT
	}
	return op.AccessTokenTypeBearer
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
