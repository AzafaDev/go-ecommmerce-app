import { useState } from 'react'
import { ImageOff } from 'lucide-react'

export function ProductImage({
  src,
  alt,
  className = '',
}: {
  src: string
  alt: string
  className?: string
}) {
  const [failed, setFailed] = useState(false)
  const showPlaceholder = !src || failed

  return (
    <div className={`overflow-hidden rounded-2xl ${className}`}>
      {showPlaceholder ? (
        <div className="flex h-full w-full items-center justify-center bg-paper text-ink-muted">
          <ImageOff size={28} strokeWidth={1.75} />
        </div>
      ) : (
        <img
          src={src}
          alt={alt}
          className="h-full w-full object-cover"
          onError={() => setFailed(true)}
        />
      )}
    </div>
  )
}
