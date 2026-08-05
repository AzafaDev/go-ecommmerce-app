import { apiFetch } from '../../lib/http'
import type { Product, ProductListMeta } from './types'

export type ListProductsParams = {
  search?: string
  page?: number
  limit?: number
}

export type ProductInput = {
  name: string
  description: string
  price: number
  stock: number
  sku: string
  category: string
}

// SKU is immutable after create — the API doesn't accept it on update.
export type UpdateProductInput = {
  name: string
  description: string
  price: number
  stock: number
  category: string
  is_active: boolean
}

function buildQuery(params: ListProductsParams) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.page) query.set('page', String(params.page))
  if (params.limit) query.set('limit', String(params.limit))
  const qs = query.toString()
  return qs ? `?${qs}` : ''
}

// Public catalog — only ever returns active products.
export function listProducts(params: ListProductsParams) {
  return apiFetch<{ products: Product[]; meta: ProductListMeta }>(
    `/products${buildQuery(params)}`,
  )
}

export function getProduct(id: string) {
  return apiFetch<{ product: Product }>(`/products/${id}`)
}

export function listCategories() {
  return apiFetch<{ categories: string[] }>('/products/categories')
}

// Admin catalog — includes inactive (soft-deleted) products, requires admin auth.
export function listAdminProducts(params: ListProductsParams) {
  return apiFetch<{ products: Product[]; meta: ProductListMeta }>(
    `/admin/products${buildQuery(params)}`,
  )
}

export function getAdminProduct(id: string) {
  return apiFetch<{ product: Product }>(`/admin/products/${id}`)
}

export function listAdminCategories() {
  return apiFetch<{ categories: string[] }>('/admin/products/categories')
}

export function createProduct(input: ProductInput) {
  return apiFetch<{ product: Product }>('/products', { method: 'POST', body: input })
}

export function updateProduct(id: string, input: UpdateProductInput) {
  return apiFetch<{ product: Product }>(`/products/${id}`, { method: 'PUT', body: input })
}

export function deleteProduct(id: string) {
  return apiFetch<{ message: string }>(`/products/${id}`, { method: 'DELETE' })
}
