import type { ReactNode } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Logo } from './Logo'
import { Card } from './Card'

type AuthCardProps = {
  eyebrow: string
  title: string
  subtitle: string
  icon?: ReactNode
  centered?: boolean
  children: ReactNode
  footer?: ReactNode
}

export function AuthCard({
  eyebrow,
  title,
  subtitle,
  icon,
  centered,
  children,
  footer,
}: AuthCardProps) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center bg-paper px-4 py-12">
      <div className="w-full max-w-[420px] animate-rise">
        <div className="mb-4 flex justify-center">
          <Link
            to="/"
            className="inline-flex items-center gap-1.5 text-[13px] font-bold text-ink-muted transition-colors hover:text-ink"
          >
            <ArrowLeft size={14} strokeWidth={2.5} />
            Back to home
          </Link>
        </div>

        <div className="mb-7 flex justify-center">
          <Logo />
        </div>

        <Card eyebrow={eyebrow} centered={centered}>
          {icon && <div className={`mb-5 ${centered ? 'flex justify-center' : ''}`}>{icon}</div>}
          <h1 className="text-[28px] font-extrabold tracking-tight text-ink">
            {title}
          </h1>
          <p className="mt-1.5 text-[14.5px] leading-relaxed text-ink-muted">
            {subtitle}
          </p>

          <div className="mt-7 flex flex-col gap-5">{children}</div>

          {footer && (
            <>
              <div className="my-6 border-t-2 border-line" />
              <div className="text-center text-[14px] text-ink-muted">
                {footer}
              </div>
            </>
          )}
        </Card>
      </div>
    </div>
  )
}
