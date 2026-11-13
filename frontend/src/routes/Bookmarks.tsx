import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  Bookmark as BookmarkIcon,
  CheckCircle2,
  Clock,
  Layers,
  Loader2,
  Trash2,
  ChevronRight,
  ArrowRight,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { Bookmark } from '../types'

export default function Bookmarks() {
  const [bookmarks, setBookmarks] = useState<Bookmark[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadBookmarks = async () => {
    try {
      setLoading(true)
      const data = await apiRequest<Bookmark[]>('/bookmarks')
      setBookmarks(data || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load bookmarks')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadBookmarks()
  }, [])

  const handleRemoveBookmark = async (bookmarkId: string) => {
    try {
      await apiRequest(`/bookmarks/${bookmarkId}`, { method: 'DELETE' })
      setBookmarks((prev) => prev.filter((b) => b.id !== bookmarkId))
    } catch (err: any) {
      setError(err.message || 'Failed to remove bookmark')
    }
  }

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  return (
    <div className="space-y-8 max-w-5xl mx-auto pb-16">
      {/* Top Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2.5">
          <BookmarkIcon className="h-6 w-6 text-indigo-400 fill-indigo-400" />
          Bookmarked Lessons
        </h1>
        <p className="text-xs text-neutral-400 mt-1">
          Quickly access your saved lessons, key reference concepts, and tricky topics.
        </p>
      </div>

      {error && (
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-4 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Bookmarks List */}
      {bookmarks.length === 0 ? (
        <div className="rounded-xl border border-dashed border-neutral-800 p-12 text-center space-y-4">
          <BookmarkIcon className="h-8 w-8 text-neutral-500 mx-auto" />
          <div>
            <p className="text-neutral-300 text-sm font-semibold">No bookmarked lessons yet.</p>
            <p className="text-neutral-500 text-xs mt-1">
              While reading any lesson, click the bookmark icon to save it here for fast revision.
            </p>
          </div>
          <div>
            <Link
              to="/learning-paths"
              className="inline-flex items-center gap-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 px-4 py-2 text-xs font-semibold text-white transition"
            >
              <span>Explore Curriculum</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {bookmarks.map((bm) => {
            const isCompleted = bm.progress_status === 'completed'
            const isStarted = bm.progress_status === 'started'

            return (
              <div
                key={bm.id}
                className="group flex flex-col md:flex-row md:items-center justify-between gap-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-5 hover:border-neutral-700 hover:bg-neutral-900/70 transition"
              >
                <div className="space-y-1.5 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 text-xs text-neutral-500">
                    <span className="flex items-center gap-1 text-indigo-400 font-medium">
                      <Layers className="h-3.5 w-3.5" /> {bm.course_title}
                    </span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" /> {bm.lesson_estimated_minutes} min
                    </span>
                    <span>•</span>
                    <span className="rounded bg-neutral-800 px-1.5 py-0.5 text-[10px] uppercase font-semibold text-neutral-400">
                      {bm.lesson_difficulty}
                    </span>
                  </div>

                  <Link
                    to={`/lessons/${bm.lesson_slug}`}
                    className="block text-base font-bold text-white hover:text-indigo-400 transition truncate"
                  >
                    {bm.lesson_title}
                  </Link>

                  {bm.lesson_summary && (
                    <p className="text-xs text-neutral-400 line-clamp-1 max-w-2xl">
                      {bm.lesson_summary}
                    </p>
                  )}
                </div>

                <div className="flex items-center gap-4 self-end md:self-auto shrink-0">
                  {isCompleted ? (
                    <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-semibold text-emerald-400 border border-emerald-500/20">
                      <CheckCircle2 className="h-3.5 w-3.5" /> Completed
                    </span>
                  ) : isStarted ? (
                    <span className="rounded-full bg-indigo-500/10 px-2.5 py-0.5 text-xs font-medium text-indigo-400 border border-indigo-500/20">
                      In Progress
                    </span>
                  ) : (
                    <span className="rounded-full bg-neutral-800 px-2 py-0.5 text-xs text-neutral-400">
                      Not Started
                    </span>
                  )}

                  <Link
                    to={`/lessons/${bm.lesson_slug}`}
                    className="inline-flex items-center gap-1 rounded-lg bg-neutral-800 hover:bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-neutral-200 hover:text-white transition"
                  >
                    <span>Read</span>
                    <ChevronRight className="h-3.5 w-3.5" />
                  </Link>

                  <button
                    onClick={() => handleRemoveBookmark(bm.id)}
                    className="p-1.5 rounded text-neutral-500 hover:text-rose-400 transition"
                    title="Remove Bookmark"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
