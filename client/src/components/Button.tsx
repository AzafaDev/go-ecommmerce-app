import type { ButtonHTMLAttributes, ReactNode } from 'react'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary'
  children: ReactNode
}

export function Button({
  variant = 'primary',
  className = '',
  children,
  ...rest
}: ButtonProps) {
  const base =
    'inline-flex w-full items-center justify-center rounded-full border-2 border-ink px-6 py-3.5 text-[15px] font-bold shadow-hard-sm transition-all hover:-translate-y-0.5 active:translate-x-[2px] active:translate-y-[2px] active:shadow-hard-press focus-visible:outline-offset-2'
  const styles =
    variant === 'primary'
      ? 'bg-brand-600 text-white hover:bg-brand-700 active:bg-brand-800'
      : 'bg-white text-ink hover:bg-paper'

  return (
    <button className={`${base} ${styles} ${className}`} {...rest}>
      {children}
    </button>
  )
}
