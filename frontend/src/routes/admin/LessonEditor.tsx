import { useState, useEffect } from 'react'
import { useParams, useSearchParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Eye, FileText, Loader2, Save } from 'lucide-react'
import { apiRequest } from '../../lib/api'
import { DifficultyLevel, ContentStatus, Lesson } from '../../types'
import MarkdownViewer from '../../components/content/MarkdownViewer'

export default function LessonEditor() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const moduleIdQuery = searchParams.get('moduleId')
  const navigate = useNavigate()

  const isEditing = Boolean(id)

  const [moduleId, setModuleId] = useState(moduleIdQuery || '')
  const [title, setTitle] = useState('')
  const [slug, setSlug] = useState('')
  const [summary, setSummary] = useState('')
  const [content, setContent] = useState(
    '# Lesson Title\n\nExplain the Go concepts here with examples:\n\n```go\npackage main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello, Go Academy!")\n}\n```\n',
  )
  const [difficulty, setDifficulty] = useState<DifficultyLevel>('beginner')
  const [estimatedMinutes, setEstimatedMinutes] = useState(10)
  const [status, setStatus] = useState<ContentStatus>('draft')
  const [sortOrder, setSortOrder] = useState(1)

  const [loading, setLoading] = useState(isEditing)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'split' | 'edit' | 'preview'>('split')

  useEffect(() => {
    async function loadLesson() {
      if (!id) return
      try {
        setLoading(true)
        const res = await apiRequest<{ lesson: Lesson; prerequisite_ids: string[] }>(`/admin/lessons/${id}`)
        const l = res.lesson
        setTitle(l.title)
        setSlug(l.slug)
        setModuleId(l.module_id)
        setSummary(l.summary || '')
        setContent(l.content)
        setDifficulty(l.difficulty)
        setEstimatedMinutes(l.estimated_minutes)
        setStatus(l.status)
        setSortOrder(l.sort_order)
      } catch (err: any) {
        setError(err.message || 'Failed to load lesson')
      } finally {
        setLoading(false)
      }
    }
    loadLesson()
  }, [id])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title || !content) {
      setError('Title and Content are required')
      return
    }

    try {
      setSaving(true)
      setError(null)

      if (isEditing) {
        await apiRequest(`/admin/lessons/${id}`, {
          method: 'PUT',
          body: JSON.stringify({
            title,
            slug: slug || undefined,
            summary,
            content,
            difficulty,
            estimated_minutes: Number(estimatedMinutes),
            sort_order: Number(sortOrder),
            prerequisite_ids: [],
          }),
        })
      } else {
        if (!moduleId) {
          setError('Module ID is required')
          setSaving(false)
          return
        }
        await apiRequest('/admin/lessons', {
          method: 'POST',
          body: JSON.stringify({
            module_id: moduleId,
            title,
            slug: slug || undefined,
            summary,
            content,
            difficulty,
            estimated_minutes: Number(estimatedMinutes),
            status,
            sort_order: Number(sortOrder),
            prerequisite_ids: [],
          }),
        })
      }

      navigate('/admin/content')
    } catch (err: any) {
      setError(err.message || 'Failed to save lesson')
    } finally {
      setSaving(false)
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
    <div className="space-y-6 max-w-6xl mx-auto pb-16">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-neutral-800 pb-4">
        <div className="flex items-center gap-3">
          <Link
            to="/admin/content"
            className="p-2 rounded-lg bg-neutral-900 border border-neutral-800 hover:bg-neutral-800 text-neutral-400 hover:text-white transition"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
              <FileText className="h-5 w-5 text-indigo-400" />
              {isEditing ? `Edit Lesson: ${title}` : 'Create New Lesson'}
            </h1>
            <p className="text-xs text-neutral-400">Author lesson text in Markdown with real-time preview.</p>
          </div>
        </div>

        {/* View mode toggle */}
        <div className="flex items-center gap-2 bg-neutral-900/80 border border-neutral-800 p-1 rounded-lg">
          <button
            type="button"
            onClick={() => setActiveTab('split')}
            className={`px-3 py-1 rounded text-xs font-medium transition ${
              activeTab === 'split' ? 'bg-neutral-800 text-white' : 'text-neutral-400 hover:text-white'
            }`}
          >
            Split
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('edit')}
            className={`px-3 py-1 rounded text-xs font-medium transition ${
              activeTab === 'edit' ? 'bg-neutral-800 text-white' : 'text-neutral-400 hover:text-white'
            }`}
          >
            Editor Only
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('preview')}
            className={`px-3 py-1 rounded text-xs font-medium transition ${
              activeTab === 'preview' ? 'bg-neutral-800 text-white' : 'text-neutral-400 hover:text-white'
            }`}
          >
            Preview Only
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-4 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Lesson Details Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-4">
          <div className="sm:col-span-2">
            <label className="text-xs font-medium text-neutral-300">Title</label>
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Buffered Channels and Deadlocks"
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Slug (Optional)</label>
            <input
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="auto-generated from title"
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Difficulty</label>
            <select
              value={difficulty}
              onChange={(e) => setDifficulty(e.target.value as DifficultyLevel)}
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white focus:border-indigo-500 focus:outline-none"
            >
              <option value="beginner">Beginner</option>
              <option value="intermediate">Intermediate</option>
              <option value="advanced">Advanced</option>
            </select>
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Estimated Minutes</label>
            <input
              type="number"
              min={1}
              value={estimatedMinutes}
              onChange={(e) => setEstimatedMinutes(Number(e.target.value))}
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value as ContentStatus)}
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white focus:border-indigo-500 focus:outline-none"
            >
              <option value="draft">Draft</option>
              <option value="published">Published</option>
              <option value="archived">Archived</option>
            </select>
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Sort Order</label>
            <input
              type="number"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value))}
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-neutral-300">Module ID</label>
            <input
              type="text"
              required
              disabled={isEditing}
              value={moduleId}
              onChange={(e) => setModuleId(e.target.value)}
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white disabled:opacity-50 focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div className="sm:col-span-4">
            <label className="text-xs font-medium text-neutral-300">Summary</label>
            <input
              type="text"
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="Short summary displayed in the course syllabus..."
              className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-1.5 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
            />
          </div>
        </div>

        {/* Markdown Editor & Live Preview */}
        <div className="grid gap-6 md:grid-cols-2">
          {(activeTab === 'split' || activeTab === 'edit') && (
            <div className={`space-y-2 ${activeTab === 'edit' ? 'md:col-span-2' : ''}`}>
              <div className="flex items-center justify-between text-xs font-semibold text-neutral-400 px-1">
                <span>Markdown Editor</span>
                <span className="font-mono text-[11px] text-neutral-500">{content.length} chars</span>
              </div>
              <textarea
                rows={22}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Write lesson markdown here..."
                className="w-full rounded-xl border border-neutral-800 bg-neutral-950 p-4 font-mono text-xs text-neutral-200 leading-relaxed focus:border-indigo-500 focus:outline-none"
              />
            </div>
          )}

          {(activeTab === 'split' || activeTab === 'preview') && (
            <div className={`space-y-2 ${activeTab === 'preview' ? 'md:col-span-2' : ''}`}>
              <div className="flex items-center justify-between text-xs font-semibold text-neutral-400 px-1">
                <span className="flex items-center gap-1.5">
                  <Eye className="h-3.5 w-3.5" /> Live Preview
                </span>
                <span className="text-[11px] text-neutral-500">Rendered GitHub Flavored Markdown</span>
              </div>
              <div className="h-[500px] overflow-y-auto rounded-xl border border-neutral-800 bg-neutral-900/50 p-6 scrollbar-thin scrollbar-thumb-neutral-700">
                <MarkdownViewer content={content} />
              </div>
            </div>
          )}
        </div>

        {/* Save Bar */}
        <div className="flex justify-end gap-3 pt-4 border-t border-neutral-800">
          <Link
            to="/admin/content"
            className="rounded-lg px-4 py-2 text-xs font-semibold text-neutral-400 hover:text-white"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={saving}
            className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-5 py-2 text-xs font-semibold text-white shadow shadow-indigo-600/30 transition"
          >
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            <span>{saving ? 'Saving...' : 'Save Lesson'}</span>
          </button>
        </div>
      </form>
    </div>
  )
}
