CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS lojas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    domain_custom VARCHAR(255),
    cnpj_cpf VARCHAR(20),
    email VARCHAR(255),
    phone VARCHAR(20),
    whatsapp VARCHAR(20),
    logo_url TEXT,
    banner_url TEXT,
    template VARCHAR(50) NOT NULL DEFAULT 'generico'
        CHECK (template IN ('moda', 'semijoias', 'festas', 'artesanato', 'generico')),
    colors_json JSONB DEFAULT '{}',
    description TEXT,
    operating_hours JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lojas_owner_id ON lojas(owner_id);
CREATE INDEX idx_lojas_slug ON lojas(slug);
