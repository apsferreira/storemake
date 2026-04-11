-- Migration 018: Criar tabela de planos de StoreMake

CREATE TABLE IF NOT EXISTS planos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0,
    max_products INTEGER NOT NULL,
    custom_domain BOOLEAN NOT NULL DEFAULT false,
    support_level VARCHAR(50) NOT NULL DEFAULT 'email', -- 'email', 'priority', 'dedicated'
    features JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_planos_slug ON planos(slug);
CREATE INDEX idx_planos_is_active ON planos(is_active);

COMMENT ON TABLE planos IS 'Planos de preço do StoreMake';
COMMENT ON COLUMN planos.slug IS 'Identificador único do plano (free, starter, pro)';
COMMENT ON COLUMN planos.price_cents IS 'Preço mensal em centavos (ex: 7900 = R$79,00)';
COMMENT ON COLUMN planos.max_products IS 'Número máximo de produtos';
COMMENT ON COLUMN planos.custom_domain IS 'Permite domínio customizado';
COMMENT ON COLUMN planos.support_level IS 'Nível de suporte';
COMMENT ON COLUMN planos.features IS 'Features adicionais em JSON';

-- Inserir planos padrão
INSERT INTO planos (slug, name, price_cents, max_products, custom_domain, support_level, features) VALUES
    ('free', 'Free', 0, 10, false, 'community', '{"frete": false, "cupons": false, "whatsapp": false, "crm": false}'),
    ('starter', 'Starter', 7900, 200, true, 'email', '{"frete": true, "cupons": true, "whatsapp": false, "crm": false}'),
    ('pro', 'Pro', 14900, 99999, true, 'priority', '{"frete": true, "cupons": true, "whatsapp": true, "crm": true}')
ON CONFLICT DO NOTHING;

-- Adicionar coluna plano_id à tabela lojas
ALTER TABLE lojas
    ADD COLUMN IF NOT EXISTS plano_id UUID REFERENCES planos(id) ON DELETE RESTRICT;

-- Atribuir plano_id padrão (free) para lojas existentes
UPDATE lojas SET plano_id = (SELECT id FROM planos WHERE slug = 'free')
WHERE plano_id IS NULL;

-- Fazer plano_id NOT NULL após atribuir padrão
ALTER TABLE lojas
    ALTER COLUMN plano_id SET NOT NULL;

CREATE INDEX idx_lojas_plano_id ON lojas(plano_id);
