import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../features/auth/auth-context'
import { Splash } from './Splash'

export function AdminRoute() {
  const { isLoading, isAuthenticated, user } = useAuth()

  if (isLoading) return <Splash />
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'admin') return <Navigate to="/products" replace />

  return <Outlet />
}
