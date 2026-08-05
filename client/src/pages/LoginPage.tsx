import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'
import { Alert } from '../components/Alert'
import { useAuth } from '../features/auth/auth-context'
import { useResendVerification } from '../features/auth/hooks'
import { loginSchema, type LoginInput } from '../features/auth/schemas'
import { ApiError } from '../lib/http'

type FormError = {
  message: string
  unverified: boolean
}

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const resendVerification = useResendVerification()

  const [formError, setFormError] = useState<FormError | null>(null)

  const passwordResetConfirmed = Boolean(
    (location.state as { passwordReset?: boolean } | null)?.passwordReset,
  )

  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginInput) => {
    setFormError(null)
    try {
      await login(data)
      navigate('/profile', { replace: true })
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError({ message: err.message, unverified: err.status === 403 })
        return
      }
      setFormError({ message: 'Something went wrong. Please try again.', unverified: false })
    }
  }

  return (
    <AuthCard
      eyebrow="go-commerce / sign in"
      title="Welcome back"
      subtitle="Sign in to continue to your go-commerce account."
      footer={
        <>
          Don&apos;t have an account?{' '}
          <Link to="/register" className="font-medium text-brand-600 hover:text-brand-700">
            Create one
          </Link>
        </>
      }
    >
      <form className="flex flex-col gap-5" onSubmit={handleSubmit(onSubmit)} noValidate>
        {passwordResetConfirmed && !formError && (
          <Alert variant="success">
            Your password has been updated. Sign in with your new password.
          </Alert>
        )}

        {formError && (
          <Alert
            variant="error"
            action={
              formError.unverified ? (
                resendVerification.isSuccess ? (
                  <p className="text-[13px] text-brand-700">
                    If that email exists, we sent a link.
                  </p>
                ) : (
                  <button
                    type="button"
                    onClick={() => resendVerification.mutate(getValues('email'))}
                    disabled={resendVerification.isPending}
                    className="text-[13px] font-bold text-danger-700 underline underline-offset-2 hover:opacity-80 disabled:opacity-60"
                  >
                    {resendVerification.isPending ? 'Sending…' : 'Resend verification email'}
                  </button>
                )
              ) : undefined
            }
          >
            {formError.message}
          </Alert>
        )}

        <Field
          label="Email"
          type="email"
          placeholder="you@example.com"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />
        <div className="flex flex-col gap-1.5">
          <Field
            label="Password"
            type="password"
            placeholder="••••••••••••"
            autoComplete="current-password"
            error={errors.password?.message}
            {...register('password')}
          />
          <div className="flex justify-end">
            <Link
              to="/forgot-password"
              className="text-[13px] font-medium text-brand-600 hover:text-brand-700"
            >
              Forgot password?
            </Link>
          </div>
        </div>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Signing in…' : 'Sign in'}
        </Button>
      </form>
    </AuthCard>
  )
}
