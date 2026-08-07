import { createContext, useContext } from 'react'

export type ToastVariant = 'success' | 'error' | 'info'

export type ToastAction = {
  label: string
  onClick: () => void
}

export type ToastInput = {
  variant: ToastVariant
  message: string
  action?: ToastAction
}

export type Toast = ToastInput & { id: string }

export type ToastContextValue = {
  push: (toast: ToastInput) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (!ctx) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return ctx
}
