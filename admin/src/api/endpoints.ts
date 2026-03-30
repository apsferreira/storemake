import { api } from './client'
import type {
  Product,
  Category,
  Order,
  Coupon,
  PaginatedResponse,
  CreateProductRequest,
  UpdateProductRequest,
  CreateCouponRequest,
  UpdateCouponRequest,
  CreateCategoryRequest,
  UpdateOrderStatusRequest,
} from '../types'

// Products
export const productsApi = {
  list: (params?: { page?: number; per_page?: number; search?: string; categoria_id?: string }) => {
    const qs = new URLSearchParams()
    if (params?.page) qs.set('page', String(params.page))
    if (params?.per_page) qs.set('per_page', String(params.per_page))
    if (params?.search) qs.set('search', params.search)
    if (params?.categoria_id) qs.set('categoria_id', params.categoria_id)
    return api.get<PaginatedResponse<Product>>(`/products?${qs}`)
  },
  get: (id: string) => api.get<Product>(`/products/${id}`),
  create: (data: CreateProductRequest) => api.post<Product>('/products', data),
  update: (id: string, data: UpdateProductRequest) => api.put<Product>(`/products/${id}`, data),
  delete: (id: string) => api.delete<void>(`/products/${id}`),
  uploadPhotos: (id: string, formData: FormData) =>
    api.upload<{ data: unknown[] }>(`/products/${id}/photos`, formData),
}

// Categories
export const categoriesApi = {
  list: () => api.get<PaginatedResponse<Category>>('/categories'),
  get: (id: string) => api.get<Category>(`/categories/${id}`),
  create: (data: CreateCategoryRequest) => api.post<Category>('/categories', data),
  update: (id: string, data: Partial<CreateCategoryRequest>) => api.put<Category>(`/categories/${id}`, data),
  delete: (id: string) => api.delete<void>(`/categories/${id}`),
}

// Orders
export const ordersApi = {
  list: (params?: { page?: number; per_page?: number; status?: string; search?: string }) => {
    const qs = new URLSearchParams()
    if (params?.page) qs.set('page', String(params.page))
    if (params?.per_page) qs.set('per_page', String(params.per_page))
    if (params?.status) qs.set('status', params.status)
    if (params?.search) qs.set('search', params.search)
    return api.get<PaginatedResponse<Order>>(`/orders?${qs}`)
  },
  get: (id: string) => api.get<Order>(`/orders/${id}`),
  updateStatus: (id: string, data: UpdateOrderStatusRequest) =>
    api.put<Order>(`/orders/${id}/status`, data),
}

// Coupons
export const couponsApi = {
  list: () => api.get<{ data: Coupon[] }>('/coupons'),
  get: (id: string) => api.get<Coupon>(`/coupons/${id}`),
  create: (data: CreateCouponRequest) => api.post<Coupon>('/coupons', data),
  update: (id: string, data: UpdateCouponRequest) => api.put<Coupon>(`/coupons/${id}`, data),
  delete: (id: string) => api.delete<void>(`/coupons/${id}`),
}

// Stock alerts
export const stockApi = {
  lowStock: () => api.get<{ data: Product[] }>('/stock/alerts'),
}
