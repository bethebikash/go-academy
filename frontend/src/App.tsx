import { BrowserRouter, Routes, Route, Navigate, Link } from 'react-router-dom'
import { AuthProvider } from './features/auth/AuthContext'
import ProtectedRoute from './components/shared/ProtectedRoute'
import AppLayout from './components/layout/AppLayout'
import Login from './routes/Login'
import Register from './routes/Register'
import Dashboard from './routes/Dashboard'
import AdminInvitations from './routes/admin/Invitations'
import AdminUsers from './routes/admin/Users'
import { Construction, ArrowLeft } from 'lucide-react'

function ComingSoon({ title, phase }: { title: string; phase: string }) {
  return (
    <div className="py-16 text-center space-y-4 max-w-md mx-auto">
      <div className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-neutral-900 border border-neutral-800 text-neutral-400">
        <Construction className="h-6 w-6" />
      </div>
      <div>
        <h2 className="text-xl font-bold text-white">{title}</h2>
        <p className="text-sm text-neutral-400 mt-1">This module is scheduled for {phase}.</p>
      </div>
      <div>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-2 text-xs font-medium text-indigo-400 hover:text-indigo-300"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to Dashboard
        </Link>
      </div>
    </div>
  )
}

function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-950 text-neutral-100 p-4">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-extrabold text-white">404</h1>
        <p className="text-neutral-400 text-sm">The requested page does not exist.</p>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 text-white text-xs font-medium py-2 px-4 transition"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Return to Dashboard
        </Link>
      </div>
    </div>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Public Auth Routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          {/* Protected Application Shell */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="learning-paths" element={<ComingSoon title="Learning Paths" phase="Phase 2" />} />
            <Route path="challenges" element={<ComingSoon title="Coding Challenges" phase="Phase 4" />} />
            <Route path="notes" element={<ComingSoon title="Personal Notes" phase="Phase 3" />} />
            <Route path="bookmarks" element={<ComingSoon title="Bookmarks" phase="Phase 3" />} />
            <Route path="progress" element={<ComingSoon title="Detailed Progress" phase="Phase 6" />} />

            {/* Protected Admin Routes */}
            <Route
              path="admin/invitations"
              element={
                <ProtectedRoute requiredRole="admin">
                  <AdminInvitations />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/users"
              element={
                <ProtectedRoute requiredRole="admin">
                  <AdminUsers />
                </ProtectedRoute>
              }
            />
          </Route>

          {/* Catch-all 404 */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
