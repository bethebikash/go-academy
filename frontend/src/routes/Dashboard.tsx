import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Flame, BookOpen, Code2, Award, CheckCircle2, ArrowRight } from 'lucide-react'
import { useAuth } from '../features/auth/AuthContext'
import { apiRequest } from '../lib/api'
import { UserSummaryStats } from '../types'

export default function Dashboard() {
  const { user } = useAuth()
  const [stats, setStats] = useState<UserSummaryStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let isMounted = true
    apiRequest<UserSummaryStats>('/analytics/summary')
      .then((data) => {
        if (isMounted) {
          setStats(data)
        }
      })
      .catch((err) => {
        console.error('Failed to load dashboard metrics', err)
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false)
        }
      })
    return () => {
      isMounted = false
    }
  }, [])

  const currentStreak = stats?.streak?.current_streak ?? 0
  const longestStreak = stats?.streak?.longest_streak ?? 0
  const lessonsCompleted = stats?.completed_lessons ?? 0
  const challengesSolved = stats?.solved_challenges ?? 0
  const badgesUnlocked = stats?.earned_badges_count ?? 0

  return (
    <div className="max-w-6xl space-y-8">
      {/* Welcome Hero */}
      <div className="rounded-2xl border border-neutral-800 bg-gradient-to-r from-neutral-900/80 via-neutral-900/40 to-neutral-950 p-8 shadow-xl">
        <div className="max-w-2xl space-y-3">
          <div className="inline-flex items-center gap-2 rounded-full bg-indigo-500/10 px-3 py-1 text-xs font-medium text-indigo-400 border border-indigo-500/20">
            <span>Go 1.24+ Interactive Platform</span>
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">
            Welcome back, {user?.username || 'Learner'}
          </h1>
          <p className="text-neutral-400 text-sm leading-relaxed">
            Your personalized space for structured Go mastery. Understand core memory semantics, concurrency
            primitives, interfaces, and system patterns with live in-browser feedback.
          </p>
        </div>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
        <Link
          to="/progress"
          className="rounded-xl border border-neutral-800 bg-neutral-900/30 hover:bg-neutral-900/60 transition p-5 space-y-3 group"
        >
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Current Streak</span>
            <Flame className={`h-4 w-4 ${currentStreak > 0 ? 'text-amber-500 animate-pulse' : 'text-neutral-500'}`} />
          </div>
          <div className="text-3xl font-bold text-white">
            {loading ? '—' : `${currentStreak} Days`}
          </div>
          <p className="text-xs text-neutral-500 group-hover:text-amber-400 transition">
            {currentStreak > 0
              ? `Personal best: ${longestStreak} days`
              : 'Complete a lesson or challenge today'}
          </p>
        </Link>

        <Link
          to="/learning-paths"
          className="rounded-xl border border-neutral-800 bg-neutral-900/30 hover:bg-neutral-900/60 transition p-5 space-y-3 group"
        >
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Lessons Finished</span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="text-3xl font-bold text-white">
            {loading ? '—' : `${lessonsCompleted} Done`}
          </div>
          <p className="text-xs text-neutral-500 group-hover:text-emerald-400 transition">
            Interactive curriculum
          </p>
        </Link>

        <Link
          to="/challenges"
          className="rounded-xl border border-neutral-800 bg-neutral-900/30 hover:bg-neutral-900/60 transition p-5 space-y-3 group"
        >
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Challenges Solved</span>
            <Code2 className="h-4 w-4 text-indigo-500" />
          </div>
          <div className="text-3xl font-bold text-white">
            {loading ? '—' : `${challengesSolved} Solved`}
          </div>
          <p className="text-xs text-neutral-500 group-hover:text-indigo-400 transition">
            Monaco editor & unit test runner
          </p>
        </Link>

        <Link
          to="/progress"
          className="rounded-xl border border-neutral-800 bg-neutral-900/30 hover:bg-neutral-900/60 transition p-5 space-y-3 group"
        >
          <div className="flex items-center justify-between text-neutral-400">
            <span className="text-xs font-semibold uppercase tracking-wider">Badges Earned</span>
            <Award className="h-4 w-4 text-amber-400" />
          </div>
          <div className="text-3xl font-bold text-white">
            {loading ? '—' : `${badgesUnlocked} Badges`}
          </div>
          <p className="text-xs text-neutral-500 group-hover:text-amber-400 transition">
            View achievements & certificates
          </p>
        </Link>
      </div>

      {/* Featured Learning Path */}
      <div className="rounded-2xl border border-neutral-800 bg-neutral-900/30 p-6 space-y-6">
        <div className="flex items-center justify-between border-b border-neutral-800/80 pb-4">
          <div>
            <h2 className="text-lg font-bold text-white">Curriculum & Tracks</h2>
            <p className="text-xs text-neutral-400">Structured sequence from zero values to production concurrency</p>
          </div>
          <Link
            to="/learning-paths"
            className="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600/10 hover:bg-indigo-600/20 text-indigo-400 border border-indigo-500/20 px-3 py-1.5 text-xs font-medium transition"
          >
            <span>Explore All Paths</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          <Link
            to="/learning-paths"
            className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3 hover:border-neutral-700 transition block group"
          >
            <div className="flex items-center gap-2 text-indigo-400">
              <BookOpen className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Foundations</span>
            </div>
            <h3 className="font-semibold text-white text-sm group-hover:text-indigo-300 transition">
              Language Fundamentals
            </h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Types, pointers, memory layout, struct composition, slices internals, and package design.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-indigo-400 font-medium">
              <span>Start Learning</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </div>
          </Link>

          <Link
            to="/challenges"
            className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3 hover:border-neutral-700 transition block group"
          >
            <div className="flex items-center gap-2 text-emerald-400">
              <Code2 className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Coding Sandbox</span>
            </div>
            <h3 className="font-semibold text-white text-sm group-hover:text-emerald-300 transition">
              Interactive Challenges
            </h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Solve real-world Go challenges with in-memory tests, sub-second execution, and hidden test suites.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-emerald-400 font-medium">
              <span>Practice Challenges</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </div>
          </Link>

          <Link
            to="/progress"
            className="rounded-xl border border-neutral-800/80 bg-neutral-950/60 p-5 space-y-3 hover:border-neutral-700 transition block group"
          >
            <div className="flex items-center gap-2 text-amber-400">
              <Award className="h-4 w-4" />
              <span className="text-xs font-semibold uppercase tracking-wider">Gamification</span>
            </div>
            <h3 className="font-semibold text-white text-sm group-hover:text-amber-300 transition">
              Badges & Certificates
            </h3>
            <p className="text-xs text-neutral-500 leading-relaxed">
              Track your 90-day activity matrix, maintain streaks, unlock badges, and earn verifiable credentials.
            </p>
            <div className="pt-2 flex items-center gap-1 text-xs text-amber-400 font-medium">
              <span>View Progress</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </div>
          </Link>
        </div>
      </div>
    </div>
  )
}
