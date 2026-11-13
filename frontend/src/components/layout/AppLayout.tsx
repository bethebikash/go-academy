import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import {
  Terminal,
  BookOpen,
  Code2,
  FileText,
  Bookmark,
  Award,
  Shield,
  Activity,
  Users,
  Mail,
  Layers,
  LogOut,
  CheckCircle2,
} from 'lucide-react'
import { useAuth } from '../../features/auth/AuthContext'

export default function AppLayout() {
  const { user, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  const navClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition ${
      isActive
        ? 'bg-neutral-800 text-white font-semibold shadow-sm'
        : 'text-neutral-400 hover:bg-neutral-800/60 hover:text-white'
    }`

  return (
    <div className="flex min-h-screen bg-neutral-950 text-neutral-100">
      {/* Sidebar */}
      <aside className="w-64 border-r border-neutral-800/80 bg-neutral-900/40 flex flex-col justify-between p-4 shrink-0">
        <div className="space-y-6">
          {/* Logo */}
          <Link to="/dashboard" className="flex items-center gap-3 px-3 py-2 border-b border-neutral-800/80 pb-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-600 text-white shadow-md shadow-indigo-600/30 font-bold">
              <Terminal className="h-5 w-5" />
            </div>
            <div>
              <span className="font-semibold tracking-tight text-white block">Go Academy</span>
              <span className="text-xs text-neutral-500">Interactive Mastery</span>
            </div>
          </Link>

          {/* Navigation Items */}
          <nav className="space-y-6">
            <div>
              <span className="text-xs font-semibold uppercase tracking-wider text-neutral-500 px-3">Learn</span>
              <div className="mt-2 space-y-1">
                <NavLink to="/dashboard" className={navClass}>
                  <Activity className="h-4 w-4 text-indigo-400" />
                  Dashboard
                </NavLink>
                <NavLink to="/learning-paths" className={navClass}>
                  <BookOpen className="h-4 w-4" />
                  Learning Paths
                </NavLink>
                <NavLink to="/challenges" className={navClass}>
                  <Code2 className="h-4 w-4" />
                  Challenges
                </NavLink>
                <NavLink to="/notes" className={navClass}>
                  <FileText className="h-4 w-4" />
                  My Notes
                </NavLink>
                <NavLink to="/bookmarks" className={navClass}>
                  <Bookmark className="h-4 w-4" />
                  Bookmarks
                </NavLink>
                <NavLink to="/progress" className={navClass}>
                  <Award className="h-4 w-4" />
                  My Progress
                </NavLink>
              </div>
            </div>

            {/* Admin Section (Conditional on Admin Role) */}
            {isAdmin && (
              <div>
                <span className="text-xs font-semibold uppercase tracking-wider text-amber-500/80 px-3 flex items-center gap-1.5">
                  <Shield className="h-3 w-3" />
                  Administration
                </span>
                <div className="mt-2 space-y-1">
                  <NavLink to="/admin/content" className={navClass}>
                    <Layers className="h-4 w-4 text-amber-400" />
                    Content Manager
                  </NavLink>
                  <NavLink to="/admin/invitations" className={navClass}>
                    <Mail className="h-4 w-4 text-amber-400" />
                    Invitations
                  </NavLink>
                  <NavLink to="/admin/users" className={navClass}>
                    <Users className="h-4 w-4 text-amber-400" />
                    Users
                  </NavLink>
                </div>
              </div>
            )}
          </nav>
        </div>

        {/* User Card & Logout */}
        <div className="pt-4 border-t border-neutral-800/80 space-y-3">
          <div className="flex items-center justify-between px-2">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="h-8 w-8 rounded-full bg-neutral-800 border border-neutral-700 flex items-center justify-center text-xs font-bold text-white shrink-0">
                {user?.username?.substring(0, 2).toUpperCase() || 'U'}
              </div>
              <div className="min-w-0">
                <div className="text-xs font-semibold text-white truncate">{user?.username}</div>
                <div className="text-[11px] text-neutral-500 truncate">{user?.email}</div>
              </div>
            </div>
            <button
              onClick={handleLogout}
              title="Sign Out"
              className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-neutral-800 rounded-md transition cursor-pointer"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-16 border-b border-neutral-800/80 px-8 flex items-center justify-between bg-neutral-900/30 shrink-0">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 text-xs text-emerald-400 font-medium">
              <CheckCircle2 className="h-3.5 w-3.5" />
              <span>Platform Online</span>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span
              className={`rounded-full px-2.5 py-0.5 text-xs font-medium border ${
                user?.role === 'admin'
                  ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                  : 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20'
              }`}
            >
              {user?.role === 'admin' ? 'Administrator' : 'Learner'}
            </span>
          </div>
        </header>

        <main className="flex-1 p-8 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
