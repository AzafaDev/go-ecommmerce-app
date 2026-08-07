import { Link, useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { MouseEvent } from 'react'
import { Badge } from '../../components/Badge'
import { formatIDR } from '../../features/products/format'
import type { Product } from '../../features/products/types'
import { useAddCartItem } from '../../features/cart/hooks'
import { useAuth } from '../../features/auth/auth-context'
import { useToast } from '../../components/toast/ToastContext'
import { ApiError } from '../../lib/http'
import { ProductImage } from './ProductImage'

export function ProductCard({ product }: { product: Product }) {
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const addCartItem = useAddCartItem()
  const { push } = useToast()

  const outOfStock = product.stock === 0

  const handleAddToCart = (e: MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()

    if (!isAuthenticated) {
      navigate('/login')
      return
    }

    addCartItem.mutate(
      { product_id: product.id, quantity: 1 },
      {
        onSuccess: () => {
          push({
            variant: 'success',
            message: 'Added to cart',
            action: { label: 'View cart', onClick: () => navigate('/cart') },
          })
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
    <Link
      to={`/products/${product.id}`}
      className="flex flex-col gap-3 rounded-3xl border-2 border-ink bg-white p-6 shadow-hard transition-all hover:-translate-y-0.5"
    >
      <div className="relative">
        <ProductImage
          src={product.image_url}
          alt={product.name}
          className="h-40 w-full border-2 border-ink"
        />
        {!outOfStock && (
          <button
            type="button"
            aria-label="Add to cart"
            onClick={handleAddToCart}
            disabled={addCartItem.isPending}
            className="absolute bottom-2 right-2 flex h-9 w-9 items-center justify-center rounded-full border-2 border-ink bg-brand-600 text-white shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-brand-700 disabled:opacity-50"
          >
            <Plus size={16} />
          </button>
        )}
      </div>
      <div className="flex items-start justify-between gap-2">
        <p className="text-[15px] font-extrabold text-ink">{product.name}</p>
        {outOfStock && <Badge variant="danger">Out of stock</Badge>}
      </div>
      {product.category && (
        <span className="w-fit rounded-full border-2 border-ink bg-paper px-2.5 py-0.5 text-[11px] font-bold uppercase tracking-wide text-ink-muted">
          {product.category}
        </span>
      )}
      <p className="mt-auto text-[17px] font-extrabold text-brand-700">
        {formatIDR(product.price)}
      </p>
    </Link>
  )
}
