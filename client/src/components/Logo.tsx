import { ShoppingBag } from 'lucide-react'

export function Logo() {
  return (
    <span className="inline-flex items-center gap-2 text-[15px] font-extrabold tracking-tight text-ink">
      <span className="flex h-7 w-7 items-center justify-center rounded-xl border-2 border-ink bg-brand-600 text-white shadow-hard-sm transition-transform hover:-rotate-6">
        <ShoppingBag size={15} strokeWidth={2.5} />
      </span>
      go-commerce
    </span>
  )
}
