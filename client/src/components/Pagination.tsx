import { ChevronLeft, ChevronRight } from 'lucide-react'

type PaginationProps = {
  page: number
  totalPages: number
  onPageChange: (page: number) => void
}

const navButtonClass =
  'inline-flex items-center justify-center rounded-full border-2 border-ink bg-white p-2 shadow-hard-sm transition-all hover:-translate-y-0.5 active:translate-x-[2px] active:translate-y-[2px] active:shadow-hard-press disabled:pointer-events-none disabled:opacity-40'

export function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  if (totalPages <= 1) return null

  return (
    <div className="flex items-center justify-center gap-3">
      <button
        type="button"
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        className={navButtonClass}
        aria-label="Previous page"
      >
        <ChevronLeft size={16} />
      </button>
      <span className="text-[13px] font-bold text-ink-muted">
        Page {page} of {totalPages}
      </span>
      <button
        type="button"
        onClick={() => onPageChange(page + 1)}
        disabled={page >= totalPages}
        className={navButtonClass}
        aria-label="Next page"
      >
        <ChevronRight size={16} />
      </button>
    </div>
  )
}
