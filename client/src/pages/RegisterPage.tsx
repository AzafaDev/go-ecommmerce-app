import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { Field } from '../components/Input'
import { Button } from '../components/Button'

export function RegisterPage() {
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
      <form className="flex flex-col gap-5">
        <Field label="Full name" type="text" placeholder="Jane Doe" autoComplete="name" />
        <Field label="Email" type="email" placeholder="you@example.com" autoComplete="email" />
        <Field
          label="Password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
          hint="At least 12 characters."
        />
        <Field
          label="Confirm password"
          type="password"
          placeholder="••••••••••••"
          autoComplete="new-password"
        />
        <Button type="submit">Create account</Button>
      </form>
    </AuthCard>
  )
}
