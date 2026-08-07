import { Link } from 'react-router-dom'
import { ShoppingCart } from 'lucide-react'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { formatIDR } from '../../features/products/format'
import { useCart, useClearCart } from '../../features/cart/hooks'
import { useCheckoutFlow } from '../../features/orders/useCheckoutFlow'
import { useToast } from '../../components/toast/ToastContext'
import { ApiError } from '../../lib/http'
import { CartLineItem } from './CartLineItem'

const CHECKOUT_LABEL = {
  idle: 'Checkout',
  'creating-order': 'Placing order…',
  'awaiting-payment': 'Opening payment…',
} as const

export function CartPage() {
  const { data, isLoading, isError } = useCart()
  const clearCart = useClearCart()
  const checkoutFlow = useCheckoutFlow()
  const { push } = useToast()

  if (isLoading) return <p className="text-[14px] text-ink-muted">Loading cart…</p>

  if (isError || !data) {
    return <p className="text-[14px] text-danger-600">Could not load your cart.</p>
  }

  const { cart } = data

  if (cart.items.length === 0) {
    return (
      <div className="mx-auto max-w-md">
        <Card centered>
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border-2 border-ink bg-paper text-ink-muted">
            <ShoppingCart size={22} />
          </div>
          <h1 className="text-[18px] font-extrabold text-ink">Your cart is empty</h1>
          <p className="mt-2 text-[13.5px] text-ink-muted">
            Browse the catalog and add something you like.
          </p>
          <div className="mt-6">
            <Link to="/products">
              <Button type="button">Browse products</Button>
            </Link>
          </div>
        </Card>
      </div>
    )
  }

  const handleClearCart = () => {
    clearCart.mutate(undefined, {
      onError: (err) => {
        push({
          variant: 'error',
          message: err instanceof ApiError ? err.message : 'Could not clear cart.',
        })
      },
    })
  }

  return (
    <div>
      <h1 className="mb-6 text-[22px] font-extrabold text-ink">Cart</h1>
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-[1fr_320px]">
        <div className="rounded-3xl border-2 border-ink bg-white px-6 shadow-hard">
          {cart.items.map((item) => (
            <CartLineItem key={item.product_id} item={item} />
          ))}
        </div>

        <div className="lg:sticky lg:top-8 lg:self-start">
          <Card eyebrow="go-commerce / summary">
            <div className="flex items-center justify-between text-[13.5px] text-ink-muted">
              <span>Items</span>
              <span className="font-bold text-ink">{cart.total_items}</span>
            </div>
            <div className="mt-2 flex items-center justify-between border-t-2 border-line pt-4 text-[15px]">
              <span className="font-bold text-ink">Total</span>
              <span className="text-[18px] font-extrabold text-brand-700">
                {formatIDR(cart.total)}
              </span>
            </div>
            <div className="mt-6 flex flex-col gap-2 rounded-2xl border-2 border-dashed border-ink/25 bg-paper px-4 py-3">
              <p className="text-[12px] font-bold text-ink-muted">
                Testing checkout? This is Midtrans{' '}
                <span className="text-ink">sandbox</span> — no real money moves.
              </p>
              <p className="text-[11.5px] leading-relaxed text-ink-muted">
                Pick <span className="font-bold text-ink">Card Payment</span> and use{' '}
                <span className="font-mono">4811 1111 1111 1114</span>, any future expiry, CVV{' '}
                <span className="font-mono">123</span>, OTP{' '}
                <span className="font-mono">112233</span>.
              </p>
            </div>

            <div className="mt-4 flex flex-col gap-3">
              <Button
                type="button"
                onClick={checkoutFlow.start}
                disabled={checkoutFlow.phase !== 'idle'}
              >
                {CHECKOUT_LABEL[checkoutFlow.phase]}
              </Button>
              <button
                type="button"
                onClick={handleClearCart}
                disabled={clearCart.isPending}
                className="text-[12.5px] font-bold text-ink-muted hover:text-danger-600 disabled:opacity-40"
              >
                Clear cart
              </button>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
