-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE user_status AS ENUM (
    'active',
    'disabled',
    'locked',
    'pending_verification',
    'invited',
    'deleted'
);

CREATE TYPE token_type AS ENUM (
    'password_reset',
    'email_verification',
    'invite',
    'magic_link',
    'mfa',
    'device_verification'
);

CREATE TYPE audit_event_type AS ENUM (
    'user.registered',
    'user.login',
    'user.logout',
    'user.password_changed',
    'user.password_reset_requested',
    'user.email_verified',
    'user.disabled',
    'user.deleted',
    'user.updated',
    'session.created',
    'session.revoked',
    'token.revoked',
    'oidc_client.created',
    'oidc_client.updated',
    'oidc_client.deleted',
    'idp.created',
    'idp.updated'
);

CREATE TYPE idp_protocol AS ENUM (
    'oidc',
    'oauth2',
    'saml',
    'ldap',
    'scim'
);
-- +goose StatementEnd

-- +goose Down
DROP TYPE IF EXISTS idp_protocol;
DROP TYPE IF EXISTS audit_event_type;
DROP TYPE IF EXISTS token_type;
DROP TYPE IF EXISTS user_status;
DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pgcrypto";
