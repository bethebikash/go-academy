import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, BookOpen, CheckCircle2, ChevronRight, Layers, Loader2 } from 'lucide-react'
import { apiRequest } from '../lib/api'
import { Course, LearningPath } from '../types'

interface PathDetailResponse {
  path: LearningPath
  courses: Course[]
}

export default function PathDetail() {
  const { slug } = useParams<{ slug: string }>()
  const [data, setData] = useState<PathDetailResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadPath() {
      try {
        setLoading(true)
        const res = await apiRequest<PathDetailResponse>(`/learning-paths/${slug}`)
        setData(res)
      } catch (err: any) {
        setError(err.message || 'Failed to load learning path')
      } finally {
        setLoading(false)
      }
    }
    if (slug) {
      loadPath()
    }
  }, [slug])

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="space-y-4 max-w-2xl mx-auto py-12 text-center">
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-6 text-sm text-rose-300">
          {error || 'Learning path not found.'}
        </div>
        <Link
          to="/learning-paths"
          className="inline-flex items-center gap-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300"
        >
          <ArrowLeft className="h-3.5 w-3.5" /> Back to Learning Paths
        </Link>
      </div>
    )
  }

  const { path, courses } = data
  const progress = path.progress_percent || 0

  return (
    <div className="space-y-8 max-w-5xl mx-auto pb-12">
      {/* Back button */}
      <div>
        <Link
          to="/learning-paths"
          className="inline-flex items-center gap-2 text-xs font-medium text-neutral-400 hover:text-white transition"
        >
          <ArrowLeft className="h-3.5 w-3.5" /> All Learning Paths
        </Link>
      </div>

      {/* Path Header */}
      <div className="rounded-2xl border border-neutral-800 bg-neutral-900/40 p-8 space-y-6">
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-xs font-medium text-indigo-400">
            <BookOpen className="h-4 w-4" /> Learning Path
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-white">{path.title}</h1>
          {path.description && (
            <p className="text-sm text-neutral-300 leading-relaxed max-w-3xl">{path.description}</p>
          )}
        </div>

        {/* Progress summary */}
        <div className="pt-4 border-t border-neutral-800/80 max-w-md">
          <div className="flex justify-between text-xs text-neutral-400 mb-2">
            <span>Overall Path Progress</span>
            <span>{path.completed_lessons || 0} of {path.total_lessons || 0} lessons ({progress}%)</span>
          </div>
          <div className="h-2.5 w-full rounded-full bg-neutral-800 overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      </div>

      {/* Courses in this Path */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-white flex items-center gap-2">
          <Layers className="h-5 w-5 text-indigo-400" />
          Courses in this Path ({courses.length})
        </h2>

        {courses.length === 0 ? (
          <div className="rounded-xl border border-dashed border-neutral-800 p-8 text-center text-neutral-400 text-sm">
            No courses have been published in this path yet.
          </div>
        ) : (
          <div className="space-y-3">
            {courses.map((course, idx) => {
              const courseProgress = course.progress_percent || 0
              const isCompleted = courseProgress === 100

              return (
                <Link
                  key={course.id}
                  to={`/courses/${course.slug}`}
                  className="group flex flex-col md:flex-row md:items-center justify-between gap-4 rounded-xl border border-neutral-800 bg-neutral-900/30 p-5 hover:border-neutral-700 hover:bg-neutral-900/70 transition"
                >
                  <div className="flex items-start gap-4">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-neutral-800 text-xs font-bold text-neutral-300 group-hover:bg-indigo-600 group-hover:text-white transition">
                      {idx + 1}
                    </div>
                    <div>
                      <h3 className="text-base font-semibold text-white group-hover:text-indigo-400 transition">
                        {course.title}
                      </h3>
                      {course.description && (
                        <p className="mt-1 text-xs text-neutral-400 line-clamp-2 max-w-xl">
                          {course.description}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-6 self-end md:self-auto shrink-0">
                    <div className="text-right">
                      <div className="text-xs font-medium text-neutral-300">
                        {course.completed_lessons || 0} / {course.total_lessons || 0} lessons
                      </div>
                      <div className="text-[11px] text-neutral-500">{courseProgress}% completed</div>
                    </div>
                    {isCompleted ? (
                      <CheckCircle2 className="h-5 w-5 text-emerald-400" />
                    ) : (
                      <ChevronRight className="h-5 w-5 text-neutral-500 group-hover:text-indigo-400 transition" />
                    )}
                  </div>
                </Link>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
