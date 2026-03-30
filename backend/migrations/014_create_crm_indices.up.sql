-- Índices para aggregations do CRM e relatórios

-- Índice para aggregations de vendas por período
CREATE INDEX IF NOT EXISTS idx_pedidos_loja_created_at ON pedidos(loja_id, created_at DESC);

-- Índice para aggregations de vendas por status + período
CREATE INDEX IF NOT EXISTS idx_pedidos_loja_status_created ON pedidos(loja_id, status, created_at DESC);

-- Índice para total de vendas (soma de total_cents por loja)
CREATE INDEX IF NOT EXISTS idx_pedidos_loja_total ON pedidos(loja_id, total_cents);

-- Índice para top produtos por vendas
CREATE INDEX IF NOT EXISTS idx_pedido_itens_produto ON pedido_itens(produto_id);

-- Índice para busca de clientes por loja com ordenação
CREATE INDEX IF NOT EXISTS idx_clientes_loja_created ON clientes(loja_id, created_at DESC);

-- Índice para estoque baixo (já existe parcialmente via produtos, mas adicionamos composto)
CREATE INDEX IF NOT EXISTS idx_produtos_stock_alert ON produtos(loja_id, is_active, stock_quantity, stock_alert_threshold);
