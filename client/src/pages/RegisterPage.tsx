import { useState } from 'react'
import { Link } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { MailCheck } from 'lucide-react'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'
import { Alert } from '../components/Alert'
import { useRegister, useResendVerification } from '../features/auth/hooks'
import { registerSchema, type RegisterInput } from '../features/auth/schemas'
import { ApiError } from '../lib/http'

export function RegisterPage() {
  const registerMutation = useRegister()
  const resendVerification = useResendVerification()
  const [registeredEmail, setRegisteredEmail] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterInput) => {
    setFormError(null)
    try {
      await registerMutation.mutateAsync({
        full_name: data.full_name,
        email: data.email,
        password: data.password,
      })
      setRegisteredEmail(data.email)
    } catch (err) {
      setFormError(
        err instanceof ApiError ? err.message : 'Something went wrong. Please try again.',
      )
    }
  }

  if (registeredEmail) {
    return (
      <AuthCard
        centered
        icon={
          <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-brand-600 bg-brand-50 text-brand-600">
            <MailCheck size={28} strokeWidth={2} />
          </span>
        }
        eyebrow="go-commerce / almost there"
        title="Check your inbox"
        subtitle={`We sent a verification link to ${registeredEmail}. Click it to activate your account.`}
      >
        <div className="flex flex-col items-center gap-4">
          {resendVerification.isSuccess ? (
            <p className="text-[13px] text-brand-700">If that email exists, we sent a link.</p>
          ) : (
            <button
              type="button"
              onClick={() => resendVerification.mutate(registeredEmail)}
              disabled={resendVerification.isPending}
              className="text-[13px] font-bold text-brand-600 underline underline-offset-2 hover:text-brand-700 disabled:opacity-60"
            >
              {resendVerification.isPending ? 'Sending…' : "Didn't get it? Resend email"}
            </button>
          )}
          <Link to="/login" className="w-full">
            <Button type="button" variant="secondary">
              Back to sign in
            </Button>
          </Link>
        </div>
      </AuthCard>
    )
  }

  return (
    <AuthCard
      eyebrow="go-commerce / create account"
      title="Create your account"
      subtitle="Start shopping with go-commerce in less than a minute."
      footer={
        <>
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-brand-600 hover:text-brand-700">
            Sign in
          </Link>
        </>
      }
    >
      <form className="flex flex-col gap-5" onSubmit={handleSubmit(onSubmit)} noValidate>
        {formError && <Alert variant="error">{formError}</Alert>}

        <Field
          label="Full name"
          type="text"
          placeholder="Jane Doe"
          autoComplete="name"
          error={errors.full_name?.message}
          {...register('full_name')}
        />
        <Field
          label="Email"
          type="email"
          placeholder="you@example.com"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />
        <Field
          label="Password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          hint="At least 12 characters."
          error={errors.password?.message}
          {...register('password')}
        />
        <Field
          label="Confirm password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating account…' : 'Create account'}
        </Button>
      </form>
    </AuthCard>
  )
}
