import { apiFetch } from '../../lib/http'
import type { Cart, CartItem } from './types'

export function getCart() {
  return apiFetch<{ cart: Cart }>('/cart/')
}

export function clearCart() {
  return apiFetch<{ message: string }>('/cart/', { method: 'DELETE' })
}

export function addCartItem(input: { product_id: string; quantity: number }) {
  return apiFetch<{ item: CartItem }>('/cart/items/', { method: 'POST', body: input })
}

// Sets the absolute quantity — does not increment.
export function updateCartItemQuantity(productId: string, quantity: number) {
  return apiFetch<{ item: CartItem }>(`/cart/items/${productId}`, {
    method: 'PATCH',
    body: { quantity },
  })
}

export function removeCartItem(productId: string) {
  return apiFetch<{ message: string }>(`/cart/items/${productId}`, { method: 'DELETE' })
}
