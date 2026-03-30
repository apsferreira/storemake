import { api } from './client';
import type { CartResponse, CartItem } from '../types';

export async function fetchCart(storeId: string): Promise<CartResponse> {
  return api.get<CartResponse>(`/api/v1/cart?store_id=${encodeURIComponent(storeId)}`);
}

export async function addToCart(
  storeId: string,
  productId: string,
  quantity: number,
  variantId?: string,
): Promise<{ item: CartItem; session_id: string }> {
  return api.post('/api/v1/cart/add', {
    store_id: storeId,
    product_id: productId,
    variant_id: variantId || undefined,
    quantity,
  });
}

export async function updateCartItem(
  storeId: string,
  itemId: string,
  quantity: number,
): Promise<CartItem> {
  return api.put<CartItem>(
    `/api/v1/cart/update/${encodeURIComponent(itemId)}?store_id=${encodeURIComponent(storeId)}`,
    { quantity },
  );
}

export async function removeCartItem(storeId: string, itemId: string): Promise<void> {
  await api.del(`/api/v1/cart/remove/${encodeURIComponent(itemId)}?store_id=${encodeURIComponent(storeId)}`);
}
