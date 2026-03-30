import { api } from './client';
import type { PaginatedResponse, Product } from '../types';

interface CatalogParams {
  storeId: string;
  page?: number;
  perPage?: number;
  categoriaId?: string;
  search?: string;
  minPrice?: number;
  maxPrice?: number;
  inStock?: boolean;
}

export async function fetchCatalog(params: CatalogParams): Promise<PaginatedResponse<Product>> {
  const qs = new URLSearchParams();
  qs.set('store_id', params.storeId);
  if (params.page) qs.set('page', String(params.page));
  if (params.perPage) qs.set('per_page', String(params.perPage));
  if (params.categoriaId) qs.set('categoria_id', params.categoriaId);
  if (params.search) qs.set('search', params.search);
  if (params.minPrice !== undefined) qs.set('min_price', String(params.minPrice));
  if (params.maxPrice !== undefined) qs.set('max_price', String(params.maxPrice));
  if (params.inStock) qs.set('in_stock', 'true');

  return api.get<PaginatedResponse<Product>>(`/api/v1/public/catalog?${qs.toString()}`);
}
