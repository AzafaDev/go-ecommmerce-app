import { useState } from 'react'
import { Link } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { MailCheck } from 'lucide-react'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'
import { useForgotPassword } from '../features/auth/hooks'
import { forgotPasswordSchema, type ForgotPasswordInput } from '../features/auth/schemas'

export function ForgotPasswordPage() {
  const forgotPassword = useForgotPassword()
  const [submittedEmail, setSubmittedEmail] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordInput>({
    resolver: zodResolver(forgotPasswordSchema),
  })

  const onSubmit = async (data: ForgotPasswordInput) => {
    await forgotPassword.mutateAsync(data.email).catch(() => undefined)
    setSubmittedEmail(data.email)
  }

  if (submittedEmail) {
    return (
      <AuthCard
        centered
        icon={
          <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-brand-600 bg-brand-50 text-brand-600">
            <MailCheck size={28} strokeWidth={2} />
          </span>
        }
        eyebrow="go-commerce / reset requested"
        title="Check your inbox"
        subtitle={`If ${submittedEmail} is linked to a go-commerce account, we've sent a reset link to it.`}
      >
        <Link to="/login">
          <Button type="button">Back to sign in</Button>
        </Link>
      </AuthCard>
    )
  }

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
      <form className="flex flex-col gap-5" onSubmit={handleSubmit(onSubmit)} noValidate>
        <Field
          label="Email"
          type="email"
          placeholder="you@example.com"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Sending…' : 'Send reset link'}
        </Button>
      </form>
    </AuthCard>
  )
}
