import { api } from './client';
import type { CheckoutRequest, Order, ValidateCouponResponse } from '../types';

export async function submitCheckout(req: CheckoutRequest): Promise<{ order: Order; message: string }> {
  return api.post('/api/v1/checkout', req);
}

export async function validateCoupon(
  storeId: string,
  code: string,
  subtotalCents: number,
): Promise<ValidateCouponResponse> {
  return api.post('/api/v1/coupons/validate', {
    store_id: storeId,
    code,
    subtotal_cents: subtotalCents,
  });
}
