import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { Card } from '../../components/Card'
import { Field } from '../../components/Input'
import { Button } from '../../components/Button'
import { Alert } from '../../components/Alert'
import { useAdminProduct, useAdminCategories, useCreateProduct, useUpdateProduct } from '../../features/products/hooks'
import { productSchema, type ProductFormInput } from '../../features/products/schemas'
import { ApiError } from '../../lib/http'

const NEW_CATEGORY_OPTION = '__new__'

export function AdminProductFormPage() {
  const { id } = useParams<{ id: string }>()
  const isEditMode = !!id
  const navigate = useNavigate()

  const { data: existing, isLoading: isLoadingProduct } = useAdminProduct(id)
  const { data: categoriesData } = useAdminCategories()
  const createProduct = useCreateProduct()
  const updateProduct = useUpdateProduct(id ?? '')

  const [isActive, setIsActive] = useState(true)
  const [isNewCategory, setIsNewCategory] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<ProductFormInput>({
    resolver: zodResolver(productSchema),
  })

  const categoryValue = watch('category')

  useEffect(() => {
    if (existing) {
      reset({
        name: existing.product.name,
        description: existing.product.description,
        price: Number(existing.product.price),
        stock: existing.product.stock,
        sku: existing.product.sku,
        category: existing.product.category,
      })
      setIsActive(existing.product.is_active)
    }
  }, [existing, reset])

  // Categories come from distinct existing values; make sure the product's
  // own category still shows up in the dropdown even if it hasn't loaded yet.
  const categoryOptions = useMemo(() => {
    const options = new Set(categoriesData?.categories ?? [])
    if (existing?.product.category) options.add(existing.product.category)
    return Array.from(options).sort()
  }, [categoriesData, existing])

  const handleCategorySelect = (value: string) => {
    if (value === NEW_CATEGORY_OPTION) {
      setIsNewCategory(true)
      setValue('category', '', { shouldValidate: true })
      return
    }
    setValue('category', value, { shouldValidate: true })
  }

  const onSubmit = async (data: ProductFormInput) => {
    setFormError(null)
    try {
      if (isEditMode) {
        // sku is intentionally dropped — it's immutable after create and
        // the update endpoint doesn't accept it.
        const { sku: _sku, ...rest } = data
        await updateProduct.mutateAsync({ ...rest, is_active: isActive })
      } else {
        await createProduct.mutateAsync(data)
      }
      navigate('/admin/products')
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message)
        return
      }
      setFormError('Something went wrong. Please try again.')
    }
  }

  if (isEditMode && isLoadingProduct) {
    return <p className="text-[14px] text-ink-muted">Loading product…</p>
  }

  return (
    <div className="mx-auto max-w-lg">
      <Card eyebrow={isEditMode ? 'go-commerce / edit product' : 'go-commerce / new product'}>
        <h1 className="mb-6 text-[20px] font-extrabold text-ink">
          {isEditMode ? 'Edit product' : 'Create product'}
        </h1>
        <form className="flex flex-col gap-5" onSubmit={handleSubmit(onSubmit)} noValidate>
          {formError && <Alert variant="error">{formError}</Alert>}

          <Field label="Name" error={errors.name?.message} {...register('name')} />
          <Field
            label="Description"
            error={errors.description?.message}
            {...register('description')}
          />
          <div className="grid grid-cols-2 gap-4">
            <Field
              label="Price (IDR)"
              type="number"
              step="1"
              error={errors.price?.message}
              {...register('price')}
            />
            <Field
              label="Stock"
              type="number"
              step="1"
              error={errors.stock?.message}
              {...register('stock')}
            />
          </div>
          <Field
            label="SKU"
            error={errors.sku?.message}
            readOnly={isEditMode}
            hint={isEditMode ? 'SKU cannot be changed after a product is created.' : undefined}
            {...register('sku')}
          />

          <div className="flex flex-col gap-1.5">
            <label className="text-[13px] font-bold text-ink" htmlFor="category-field">
              Category
            </label>
            {isNewCategory ? (
              <input
                id="category-field"
                autoFocus
                placeholder="New category name"
                className={`w-full rounded-2xl border-2 bg-white px-4 py-3 text-[15px] text-ink placeholder:text-ink-muted/60 outline-none transition-colors focus:ring-4 ${
                  errors.category
                    ? 'border-danger-600 focus:border-danger-600 focus:ring-danger-600/10'
                    : 'border-ink focus:border-brand-600 focus:ring-brand-600/10'
                }`}
                {...register('category')}
              />
            ) : (
              <select
                id="category-field"
                value={categoryValue || ''}
                onChange={(e) => handleCategorySelect(e.target.value)}
                className={`w-full rounded-2xl border-2 bg-white px-4 py-3 text-[15px] text-ink outline-none transition-colors focus:ring-4 ${
                  errors.category
                    ? 'border-danger-600 focus:border-danger-600 focus:ring-danger-600/10'
                    : 'border-ink focus:border-brand-600 focus:ring-brand-600/10'
                }`}
              >
                <option value="" disabled>
                  Select a category…
                </option>
                {categoryOptions.map((category) => (
                  <option key={category} value={category}>
                    {category}
                  </option>
                ))}
                <option value={NEW_CATEGORY_OPTION}>+ Add new category</option>
              </select>
            )}
            {isNewCategory && (
              <button
                type="button"
                onClick={() => {
                  setIsNewCategory(false)
                  setValue('category', '', { shouldValidate: true })
                }}
                className="self-start text-[12.5px] font-bold text-brand-600 hover:text-brand-700"
              >
                Choose from existing categories instead
              </button>
            )}
            {errors.category && (
              <p className="text-[12.5px] font-medium text-danger-600">{errors.category.message}</p>
            )}
          </div>

          {isEditMode && (
            <label className="flex items-center gap-2 text-[13px] font-bold text-ink">
              <input
                type="checkbox"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
                className="h-4 w-4 rounded border-2 border-ink"
              />
              Active
            </label>
          )}

          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving…' : isEditMode ? 'Save changes' : 'Create product'}
          </Button>
        </form>
      </Card>
    </div>
  )
}
