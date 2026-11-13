import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Code2,
  Plus,
  Trash2,
  Edit,
  Clock,
  AlertCircle,
  Eye,
  Archive,
  CheckCircle,
} from 'lucide-react'
import { apiRequest } from '../../lib/api'
import { ChallengeSummary, ContentStatus, DifficultyLevel } from '../../types'

export default function AdminChallenges() {
  const [challenges, setChallenges] = useState<ChallengeSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadChallenges()
  }, [])

  const loadChallenges = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiRequest<ChallengeSummary[]>('/admin/challenges')
      setChallenges(data || [])
    } catch (err: any) {
      setError(err?.message || 'Failed to load challenges')
    } finally {
      setLoading(false)
    }
  }

  const handleStatusChange = async (id: string, newStatus: ContentStatus) => {
    try {
      await apiRequest<ChallengeSummary>(`/admin/challenges/${id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status: newStatus }),
      })
      await loadChallenges()
    } catch (err: any) {
      alert('Failed to update status: ' + (err?.message || 'Unknown error'))
    }
  }

  const handleDelete = async (id: string, title: string) => {
    if (!window.confirm(`Are you sure you want to permanently delete "${title}"?`)) {
      return
    }
    try {
      await apiRequest(`/admin/challenges/${id}`, { method: 'DELETE' })
      setChallenges((prev) => prev.filter((c) => c.id !== id))
    } catch (err: any) {
      alert('Failed to delete challenge: ' + (err?.message || 'Unknown error'))
    }
  }

  const getDifficultyColor = (difficulty: DifficultyLevel) => {
    switch (difficulty) {
      case 'beginner':
        return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
      case 'intermediate':
        return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
      case 'advanced':
        return 'text-rose-400 bg-rose-500/10 border-rose-500/20'
      default:
        return 'text-neutral-400 bg-neutral-800 border-neutral-700'
    }
  }

  const getStatusColor = (status: ContentStatus) => {
    switch (status) {
      case 'published':
        return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
      case 'draft':
        return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
      case 'archived':
        return 'text-neutral-400 bg-neutral-800 border-neutral-700'
      default:
        return 'text-neutral-400 bg-neutral-800'
    }
  }

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-neutral-800 pb-6">
        <div>
          <div className="flex items-center gap-2 text-amber-400 text-xs font-semibold tracking-wider uppercase mb-1">
            <Code2 className="h-3.5 w-3.5" />
            Curriculum Administration
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Coding Challenges</h1>
          <p className="text-sm text-neutral-400 mt-1">
            Create and maintain coding problems, starter templates, and evaluation test cases.
          </p>
        </div>

        <Link
          to="/admin/challenges/new"
          className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-lg shadow-md shadow-indigo-600/30 transition cursor-pointer"
        >
          <Plus className="h-4 w-4" />
          <span>New Challenge</span>
        </Link>
      </div>

      {/* Error state */}
      {error && (
        <div className="p-4 rounded-xl bg-red-950/40 border border-red-800/60 text-red-300 text-sm flex items-center gap-3">
          <AlertCircle className="h-5 w-5 shrink-0 text-red-400" />
          <span>{error}</span>
          <button
            onClick={loadChallenges}
            className="ml-auto text-xs underline hover:text-white cursor-pointer"
          >
            Retry
          </button>
        </div>
      )}

      {/* Table list */}
      {loading ? (
        <div className="py-16 text-center text-sm text-neutral-400">Loading challenges...</div>
      ) : challenges.length === 0 ? (
        <div className="py-16 text-center space-y-3 bg-neutral-900/30 rounded-2xl border border-neutral-800/80">
          <div className="h-12 w-12 rounded-xl bg-neutral-800 text-neutral-400 mx-auto flex items-center justify-center">
            <Code2 className="h-6 w-6" />
          </div>
          <h3 className="text-base font-semibold text-white">No challenges created</h3>
          <p className="text-xs text-neutral-400">Get started by creating your first coding challenge.</p>
        </div>
      ) : (
        <div className="bg-neutral-900/40 border border-neutral-800 rounded-xl overflow-hidden">
          <table className="w-full text-left text-xs">
            <thead className="bg-neutral-950/60 text-neutral-400 border-b border-neutral-800 uppercase tracking-wider font-semibold">
              <tr>
                <th className="px-6 py-3.5">Title & Slug</th>
                <th className="px-4 py-3.5">Difficulty</th>
                <th className="px-4 py-3.5">Duration</th>
                <th className="px-4 py-3.5">Status</th>
                <th className="px-6 py-3.5 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800/60">
              {challenges.map((c) => (
                <tr key={c.id} className="hover:bg-neutral-850/40 transition">
                  <td className="px-6 py-4">
                    <div className="font-semibold text-white text-sm">{c.title}</div>
                    <div className="text-neutral-500 font-mono text-[11px] mt-0.5">{c.slug}</div>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium border capitalize ${getDifficultyColor(
                        c.difficulty,
                      )}`}
                    >
                      {c.difficulty}
                    </span>
                  </td>
                  <td className="px-4 py-4 text-neutral-300">
                    <div className="flex items-center gap-1.5">
                      <Clock className="h-3.5 w-3.5 text-neutral-500" />
                      <span>{c.estimated_minutes}m</span>
                    </div>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium border capitalize ${getStatusColor(
                        c.status,
                      )}`}
                    >
                      {c.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {c.status === 'draft' ? (
                        <button
                          onClick={() => handleStatusChange(c.id, 'published')}
                          title="Publish Challenge"
                          className="p-1.5 text-emerald-400 hover:bg-neutral-800 rounded transition cursor-pointer"
                        >
                          <CheckCircle className="h-4 w-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => handleStatusChange(c.id, 'draft')}
                          title="Unpublish (Revert to Draft)"
                          className="p-1.5 text-amber-400 hover:bg-neutral-800 rounded transition cursor-pointer"
                        >
                          <Archive className="h-4 w-4" />
                        </button>
                      )}

                      <Link
                        to={`/challenges/${c.slug}`}
                        title="Preview as Learner"
                        className="p-1.5 text-neutral-400 hover:text-white hover:bg-neutral-800 rounded transition cursor-pointer"
                      >
                        <Eye className="h-4 w-4" />
                      </Link>

                      <Link
                        to={`/admin/challenges/${c.id}/edit`}
                        title="Edit Challenge"
                        className="p-1.5 text-neutral-400 hover:text-white hover:bg-neutral-800 rounded transition cursor-pointer"
                      >
                        <Edit className="h-4 w-4" />
                      </Link>

                      <button
                        onClick={() => handleDelete(c.id, c.title)}
                        title="Delete Challenge"
                        className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-neutral-800 rounded transition cursor-pointer"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
