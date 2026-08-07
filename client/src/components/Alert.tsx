import type { ReactNode } from 'react'
import { AlertCircle, AlertTriangle, CheckCircle2 } from 'lucide-react'

type AlertProps = {
  variant: 'error' | 'success' | 'warning'
  children: ReactNode
  action?: ReactNode
}

const VARIANT_STYLES = {
  error: 'border-danger-600 bg-danger-50 text-danger-700',
  success: 'border-brand-600 bg-brand-50 text-brand-700',
  warning: 'border-warning-600 bg-warning-50 text-warning-700',
}

const VARIANT_ICON = {
  error: AlertTriangle,
  success: CheckCircle2,
  warning: AlertCircle,
}

export function Alert({ variant, children, action }: AlertProps) {
  const Icon = VARIANT_ICON[variant]

  return (
    <div
      className={`flex flex-col gap-2 rounded-2xl border-2 px-4 py-3 text-[13.5px] leading-relaxed ${VARIANT_STYLES[variant]}`}
    >
      <div className="flex items-start gap-2">
        <Icon size={16} className="mt-0.5 shrink-0" />
        <p className="font-medium">{children}</p>
      </div>
      {action && <div className="pl-6">{action}</div>}
    </div>
  )
}
