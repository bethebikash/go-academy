import { ApiErrorResponse } from '../types'

let inMemoryToken: string | null = null

export function setAccessToken(token: string | null) {
  inMemoryToken = token
}

export function getAccessToken(): string | null {
  return inMemoryToken
}

export class ApiError extends Error {
  code: string
  details?: { field?: string; message: string }[]
  status: number

  constructor(status: number, code: string, message: string, details?: { field?: string; message: string }[]) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export async function apiRequest<T = any>(
  endpoint: string,
  options: RequestInit = {},
  isRetry = false,
): Promise<T> {
  const headers = new Headers(options.headers || {})

  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (inMemoryToken && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${inMemoryToken}`)
  }

  const url = endpoint.startsWith('http') ? endpoint : `/api/v1${endpoint.startsWith('/') ? '' : '/'}${endpoint}`

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include', // Includes HttpOnly cookies (refresh_token)
  })

  // Handle 401 Unauthorized by attempting transparent token refresh once
  if (response.status === 401 && !isRetry && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/refresh')) {
    try {
      const refreshRes = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      })

      if (refreshRes.ok) {
        const refreshData = await refreshRes.json()
        const newAccessToken = refreshData.data?.access_token
        if (newAccessToken) {
          setAccessToken(newAccessToken)
          return apiRequest<T>(endpoint, options, true)
        }
      } else {
        setAccessToken(null)
      }
    } catch {
      setAccessToken(null)
    }
  }

  if (response.status === 204) {
    return {} as T
  }

  let data: any
  try {
    data = await response.json()
  } catch {
    data = {}
  }

  if (!response.ok) {
    const errorData = data as ApiErrorResponse
    const code = errorData?.error?.code || 'UNKNOWN_ERROR'
    const message = errorData?.error?.message || `Request failed with HTTP status ${response.status}`
    const details = errorData?.error?.details
    throw new ApiError(response.status, code, message, details)
  }

  return (data?.data !== undefined ? data.data : data) as T
}
