import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'

export function LoginPage() {
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
      <form className="flex flex-col gap-5">
        <Field label="Email" type="email" placeholder="you@example.com" autoComplete="email" />
        <div className="flex flex-col gap-1.5">
          <Field label="Password" type="password" placeholder="••••••••••••" autoComplete="current-password" />
          <div className="flex justify-end">
            <Link
              to="/forgot-password"
              className="text-[13px] font-medium text-brand-600 hover:text-brand-700"
            >
              Forgot password?
            </Link>
          </div>
        </div>
        <Button type="submit">Sign in</Button>
      </form>
    </AuthCard>
  )
}
