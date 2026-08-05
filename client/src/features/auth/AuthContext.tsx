import { useMemo, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as authApi from './api'
import { AuthContext, type AuthContextValue } from './auth-context'
import { clearAccessToken, setAccessToken } from './token-store'
import type { User } from './types'

const SESSION_QUERY_KEY = ['session'] as const

async function bootSession(): Promise<User | null> {
  try {
    const { accessToken } = await authApi.refresh()
    setAccessToken(accessToken)
    const { user } = await authApi.me()
    return user
  } catch {
    clearAccessToken()
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()

  const sessionQuery = useQuery({
    queryKey: SESSION_QUERY_KEY,
    queryFn: bootSession,
    staleTime: Infinity,
    retry: false,
  })

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: ({ accessToken, user }) => {
      setAccessToken(accessToken)
      queryClient.setQueryData(SESSION_QUERY_KEY, user)
    },
  })

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSettled: () => {
      clearAccessToken()
      queryClient.setQueryData(SESSION_QUERY_KEY, null)
    },
  })

  const value = useMemo<AuthContextValue>(
    () => ({
      user: sessionQuery.data ?? null,
      isLoading: sessionQuery.isLoading,
      isAuthenticated: !!sessionQuery.data,
      login: async (input) => {
        const result = await loginMutation.mutateAsync(input)
        return result.user
      },
      logout: async () => {
        await logoutMutation.mutateAsync().catch(() => undefined)
      },
    }),
    [sessionQuery.data, sessionQuery.isLoading, loginMutation, logoutMutation],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
