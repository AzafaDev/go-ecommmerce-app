import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Field } from '../../components/Input'
import { Pagination } from '../../components/Pagination'
import { CategoryChips } from '../../components/CategoryChips'
import { useCategories, useProducts } from '../../features/products/hooks'
import { ProductCard } from './ProductCard'

export function ProductListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const search = searchParams.get('search') ?? ''
  const category = searchParams.get('category') ?? ''
  const page = Number(searchParams.get('page') ?? '1')

  const [searchInput, setSearchInput] = useState(search)
  const { data: categoriesData } = useCategories()

  useEffect(() => {
    const timeout = setTimeout(() => {
      setSearchParams((prev) => {
        if ((prev.get('search') ?? '') === searchInput) return prev
        const next = new URLSearchParams(prev)
        if (searchInput) next.set('search', searchInput)
        else next.delete('search')
        next.delete('page')
        return next
      })
    }, 350)
    return () => clearTimeout(timeout)
  }, [searchInput, setSearchParams])

  const { data, isLoading, isError } = useProducts({
    search: search || undefined,
    category: category || undefined,
    page,
  })

  const handlePageChange = (nextPage: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.set('page', String(nextPage))
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

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-[26px] font-extrabold text-ink">Products</h1>
        <p className="text-[14px] text-ink-muted">Browse the go-commerce catalog.</p>
      </div>

      <div className="flex flex-col gap-4">
        <div className="max-w-sm">
          <Field
            label="Search"
            placeholder="Search products…"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
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
      {isError && (
        <p className="text-[14px] text-danger-600">Something went wrong. Please try again.</p>
      )}

      {data && data.products.length === 0 && (
        <p className="text-[14px] text-ink-muted">No products found.</p>
      )}

      {data && data.products.length > 0 && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {data.products.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
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
