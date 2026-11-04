ALTER TABLE organizations ADD COLUMN IF NOT EXISTS archived_at timestamptz NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_archived_at ON organizations (archived_at);
