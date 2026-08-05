import { Link, useSearchParams } from 'react-router-dom'
import { AlertTriangle, CheckCircle2, Loader2 } from 'lucide-react'
import { AuthCard } from '../components/AuthCard'
import { Button } from '../components/Button'
import { useVerifyEmail } from '../features/auth/hooks'

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const verifyEmail = useVerifyEmail(token)

  if (!token || verifyEmail.isError) {
    return (
      <AuthCard
        centered
        icon={
          <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-danger-600 bg-danger-50 text-danger-600">
            <AlertTriangle size={28} strokeWidth={2} />
          </span>
        }
        eyebrow="go-commerce / verification failed"
        title="Invalid or expired link"
        subtitle="This verification link no longer works. Try signing in to request a new one."
      >
        <Link to="/login">
          <Button type="button">Back to sign in</Button>
        </Link>
      </AuthCard>
    )
  }

  if (verifyEmail.isSuccess) {
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

  return (
    <AuthCard
      centered
      icon={
        <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-ink bg-paper text-ink">
          <Loader2 size={28} strokeWidth={2} className="animate-spin" />
        </span>
      }
      eyebrow="go-commerce / verifying"
      title="Verifying your email"
      subtitle="Hang tight, this only takes a moment."
    >
      {null}
    </AuthCard>
  )
}
