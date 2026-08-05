import { apiFetch } from '../../lib/http'
import type { User } from './types'

export function register(input: { full_name: string; email: string; password: string }) {
  return apiFetch<{ user: User }>('/users/register', { method: 'POST', body: input })
}

export function login(input: { email: string; password: string }) {
  return apiFetch<{ accessToken: string; user: User }>('/users/login', {
    method: 'POST',
    body: input,
  })
}

export function refresh() {
  return apiFetch<{ accessToken: string }>('/users/refresh', { method: 'POST' })
}

export function logout() {
  return apiFetch<{ message: string }>('/users/logout', { method: 'POST' })
}

export function me() {
  return apiFetch<{ user: User }>('/users/me')
}

export function verifyEmail(token: string) {
  return apiFetch<{ message: string }>(
    `/users/verify-email?token=${encodeURIComponent(token)}`,
    { method: 'POST' },
  )
}

export function resendVerification(email: string) {
  return apiFetch<{ message: string }>('/users/resend-verification', {
    method: 'POST',
    body: { email },
  })
}

export function forgotPassword(email: string) {
  return apiFetch<{ message: string }>('/users/forgot-password', {
    method: 'POST',
    body: { email },
  })
}

export function resetPassword(token: string, password: string) {
  return apiFetch<{ message: string }>(
    `/users/reset-password?token=${encodeURIComponent(token)}`,
    { method: 'POST', body: { password } },
  )
}
