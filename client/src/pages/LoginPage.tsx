import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { ShieldCheck, User } from 'lucide-react'
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

// Seeded via `make seed` (api/cmd/seed) — safe to expose, demo data only.
const DEMO_ACCOUNTS = {
  customer: { email: 'customer1@go-commerce.local', password: 'password1234' },
  admin: { email: 'admin@go-commerce.local', password: 'password1234' },
} as const

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const resendVerification = useResendVerification()

  const [formError, setFormError] = useState<FormError | null>(null)
  const [demoLoading, setDemoLoading] = useState<'customer' | 'admin' | null>(null)

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

  const handleDemoLogin = async (role: 'customer' | 'admin') => {
    setFormError(null)
    setDemoLoading(role)
    try {
      await login(DEMO_ACCOUNTS[role])
      navigate(role === 'admin' ? '/admin/products' : '/profile', { replace: true })
    } catch {
      setFormError({ message: 'Demo login is unavailable right now.', unverified: false })
    } finally {
      setDemoLoading(null)
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
      <div className="flex flex-col gap-3 rounded-2xl border-2 border-dashed border-ink/25 bg-paper px-4 py-4">
        <p className="text-[13px] font-bold text-ink-muted">
          Recruiter? Skip the form — jump in with a live demo account.
        </p>
        <div className="flex gap-2.5">
          <button
            type="button"
            onClick={() => handleDemoLogin('customer')}
            disabled={demoLoading !== null}
            className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-full border-2 border-ink bg-white px-3 py-2.5 text-[13px] font-bold text-ink shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-paper active:translate-x-[2px] active:translate-y-[2px] active:shadow-hard-press disabled:pointer-events-none disabled:opacity-60"
          >
            <User size={14} />
            {demoLoading === 'customer' ? 'Signing in…' : 'Customer demo'}
          </button>
          <button
            type="button"
            onClick={() => handleDemoLogin('admin')}
            disabled={demoLoading !== null}
            className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-full border-2 border-ink bg-brand-600 px-3 py-2.5 text-[13px] font-bold text-white shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-brand-700 active:translate-x-[2px] active:translate-y-[2px] active:shadow-hard-press disabled:pointer-events-none disabled:opacity-60"
          >
            <ShieldCheck size={14} />
            {demoLoading === 'admin' ? 'Signing in…' : 'Admin demo'}
          </button>
        </div>
        <p className="text-[11.5px] leading-relaxed text-ink-muted">
          Both accounts use <span className="font-mono">password1234</span> — feel free to sign in
          manually with{' '}
          <span className="font-mono">{DEMO_ACCOUNTS.customer.email}</span> or{' '}
          <span className="font-mono">{DEMO_ACCOUNTS.admin.email}</span> below too.
        </p>
      </div>

      <div className="flex items-center gap-3">
        <div className="h-px flex-1 bg-line" />
        <span className="text-[11px] font-bold uppercase tracking-wide text-ink-muted">
          or sign in manually
        </span>
        <div className="h-px flex-1 bg-line" />
      </div>

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
