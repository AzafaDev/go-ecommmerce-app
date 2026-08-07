import { apiFetch } from '../../lib/http'
import type { Order } from './types'

export function checkout() {
  return apiFetch<{ order: Order }>('/orders/checkout', { method: 'POST' })
}

export function listOrders() {
  return apiFetch<{ orders: Order[] }>('/orders/')
}

export function getOrder(id: string) {
  return apiFetch<{ order: Order }>(`/orders/${id}`)
}
