import { Link } from 'react-router-dom'
import { CheckCircle2 } from 'lucide-react'
import { AuthCard } from '../components/AuthCard'
import { Button } from '../components/Button'

export function VerifyEmailPage() {
  return (
    <AuthCard
      centered
      icon={
        <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-brand-600 bg-brand-50 text-brand-600">
          <CheckCircle2 size={28} strokeWidth={2} />
        </span>
      }
      eyebrow="go-commerce / verified"
      title="Email verified"
      subtitle="Your email has been confirmed. You can now sign in to your go-commerce account."
    >
      <Link to="/login">
        <Button type="button">Continue to sign in</Button>
      </Link>
    </AuthCard>
  )
}
