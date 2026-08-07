import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card } from '../../components/Card'
import { DataRow } from '../../components/DataRow'
import { Badge } from '../../components/Badge'
import { Button } from '../../components/Button'
import { QuantityStepper } from '../../components/QuantityStepper'
import { useProduct } from '../../features/products/hooks'
import { formatIDR } from '../../features/products/format'
import { useAddCartItem } from '../../features/cart/hooks'
import { useAuth } from '../../features/auth/auth-context'
import { useToast } from '../../components/toast/ToastContext'
import { ApiError } from '../../lib/http'
import { ProductImage } from './ProductImage'

export function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data, isLoading, isError } = useProduct(id)
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const addCartItem = useAddCartItem()
  const { push } = useToast()
  const [quantity, setQuantity] = useState(1)

  if (isLoading) return <p className="text-[14px] text-ink-muted">Loading product…</p>

  if (isError || !data) {
    return <p className="text-[14px] text-danger-600">Product not found.</p>
  }

  const { product } = data
  const outOfStock = product.stock === 0

  const handleAddToCart = () => {
    if (!isAuthenticated) {
      navigate('/login')
      return
    }

    addCartItem.mutate(
      { product_id: product.id, quantity },
      {
        onSuccess: () => {
          push({
            variant: 'success',
            message: 'Added to cart',
            action: { label: 'View cart', onClick: () => navigate('/cart') },
          })
          setQuantity(1)
        },
        onError: (err) => {
          push({
            variant: 'error',
            message: err instanceof ApiError ? err.message : 'Could not add to cart.',
          })
        },
      },
    )
  }

  return (
    <div className="mx-auto max-w-xl">
      <Card eyebrow={product.category || 'go-commerce / product'}>
        <ProductImage
          src={product.image_url}
          alt={product.name}
          className="h-64 w-full border-2 border-ink"
        />
        <div className="mt-6 flex items-start justify-between gap-3">
          <h1 className="text-[22px] font-extrabold text-ink">{product.name}</h1>
          {outOfStock && <Badge variant="danger">Out of stock</Badge>}
        </div>
        <p className="mt-3 text-[20px] font-extrabold text-brand-700">
          {formatIDR(product.price)}
        </p>
        {product.description && (
          <p className="mt-4 text-[14px] leading-relaxed text-ink-muted">
            {product.description}
          </p>
        )}

        <div className="mt-6 flex flex-col">
          <DataRow label="SKU" value={product.sku} />
          <DataRow label="Stock" value={product.stock} />
          <DataRow label="Category" value={product.category || '—'} />
        </div>

        {!outOfStock && (
          <div className="mt-6 flex items-center gap-4">
            <QuantityStepper
              value={quantity}
              max={product.stock}
              onChange={setQuantity}
              disabled={addCartItem.isPending}
            />
            <div className="flex-1">
              <Button type="button" onClick={handleAddToCart} disabled={addCartItem.isPending}>
                Add to cart
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
