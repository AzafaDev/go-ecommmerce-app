import { Minus, Plus } from 'lucide-react'

type QuantityStepperProps = {
  value: number
  min?: number
  max?: number
  onChange: (next: number) => void
  disabled?: boolean
}

export function QuantityStepper({ value, min = 1, max, onChange, disabled }: QuantityStepperProps) {
  const canDecrement = !disabled && value > min
  const canIncrement = !disabled && (max === undefined || value < max)

  return (
    <div className="inline-flex items-center gap-3 rounded-full border-2 border-ink bg-white px-1 py-1 shadow-hard-sm">
      <button
        type="button"
        aria-label="Decrease quantity"
        disabled={!canDecrement}
        onClick={() => onChange(value - 1)}
        className="flex h-6 w-6 items-center justify-center rounded-full text-ink transition-colors hover:bg-paper disabled:opacity-30"
      >
        <Minus size={13} />
      </button>
      <span className="min-w-4 text-center text-[13.5px] font-bold text-ink">{value}</span>
      <button
        type="button"
        aria-label="Increase quantity"
        disabled={!canIncrement}
        onClick={() => onChange(value + 1)}
        className="flex h-6 w-6 items-center justify-center rounded-full text-ink transition-colors hover:bg-paper disabled:opacity-30"
      >
        <Plus size={13} />
      </button>
    </div>
  )
}
