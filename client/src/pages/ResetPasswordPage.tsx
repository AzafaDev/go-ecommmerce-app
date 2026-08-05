import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'

export function ResetPasswordPage() {
  return (
    <AuthCard
      eyebrow="go-commerce / new password"
      title="Set a new password"
      subtitle="Choose a strong password for your go-commerce account."
      footer={
        <>
          Back to{' '}
          <Link to="/login" className="font-medium text-brand-600 hover:text-brand-700">
            sign in
          </Link>
        </>
      }
    >
      <form className="flex flex-col gap-5">
        <Field
          label="New password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          hint="At least 12 characters."
        />
        <Field
          label="Confirm new password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
        />
        <Button type="submit">Reset password</Button>
      </form>
    </AuthCard>
  )
}
