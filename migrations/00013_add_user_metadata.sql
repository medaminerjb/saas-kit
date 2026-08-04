-- +goose Up
-- Add metadata columns to users table
ALTER TABLE users 
ADD COLUMN metadata_public JSONB DEFAULT '{}'::jsonb,
ADD COLUMN metadata_private JSONB DEFAULT '{}'::jsonb;

-- Add check constraint for 32KB max size (32 * 1024 = 32768 bytes)
ALTER TABLE users 
ADD CONSTRAINT chk_metadata_public_size CHECK (octet_length(metadata_public::text) <= 32768),
ADD CONSTRAINT chk_metadata_private_size CHECK (octet_length(metadata_private::text) <= 32768);

-- Create GIN indexes for efficient JSONB queries
CREATE INDEX idx_users_metadata_public ON users USING GIN (metadata_public);
CREATE INDEX idx_users_metadata_private ON users USING GIN (metadata_private);

-- +goose Down
DROP INDEX IF EXISTS idx_users_metadata_public;
DROP INDEX IF EXISTS idx_users_metadata_private;
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS chk_metadata_public_size,
DROP CONSTRAINT IF EXISTS chk_metadata_private_size;
ALTER TABLE users 
DROP COLUMN IF EXISTS metadata_public,
DROP COLUMN IF EXISTS metadata_private;
