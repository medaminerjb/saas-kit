-- +goose Up
CREATE TABLE feature_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT NOT NULL UNIQUE,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Reserved for Phase 2: tenant-scoped feature flag overrides
-- CREATE TABLE tenant_feature_flags (
--     id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     tenant_id       UUID NOT NULL,
--     feature_flag_id UUID NOT NULL REFERENCES feature_flags(id),
--     enabled         BOOLEAN NOT NULL,
--     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     UNIQUE (tenant_id, feature_flag_id)
-- );

-- +goose Down
DROP TABLE IF EXISTS feature_flags;
