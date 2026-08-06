import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { Field } from '../../components/Input'
import { Badge } from '../../components/Badge'
import { Pagination } from '../../components/Pagination'
import { CategoryChips } from '../../components/CategoryChips'
import {
  useAdminCategories,
  useAdminProducts,
  useDeleteProduct,
} from '../../features/products/hooks'
import { formatIDR } from '../../features/products/format'

export function AdminProductListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const search = searchParams.get('search') ?? ''
  const category = searchParams.get('category') ?? ''
  const page = Number(searchParams.get('page') ?? '1')

  const { data, isLoading } = useAdminProducts({
    search: search || undefined,
    category: category || undefined,
    page,
  })
  const { data: categoriesData } = useAdminCategories()
  const deleteProduct = useDeleteProduct()
  const [confirmingId, setConfirmingId] = useState<string | null>(null)

  const handlePageChange = (nextPage: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.set('page', String(nextPage))
      return next
    })
  }

  const handleSearchChange = (value: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      if (value) next.set('search', value)
      else next.delete('search')
      next.delete('page')
      return next
    })
  }

  const handleCategoryChange = (nextCategory: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      if (nextCategory) next.set('category', nextCategory)
      else next.delete('category')
      next.delete('page')
      return next
    })
  }

  const handleDelete = (id: string) => {
    if (confirmingId !== id) {
      setConfirmingId(id)
      return
    }
    deleteProduct.mutate(id, { onSettled: () => setConfirmingId(null) })
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-[26px] font-extrabold text-ink">Manage products</h1>
          <p className="text-[14px] text-ink-muted">Create, edit, and remove catalog products.</p>
        </div>
        <Link
          to="/admin/products/new"
          className="inline-flex items-center gap-1.5 rounded-full border-2 border-ink bg-brand-600 px-5 py-2.5 text-[14px] font-bold text-white shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-brand-700"
        >
          <Plus size={15} />
          New product
        </Link>
      </div>

      <div className="flex flex-col gap-4">
        <div className="max-w-sm">
          <Field
            label="Search"
            placeholder="Search products…"
            defaultValue={search}
            onChange={(e) => handleSearchChange(e.target.value)}
          />
        </div>

        {categoriesData && categoriesData.categories.length > 0 && (
          <CategoryChips
            categories={categoriesData.categories}
            value={category}
            onChange={handleCategoryChange}
          />
        )}
      </div>

      {isLoading && <p className="text-[14px] text-ink-muted">Loading products…</p>}

      {data && data.products.length === 0 && (
        <p className="text-[14px] text-ink-muted">No products found.</p>
      )}

      {data && data.products.length > 0 && (
        <div className="overflow-x-auto rounded-3xl border-2 border-ink bg-white shadow-hard">
          <table className="w-full text-left text-[13.5px]">
            <thead>
              <tr className="border-b-2 border-ink text-[11px] font-bold uppercase tracking-wide text-ink-muted">
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3">SKU</th>
                <th className="px-4 py-3">Category</th>
                <th className="px-4 py-3">Price</th>
                <th className="px-4 py-3">Stock</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {data.products.map((product) => (
                <tr key={product.id} className="border-b-2 border-line last:border-b-0">
                  <td className="px-4 py-3 font-bold text-ink">{product.name}</td>
                  <td className="px-4 py-3 font-mono text-ink-muted">{product.sku}</td>
                  <td className="px-4 py-3 text-ink-muted">{product.category || '—'}</td>
                  <td className="px-4 py-3 text-ink">{formatIDR(product.price)}</td>
                  <td className="px-4 py-3 text-ink">{product.stock}</td>
                  <td className="px-4 py-3">
                    <Badge variant={product.is_active ? 'success' : 'neutral'}>
                      {product.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Link
                        to={`/admin/products/${product.id}/edit`}
                        className="font-bold text-brand-600 hover:text-brand-700"
                      >
                        Edit
                      </Link>
                      <button
                        type="button"
                        onClick={() => handleDelete(product.id)}
                        disabled={deleteProduct.isPending}
                        className="font-bold text-danger-600 hover:text-danger-700 disabled:opacity-60"
                      >
                        {confirmingId === product.id ? 'Confirm delete?' : 'Delete'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {data && (
        <Pagination
          page={data.meta.page}
          totalPages={data.meta.total_pages}
          onPageChange={handlePageChange}
        />
      )}
    </div>
  )
}
