import { Loader2 } from 'lucide-react'
import { Logo } from '../components/Logo'

export function Splash() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-paper px-4">
      <Logo />
      <Loader2 className="animate-spin text-ink-muted" size={22} />
    </div>
  )
}
