import { Flame, BookOpen, Code2, Clock, CheckCircle2, ArrowRight } from 'lucide-react'
import { useAuth } from '../features/auth/AuthContext'

export default function Dashboard() {
  const { user } = useAuth()

  return (
    <div className="max-w-6xl space-y-8">
      {/* Welcome Hero */}
      <div className="rounded-2xl border border-neutral-800 bg-gradient-to-r from-neutral-900/80 via-neutral-900/40 to-neutral-950 p-8 shadow-xl">
        <div className="max-w-2xl space-y-3">
          <div className="inline-flex items-center gap-2 rounded-full bg-indigo-500/10 px-3 py-1 text-xs font-medium text-indigo-400 border border-indigo-500/20">
            <span>Go 1.24+ Curriculum</span>
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">
            Welcome back, {user?.username || 'Learner'}
          </h1>
          <p className="text-neutral-400 text-sm leading-relaxed">
            Your personalized space for structured Go mastery. Understand core memory semantics, concurrency
            primitives, interfaces, and system patterns.
          </p>
        </div>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
        <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 p-5 space-y-3">
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Current Streak</span>
            <Flame className="h-4 w-4 text-amber-500" />
          </div>
          <div className="text-3xl font-bold text-white">0 Days</div>
          <p className="text-xs text-neutral-500">Complete a lesson today to ignite your streak</p>
        </div>

        <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 p-5 space-y-3">
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Lessons Finished</span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="text-3xl font-bold text-white">0 / 48</div>
          <p className="text-xs text-neutral-500">Phase 1 Curriculum Ready</p>
        </div>

        <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 p-5 space-y-3">
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Challenges Solved</span>
            <Code2 className="h-4 w-4 text-indigo-500" />
          </div>
          <div className="text-3xl font-bold text-white">0 Solved</div>
          <p className="text-xs text-neutral-500">Practice tests & Monaco editor</p>
        </div>

        <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 p-5 space-y-3">
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Time Invested</span>
            <Clock className="h-4 w-4 text-neutral-400" />
          </div>
          <div className="text-3xl font-bold text-white">0 hrs</div>
          <p className="text-xs text-neutral-500">Tracked automatically per session</p>
        </div>
      </div>

      {/* Featured Learning Path */}
      <div className="rounded-2xl border border-neutral-800 bg-neutral-900/30 p-6 space-y-6">
        <div className="flex items-center justify-between border-b border-neutral-800/80 pb-4">
          <div>
            <h2 className="text-lg font-bold text-white">Primary Learning Path</h2>
            <p className="text-xs text-neutral-400">Structured sequence from zero values to concurrency</p>
          </div>
          <span className="rounded-md bg-neutral-800 px-2.5 py-1 text-xs font-medium text-neutral-300">
            Phase 2 Content
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          <div className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3">
            <div className="flex items-center gap-2 text-indigo-400">
              <BookOpen className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Module 1</span>
            </div>
            <h3 className="font-semibold text-white text-sm">Language Fundamentals</h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Types, pointers, memory layout, struct composition, and slices internals.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-indigo-400 font-medium">
              <span>Coming in Phase 2</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </div>
          </div>

          <div className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3">
            <div className="flex items-center gap-2 text-emerald-400">
              <Code2 className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Module 2</span>
            </div>
            <h3 className="font-semibold text-white text-sm">Interfaces & Composition</h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Implicit interfaces, empty interface (`any`), type assertions, and clean design.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-neutral-500 font-medium">
              <span>Locked</span>
            </div>
          </div>

          <div className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3">
            <div className="flex items-center gap-2 text-amber-400">
              <Flame className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Module 3</span>
            </div>
            <h3 className="font-semibold text-white text-sm">Concurrency & Channels</h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Goroutines, buffered/unbuffered channels, select multiplexing, sync primitives.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-neutral-500 font-medium">
              <span>Locked</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
