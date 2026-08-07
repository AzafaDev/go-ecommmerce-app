import { useState } from 'react'
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import { LogOut, Menu, ShoppingCart, X } from 'lucide-react'
import { Logo } from './Logo'
import { useAuth } from '../features/auth/auth-context'
import { useCartCount } from '../features/cart/hooks'

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `text-[13.5px] font-bold transition-colors ${isActive ? 'text-brand-700' : 'text-ink-muted hover:text-ink'}`

const mobileNavLinkClass = ({ isActive }: { isActive: boolean }) =>
  `text-[15px] font-bold transition-colors ${isActive ? 'text-brand-700' : 'text-ink hover:text-brand-700'}`

export function AppShell() {
  const { user, isAuthenticated, logout } = useAuth()
  const navigate = useNavigate()
  const cartCount = useCartCount(isAuthenticated)
  const [menuOpen, setMenuOpen] = useState(false)
  const closeMenu = () => setMenuOpen(false)

  const handleLogout = async () => {
    closeMenu()
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-svh bg-paper">
      <header className="border-b-2 border-ink bg-white">
        <div className="mx-auto flex max-w-5xl items-center gap-4 px-4 py-4 md:gap-6">
          <Link to="/products" className="flex items-center" onClick={closeMenu}>
            <Logo />
          </Link>
          <nav className="hidden flex-1 items-center gap-5 md:flex">
            <NavLink to="/products" className={navLinkClass}>
              Products
            </NavLink>
            {isAuthenticated && (
              <NavLink to="/profile" className={navLinkClass}>
                Profile
              </NavLink>
            )}
            {isAuthenticated && (
              <NavLink to="/orders" className={navLinkClass}>
                Orders
              </NavLink>
            )}
            {user?.role === 'admin' && (
              <NavLink to="/admin/products" className={navLinkClass}>
                Admin
              </NavLink>
            )}
          </nav>
          <div className="ml-auto flex items-center gap-3 md:ml-0">
            {isAuthenticated && (
              <Link
                to="/cart"
                aria-label="Cart"
                onClick={closeMenu}
                className="relative inline-flex items-center justify-center rounded-full border-2 border-ink bg-white p-2.5 shadow-hard-sm transition-all hover:-translate-y-0.5"
              >
                <ShoppingCart size={16} />
                {cartCount > 0 && (
                  <span className="absolute -right-1.5 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full border-2 border-ink bg-brand-600 px-1 text-[10.5px] font-bold text-white">
                    {cartCount > 99 ? '99+' : cartCount}
                  </span>
                )}
              </Link>
            )}
            <div className="hidden md:block">
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
            <button
              type="button"
              aria-label={menuOpen ? 'Close menu' : 'Open menu'}
              aria-expanded={menuOpen}
              onClick={() => setMenuOpen((open) => !open)}
              className="inline-flex items-center justify-center rounded-full border-2 border-ink bg-white p-2.5 shadow-hard-sm transition-all hover:-translate-y-0.5 md:hidden"
            >
              {menuOpen ? <X size={16} /> : <Menu size={16} />}
            </button>
          </div>
        </div>

        {menuOpen && (
          <nav className="border-t-2 border-ink bg-white px-4 py-4 md:hidden">
            <div className="flex flex-col gap-4">
              <NavLink to="/products" className={mobileNavLinkClass} onClick={closeMenu}>
                Products
              </NavLink>
              {isAuthenticated && (
                <NavLink to="/profile" className={mobileNavLinkClass} onClick={closeMenu}>
                  Profile
                </NavLink>
              )}
              {isAuthenticated && (
                <NavLink to="/orders" className={mobileNavLinkClass} onClick={closeMenu}>
                  Orders
                </NavLink>
              )}
              {user?.role === 'admin' && (
                <NavLink to="/admin/products" className={mobileNavLinkClass} onClick={closeMenu}>
                  Admin
                </NavLink>
              )}
              <div className="mt-1 border-t-2 border-line pt-4">
                {isAuthenticated ? (
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="inline-flex items-center gap-1.5 text-[15px] font-bold text-ink-muted hover:text-ink"
                  >
                    <LogOut size={15} />
                    Log out
                  </button>
                ) : (
                  <Link
                    to="/login"
                    onClick={closeMenu}
                    className="inline-flex items-center rounded-full border-2 border-ink bg-brand-600 px-4 py-2.5 text-[14px] font-bold text-white shadow-hard-sm"
                  >
                    Sign in
                  </Link>
                )}
              </div>
            </div>
          </nav>
        )}
      </header>
      <main className="mx-auto max-w-5xl px-4 py-6 md:py-10">
        <Outlet />
      </main>
    </div>
  )
}
