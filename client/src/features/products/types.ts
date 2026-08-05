export type Product = {
  id: string
  name: string
  description: string
  price: string
  stock: number
  sku: string
  category: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type ProductListMeta = {
  page: number
  limit: number
  total_items: number
  total_pages: number
}
