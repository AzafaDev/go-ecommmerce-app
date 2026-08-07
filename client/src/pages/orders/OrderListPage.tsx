import { Link } from 'react-router-dom'
import { ShoppingBag } from 'lucide-react'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Badge } from '../../components/Badge'
import { formatIDR } from '../../features/products/format'
import { useOrders } from '../../features/orders/hooks'
import { useResumePayment } from '../../features/orders/useResumePayment'
import { formatOrderDate, formatOrderId } from '../../features/orders/format'
import { ORDER_STATUS_BADGE_VARIANT, ORDER_STATUS_LABEL } from '../../features/orders/status'

export function OrderListPage() {
  const { data, isLoading, isError } = useOrders()
  const { resume, pendingOrderId } = useResumePayment()

  if (isLoading) return <p className="text-[14px] text-ink-muted">Loading orders…</p>

  if (isError || !data) {
    return <p className="text-[14px] text-danger-600">Could not load your orders.</p>
  }

  const { orders } = data

  if (orders.length === 0) {
    return (
      <div className="mx-auto max-w-md">
        <Card centered>
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border-2 border-ink bg-paper text-ink-muted">
            <ShoppingBag size={22} />
          </div>
          <h1 className="text-[18px] font-extrabold text-ink">No orders yet</h1>
          <p className="mt-2 text-[13.5px] text-ink-muted">
            You haven't placed any orders yet.
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

  return (
    <div>
      <h1 className="mb-6 text-[22px] font-extrabold text-ink">Order History</h1>
      <div className="flex flex-col gap-4">
        {orders.map((order) => (
          <div
            key={order.id}
            className="rounded-3xl border-2 border-ink bg-white p-6 shadow-hard transition-all hover:-translate-y-0.5"
          >
            <Link to={`/orders/${order.id}`} className="flex items-center justify-between gap-4">
              <div className="min-w-0">
                <div className="flex items-center gap-2.5">
                  <span className="font-mono text-[12.5px] font-bold text-ink-muted">
                    {formatOrderId(order.id)}
                  </span>
                  <Badge variant={ORDER_STATUS_BADGE_VARIANT[order.status]}>
                    {ORDER_STATUS_LABEL[order.status]}
                  </Badge>
                </div>
                <p className="mt-1.5 text-[12.5px] text-ink-muted">
                  {formatOrderDate(order.created_at)} · {order.items.length}{' '}
                  {order.items.length === 1 ? 'item' : 'items'}
                </p>
              </div>
              <p className="shrink-0 text-[16px] font-extrabold text-ink">
                {formatIDR(order.total_amount)}
              </p>
            </Link>
            {order.status === 'pending_payment' && order.snap_token && (
              <div className="mt-4 border-t-2 border-line pt-4">
                <button
                  type="button"
                  onClick={() => resume(order.id, order.snap_token!)}
                  disabled={pendingOrderId === order.id}
                  className="inline-flex items-center rounded-full border-2 border-ink bg-brand-600 px-4 py-2 text-[12.5px] font-bold text-white shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-brand-700 disabled:opacity-50"
                >
                  {pendingOrderId === order.id ? 'Opening…' : 'Pay Now'}
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
