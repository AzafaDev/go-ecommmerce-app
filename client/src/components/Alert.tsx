import type { ReactNode } from 'react'
import { AlertTriangle, CheckCircle2 } from 'lucide-react'

type AlertProps = {
  variant: 'error' | 'success'
  children: ReactNode
  action?: ReactNode
}

export function Alert({ variant, children, action }: AlertProps) {
  const isError = variant === 'error'

  return (
    <div
      className={`flex flex-col gap-2 rounded-2xl border-2 px-4 py-3 text-[13.5px] leading-relaxed ${
        isError
          ? 'border-danger-600 bg-danger-50 text-danger-700'
          : 'border-brand-600 bg-brand-50 text-brand-700'
      }`}
    >
      <div className="flex items-start gap-2">
        {isError ? (
          <AlertTriangle size={16} className="mt-0.5 shrink-0" />
        ) : (
          <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
        )}
        <p className="font-medium">{children}</p>
      </div>
      {action && <div className="pl-6">{action}</div>}
    </div>
  )
}
