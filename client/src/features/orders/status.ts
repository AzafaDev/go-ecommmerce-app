import type { OrderStatus } from './types'

export const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  pending_payment: 'Payment pending',
  paid: 'Paid',
  cancelled: 'Cancelled',
  expired: 'Expired',
}

export const ORDER_STATUS_BADGE_VARIANT: Record<
  OrderStatus,
  'neutral' | 'success' | 'danger' | 'warning'
> = {
  pending_payment: 'warning',
  paid: 'success',
  cancelled: 'danger',
  expired: 'danger',
}
