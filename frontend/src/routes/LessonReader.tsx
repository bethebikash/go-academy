import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Clock,
  Lock,
  Loader2,
  ShieldAlert,
  Sparkles,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { LessonDetail } from '../types'
import MarkdownViewer from '../components/content/MarkdownViewer'

export default function LessonReader() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [lesson, setLesson] = useState<LessonDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [completing, setCompleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadLesson() {
      try {
        setLoading(true)
        setError(null)
        const data = await apiRequest<LessonDetail>(`/lessons/${slug}`)
        setLesson(data)
      } catch (err: any) {
        setError(err.message || 'Failed to load lesson')
      } finally {
        setLoading(false)
      }
    }
    if (slug) {
      loadLesson()
    }
  }, [slug])

  const handleMarkComplete = async () => {
    if (!lesson || lesson.is_completed) return

    try {
      setCompleting(true)
      await apiRequest(`/lessons/${lesson.id}/progress`, {
        method: 'PUT',
        body: JSON.stringify({ status: 'completed' }),
      })
      setLesson((prev) => (prev ? { ...prev, is_completed: true, progress_status: 'completed' } : null))
    } catch (err: any) {
      setError(err.message || 'Failed to update lesson progress')
    } finally {
      setCompleting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  if (error && !lesson) {
    return (
      <div className="space-y-4 max-w-2xl mx-auto py-12 text-center">
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-6 text-sm text-rose-300">
          {error}
        </div>
        <Link
          to="/learning-paths"
          className="inline-flex items-center gap-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300"
        >
          <ArrowLeft className="h-3.5 w-3.5" /> Back to Curriculum
        </Link>
      </div>
    )
  }

  if (!lesson) return null

  return (
    <div className="max-w-4xl mx-auto space-y-8 pb-16">
      {/* Navigation Breadcrumb */}
      <div className="flex flex-wrap items-center gap-2 text-xs text-neutral-400">
        <Link to="/learning-paths" className="hover:text-white transition">
          Curriculum
        </Link>
        <span>/</span>
        <Link to={`/courses/${lesson.course_slug}`} className="hover:text-white transition">
          {lesson.course_title}
        </Link>
        <span>/</span>
        <span className="text-neutral-500">{lesson.module_title}</span>
        <span>/</span>
        <span className="text-neutral-200 font-medium truncate max-w-xs">{lesson.title}</span>
      </div>

      {/* Lesson Header */}
      <div className="space-y-4 border-b border-neutral-800 pb-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <span className="rounded bg-indigo-500/10 border border-indigo-500/20 px-2.5 py-0.5 text-xs font-medium text-indigo-400 capitalize">
              {lesson.difficulty}
            </span>
            <span className="flex items-center gap-1.5 text-xs text-neutral-400">
              <Clock className="h-3.5 w-3.5 text-neutral-500" />
              {lesson.estimated_minutes} min read
            </span>
          </div>

          {lesson.is_completed && (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 px-3 py-1 text-xs font-semibold text-emerald-400">
              <CheckCircle2 className="h-4 w-4" /> Lesson Completed
            </span>
          )}
        </div>

        <h1 className="text-3xl font-extrabold tracking-tight text-white">{lesson.title}</h1>

        {lesson.summary && (
          <p className="text-sm text-neutral-300 leading-relaxed italic border-l-2 border-neutral-700 pl-3">
            {lesson.summary}
          </p>
        )}
      </div>

      {/* Locked State Warning */}
      {lesson.is_locked && (
        <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-5 space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold text-amber-400">
            <Lock className="h-4 w-4" />
            <span>This lesson has prerequisite requirements</span>
          </div>
          <p className="text-xs text-neutral-300 leading-relaxed">
            To gain the most from this topic, please complete the following foundational lessons first:
          </p>
          <ul className="space-y-1.5 pt-1">
            {lesson.missing_prerequisites.map((prereq) => (
              <li key={prereq.id}>
                <Link
                  to={`/lessons/${prereq.slug}`}
                  className="inline-flex items-center gap-1.5 text-xs font-medium text-amber-300 hover:text-amber-200 underline underline-offset-4"
                >
                  <ShieldAlert className="h-3.5 w-3.5" />
                  {prereq.title}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Lesson Content */}
      <div className="bg-neutral-900/30 border border-neutral-800/80 rounded-2xl p-6 md:p-8">
        <MarkdownViewer content={lesson.content} />
      </div>

      {/* Action and Navigation Footer */}
      <div className="rounded-2xl border border-neutral-800 bg-neutral-900/40 p-6 flex flex-col sm:flex-row items-center justify-between gap-4">
        {/* Previous Lesson */}
        <div>
          {lesson.prev_lesson_slug ? (
            <Link
              to={`/lessons/${lesson.prev_lesson_slug}`}
              className="inline-flex items-center gap-2 rounded-lg border border-neutral-800 bg-neutral-900/80 hover:bg-neutral-800 px-4 py-2 text-xs font-semibold text-neutral-300 hover:text-white transition"
            >
              <ArrowLeft className="h-3.5 w-3.5" /> Previous Lesson
            </Link>
          ) : (
            <Link
              to={`/courses/${lesson.course_slug}`}
              className="inline-flex items-center gap-2 text-xs font-medium text-neutral-500 hover:text-neutral-400 transition"
            >
              <ArrowLeft className="h-3.5 w-3.5" /> Course Syllabus
            </Link>
          )}
        </div>

        {/* Complete Lesson Action */}
        <div>
          {!lesson.is_completed ? (
            <button
              onClick={handleMarkComplete}
              disabled={completing || lesson.is_locked}
              className={`inline-flex items-center gap-2 rounded-lg px-5 py-2.5 text-xs font-semibold text-white shadow-md transition ${
                lesson.is_locked
                  ? 'bg-neutral-800 text-neutral-500 cursor-not-allowed'
                  : 'bg-emerald-600 hover:bg-emerald-500 shadow-emerald-600/20'
              }`}
            >
              {completing ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <CheckCircle2 className="h-4 w-4" />
              )}
              <span>{completing ? 'Marking Complete...' : 'Mark as Complete'}</span>
            </button>
          ) : (
            <div className="flex items-center gap-2 text-xs font-semibold text-emerald-400">
              <CheckCircle2 className="h-4 w-4" />
              <span>Completed</span>
            </div>
          )}
        </div>

        {/* Next Lesson */}
        <div>
          {lesson.next_lesson_slug ? (
            <button
              onClick={() => navigate(`/lessons/${lesson.next_lesson_slug}`)}
              className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow-md shadow-indigo-600/20 transition"
            >
              <span>Next Lesson</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          ) : (
            <Link
              to={`/courses/${lesson.course_slug}`}
              className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white transition"
            >
              <Sparkles className="h-3.5 w-3.5" />
              <span>Finish Course</span>
            </Link>
          )}
        </div>
      </div>
    </div>
  )
}
