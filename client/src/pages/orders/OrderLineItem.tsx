import { formatIDR } from '../../features/products/format'
import type { OrderItem } from '../../features/orders/types'

export function OrderLineItem({ item }: { item: OrderItem }) {
  return (
    <div className="flex items-center gap-4 border-b-2 border-line py-3 last:border-b-0">
      <div className="min-w-0 flex-1">
        <p className="truncate text-[14px] font-bold text-ink">{item.product_name}</p>
        <p className="mt-0.5 text-[12.5px] text-ink-muted">
          {item.quantity} × {formatIDR(item.price)}
        </p>
      </div>
      <p className="shrink-0 text-[14px] font-extrabold text-ink">{formatIDR(item.subtotal)}</p>
    </div>
  )
}
