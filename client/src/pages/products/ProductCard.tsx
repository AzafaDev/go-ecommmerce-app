import { Link } from 'react-router-dom'
import { Badge } from '../../components/Badge'
import { formatIDR } from '../../features/products/format'
import type { Product } from '../../features/products/types'
import { ProductImage } from './ProductImage'

export function ProductCard({ product }: { product: Product }) {
  return (
    <Link
      to={`/products/${product.id}`}
      className="flex flex-col gap-3 rounded-3xl border-2 border-ink bg-white p-6 shadow-hard transition-all hover:-translate-y-0.5"
    >
      <ProductImage src={product.image_url} alt={product.name} className="h-40 w-full border-2 border-ink" />
      <div className="flex items-start justify-between gap-2">
        <p className="text-[15px] font-extrabold text-ink">{product.name}</p>
        {product.stock === 0 && <Badge variant="danger">Out of stock</Badge>}
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
