import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { AlertTriangle } from 'lucide-react'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'
import { Alert } from '../components/Alert'
import { useResetPassword } from '../features/auth/hooks'
import { resetPasswordSchema, type ResetPasswordInput } from '../features/auth/schemas'
import { ApiError } from '../lib/http'

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const navigate = useNavigate()
  const resetPassword = useResetPassword()
  const [formError, setFormError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordInput>({
    resolver: zodResolver(resetPasswordSchema),
  })

  if (!token) {
    return (
      <AuthCard
        centered
        icon={
          <span className="inline-flex h-14 w-14 items-center justify-center rounded-full border-2 border-danger-600 bg-danger-50 text-danger-600">
            <AlertTriangle size={28} strokeWidth={2} />
          </span>
        }
        eyebrow="go-commerce / invalid link"
        title="This link is invalid"
        subtitle="The password reset link is missing or malformed. Request a new one to continue."
      >
        <Link to="/forgot-password">
          <Button type="button">Request a new link</Button>
        </Link>
      </AuthCard>
    )
  }

  const onSubmit = async (data: ResetPasswordInput) => {
    setFormError(null)
    try {
      await resetPassword.mutateAsync({ token, password: data.password })
      navigate('/login', { replace: true, state: { passwordReset: true } })
    } catch (err) {
      setFormError(
        err instanceof ApiError ? err.message : 'Something went wrong. Please try again.',
      )
    }
  }

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
      <form className="flex flex-col gap-5" onSubmit={handleSubmit(onSubmit)} noValidate>
        {formError && <Alert variant="error">{formError}</Alert>}

        <Field
          label="New password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          hint="At least 12 characters."
          error={errors.password?.message}
          {...register('password')}
        />
        <Field
          label="Confirm new password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Resetting…' : 'Reset password'}
        </Button>
      </form>
    </AuthCard>
  )
}
