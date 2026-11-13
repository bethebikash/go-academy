import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  BookOpen,
  Edit,
  FileText,
  Loader2,
  Pin,
  Plus,
  Search,
  Trash2,
  X,
  ExternalLink,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { Note, Tag } from '../types'
import MarkdownViewer from '../components/content/MarkdownViewer'

export default function Notes() {
  const [notes, setNotes] = useState<Note[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [selectedTag, setSelectedTag] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Note Modal state (for create or edit)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingNote, setEditingNote] = useState<Note | null>(null)
  const [formTitle, setFormTitle] = useState('')
  const [formContent, setFormContent] = useState('')
  const [formTags, setFormTags] = useState('')
  const [formPinned, setFormPinned] = useState(false)
  const [previewMode, setPreviewMode] = useState(false)
  const [saving, setSaving] = useState(false)
  const [modalError, setModalError] = useState<string | null>(null)

  const loadTags = async () => {
    try {
      const data = await apiRequest<Tag[]>('/tags')
      setTags(data || [])
    } catch (err) {
      console.error('Failed to load tags', err)
    }
  }

  const loadNotes = async () => {
    try {
      setLoading(true)
      const params = new URLSearchParams()
      if (selectedTag) params.set('tag', selectedTag)
      if (searchQuery.trim()) params.set('q', searchQuery.trim())

      const endpoint = `/notes${params.toString() ? '?' + params.toString() : ''}`
      const data = await apiRequest<Note[]>(endpoint)
      setNotes(data || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load notes')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTags()
  }, [])

  useEffect(() => {
    loadNotes()
  }, [selectedTag, searchQuery])

  const openCreateModal = () => {
    setEditingNote(null)
    setFormTitle('')
    setFormContent('')
    setFormTags('')
    setFormPinned(false)
    setPreviewMode(false)
    setModalError(null)
    setModalOpen(true)
  }

  const openEditModal = (note: Note) => {
    setEditingNote(note)
    setFormTitle(note.title)
    setFormContent(note.content)
    setFormTags(note.tags?.join(', ') || '')
    setFormPinned(note.pinned)
    setPreviewMode(false)
    setModalError(null)
    setModalOpen(true)
  }

  const handleSaveNote = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formTitle.trim()) {
      setModalError('Title is required')
      return
    }

    const tagsArray = formTags
      .split(',')
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean)

    try {
      setSaving(true)
      setModalError(null)

      if (editingNote) {
        // Optimistic locking update
        await apiRequest(`/notes/${editingNote.id}`, {
          method: 'PATCH',
          body: JSON.stringify({
            title: formTitle.trim(),
            content: formContent,
            pinned: formPinned,
            version: editingNote.version,
            tags: tagsArray,
          }),
        })
      } else {
        // Create new note
        await apiRequest('/notes', {
          method: 'POST',
          body: JSON.stringify({
            title: formTitle.trim(),
            content: formContent,
            pinned: formPinned,
            tags: tagsArray,
          }),
        })
      }

      setModalOpen(false)
      await loadNotes()
      await loadTags()
    } catch (err: any) {
      if (err.status === 409) {
        setModalError('Note was updated elsewhere. Please close, refresh, and try again.')
      } else {
        setModalError(err.message || 'Failed to save note')
      }
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteNote = async (noteId: string) => {
    if (!confirm('Are you sure you want to delete this note?')) return

    try {
      await apiRequest(`/notes/${noteId}`, { method: 'DELETE' })
      setNotes((prev) => prev.filter((n) => n.id !== noteId))
      await loadTags()
    } catch (err: any) {
      setError(err.message || 'Failed to delete note')
    }
  }

  return (
    <div className="space-y-8 max-w-6xl mx-auto pb-16">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2.5">
            <FileText className="h-6 w-6 text-indigo-400" />
            Personal Notes & Annotations
          </h1>
          <p className="text-xs text-neutral-400 mt-1">
            Capture Go snippets, mental models, and learnings with Markdown formatting and fast tag filtering.
          </p>
        </div>

        <button
          onClick={openCreateModal}
          className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow shadow-indigo-600/30 transition self-start sm:self-auto"
        >
          <Plus className="h-4 w-4" /> New Note
        </button>
      </div>

      {/* Search & Tag Filter Bar */}
      <div className="space-y-3">
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-neutral-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search note titles or content..."
            className="w-full rounded-xl border border-neutral-800 bg-neutral-900/60 pl-9 pr-8 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-2.5 text-neutral-500 hover:text-white"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Tag chips */}
        {tags.length > 0 && (
          <div className="flex flex-wrap items-center gap-1.5 pt-1">
            <span className="text-[11px] font-semibold text-neutral-500 uppercase tracking-wider mr-1">
              Tags:
            </span>
            <button
              onClick={() => setSelectedTag('')}
              className={`rounded-full px-2.5 py-1 text-xs font-medium transition ${
                selectedTag === ''
                  ? 'bg-indigo-600 text-white font-semibold'
                  : 'bg-neutral-800 text-neutral-400 hover:text-white'
              }`}
            >
              All
            </button>
            {tags.map((tag) => (
              <button
                key={tag.id}
                onClick={() => setSelectedTag(selectedTag === tag.name ? '' : tag.name)}
                className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium transition ${
                  selectedTag === tag.name
                    ? 'bg-indigo-600 text-white font-semibold'
                    : 'bg-neutral-800 text-neutral-400 hover:text-white'
                }`}
              >
                <span>#{tag.name}</span>
                <span className="text-[10px] opacity-75">({tag.note_count})</span>
              </button>
            ))}
          </div>
        )}
      </div>

      {error && (
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-4 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Notes Grid */}
      {loading ? (
        <div className="flex h-64 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
        </div>
      ) : notes.length === 0 ? (
        <div className="rounded-xl border border-dashed border-neutral-800 p-12 text-center space-y-3">
          <FileText className="h-8 w-8 text-neutral-500 mx-auto" />
          <p className="text-neutral-400 text-sm font-medium">No notes found.</p>
          <p className="text-neutral-500 text-xs">
            {selectedTag || searchQuery
              ? 'Try clearing the search or tag filters.'
              : 'Click "New Note" to record your first insight!'}
          </p>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {notes.map((note) => (
            <div
              key={note.id}
              className={`group flex flex-col justify-between rounded-xl border bg-neutral-900/40 p-5 transition hover:border-neutral-700 hover:bg-neutral-900/70 shadow-sm ${
                note.pinned ? 'border-amber-500/40 bg-amber-950/10' : 'border-neutral-800'
              }`}
            >
              <div className="space-y-3">
                {/* Note Header */}
                <div className="flex items-start justify-between gap-2">
                  <h3 className="font-bold text-white text-base leading-snug group-hover:text-indigo-400 transition">
                    {note.title}
                  </h3>
                  {note.pinned && (
                    <Pin className="h-4 w-4 text-amber-400 fill-amber-400 shrink-0" />
                  )}
                </div>

                {/* Linked Lesson context if any */}
                {note.lesson_slug && (
                  <Link
                    to={`/lessons/${note.lesson_slug}`}
                    className="inline-flex items-center gap-1.5 text-[11px] text-indigo-400 hover:text-indigo-300 truncate max-w-full"
                  >
                    <BookOpen className="h-3 w-3 shrink-0" />
                    <span className="truncate">{note.lesson_title || 'Lesson Link'}</span>
                    <ExternalLink className="h-2.5 w-2.5 shrink-0" />
                  </Link>
                )}

                {/* Note Content preview */}
                <div className="text-xs text-neutral-300 line-clamp-4 leading-relaxed font-sans">
                  {note.content}
                </div>
              </div>

              {/* Note Footer */}
              <div className="mt-5 pt-3 border-t border-neutral-800/80 space-y-3">
                {/* Tags */}
                {note.tags && note.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {note.tags.map((t) => (
                      <span
                        key={t}
                        className="rounded bg-neutral-800 px-1.5 py-0.5 text-[10px] font-medium text-neutral-400"
                      >
                        #{t}
                      </span>
                    ))}
                  </div>
                )}

                <div className="flex items-center justify-between text-[11px] text-neutral-500">
                  <span>{new Date(note.updated_at).toLocaleDateString()}</span>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => openEditModal(note)}
                      className="p-1.5 rounded hover:bg-neutral-800 text-neutral-400 hover:text-white transition"
                      title="Edit note"
                    >
                      <Edit className="h-3.5 w-3.5" />
                    </button>
                    <button
                      onClick={() => handleDeleteNote(note.id)}
                      className="p-1.5 rounded hover:bg-neutral-800 text-neutral-400 hover:text-rose-400 transition"
                      title="Delete note"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create / Edit Note Modal */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-2xl rounded-2xl border border-neutral-800 bg-neutral-900 p-6 space-y-5 shadow-2xl">
            <div className="flex items-center justify-between border-b border-neutral-800 pb-3">
              <h2 className="text-lg font-bold text-white flex items-center gap-2">
                <FileText className="h-5 w-5 text-indigo-400" />
                {editingNote ? 'Edit Note' : 'Create New Note'}
              </h2>
              <button
                onClick={() => setModalOpen(false)}
                className="text-neutral-500 hover:text-white"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {modalError && (
              <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-3 text-xs text-rose-300">
                {modalError}
              </div>
            )}

            <form onSubmit={handleSaveNote} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-neutral-300">Title</label>
                <input
                  type="text"
                  required
                  value={formTitle}
                  onChange={(e) => setFormTitle(e.target.value)}
                  placeholder="e.g. Channel Synchronization Idiom"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-neutral-300">
                  Tags (comma separated)
                </label>
                <input
                  type="text"
                  value={formTags}
                  onChange={(e) => setFormTags(e.target.value)}
                  placeholder="e.g. concurrency, goroutines, channels"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>

              {/* Content editor & preview toggle */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-medium text-neutral-300">
                    Content (Markdown)
                  </label>
                  <button
                    type="button"
                    onClick={() => setPreviewMode(!previewMode)}
                    className="text-xs text-indigo-400 hover:text-indigo-300 font-medium"
                  >
                    {previewMode ? 'Switch to Editor' : 'Live Preview'}
                  </button>
                </div>

                {previewMode ? (
                  <div className="h-60 overflow-y-auto rounded-lg border border-neutral-750 bg-neutral-950 p-4 text-xs">
                    <MarkdownViewer content={formContent || '*No content yet...*'} />
                  </div>
                ) : (
                  <textarea
                    rows={10}
                    value={formContent}
                    onChange={(e) => setFormContent(e.target.value)}
                    placeholder="Write your Markdown notes here..."
                    className="w-full rounded-lg border border-neutral-700 bg-neutral-950 p-3 font-mono text-xs text-white placeholder-neutral-500 leading-relaxed focus:border-indigo-500 focus:outline-none"
                  />
                )}
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="formPinned"
                  checked={formPinned}
                  onChange={(e) => setFormPinned(e.target.checked)}
                  className="h-4 w-4 rounded border-neutral-700 bg-neutral-950 text-indigo-600 focus:ring-indigo-500"
                />
                <label htmlFor="formPinned" className="text-xs text-neutral-300 select-none">
                  Pin this note to the top
                </label>
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-neutral-800">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className="rounded-lg px-4 py-2 text-xs font-semibold text-neutral-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-5 py-2 text-xs font-semibold text-white shadow shadow-indigo-600/30 transition"
                >
                  {saving && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                  <span>{saving ? 'Saving...' : 'Save Note'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
