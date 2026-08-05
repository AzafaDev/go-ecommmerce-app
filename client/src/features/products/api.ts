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

export type UpdateProductInput = ProductInput & { is_active: boolean }

function buildQuery(params: ListProductsParams) {
  const query = new URLSearchParams()
  if (params.search) query.set('search', params.search)
  if (params.page) query.set('page', String(params.page))
  if (params.limit) query.set('limit', String(params.limit))
  const qs = query.toString()
  return qs ? `?${qs}` : ''
}

export function listProducts(params: ListProductsParams) {
  return apiFetch<{ products: Product[]; meta: ProductListMeta }>(
    `/products${buildQuery(params)}`,
  )
}

export function getProduct(id: string) {
  return apiFetch<{ product: Product }>(`/products/${id}`)
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
