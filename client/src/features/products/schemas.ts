import { z } from 'zod'

// Mirrors the API's custom "price" validator (pkg/validation): up to 2
// decimal digits, no negative sign — kept in sync deliberately so the form
// rejects the same inputs the server would, instead of round-tripping a 400.
const priceRegex = /^\d+(\.\d{1,2})?$/

export const productSchema = z.object({
  name: z.string().min(3, 'At least 3 characters').max(255, 'Too long'),
  description: z.string().max(2000, 'Too long'),
  price: z
    .string()
    .regex(priceRegex, 'Use up to 2 decimal digits, e.g. 19.99')
    .refine((val) => Number(val) > 0, 'Must be greater than 0'),
  stock: z.coerce.number().int('Must be a whole number').gte(0, 'Cannot be negative'),
  sku: z.string().min(3, 'At least 3 characters').max(50, 'Too long'),
  category: z.string().min(1, 'Required').max(100, 'Too long'),
  // Empty string means "no image" — mirrors the backend's PUT semantics
  // where image_url: "" clears the image.
  image_url: z.union([z.literal(''), z.string().url('Must be a valid URL')]),
})
export type ProductFormInput = z.infer<typeof productSchema>
