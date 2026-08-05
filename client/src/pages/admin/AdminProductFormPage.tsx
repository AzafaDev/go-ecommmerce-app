import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { Card } from '../../components/Card'
import { Field } from '../../components/Input'
import { Button } from '../../components/Button'
import { Alert } from '../../components/Alert'
import { useProduct, useCreateProduct, useUpdateProduct } from '../../features/products/hooks'
import { productSchema, type ProductFormInput } from '../../features/products/schemas'
import { ApiError } from '../../lib/http'

export function AdminProductFormPage() {
  const { id } = useParams<{ id: string }>()
  const isEditMode = !!id
  const navigate = useNavigate()

  const { data: existing, isLoading: isLoadingProduct } = useProduct(id)
  const createProduct = useCreateProduct()
  const updateProduct = useUpdateProduct(id ?? '')

  const [isActive, setIsActive] = useState(true)
  const [formError, setFormError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ProductFormInput>({
    resolver: zodResolver(productSchema),
  })

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

  const onSubmit = async (data: ProductFormInput) => {
    setFormError(null)
    try {
      if (isEditMode) {
        await updateProduct.mutateAsync({ ...data, is_active: isActive })
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
          <Field label="SKU" error={errors.sku?.message} {...register('sku')} />
          <Field label="Category" error={errors.category?.message} {...register('category')} />

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
