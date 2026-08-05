import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { LogOut } from 'lucide-react'
import { Logo } from './Logo'
import { useAuth } from '../features/auth/auth-context'

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `text-[13.5px] font-bold transition-colors ${isActive ? 'text-brand-700' : 'text-ink-muted hover:text-ink'}`

export function AppShell() {
  const { user, isAuthenticated, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-svh bg-paper">
      <header className="border-b-2 border-ink bg-white">
        <div className="mx-auto flex max-w-5xl items-center gap-6 px-4 py-4">
          <Link to="/products" className="flex items-center">
            <Logo />
          </Link>
          <nav className="flex flex-1 items-center gap-5">
            <NavLink to="/products" className={navLinkClass}>
              Products
            </NavLink>
            {isAuthenticated && (
              <NavLink to="/profile" className={navLinkClass}>
                Profile
              </NavLink>
            )}
            {user?.role === 'admin' && (
              <NavLink to="/admin/products" className={navLinkClass}>
                Admin
              </NavLink>
            )}
          </nav>
          {isAuthenticated ? (
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 text-[13.5px] font-bold text-ink-muted hover:text-ink"
            >
              <LogOut size={14} />
              Log out
            </button>
          ) : (
            <Link
              to="/login"
              className="inline-flex items-center rounded-full border-2 border-ink bg-brand-600 px-4 py-2 text-[13px] font-bold text-white shadow-hard-sm transition-all hover:-translate-y-0.5 hover:bg-brand-700"
            >
              Sign in
            </Link>
          )}
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-10">
        <Outlet />
      </main>
    </div>
  )
}
