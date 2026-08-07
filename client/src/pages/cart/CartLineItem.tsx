import { Link } from 'react-router-dom'
import { Trash2 } from 'lucide-react'
import { Badge } from '../../components/Badge'
import { QuantityStepper } from '../../components/QuantityStepper'
import { ProductImage } from '../products/ProductImage'
import { formatIDR } from '../../features/products/format'
import { useRemoveCartItem, useUpdateCartItemQuantity } from '../../features/cart/hooks'
import { useToast } from '../../components/toast/ToastContext'
import { ApiError } from '../../lib/http'
import type { CartItem } from '../../features/cart/types'

export function CartLineItem({ item }: { item: CartItem }) {
  const updateQuantity = useUpdateCartItemQuantity()
  const removeItem = useRemoveCartItem()
  const { push } = useToast()

  const busy = updateQuantity.isPending || removeItem.isPending

  const handleQuantityChange = (next: number) => {
    updateQuantity.mutate(
      { productId: item.product_id, quantity: next },
      {
        onError: (err) => {
          push({
            variant: 'error',
            message: err instanceof ApiError ? err.message : 'Could not update quantity.',
          })
        },
      },
    )
  }

  const handleRemove = () => {
    removeItem.mutate(item.product_id, {
      onError: (err) => {
        push({
          variant: 'error',
          message: err instanceof ApiError ? err.message : 'Could not remove item.',
        })
      },
    })
  }

  return (
    <div className="flex items-center gap-4 border-b-2 border-line py-4 last:border-b-0">
      <ProductImage
        src={item.image_url}
        alt={item.name}
        className="h-20 w-20 shrink-0 border-2 border-ink"
      />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <Link
            to={`/products/${item.product_id}`}
            className="truncate text-[14.5px] font-extrabold text-ink hover:text-brand-700"
          >
            {item.name}
          </Link>
          {!item.is_active && <Badge variant="danger">No longer available</Badge>}
        </div>
        <p className="mt-0.5 text-[12.5px] text-ink-muted">{formatIDR(item.price)} / item</p>
        <div className="mt-3">
          <QuantityStepper
            value={item.quantity}
            onChange={handleQuantityChange}
            disabled={busy || !item.is_active}
          />
        </div>
      </div>
      <div className="flex shrink-0 flex-col items-end gap-3">
        <p className="text-[15px] font-extrabold text-ink">{formatIDR(item.subtotal)}</p>
        <button
          type="button"
          aria-label="Remove item"
          onClick={handleRemove}
          disabled={busy}
          className="text-ink-muted transition-colors hover:text-danger-600 disabled:opacity-40"
        >
          <Trash2 size={16} />
        </button>
      </div>
    </div>
  )
}
