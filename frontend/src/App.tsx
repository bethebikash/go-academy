import { BrowserRouter, Routes, Route, Navigate, Link } from 'react-router-dom'
import { AuthProvider } from './features/auth/AuthContext'
import ProtectedRoute from './components/shared/ProtectedRoute'
import AppLayout from './components/layout/AppLayout'
import Login from './routes/Login'
import Register from './routes/Register'
import Dashboard from './routes/Dashboard'
import LearningPaths from './routes/LearningPaths'
import PathDetail from './routes/PathDetail'
import CourseDetail from './routes/CourseDetail'
import LessonReader from './routes/LessonReader'
import Challenges from './routes/Challenges'
import ChallengeWorkspace from './routes/ChallengeWorkspace'
import Notes from './routes/Notes'
import Bookmarks from './routes/Bookmarks'
import Progress from './routes/Progress'
import VerifyCertificate from './routes/VerifyCertificate'
import AdminInvitations from './routes/admin/Invitations'
import AdminUsers from './routes/admin/Users'
import ContentManager from './routes/admin/ContentManager'
import LessonEditor from './routes/admin/LessonEditor'
import AdminChallenges from './routes/admin/AdminChallenges'
import ChallengeEditor from './routes/admin/ChallengeEditor'
import { ArrowLeft } from 'lucide-react'

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
          <Route path="/certificates/verify/:code" element={<VerifyCertificate />} />

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
            <Route path="learning-paths" element={<LearningPaths />} />
            <Route path="learning-paths/:slug" element={<PathDetail />} />
            <Route path="courses/:slug" element={<CourseDetail />} />
            <Route path="lessons/:slug" element={<LessonReader />} />
            <Route path="challenges" element={<Challenges />} />
            <Route path="challenges/:slug" element={<ChallengeWorkspace />} />
            <Route path="notes" element={<Notes />} />
            <Route path="bookmarks" element={<Bookmarks />} />
            <Route path="progress" element={<Progress />} />

            {/* Protected Admin Routes */}
            <Route
              path="admin/content"
              element={
                <ProtectedRoute requiredRole="admin">
                  <ContentManager />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/challenges"
              element={
                <ProtectedRoute requiredRole="admin">
                  <AdminChallenges />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/challenges/new"
              element={
                <ProtectedRoute requiredRole="admin">
                  <ChallengeEditor />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/challenges/:id/edit"
              element={
                <ProtectedRoute requiredRole="admin">
                  <ChallengeEditor />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/lessons/new"
              element={
                <ProtectedRoute requiredRole="admin">
                  <LessonEditor />
                </ProtectedRoute>
              }
            />
            <Route
              path="admin/lessons/:id/edit"
              element={
                <ProtectedRoute requiredRole="admin">
                  <LessonEditor />
                </ProtectedRoute>
              }
            />
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
