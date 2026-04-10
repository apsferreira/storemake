export interface Product {
  id: string
  loja_id: string
  categoria_id: string | null
  name: string
  slug: string
  description: string | null
  price_cents: number
  compare_price_cents: number
  sku: string | null
  stock_quantity: number
  stock_alert_threshold: number
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
  variations?: ProductVariation[]
  photos?: ProductPhoto[]
}

export interface ProductPhoto {
  id: string
  produto_id: string
  url: string
  sort_order: number
  created_at: string
}

export interface ProductVariation {
  id: string
  produto_id: string
  name: string
  value: string
  price_adjustment_cents: number
  stock_quantity: number
  created_at: string
}

export interface Category {
  id: string
  loja_id: string
  name: string
  slug: string
  sort_order: number
  created_at: string
}

export interface Order {
  id: string
  store_id: string
  customer_name: string
  customer_email: string
  customer_phone: string
  address: unknown
  status: OrderStatus
  subtotal_cents: number
  shipping_cents: number
  discount_cents: number
  total_cents: number
  payment_method: string
  payment_id: string
  tracking_code: string
  notes: string
  items?: OrderItem[]
  created_at: string
  updated_at: string
}

export type OrderStatus = 'pendente' | 'pago' | 'preparando' | 'enviado' | 'entregue' | 'cancelado'

export interface OrderItem {
  id: string
  order_id: string
  product_id: string
  product_name: string
  variant: string
  quantity: number
  unit_price_cents: number
  created_at: string
}

export interface Coupon {
  id: string
  store_id: string
  code: string
  discount_type: 'percent' | 'fixed'
  discount_value: number
  min_order_cents: number
  max_uses: number
  used_count: number
  valid_until: string | null
  is_active: boolean
  created_at: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface CreateProductRequest {
  categoria_id?: string
  name: string
  slug: string
  description?: string
  price_cents: number
  compare_price_cents?: number
  sku?: string
  stock_quantity: number
  stock_alert_threshold: number
  is_active?: boolean
  sort_order?: number
  variations?: CreateVariationRequest[]
}

export interface UpdateProductRequest {
  categoria_id?: string
  name?: string
  slug?: string
  description?: string
  price_cents?: number
  compare_price_cents?: number
  sku?: string
  stock_quantity?: number
  stock_alert_threshold?: number
  is_active?: boolean
  sort_order?: number
}

export interface CreateVariationRequest {
  name: string
  value: string
  price_adjustment_cents: number
  stock_quantity: number
}

export interface CreateCouponRequest {
  code: string
  discount_type: 'percent' | 'fixed'
  discount_value: number
  min_order_cents: number
  max_uses: number
  valid_until?: string
  is_active?: boolean
}

export interface UpdateCouponRequest {
  code?: string
  discount_type?: string
  discount_value?: number
  min_order_cents?: number
  max_uses?: number
  valid_until?: string
  is_active?: boolean
}

export interface CreateCategoryRequest {
  name: string
  slug: string
  sort_order?: number
}

export interface UpdateOrderStatusRequest {
  status: OrderStatus
  tracking_code?: string
}

// --- BKL-900: Inventário multi-loja centralizado ---

export interface InventoryMaster {
  id: string
  tenant_id: string
  produto_id: string | null
  sku_global: string
  nome: string
  descricao: string
  unidade: string
  custo_unitario_cents: number
  quantity_total: number
  quantity_reserved: number
  reorder_point: number
  reorder_quantity: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface StoreAllocation {
  id: string
  master_id: string
  loja_id: string
  quantity_allocated: number
  quantity_sold: number
  profit_share_pct: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface InventoryMovement {
  id: string
  master_id: string
  loja_id: string | null
  pedido_id: string | null
  movement_type: 'entrada' | 'saida_venda' | 'saida_perda' | 'transferencia' | 'ajuste' | 'devolucao'
  quantity: number
  quantity_before: number
  quantity_after: number
  custo_unitario_cents: number | null
  observacao: string
  created_by: string | null
  created_at: string
}

export interface InventoryAlert {
  id: string
  master_id: string
  quantity_current: number
  quantity_reorder: number
  alert_type: 'low_stock' | 'out_of_stock'
  acknowledged: boolean
  acknowledged_by: string | null
  acknowledged_at: string | null
  created_at: string
}

export interface SupplierOrder {
  id: string
  tenant_id: string
  master_id: string
  status: 'rascunho' | 'enviado' | 'confirmado' | 'recebido' | 'cancelado'
  quantity_ordered: number
  quantity_received: number
  custo_total_cents: number
  fornecedor_nome: string
  fornecedor_contato: string
  observacao: string
  expected_at: string | null
  received_at: string | null
  created_at: string
  updated_at: string
}

export interface CreateInventoryMasterRequest {
  sku_global?: string
  nome: string
  descricao?: string
  unidade?: string
  custo_unitario_cents?: number
  quantity_total: number
  reorder_point?: number
  reorder_quantity?: number
}

export interface AdjustQuantityRequest {
  delta: number
  loja_id?: string
  observacao?: string
}

export interface UpsertAllocationRequest {
  quantity_allocated: number
  profit_share_pct?: number
}

export interface CreateSupplierOrderRequest {
  quantity_ordered: number
  custo_total_cents?: number
  fornecedor_nome?: string
  observacao?: string
}
