export type CartItem = {
  product_id: string
  name: string
  price: string
  image_url: string
  is_active: boolean
  quantity: number
  subtotal: string
}

export type Cart = {
  items: CartItem[]
  total_items: number
  total: string
}
