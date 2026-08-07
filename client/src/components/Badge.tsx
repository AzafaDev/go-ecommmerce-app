import type { ReactNode } from 'react'

type BadgeProps = {
  variant?: 'neutral' | 'success' | 'danger' | 'warning'
  children: ReactNode
}

export function Badge({ variant = 'neutral', children }: BadgeProps) {
  const styles = {
    neutral: 'bg-paper text-ink',
    success: 'bg-brand-50 text-brand-700',
    danger: 'bg-danger-50 text-danger-700',
    warning: 'bg-warning-50 text-warning-700',
  }[variant]

  return (
    <span
      className={`inline-flex shrink-0 items-center rounded-full border-2 border-ink px-2.5 py-1 text-[11.5px] font-bold capitalize ${styles}`}
    >
      {children}
    </span>
  )
}
