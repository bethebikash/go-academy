import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { User, AuthResponse } from '../../types'
import { apiRequest, setAccessToken } from '../../lib/api'

interface AuthContextType {
  user: User | null
  loading: boolean
  isAuthenticated: boolean
  isAdmin: boolean
  login: (email: string, password: string) => Promise<User>
  register: (invitationToken: string, username: string, password: string) => Promise<User>
  logout: () => Promise<void>
  refreshProfile: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState<boolean>(true)

  const refreshProfile = useCallback(async () => {
    try {
      const userData = await apiRequest<User>('/auth/me')
      setUser(userData)
    } catch {
      setUser(null)
      setAccessToken(null)
    }
  }, [])

  useEffect(() => {
    // Attempt to hydrate session from existing cookie
    async function initAuth() {
      try {
        const refreshRes = await fetch('/api/v1/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        })
        if (refreshRes.ok) {
          const data = await refreshRes.json()
          if (data.data?.access_token) {
            setAccessToken(data.data.access_token)
            await refreshProfile()
          }
        }
      } catch {
        // Not authenticated
      } finally {
        setLoading(false)
      }
    }

    initAuth()
  }, [refreshProfile])

  const login = async (email: string, password: string): Promise<User> => {
    const data = await apiRequest<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    setAccessToken(data.access_token)
    setUser(data.user)
    return data.user
  }

  const register = async (invitationToken: string, username: string, password: string): Promise<User> => {
    const data = await apiRequest<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        invitation_token: invitationToken,
        username,
        password,
      }),
    })
    setAccessToken(data.access_token)
    setUser(data.user)
    return data.user
  }

  const logout = async () => {
    try {
      await apiRequest('/auth/logout', { method: 'POST' })
    } catch {
      // Ignore logout errors
    } finally {
      setAccessToken(null)
      setUser(null)
    }
  }

  const value: AuthContextType = {
    user,
    loading,
    isAuthenticated: !!user,
    isAdmin: user?.role === 'admin',
    login,
    register,
    logout,
    refreshProfile,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
