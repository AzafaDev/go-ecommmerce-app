import { Eye, EyeOff } from 'lucide-react'
import { forwardRef, useId, useState, type InputHTMLAttributes } from 'react'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  hint?: string
  error?: string
}

export const Field = forwardRef<HTMLInputElement, InputProps>(function Field(
  { label, hint, error, id, type, ...rest },
  ref,
) {
  const generatedId = useId()
  const fieldId = id ?? generatedId
  const [revealed, setRevealed] = useState(false)
  const isPassword = type === 'password'

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={fieldId} className="text-[13px] font-bold text-ink">
        {label}
      </label>
      <div className="relative">
        <input
          ref={ref}
          id={fieldId}
          type={isPassword && revealed ? 'text' : type}
          aria-invalid={!!error}
          className={`w-full rounded-2xl border-2 bg-white px-4 py-3 text-[15px] text-ink placeholder:text-ink-muted/60 outline-none transition-colors focus:ring-4 ${
            error
              ? 'border-danger-600 focus:border-danger-600 focus:ring-danger-600/10'
              : 'border-ink focus:border-brand-600 focus:ring-brand-600/10'
          }`}
          {...rest}
        />
        {isPassword && (
          <button
            type="button"
            onClick={() => setRevealed((v) => !v)}
            className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-ink-muted hover:text-ink"
            aria-label={revealed ? 'Hide password' : 'Show password'}
            tabIndex={-1}
          >
            {revealed ? <EyeOff size={16} /> : <Eye size={16} />}
          </button>
        )}
      </div>
      {error ? (
        <p className="text-[12.5px] font-medium text-danger-600">{error}</p>
      ) : hint ? (
        <p className="text-[12.5px] text-ink-muted">{hint}</p>
      ) : null}
    </div>
  )
})
