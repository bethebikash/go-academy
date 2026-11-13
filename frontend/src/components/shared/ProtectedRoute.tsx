import React from 'react'
import { Navigate, useLocation, Link } from 'react-router-dom'
import { Loader2, ShieldAlert, ArrowLeft } from 'lucide-react'
import { useAuth } from '../../features/auth/AuthContext'
import { UserRole } from '../../types'

interface ProtectedRouteProps {
  children: React.ReactNode
  requiredRole?: UserRole
}

export default function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { user, loading, isAuthenticated } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-neutral-950 text-neutral-400 gap-3">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
        <p className="text-sm">Authenticating session...</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (requiredRole && user?.role !== requiredRole) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-950 text-neutral-100 p-4">
        <div className="max-w-md w-full bg-neutral-900/60 p-8 rounded-2xl border border-red-500/20 text-center space-y-5">
          <div className="inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-red-950/60 text-red-400 border border-red-500/30">
            <ShieldAlert className="h-7 w-7" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">403 — Access Denied</h2>
            <p className="text-sm text-neutral-400 mt-2">
              You do not have the required administrative permissions to access this screen.
            </p>
          </div>
          <div>
            <Link
              to="/dashboard"
              className="inline-flex items-center gap-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 text-white text-sm font-medium py-2.5 px-4 transition"
            >
              <ArrowLeft className="h-4 w-4" />
              Return to Learner Dashboard
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
