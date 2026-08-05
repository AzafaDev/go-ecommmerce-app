import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as productsApi from './api'
import type { ListProductsParams, ProductInput, UpdateProductInput } from './api'

const PRODUCTS_KEY = ['products'] as const
const ADMIN_PRODUCTS_KEY = ['admin-products'] as const

export function useProducts(params: ListProductsParams) {
  return useQuery({
    queryKey: [...PRODUCTS_KEY, params],
    queryFn: () => productsApi.listProducts(params),
  })
}

export function useProduct(id: string | undefined) {
  return useQuery({
    queryKey: [...PRODUCTS_KEY, id],
    queryFn: () => productsApi.getProduct(id!),
    enabled: !!id,
  })
}

export function useCategories() {
  return useQuery({
    queryKey: [...PRODUCTS_KEY, 'categories'],
    queryFn: () => productsApi.listCategories(),
  })
}

export function useAdminProducts(params: ListProductsParams) {
  return useQuery({
    queryKey: [...ADMIN_PRODUCTS_KEY, params],
    queryFn: () => productsApi.listAdminProducts(params),
  })
}

export function useAdminProduct(id: string | undefined) {
  return useQuery({
    queryKey: [...ADMIN_PRODUCTS_KEY, id],
    queryFn: () => productsApi.getAdminProduct(id!),
    enabled: !!id,
  })
}

// Admin sees categories from inactive products too, so a name can be
// reused when reactivating/replacing a discontinued product.
export function useAdminCategories() {
  return useQuery({
    queryKey: [...ADMIN_PRODUCTS_KEY, 'categories'],
    queryFn: () => productsApi.listAdminCategories(),
  })
}

function invalidateProductQueries(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: PRODUCTS_KEY })
  queryClient.invalidateQueries({ queryKey: ADMIN_PRODUCTS_KEY })
}

export function useCreateProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProductInput) => productsApi.createProduct(input),
    onSuccess: () => invalidateProductQueries(queryClient),
  })
}

export function useUpdateProduct(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateProductInput) => productsApi.updateProduct(id, input),
    onSuccess: () => invalidateProductQueries(queryClient),
  })
}

export function useDeleteProduct() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => productsApi.deleteProduct(id),
    onSuccess: () => invalidateProductQueries(queryClient),
  })
}
