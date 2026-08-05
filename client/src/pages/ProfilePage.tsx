import { BadgeCheck, LogOut } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Logo } from '../components/Logo'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { DataRow } from '../components/DataRow'
import { useAuth } from '../features/auth/auth-context'

function initials(name: string) {
  return name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function formatDate(isoDate: string) {
  return new Date(isoDate).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export function ProfilePage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  if (!user) return null

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center bg-paper px-4 py-12">
      <div className="w-full max-w-[420px] animate-rise">
        <div className="mb-7 flex justify-center">
          <Logo />
        </div>

        <Card eyebrow="go-commerce / account">
          <div className="flex items-center gap-3.5">
            <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full border-2 border-ink bg-brand-600 text-[15px] font-bold text-white shadow-hard-sm">
              {initials(user.full_name)}
            </span>
            <div className="min-w-0">
              <p className="truncate text-[16px] font-extrabold text-ink">{user.full_name}</p>
              <p className="truncate text-[13.5px] text-ink-muted">{user.email}</p>
            </div>
            <span className="ml-auto inline-flex shrink-0 items-center rounded-full border-2 border-ink bg-brand-50 px-2.5 py-1 text-[11.5px] font-bold capitalize text-brand-700">
              {user.role}
            </span>
          </div>

          <div className="my-6 border-t-2 border-line" />

          <div className="flex flex-col">
            <DataRow label="Account ID" value={`GC-${user.id.split('-')[0].toUpperCase()}`} />
            <DataRow label="Member since" value={formatDate(user.created_at)} />
            <DataRow
              label="Email status"
              value={
                user.email_verified_at ? (
                  <span className="inline-flex items-center gap-1 text-brand-700">
                    <BadgeCheck size={13} /> Verified
                  </span>
                ) : (
                  'Unverified'
                )
              }
            />
          </div>

          <div className="mt-7">
            <Button variant="secondary" type="button" onClick={handleLogout}>
              <LogOut size={15} className="mr-2" />
              Log out
            </Button>
          </div>
        </Card>
      </div>
    </div>
  )
}
