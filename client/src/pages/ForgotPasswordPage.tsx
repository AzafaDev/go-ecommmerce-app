import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'

export function ForgotPasswordPage() {
  return (
    <AuthCard
      eyebrow="go-commerce / reset request"
      title="Forgot your password?"
      subtitle="Enter the email linked to your account and we'll send you a reset link."
      footer={
        <>
          Remembered it?{' '}
          <Link to="/login" className="font-medium text-brand-600 hover:text-brand-700">
            Sign in
          </Link>
        </>
      }
    >
      <form className="flex flex-col gap-5">
        <Field label="Email" type="email" placeholder="you@example.com" autoComplete="email" />
        <Button type="submit">Send reset link</Button>
      </form>
    </AuthCard>
  )
}
