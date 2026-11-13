import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { BookOpen, CheckCircle2, ChevronRight, Compass, Flame, Loader2, Sparkles } from 'lucide-react'
import { apiRequest } from '../lib/api'
import { LearningPath } from '../types'

export default function LearningPaths() {
  const [paths, setPaths] = useState<LearningPath[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadPaths() {
      try {
        setLoading(true)
        const data = await apiRequest<LearningPath[]>('/learning-paths')
        setPaths(data || [])
      } catch (err: any) {
        setError(err.message || 'Failed to load learning paths')
      } finally {
        setLoading(false)
      }
    }
    loadPaths()
  }, [])

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  return (
    <div className="space-y-8 max-w-6xl mx-auto pb-12">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl border border-neutral-800 bg-gradient-to-r from-neutral-900 via-neutral-900 to-indigo-950/40 p-8 shadow-sm">
        <div className="relative z-10 max-w-2xl space-y-3">
          <div className="inline-flex items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/10 px-3 py-1 text-xs font-medium text-indigo-400">
            <Sparkles className="h-3.5 w-3.5" />
            Structured Curriculum
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Go Mastery Learning Paths</h1>
          <p className="text-sm text-neutral-400 leading-relaxed">
            Curated, sequential paths designed to guide you from core Go idioms to advanced systems engineering,
            distributed concurrency, and production-grade architectures.
          </p>
        </div>
        <div className="absolute -right-12 -bottom-12 opacity-10 pointer-events-none">
          <Compass className="h-64 w-64 text-indigo-400" />
        </div>
      </div>

      {error && (
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-4 text-sm text-rose-300">
          {error}
        </div>
      )}

      {/* Path List */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-white">Available Paths</h2>
          <span className="text-xs text-neutral-400">{paths.length} Paths Available</span>
        </div>

        {paths.length === 0 ? (
          <div className="rounded-xl border border-dashed border-neutral-800 p-12 text-center space-y-3">
            <BookOpen className="h-8 w-8 text-neutral-500 mx-auto" />
            <p className="text-neutral-400 text-sm font-medium">No published learning paths yet.</p>
            <p className="text-neutral-500 text-xs">Curriculum is being prepared by our instructors.</p>
          </div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2">
            {paths.map((path) => {
              const progress = path.progress_percent || 0
              const total = path.total_lessons || 0
              const completed = path.completed_lessons || 0

              return (
                <div
                  key={path.id}
                  className="group flex flex-col justify-between rounded-xl border border-neutral-800 bg-neutral-900/40 p-6 transition hover:border-neutral-700 hover:bg-neutral-900/80 shadow-sm"
                >
                  <div className="space-y-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-600/10 border border-indigo-500/20 text-indigo-400 font-bold group-hover:scale-105 transition">
                        <BookOpen className="h-5 w-5" />
                      </div>
                      {progress === 100 ? (
                        <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-400 border border-emerald-500/20">
                          <CheckCircle2 className="h-3.5 w-3.5" /> Completed
                        </span>
                      ) : progress > 0 ? (
                        <span className="inline-flex items-center gap-1.5 rounded-full bg-indigo-500/10 px-2.5 py-0.5 text-xs font-semibold text-indigo-400 border border-indigo-500/20">
                          <Flame className="h-3.5 w-3.5" /> In Progress
                        </span>
                      ) : (
                        <span className="rounded-full bg-neutral-800 px-2.5 py-0.5 text-xs font-medium text-neutral-400">
                          Not Started
                        </span>
                      )}
                    </div>

                    <div>
                      <h3 className="text-lg font-bold text-white group-hover:text-indigo-400 transition">
                        {path.title}
                      </h3>
                      {path.description && (
                        <p className="mt-2 text-sm text-neutral-400 line-clamp-3 leading-relaxed">
                          {path.description}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="mt-6 pt-4 border-t border-neutral-800/80 space-y-4">
                    {/* Progress Bar */}
                    <div>
                      <div className="flex justify-between text-xs text-neutral-400 mb-1.5">
                        <span>Progress</span>
                        <span>{completed} / {total} lessons ({progress}%)</span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-neutral-800 overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300"
                          style={{ width: `${progress}%` }}
                        />
                      </div>
                    </div>

                    <Link
                      to={`/learning-paths/${path.slug}`}
                      className="flex items-center justify-center gap-2 w-full rounded-lg bg-neutral-800 hover:bg-indigo-600 text-neutral-200 hover:text-white px-4 py-2 text-xs font-semibold transition"
                    >
                      <span>Explore Syllabus</span>
                      <ChevronRight className="h-3.5 w-3.5" />
                    </Link>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
