import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../features/auth/auth-context'
import { Splash } from './Splash'

export function PublicOnlyRoute() {
  const { isLoading, isAuthenticated, user } = useAuth()

  if (isLoading) return <Splash />
  if (isAuthenticated) {
    return <Navigate to={user?.role === 'admin' ? '/admin/products' : '/profile'} replace />
  }

  return <Outlet />
}
