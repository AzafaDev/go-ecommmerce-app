import { useMutation, useQuery } from '@tanstack/react-query'
import * as authApi from './api'

export function useRegister() {
  return useMutation({ mutationFn: authApi.register })
}

export function useResendVerification() {
  return useMutation({ mutationFn: authApi.resendVerification })
}

export function useForgotPassword() {
  return useMutation({ mutationFn: authApi.forgotPassword })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: ({ token, password }: { token: string; password: string }) =>
      authApi.resetPassword(token, password),
  })
}

export function useVerifyEmail(token: string | null) {
  return useQuery({
    queryKey: ['verify-email', token],
    queryFn: () => authApi.verifyEmail(token!),
    enabled: !!token,
    retry: false,
    staleTime: Infinity,
  })
}
