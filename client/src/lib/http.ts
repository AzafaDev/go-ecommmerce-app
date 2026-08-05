import { clearAccessToken, getAccessToken, setAccessToken } from '../features/auth/token-store'

const BASE_URL = import.meta.env.VITE_API_URL

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type Envelope<T> = {
  success: boolean
  message?: string
  data?: T
}

async function parseResponse<T>(res: Response): Promise<T> {
  let body: Envelope<T> | undefined
  try {
    body = await res.json()
  } catch {
    body = undefined
  }

  if (!res.ok || !body?.success) {
    throw new ApiError(body?.message ?? 'something went wrong', res.status)
  }

  return body.data as T
}

let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const res = await fetch(`${BASE_URL}/users/refresh`, {
        method: 'POST',
        credentials: 'include',
      })
      const data = await parseResponse<{ accessToken: string }>(res)
      setAccessToken(data.accessToken)
      return data.accessToken
    })().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

type RequestOptions = Omit<RequestInit, 'body'> & { body?: unknown }

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, headers, ...rest } = options

  const doFetch = (token: string | null) =>
    fetch(`${BASE_URL}${path}`, {
      ...rest,
      credentials: 'include',
      headers: {
        ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...headers,
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })

  const initialToken = getAccessToken()
  let res = await doFetch(initialToken)

  // Only retry via refresh when this request actually carried a session token —
  // a 401 from an anonymous, token-in-query endpoint (verify-email, reset-password)
  // means that token is invalid, not that the session expired.
  if (res.status === 401 && initialToken && path !== '/users/refresh') {
    try {
      await refreshAccessToken()
    } catch {
      clearAccessToken()
      throw new ApiError('unauthorized', 401)
    }
    res = await doFetch(getAccessToken())
  }

  return parseResponse<T>(res)
}
