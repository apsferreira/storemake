CREATE TABLE IF NOT EXISTS cupons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    discount_percent INT NOT NULL DEFAULT 0 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    max_uses INT NOT NULL DEFAULT 0,
    used_count INT NOT NULL DEFAULT 0,
    valid_until TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(loja_id, code)
);

CREATE INDEX idx_cupons_loja_id ON cupons(loja_id);
CREATE INDEX idx_cupons_code ON cupons(loja_id, code);
