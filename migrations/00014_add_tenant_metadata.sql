-- +goose Up
-- Add metadata column to tenants table
ALTER TABLE tenants 
ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;

-- Add check constraint for 32KB max size (32 * 1024 = 32768 bytes)
ALTER TABLE tenants 
ADD CONSTRAINT chk_tenant_metadata_size CHECK (octet_length(metadata::text) <= 32768);

-- Create GIN index for efficient JSONB queries
CREATE INDEX idx_tenants_metadata ON tenants USING GIN (metadata);

-- +goose Down
DROP INDEX IF EXISTS idx_tenants_metadata;
ALTER TABLE tenants 
DROP CONSTRAINT IF EXISTS chk_tenant_metadata_size;
ALTER TABLE tenants 
DROP COLUMN IF EXISTS metadata;
