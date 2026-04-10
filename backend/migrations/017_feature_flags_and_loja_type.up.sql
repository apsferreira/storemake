-- SPEC-006-B: Feature flags de módulos por tenant + tipo de loja (filial/virtual/master)

-- Tabela de feature flags por tenant
-- Permite ativar/desativar módulos do StoreMake por cliente
CREATE TABLE IF NOT EXISTS tenant_modules (
    tenant_id   UUID        NOT NULL,
    module      VARCHAR(50) NOT NULL,       -- 'storefront' | 'inventory' | 'crm' | 'whatsapp'
    enabled     BOOLEAN     NOT NULL DEFAULT true,
    config      JSONB       DEFAULT '{}',   -- configurações específicas do módulo
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, module)
);

-- Índice para lookup por tenant
CREATE INDEX IF NOT EXISTS idx_tenant_modules_tenant ON tenant_modules(tenant_id);

-- Inserir módulos padrão para todos os tenants existentes (todos habilitados por padrão)
INSERT INTO tenant_modules (tenant_id, module, enabled)
SELECT id, unnest(ARRAY['storefront','inventory','crm','whatsapp']), true
FROM lojas
ON CONFLICT DO NOTHING;

-- Adicionar coluna loja_type à tabela lojas
-- 'virtual'  = loja online pública (catálogo + checkout)
-- 'filial'   = loja física/filial (consome estoque, sem catálogo público)
-- 'master'   = filial master que controla o estoque centralizado
ALTER TABLE lojas
    ADD COLUMN IF NOT EXISTS loja_type VARCHAR(20) NOT NULL DEFAULT 'virtual'
        CHECK (loja_type IN ('virtual', 'filial', 'master'));

CREATE INDEX IF NOT EXISTS idx_lojas_loja_type ON lojas(loja_type);

-- Atualizar função de trigger para updated_at em tenant_modules
CREATE OR REPLACE FUNCTION update_tenant_modules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenant_modules_updated_at
    BEFORE UPDATE ON tenant_modules
    FOR EACH ROW
    EXECUTE FUNCTION update_tenant_modules_updated_at();
