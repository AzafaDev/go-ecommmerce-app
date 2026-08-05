import { z } from 'zod'

export const productSchema = z.object({
  name: z.string().min(3, 'At least 3 characters').max(255, 'Too long'),
  description: z.string().max(2000, 'Too long'),
  price: z.coerce.number().gt(0, 'Must be greater than 0'),
  stock: z.coerce.number().int('Must be a whole number').gte(0, 'Cannot be negative'),
  sku: z.string().min(3, 'At least 3 characters').max(50, 'Too long'),
  category: z.string().min(1, 'Required').max(100, 'Too long'),
})
export type ProductFormInput = z.infer<typeof productSchema>
