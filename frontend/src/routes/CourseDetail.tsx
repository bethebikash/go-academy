import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  ArrowLeft,
  BookOpen,
  CheckCircle2,
  Clock,
  Layers,
  Lock,
  PlayCircle,
  Loader2,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { CourseDetail as CourseDetailType } from '../types'

export default function CourseDetail() {
  const { slug } = useParams<{ slug: string }>()
  const [course, setCourse] = useState<CourseDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedModules, setExpandedModules] = useState<Record<string, boolean>>({})

  useEffect(() => {
    async function loadCourse() {
      try {
        setLoading(true)
        const data = await apiRequest<CourseDetailType>(`/courses/${slug}`)
        setCourse(data)
        // Expand all modules by default
        if (data?.modules) {
          const initialExpanded: Record<string, boolean> = {}
          data.modules.forEach((m) => {
            initialExpanded[m.id] = true
          })
          setExpandedModules(initialExpanded)
        }
      } catch (err: any) {
        setError(err.message || 'Failed to load course details')
      } finally {
        setLoading(false)
      }
    }
    if (slug) {
      loadCourse()
    }
  }, [slug])

  const toggleModule = (moduleId: string) => {
    setExpandedModules((prev) => ({
      ...prev,
      [moduleId]: !prev[moduleId],
    }))
  }

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  if (error || !course) {
    return (
      <div className="space-y-4 max-w-2xl mx-auto py-12 text-center">
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-6 text-sm text-rose-300">
          {error || 'Course not found.'}
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

  const progress = course.progress_percent || 0

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

      {/* Course Banner */}
      <div className="rounded-2xl border border-neutral-800 bg-neutral-900/40 p-8 space-y-6">
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-xs font-medium text-indigo-400">
            <Layers className="h-4 w-4" /> Course Syllabus
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-white">{course.title}</h1>
          {course.description && (
            <p className="text-sm text-neutral-300 leading-relaxed max-w-3xl">{course.description}</p>
          )}
        </div>

        {/* Progress & Meta */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 pt-6 border-t border-neutral-800/80">
          <div className="w-full sm:max-w-md space-y-2">
            <div className="flex justify-between text-xs text-neutral-400">
              <span>Syllabus Progress</span>
              <span>{course.completed_lessons} / {course.total_lessons} lessons ({progress}%)</span>
            </div>
            <div className="h-2 w-full rounded-full bg-neutral-800 overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>

          <div className="flex items-center gap-4 text-xs text-neutral-400">
            <span className="flex items-center gap-1.5">
              <BookOpen className="h-4 w-4 text-neutral-500" />
              {course.modules?.length || 0} Modules
            </span>
            <span>•</span>
            <span className="flex items-center gap-1.5">
              <Clock className="h-4 w-4 text-neutral-500" />
              {course.total_lessons || 0} Lessons
            </span>
          </div>
        </div>
      </div>

      {/* Syllabus Modules */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-white">Course Syllabus</h2>

        {course.modules?.length === 0 ? (
          <div className="rounded-xl border border-dashed border-neutral-800 p-8 text-center text-neutral-400 text-sm">
            No modules have been published for this course yet.
          </div>
        ) : (
          <div className="space-y-4">
            {course.modules.map((mod, modIdx) => {
              const isExpanded = expandedModules[mod.id] ?? true
              const totalModLessons = mod.lessons?.length || 0
              const completedModLessons = mod.lessons?.filter((l) => l.is_completed).length || 0

              return (
                <div
                  key={mod.id}
                  className="rounded-xl border border-neutral-800 bg-neutral-900/30 overflow-hidden transition"
                >
                  {/* Module Header / Toggle */}
                  <button
                    onClick={() => toggleModule(mod.id)}
                    className="w-full flex items-center justify-between p-5 bg-neutral-900/50 hover:bg-neutral-800/50 transition text-left"
                  >
                    <div className="flex items-center gap-3">
                      <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-neutral-800 text-xs font-bold text-neutral-400">
                        {modIdx + 1}
                      </span>
                      <div>
                        <h3 className="text-base font-semibold text-white">{mod.title}</h3>
                        {mod.description && (
                          <p className="text-xs text-neutral-400 mt-0.5">{mod.description}</p>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-4 shrink-0">
                      <span className="text-xs text-neutral-400">
                        {completedModLessons} / {totalModLessons}
                      </span>
                      {isExpanded ? (
                        <ChevronDown className="h-5 w-5 text-neutral-400" />
                      ) : (
                        <ChevronRight className="h-5 w-5 text-neutral-400" />
                      )}
                    </div>
                  </button>

                  {/* Lessons List in Module */}
                  {isExpanded && (
                    <div className="divide-y divide-neutral-800/60 border-t border-neutral-800/80 bg-neutral-950/40">
                      {mod.lessons?.length === 0 ? (
                        <div className="p-4 text-xs text-neutral-500 italic">No lessons in this module.</div>
                      ) : (
                        mod.lessons.map((lesson, lessonIdx) => {
                          const isLocked = lesson.is_locked
                          const isCompleted = lesson.is_completed

                          return (
                            <div
                              key={lesson.id}
                              className={`flex items-center justify-between p-4 px-6 transition ${
                                isLocked
                                  ? 'opacity-60 bg-neutral-950/60 cursor-not-allowed'
                                  : 'hover:bg-neutral-900/40'
                              }`}
                            >
                              <div className="flex items-center gap-3.5">
                                {/* State icon */}
                                {isCompleted ? (
                                  <CheckCircle2 className="h-5 w-5 text-emerald-400 shrink-0" />
                                ) : isLocked ? (
                                  <Lock className="h-4 w-4 text-neutral-500 shrink-0" />
                                ) : (
                                  <PlayCircle className="h-5 w-5 text-indigo-400 shrink-0" />
                                )}

                                <div>
                                  <div className="flex items-center gap-2">
                                    <span className="text-xs text-neutral-500 font-mono">
                                      {modIdx + 1}.{lessonIdx + 1}
                                    </span>
                                    {isLocked ? (
                                      <span className="text-sm font-medium text-neutral-400">
                                        {lesson.title}
                                      </span>
                                    ) : (
                                      <Link
                                        to={`/lessons/${lesson.slug}`}
                                        className="text-sm font-medium text-white hover:text-indigo-400 transition"
                                      >
                                        {lesson.title}
                                      </Link>
                                    )}
                                  </div>
                                  {lesson.summary && (
                                    <p className="text-xs text-neutral-400 mt-0.5 line-clamp-1 max-w-lg">
                                      {lesson.summary}
                                    </p>
                                  )}
                                </div>
                              </div>

                              <div className="flex items-center gap-3 shrink-0">
                                <span className="inline-flex items-center gap-1 text-[11px] text-neutral-400">
                                  <Clock className="h-3 w-3" />
                                  {lesson.estimated_minutes} min
                                </span>
                                <span className="rounded bg-neutral-800/80 px-2 py-0.5 text-[11px] font-medium text-neutral-400 uppercase tracking-wider">
                                  {lesson.difficulty}
                                </span>
                                {!isLocked && (
                                  <Link
                                    to={`/lessons/${lesson.slug}`}
                                    className="rounded-lg bg-neutral-800 hover:bg-indigo-600 px-3 py-1 text-xs font-semibold text-neutral-200 hover:text-white transition"
                                  >
                                    {isCompleted ? 'Review' : 'Start'}
                                  </Link>
                                )}
                              </div>
                            </div>
                          )
                        })
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
