export interface Product {
  id: string;
  loja_id: string;
  categoria_id: string | null;
  name: string;
  slug: string;
  description: string | null;
  price_cents: number;
  compare_price_cents: number;
  sku: string | null;
  stock_quantity: number;
  stock_alert_threshold: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
  variations: ProductVariation[];
  photos: ProductPhoto[];
}

export interface ProductPhoto {
  id: string;
  produto_id: string;
  url: string;
  sort_order: number;
  created_at: string;
}

export interface ProductVariation {
  id: string;
  produto_id: string;
  name: string;
  value: string;
  price_adjustment_cents: number;
  stock_quantity: number;
  created_at: string;
}

export interface Category {
  id: string;
  loja_id: string;
  name: string;
  slug: string;
  sort_order: number;
  created_at: string;
}

export interface CartItem {
  id: string;
  cart_id: string;
  product_id: string;
  variant_id: string | null;
  quantity: number;
  unit_price_cents: number;
  product_name: string;
  variant_name: string;
  created_at: string;
}

export interface Cart {
  id: string;
  session_id: string;
  store_id: string;
  items: CartItem[];
  created_at: string;
  updated_at: string;
}

export interface CartResponse {
  cart: Cart;
  subtotal_cents: number;
  item_count: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface CheckoutRequest {
  store_id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  address: Address;
  shipping_cents: number;
  coupon_code?: string;
  payment_method: string;
  notes?: string;
}

export interface Address {
  street: string;
  number: string;
  complement?: string;
  neighborhood: string;
  city: string;
  state: string;
  zip_code: string;
}

export interface ValidateCouponResponse {
  valid: boolean;
  message?: string;
  discount_cents: number;
  coupon_id?: string;
}

export interface Order {
  id: string;
  store_id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  address: Address;
  status: string;
  subtotal_cents: number;
  shipping_cents: number;
  discount_cents: number;
  total_cents: number;
  payment_method: string;
  created_at: string;
}

export type TemplateTheme = 'moda' | 'semi-joias' | 'festas' | 'artesanato';

export interface StoreConfig {
  id: string;
  name: string;
  whatsapp: string;
  template: TemplateTheme;
  logo_url?: string;
  banner_url?: string;
}
