import { useState, useEffect } from 'react'
import {
  Terminal,
  BookOpen,
  Code2,
  FileText,
  Bookmark,
  Award,
  Shield,
  Activity,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react'

interface HealthState {
  status: string
  loading: boolean
  error?: string
}

export default function App() {
  const [health, setHealth] = useState<HealthState>({
    status: 'checking...',
    loading: true,
  })

  useEffect(() => {
    fetch('/health/live')
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then((data) => {
        setHealth({ status: data.data?.status || 'ok', loading: false })
      })
      .catch((err) => {
        setHealth({ status: 'offline', loading: false, error: err.message })
      })
  }, [])

  return (
    <div className="flex min-h-screen bg-neutral-950 text-neutral-100">
      {/* Sidebar */}
      <aside className="w-64 border-r border-neutral-800 bg-neutral-900/50 flex flex-col justify-between p-4">
        <div>
          <div className="flex items-center gap-3 px-3 py-4 mb-4 border-b border-neutral-800">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-600 text-white font-bold">
              <Terminal className="h-5 w-5" />
            </div>
            <div>
              <span className="font-semibold tracking-tight text-white block">Go Academy</span>
              <span className="text-xs text-neutral-400">Interactive Learning</span>
            </div>
          </div>

          <nav className="space-y-6">
            <div>
              <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 px-3">
                Learn
              </span>
              <div className="mt-2 space-y-1">
                <a
                  href="#dashboard"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium bg-neutral-800 text-white"
                >
                  <Activity className="h-4 w-4 text-indigo-400" />
                  Dashboard
                </a>
                <a
                  href="#learning-paths"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <BookOpen className="h-4 w-4" />
                  Learning Paths
                </a>
                <a
                  href="#challenges"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <Code2 className="h-4 w-4" />
                  Challenges
                </a>
                <a
                  href="#notes"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <FileText className="h-4 w-4" />
                  My Notes
                </a>
                <a
                  href="#bookmarks"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <Bookmark className="h-4 w-4" />
                  Bookmarks
                </a>
                <a
                  href="#progress"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <Award className="h-4 w-4" />
                  My Progress
                </a>
              </div>
            </div>

            <div>
              <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 px-3">
                Admin
              </span>
              <div className="mt-2 space-y-1">
                <a
                  href="#admin-overview"
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-neutral-400 hover:bg-neutral-800/60 hover:text-white transition"
                >
                  <Shield className="h-4 w-4 text-amber-400" />
                  Overview
                </a>
              </div>
            </div>
          </nav>
        </div>

        <div className="pt-4 border-t border-neutral-800">
          <div className="flex items-center justify-between text-xs text-neutral-400 px-2">
            <span>API Status</span>
            <div className="flex items-center gap-1.5 font-medium">
              {health.loading ? (
                <span className="text-neutral-400">checking...</span>
              ) : health.status === 'ok' ? (
                <span className="text-emerald-400 flex items-center gap-1">
                  <CheckCircle2 className="h-3.5 w-3.5" /> Online
                </span>
              ) : (
                <span className="text-amber-400 flex items-center gap-1">
                  <AlertCircle className="h-3.5 w-3.5" /> Standby
                </span>
              )}
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col">
        {/* Header */}
        <header className="h-16 border-b border-neutral-800 px-8 flex items-center justify-between bg-neutral-900/30">
          <div className="flex items-center gap-3">
            <h1 className="text-lg font-semibold text-white">Platform Overview</h1>
            <span className="rounded-full bg-indigo-500/10 px-2.5 py-0.5 text-xs font-medium text-indigo-400 border border-indigo-500/20">
              Phase 1 Baseline
            </span>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <div className="text-sm font-medium text-neutral-200">Welcome Learner</div>
              <div className="text-xs text-neutral-500">Go 1.24+ Ready</div>
            </div>
          </div>
        </header>

        {/* Dashboard Shell Content */}
        <div className="flex-1 p-8 space-y-6 max-w-6xl">
          <div className="rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
            <h2 className="text-xl font-bold text-white mb-2">Interactive Go Mastery</h2>
            <p className="text-neutral-400 text-sm max-w-2xl leading-relaxed">
              Understand the core mechanics of the Go language through structured paths, hands-on
              practice challenges, and personal Markdown notes.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            <div className="rounded-lg border border-neutral-800 bg-neutral-900/20 p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-neutral-300">Curriculum</span>
                <BookOpen className="h-4 w-4 text-indigo-400" />
              </div>
              <div className="text-2xl font-bold text-white">16 Modules</div>
              <p className="text-xs text-neutral-500 mt-1">From Fundamentals to Production Go</p>
            </div>

            <div className="rounded-lg border border-neutral-800 bg-neutral-900/20 p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-neutral-300">Challenges</span>
                <Code2 className="h-4 w-4 text-emerald-400" />
              </div>
              <div className="text-2xl font-bold text-white">Monaco Editor</div>
              <p className="text-xs text-neutral-500 mt-1">Isolated Docker Runner (Phase 5)</p>
            </div>

            <div className="rounded-lg border border-neutral-800 bg-neutral-900/20 p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-neutral-300">Personal Space</span>
                <FileText className="h-4 w-4 text-amber-400" />
              </div>
              <div className="text-2xl font-bold text-white">Markdown Notes</div>
              <p className="text-xs text-neutral-500 mt-1">Autosaved with Optimistic Locking</p>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
