import { useParams } from 'react-router-dom'
import { Card } from '../../components/Card'
import { DataRow } from '../../components/DataRow'
import { Badge } from '../../components/Badge'
import { Alert } from '../../components/Alert'
import { formatIDR } from '../../features/products/format'
import { useOrder } from '../../features/orders/hooks'
import { useResumePayment } from '../../features/orders/useResumePayment'
import { formatOrderDate, formatOrderId } from '../../features/orders/format'
import { ORDER_STATUS_BADGE_VARIANT, ORDER_STATUS_LABEL } from '../../features/orders/status'
import { OrderLineItem } from './OrderLineItem'

const INERT_STATUS_NOTE: Record<'cancelled' | 'expired', string> = {
  cancelled: 'This order was cancelled.',
  expired: "This order expired because payment wasn't completed in time.",
}

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data, isLoading, isError } = useOrder(id)
  const { resume, pendingOrderId } = useResumePayment()

  if (isLoading) return <p className="text-[14px] text-ink-muted">Loading order…</p>

  if (isError || !data) {
    return <p className="text-[14px] text-danger-600">Order not found.</p>
  }

  const { order } = data
  const isResuming = pendingOrderId === order.id

  return (
    <div className="mx-auto max-w-xl">
      <Card eyebrow="go-commerce / order">
        <div className="flex items-center justify-between gap-3">
          <span className="font-mono text-[13.5px] font-bold text-ink-muted">
            {formatOrderId(order.id)}
          </span>
          <Badge variant={ORDER_STATUS_BADGE_VARIANT[order.status]}>
            {ORDER_STATUS_LABEL[order.status]}
          </Badge>
        </div>

        {order.status === 'pending_payment' &&
          (order.snap_token ? (
            <div className="mt-4">
              <Alert
                variant="warning"
                action={
                  <button
                    type="button"
                    onClick={() => resume(order.id, order.snap_token!)}
                    disabled={isResuming}
                    className="text-[13px] font-bold text-warning-700 underline underline-offset-2 hover:opacity-80 disabled:opacity-60"
                  >
                    {isResuming ? 'Opening…' : 'Pay Now'}
                  </button>
                }
              >
                Complete your payment to secure this order.
              </Alert>
              <p className="mt-2 px-1 text-[11.5px] leading-relaxed text-ink-muted">
                Sandbox tip: pick <span className="font-bold text-ink">Card Payment</span>,{' '}
                <span className="font-mono">4811 1111 1111 1114</span>, any future expiry, CVV{' '}
                <span className="font-mono">123</span>, OTP <span className="font-mono">112233</span>.
              </p>
            </div>
          ) : (
            <p className="mt-4 text-[13px] text-ink-muted">
              Payment setup didn't complete for this order. It will be automatically cancelled
              shortly.
            </p>
          ))}

        {(order.status === 'cancelled' || order.status === 'expired') && (
          <p className="mt-4 text-[13px] text-ink-muted">{INERT_STATUS_NOTE[order.status]}</p>
        )}

        <div className="mt-6 flex flex-col">
          <DataRow label="Placed" value={formatOrderDate(order.created_at)} />
          <DataRow label="Paid" value={order.paid_at ? formatOrderDate(order.paid_at) : '—'} />
          <DataRow label="Total" value={formatIDR(order.total_amount)} />
        </div>

        <div className="mt-6 border-t-2 border-line pt-2">
          {order.items.map((item) => (
            <OrderLineItem key={item.product_id} item={item} />
          ))}
        </div>
      </Card>
    </div>
  )
}
