import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../features/auth/auth-context'
import { Splash } from './Splash'

export function ProtectedRoute() {
  const { isLoading, isAuthenticated } = useAuth()

  if (isLoading) return <Splash />
  if (!isAuthenticated) return <Navigate to="/login" replace />

  return <Outlet />
}
