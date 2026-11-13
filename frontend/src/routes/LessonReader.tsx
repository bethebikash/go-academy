import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  Bookmark as BookmarkIcon,
  CheckCircle2,
  Clock,
  FileText,
  Lock,
  Loader2,
  Plus,
  ShieldAlert,
  Sparkles,
  Trash2,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { LessonDetail, Note } from '../types'
import MarkdownViewer from '../components/content/MarkdownViewer'

export default function LessonReader() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [lesson, setLesson] = useState<LessonDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [completing, setCompleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Bookmarking state
  const [isBookmarked, setIsBookmarked] = useState(false)
  const [bookmarking, setBookmarking] = useState(false)

  // Lesson Notes state
  const [notes, setNotes] = useState<Note[]>([])
  const [notesOpen, setNotesOpen] = useState(false)
  const [newNoteTitle, setNewNoteTitle] = useState('')
  const [newNoteContent, setNewNoteContent] = useState('')
  const [creatingNote, setCreatingNote] = useState(false)

  useEffect(() => {
    async function loadLesson() {
      try {
        setLoading(true)
        setError(null)
        const data = await apiRequest<LessonDetail>(`/lessons/${slug}`)
        setLesson(data)

        // Check if lesson is bookmarked
        try {
          const bmRes = await apiRequest<{ bookmarked: boolean }>(`/bookmarks/check?lesson_id=${data.id}`)
          setIsBookmarked(bmRes?.bookmarked || false)
        } catch {
          // Non-blocking
        }

        // Fetch notes for this lesson
        try {
          const notesData = await apiRequest<Note[]>(`/notes?lesson_id=${data.id}`)
          setNotes(notesData || [])
        } catch {
          // Non-blocking
        }
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

  const handleToggleBookmark = async () => {
    if (!lesson || bookmarking) return

    const previousState = isBookmarked
    setIsBookmarked(!previousState) // Optimistic update
    setBookmarking(true)

    try {
      if (previousState) {
        await apiRequest(`/bookmarks/lessons/${lesson.id}`, { method: 'DELETE' })
      } else {
        await apiRequest('/bookmarks', {
          method: 'POST',
          body: JSON.stringify({ lesson_id: lesson.id }),
        })
      }
    } catch (err) {
      // Revert on error
      setIsBookmarked(previousState)
    } finally {
      setBookmarking(false)
    }
  }

  const handleCreateLessonNote = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!lesson || !newNoteTitle.trim()) return

    try {
      setCreatingNote(true)
      const created = await apiRequest<Note>('/notes', {
        method: 'POST',
        body: JSON.stringify({
          lesson_id: lesson.id,
          title: newNoteTitle.trim(),
          content: newNoteContent,
          pinned: false,
          tags: ['lesson-notes'],
        }),
      })
      setNotes((prev) => [created, ...prev])
      setNewNoteTitle('')
      setNewNoteContent('')
    } catch (err: any) {
      setError(err.message || 'Failed to save note')
    } finally {
      setCreatingNote(false)
    }
  }

  const handleDeleteLessonNote = async (noteId: string) => {
    try {
      await apiRequest(`/notes/${noteId}`, { method: 'DELETE' })
      setNotes((prev) => prev.filter((n) => n.id !== noteId))
    } catch (err: any) {
      setError(err.message || 'Failed to delete note')
    }
  }

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
      {/* Navigation Breadcrumb & Quick Actions */}
      <div className="flex flex-wrap items-center justify-between gap-4">
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

        {/* Action icons: Bookmark & Lesson Notes */}
        <div className="flex items-center gap-2">
          <button
            onClick={handleToggleBookmark}
            disabled={bookmarking}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
              isBookmarked
                ? 'bg-indigo-600/20 border-indigo-500/40 text-indigo-300'
                : 'bg-neutral-900 border-neutral-800 text-neutral-400 hover:text-white'
            }`}
            title={isBookmarked ? 'Remove Bookmark' : 'Bookmark this lesson'}
          >
            <BookmarkIcon
              className={`h-3.5 w-3.5 ${isBookmarked ? 'fill-indigo-400 text-indigo-400' : ''}`}
            />
            <span>{isBookmarked ? 'Bookmarked' : 'Bookmark'}</span>
          </button>

          <button
            onClick={() => setNotesOpen(!notesOpen)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
              notesOpen
                ? 'bg-neutral-800 border-neutral-700 text-white'
                : 'bg-neutral-900 border-neutral-800 text-neutral-400 hover:text-white'
            }`}
          >
            <FileText className="h-3.5 w-3.5 text-indigo-400" />
            <span>Notes {notes.length > 0 && `(${notes.length})`}</span>
          </button>
        </div>
      </div>

      {/* Slide-down Lesson Notes Drawer */}
      {notesOpen && (
        <div className="rounded-2xl border border-neutral-800 bg-neutral-900/60 p-6 space-y-5">
          <div className="flex items-center justify-between border-b border-neutral-800 pb-3">
            <h3 className="text-sm font-bold text-white flex items-center gap-2">
              <FileText className="h-4 w-4 text-indigo-400" />
              My Notes for this Lesson ({notes.length})
            </h3>
            <span className="text-xs text-neutral-500">Private annotations</span>
          </div>

          {/* New note form */}
          <form onSubmit={handleCreateLessonNote} className="space-y-3">
            <input
              type="text"
              required
              value={newNoteTitle}
              onChange={(e) => setNewNoteTitle(e.target.value)}
              placeholder="Note title or observation..."
              className="w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
            />
            <textarea
              rows={3}
              value={newNoteContent}
              onChange={(e) => setNewNoteContent(e.target.value)}
              placeholder="Jot down personal notes in Markdown..."
              className="w-full rounded-lg border border-neutral-700 bg-neutral-950 p-3 font-mono text-xs text-white placeholder-neutral-500 leading-relaxed focus:border-indigo-500 focus:outline-none"
            />
            <div className="flex justify-end">
              <button
                type="submit"
                disabled={creatingNote}
                className="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-3.5 py-1.5 text-xs font-semibold text-white shadow transition"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>{creatingNote ? 'Adding...' : 'Add Note'}</span>
              </button>
            </div>
          </form>

          {/* Notes list */}
          {notes.length > 0 && (
            <div className="space-y-2 pt-2 border-t border-neutral-800">
              {notes.map((n) => (
                <div
                  key={n.id}
                  className="flex items-start justify-between gap-3 p-3 rounded-lg border border-neutral-800/80 bg-neutral-950/60 text-xs"
                >
                  <div className="space-y-1">
                    <span className="font-semibold text-white block">{n.title}</span>
                    <p className="text-neutral-300 font-sans leading-relaxed whitespace-pre-wrap">
                      {n.content}
                    </p>
                  </div>
                  <button
                    onClick={() => handleDeleteLessonNote(n.id)}
                    className="p-1 text-neutral-500 hover:text-rose-400 shrink-0"
                    title="Delete note"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

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
