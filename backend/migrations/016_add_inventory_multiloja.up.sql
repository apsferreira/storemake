-- Migration 016: Módulo de Estoque Multi-Loja
-- Expande o controle de estoque da tabela `produtos` para suportar:
-- 1. Estoque mestre centralizado (master entity)
-- 2. Alocação de cotas por loja dentro do mesmo tenant
-- 3. Movimentações auditáveis (entradas, saídas, ajustes, transferências)
-- 4. Configuração de repasse de lucro por loja
-- 5. Pedidos de reposição ao fornecedor
--
-- DECISÃO: módulo dentro do storemake, não serviço separado.
-- Ver ADR: docs/adr/001-inventory-storemake.md

-- 1. Tabela master de estoque por tenant
-- O "master" é o tenant dono de múltiplas lojas. Controla o SKU globalmente.
CREATE TABLE IF NOT EXISTS inventory_masters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,             -- owner_id do grupo de lojas
    produto_id UUID REFERENCES produtos(id) ON DELETE CASCADE,
    sku_global VARCHAR(100),             -- SKU único no nível do grupo
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    unidade VARCHAR(20) DEFAULT 'un',    -- un, kg, l, cx, etc.
    custo_unitario_cents BIGINT DEFAULT 0,
    reorder_point INT NOT NULL DEFAULT 5,   -- ponto de reposição
    reorder_quantity INT NOT NULL DEFAULT 10, -- quantidade sugerida de reposição
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, sku_global)
);

CREATE INDEX idx_inventory_masters_tenant ON inventory_masters(tenant_id);
CREATE INDEX idx_inventory_masters_produto ON inventory_masters(produto_id);

-- 2. Quantidade total disponível por master (view consolidada)
-- quantity_available = quantity_total - quantity_reserved - sum(store_allocations.allocated)
ALTER TABLE inventory_masters
    ADD COLUMN IF NOT EXISTS quantity_total INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS quantity_reserved INT NOT NULL DEFAULT 0;

-- 3. Alocação de estoque por loja
-- Cada loja recebe uma cota do estoque master + configuração de repasse de lucro
CREATE TABLE IF NOT EXISTS store_allocations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id UUID NOT NULL REFERENCES inventory_masters(id) ON DELETE CASCADE,
    loja_id UUID NOT NULL REFERENCES lojas(id) ON DELETE CASCADE,
    quantity_allocated INT NOT NULL DEFAULT 0,  -- cota alocada para esta loja
    quantity_sold INT NOT NULL DEFAULT 0,        -- vendido por esta loja (para cálculo de giro)
    profit_share_pct NUMERIC(5,2) NOT NULL DEFAULT 100.00, -- % do lucro repassado para a loja (0-100)
    -- O restante (100 - profit_share_pct) fica com o master
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(master_id, loja_id),
    CONSTRAINT chk_profit_share CHECK (profit_share_pct >= 0 AND profit_share_pct <= 100),
    CONSTRAINT chk_quantity_allocated CHECK (quantity_allocated >= 0)
);

CREATE INDEX idx_store_allocations_master ON store_allocations(master_id);
CREATE INDEX idx_store_allocations_loja ON store_allocations(loja_id);

-- 4. Movimentações de estoque (auditoria completa)
CREATE TYPE inventory_movement_type AS ENUM (
    'entrada',         -- recebimento de fornecedor
    'saida_venda',     -- venda em loja
    'saida_perda',     -- perda, quebra, expiração
    'transferencia',   -- transferência entre lojas
    'ajuste',          -- ajuste de inventário (contagem física)
    'devolucao'        -- devolução de cliente
);

CREATE TABLE IF NOT EXISTS inventory_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id UUID NOT NULL REFERENCES inventory_masters(id),
    loja_id UUID REFERENCES lojas(id),  -- NULL = operação no master
    pedido_id UUID REFERENCES pedidos(id),  -- vínculo com pedido se for venda
    movement_type inventory_movement_type NOT NULL,
    quantity INT NOT NULL,              -- positivo = entrada, negativo = saída
    quantity_before INT NOT NULL,       -- snapshot antes da movimentação
    quantity_after INT NOT NULL,        -- snapshot depois da movimentação
    custo_unitario_cents BIGINT,        -- custo unitário na data da movimentação
    observacao TEXT,
    created_by UUID,                    -- user_id que realizou (NULL = automático)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_movements_master ON inventory_movements(master_id);
CREATE INDEX idx_inventory_movements_loja ON inventory_movements(loja_id);
CREATE INDEX idx_inventory_movements_type ON inventory_movements(movement_type);
CREATE INDEX idx_inventory_movements_created ON inventory_movements(created_at);

-- 5. Pedidos de reposição ao fornecedor
CREATE TYPE supplier_order_status AS ENUM (
    'rascunho',
    'enviado',
    'confirmado',
    'recebido',
    'cancelado'
);

CREATE TABLE IF NOT EXISTS supplier_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    master_id UUID NOT NULL REFERENCES inventory_masters(id),
    status supplier_order_status NOT NULL DEFAULT 'rascunho',
    quantity_ordered INT NOT NULL,
    quantity_received INT DEFAULT 0,
    custo_total_cents BIGINT DEFAULT 0,
    fornecedor_nome VARCHAR(255),
    fornecedor_contato TEXT,
    observacao TEXT,
    expected_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supplier_orders_tenant ON supplier_orders(tenant_id);
CREATE INDEX idx_supplier_orders_master ON supplier_orders(master_id);
CREATE INDEX idx_supplier_orders_status ON supplier_orders(status);

-- 6. Alertas de reposição (gerados automaticamente quando quantity_total <= reorder_point)
CREATE TABLE IF NOT EXISTS inventory_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id UUID NOT NULL REFERENCES inventory_masters(id),
    quantity_current INT NOT NULL,
    quantity_reorder INT NOT NULL,
    alert_type VARCHAR(50) NOT NULL DEFAULT 'low_stock',  -- low_stock, out_of_stock
    acknowledged BOOLEAN NOT NULL DEFAULT false,
    acknowledged_by UUID,
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_alerts_master ON inventory_alerts(master_id);
CREATE INDEX idx_inventory_alerts_acknowledged ON inventory_alerts(acknowledged);

-- 7. Função + trigger para auto-criar alerta quando estoque cai abaixo do reorder_point
CREATE OR REPLACE FUNCTION check_inventory_reorder()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.quantity_total <= NEW.reorder_point AND
       (OLD.quantity_total IS NULL OR OLD.quantity_total > OLD.reorder_point) THEN
        INSERT INTO inventory_alerts (master_id, quantity_current, quantity_reorder, alert_type)
        VALUES (
            NEW.id,
            NEW.quantity_total,
            NEW.reorder_point,
            CASE WHEN NEW.quantity_total = 0 THEN 'out_of_stock' ELSE 'low_stock' END
        )
        ON CONFLICT DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_inventory_reorder
    AFTER UPDATE OF quantity_total ON inventory_masters
    FOR EACH ROW
    EXECUTE FUNCTION check_inventory_reorder();
