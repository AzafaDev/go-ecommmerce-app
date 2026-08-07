import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as ordersApi from './api'
import { CART_KEY } from '../cart/hooks'

export const ORDERS_KEY = ['orders'] as const

export function useOrders() {
  return useQuery({ queryKey: ORDERS_KEY, queryFn: ordersApi.listOrders })
}

export function useOrder(id: string | undefined) {
  return useQuery({
    queryKey: [...ORDERS_KEY, id],
    queryFn: () => ordersApi.getOrder(id!),
    enabled: !!id,
  })
}

export function useCheckout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ordersApi.checkout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ORDERS_KEY })
      // Backend clears the cart server-side as part of checkout.
      queryClient.invalidateQueries({ queryKey: CART_KEY })
    },
  })
}
