import { useCallback, useMemo, useRef, useState, type ReactNode } from 'react'
import { ToastContext, type Toast as ToastData, type ToastInput } from './ToastContext'
import { Toast } from './Toast'

const DURATION_MS: Record<ToastData['variant'], number> = {
  success: 3500,
  info: 3500,
  error: 5000,
}

let nextId = 0

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastData[]>([])
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>())

  const dismiss = useCallback((id: string) => {
    setToasts((current) => current.filter((toast) => toast.id !== id))
    const timer = timers.current.get(id)
    if (timer) {
      clearTimeout(timer)
      timers.current.delete(id)
    }
  }, [])

  const push = useCallback(
    (input: ToastInput) => {
      const id = `toast-${++nextId}`
      setToasts((current) => [...current, { ...input, id }])
      timers.current.set(
        id,
        setTimeout(() => dismiss(id), DURATION_MS[input.variant]),
      )
    },
    [dismiss],
  )

  const value = useMemo(() => ({ push }), [push])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex w-[min(360px,calc(100vw-2rem))] flex-col gap-2">
        {toasts.map((toast) => (
          <Toast key={toast.id} toast={toast} onDismiss={() => dismiss(toast.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}
