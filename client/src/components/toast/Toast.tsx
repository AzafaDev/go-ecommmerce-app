import { AlertCircle, CheckCircle2, Info, X } from 'lucide-react'
import type { Toast as ToastData } from './ToastContext'

const VARIANT_STYLES = {
  success: 'border-brand-600 bg-brand-50 text-brand-700',
  error: 'border-danger-600 bg-danger-50 text-danger-700',
  info: 'border-ink bg-white text-ink',
}

const VARIANT_ICON = {
  success: CheckCircle2,
  error: AlertCircle,
  info: Info,
}

export function Toast({ toast, onDismiss }: { toast: ToastData; onDismiss: () => void }) {
  const Icon = VARIANT_ICON[toast.variant]

  return (
    <div
      className={`animate-rise flex items-start gap-2.5 rounded-2xl border-2 px-4 py-3 text-[13.5px] leading-relaxed shadow-hard ${VARIANT_STYLES[toast.variant]}`}
    >
      <Icon size={16} className="mt-0.5 shrink-0" />
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <p className="font-medium">{toast.message}</p>
        {toast.action && (
          <button
            type="button"
            onClick={toast.action.onClick}
            className="w-fit text-[12.5px] font-bold underline underline-offset-2 hover:no-underline"
          >
            {toast.action.label}
          </button>
        )}
      </div>
      <button
        type="button"
        onClick={onDismiss}
        aria-label="Dismiss"
        className="shrink-0 opacity-60 hover:opacity-100"
      >
        <X size={15} />
      </button>
    </div>
  )
}
