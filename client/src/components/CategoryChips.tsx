type CategoryChipsProps = {
  categories: string[]
  value: string
  onChange: (category: string) => void
}

export function CategoryChips({ categories, value, onChange }: CategoryChipsProps) {
  return (
    <div className="flex flex-wrap gap-2" role="group" aria-label="Filter by category">
      <Chip active={value === ''} onClick={() => onChange('')}>
        All
      </Chip>
      {categories.map((category) => (
        <Chip key={category} active={value === category} onClick={() => onChange(category)}>
          {category}
        </Chip>
      ))}
    </div>
  )
}

function Chip({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`inline-flex shrink-0 items-center rounded-full border-2 border-ink px-3.5 py-1.5 text-[13px] font-bold capitalize transition-all hover:-translate-y-0.5 active:translate-y-0 ${
        active
          ? 'bg-brand-600 text-white shadow-hard-sm'
          : 'bg-white text-ink shadow-hard-sm hover:bg-paper'
      }`}
    >
      {children}
    </button>
  )
}
