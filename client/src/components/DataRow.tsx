import type { ReactNode } from 'react'

type DataRowProps = {
  label: string
  value: ReactNode
}

export function DataRow({ label, value }: DataRowProps) {
  return (
    <div className="flex items-center justify-between border-b-2 border-line py-2.5 last:border-b-0">
      <span className="text-[13px] font-bold text-ink-muted">{label}</span>
      <span className="inline-flex items-center rounded-full border-2 border-ink bg-paper px-3 py-1 font-mono text-[12.5px] font-bold text-ink">
        {value}
      </span>
    </div>
  )
}
