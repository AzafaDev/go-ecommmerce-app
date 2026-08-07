import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as cartApi from './api'

export const CART_KEY = ['cart'] as const

export function useCart(options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: CART_KEY,
    queryFn: cartApi.getCart,
    enabled: options.enabled ?? true,
  })
}

// Derived from the cart query rather than a separate request — backs the nav badge.
export function useCartCount(enabled = true) {
  const { data } = useCart({ enabled })
  return data?.cart.total_items ?? 0
}

function invalidateCart(queryClient: ReturnType<typeof useQueryClient>) {
  return queryClient.invalidateQueries({ queryKey: CART_KEY })
}

export function useAddCartItem() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: cartApi.addCartItem,
    onSuccess: () => invalidateCart(queryClient),
  })
}

export function useUpdateCartItemQuantity() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productId, quantity }: { productId: string; quantity: number }) =>
      cartApi.updateCartItemQuantity(productId, quantity),
    onSuccess: () => invalidateCart(queryClient),
  })
}

export function useRemoveCartItem() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: cartApi.removeCartItem,
    onSuccess: () => invalidateCart(queryClient),
  })
}

export function useClearCart() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: cartApi.clearCart,
    onSuccess: () => invalidateCart(queryClient),
  })
}
