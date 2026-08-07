export type OrderStatus = 'pending_payment' | 'paid' | 'cancelled' | 'expired'

export type OrderItem = {
  product_id: string
  product_name: string
  price: string
  quantity: number
  subtotal: string
}

export type Order = {
  id: string
  status: OrderStatus
  total_amount: string
  snap_token?: string
  items: OrderItem[]
  created_at: string
  paid_at?: string
}
