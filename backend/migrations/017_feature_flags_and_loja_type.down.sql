DROP TRIGGER IF EXISTS tenant_modules_updated_at ON tenant_modules;
DROP FUNCTION IF EXISTS update_tenant_modules_updated_at();
DROP INDEX IF EXISTS idx_lojas_loja_type;
ALTER TABLE lojas DROP COLUMN IF EXISTS loja_type;
DROP INDEX IF EXISTS idx_tenant_modules_tenant;
DROP TABLE IF EXISTS tenant_modules;
