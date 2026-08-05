import type { ReactNode } from 'react'

type CardProps = {
  eyebrow?: string
  children: ReactNode
  centered?: boolean
}

export function Card({ eyebrow, children, centered }: CardProps) {
  return (
    <div
      className={`rounded-3xl border-2 border-ink bg-white p-8 shadow-hard ${centered ? 'text-center' : ''}`}
    >
      {eyebrow && (
        <p className="mb-4 inline-flex items-center rounded-full border-2 border-ink bg-brand-50 px-3 py-1 text-[11px] font-bold uppercase tracking-[0.08em] text-brand-700">
          {eyebrow}
        </p>
      )}
      {children}
    </div>
  )
}
