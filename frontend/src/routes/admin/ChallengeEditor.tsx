import { useEffect, useState } from 'react'
import { useNavigate, useParams, Link } from 'react-router-dom'
import {
  ArrowLeft,
  Save,
  Plus,
  Trash2,
  Sparkles,
  AlertCircle,
  Code2,
} from 'lucide-react'
import { apiRequest } from '../../lib/api'
import {
  AdminChallengeDetail,
  CreateTestCasePayload,
  DifficultyLevel,
  ContentStatus,
} from '../../types'
import MarkdownViewer from '../../components/content/MarkdownViewer'

export default function ChallengeEditor() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const isEditing = Boolean(id)

  const [loading, setLoading] = useState(isEditing)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [previewTab, setPreviewTab] = useState<'write' | 'preview'>('write')

  // Form Fields
  const [title, setTitle] = useState('')
  const [slug, setSlug] = useState('')
  const [difficulty, setDifficulty] = useState<DifficultyLevel>('beginner')
  const [estimatedMinutes, setEstimatedMinutes] = useState(15)
  const [status, setStatus] = useState<ContentStatus>('draft')
  const [sortOrder, setSortOrder] = useState(1)
  const [description, setDescription] = useState('')
  const [starterCode, setStarterCode] = useState(`package main

// Solve problem
func Solve() {
\t
}
`)
  const [solutionCode, setSolutionCode] = useState(`package main

func Solve() {
\t
}
`)
  const [testCases, setTestCases] = useState<CreateTestCasePayload[]>([
    {
      title: 'Sample 1',
      input: '',
      expected_output: '',
      hidden: false,
      sort_order: 1,
    },
  ])

  useEffect(() => {
    if (isEditing && id) {
      loadChallenge(id)
    }
  }, [id, isEditing])

  const loadChallenge = async (challengeId: string) => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiRequest<AdminChallengeDetail>(`/admin/challenges/${challengeId}`)
      setTitle(data.title)
      setSlug(data.slug)
      setDifficulty(data.difficulty)
      setEstimatedMinutes(data.estimated_minutes)
      setStatus(data.status)
      setSortOrder(data.sort_order)
      setDescription(data.description)
      setStarterCode(data.starter_code || '')
      setSolutionCode(data.solution_code || '')
      if (data.test_cases && data.test_cases.length > 0) {
        setTestCases(
          data.test_cases.map((tc) => ({
            title: tc.title,
            input: tc.input,
            expected_output: tc.expected_output,
            hidden: tc.hidden,
            sort_order: tc.sort_order,
          })),
        )
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to load challenge')
    } finally {
      setLoading(false)
    }
  }

  const handleAddTestCase = () => {
    setTestCases((prev) => [
      ...prev,
      {
        title: `Test Case ${prev.length + 1}`,
        input: '',
        expected_output: '',
        hidden: prev.length >= 2, // default subsequent to hidden
        sort_order: prev.length + 1,
      },
    ])
  }

  const handleRemoveTestCase = (index: number) => {
    setTestCases((prev) => prev.filter((_, idx) => idx !== index))
  }

  const handleTestCaseChange = (index: number, field: keyof CreateTestCasePayload, value: any) => {
    setTestCases((prev) => {
      const updated = [...prev]
      updated[index] = { ...updated[index], [field]: value }
      return updated
    })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) {
      setError('Title is required')
      return
    }

    try {
      setSaving(true)
      setError(null)

      if (isEditing && id) {
        await apiRequest<AdminChallengeDetail>(`/admin/challenges/${id}`, {
          method: 'PUT',
          body: JSON.stringify({
            title: title.trim(),
            slug: slug.trim() || undefined,
            description,
            difficulty,
            estimated_minutes: Number(estimatedMinutes),
            starter_code: starterCode,
            solution_code: solutionCode,
            sort_order: Number(sortOrder),
            test_cases: testCases,
          }),
        })
      } else {
        await apiRequest<AdminChallengeDetail>('/admin/challenges', {
          method: 'POST',
          body: JSON.stringify({
            title: title.trim(),
            slug: slug.trim() || undefined,
            description,
            difficulty,
            estimated_minutes: Number(estimatedMinutes),
            starter_code: starterCode,
            solution_code: solutionCode,
            status,
            sort_order: Number(sortOrder),
            test_cases: testCases,
          }),
        })
      }

      navigate('/admin/challenges')
    } catch (err: any) {
      setError(err?.message || 'Failed to save challenge')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-sm text-neutral-400">Loading challenge...</div>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-5xl mx-auto space-y-8 pb-16">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-neutral-800 pb-6">
        <div className="flex items-center gap-3">
          <Link
            to="/admin/challenges"
            className="flex items-center gap-1.5 text-xs text-neutral-400 hover:text-white px-2.5 py-1 rounded-md hover:bg-neutral-800 transition"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            Back to Challenges
          </Link>
          <div className="h-4 w-px bg-neutral-800" />
          <h1 className="text-xl font-bold text-white">
            {isEditing ? `Edit Challenge: ${title}` : 'Create New Coding Challenge'}
          </h1>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="flex items-center gap-2 px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-lg shadow-md shadow-indigo-600/30 transition cursor-pointer disabled:opacity-50"
        >
          <Save className="h-4 w-4" />
          <span>{saving ? 'Saving...' : isEditing ? 'Update Challenge' : 'Create Challenge'}</span>
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-red-950/40 border border-red-800/60 text-red-300 text-sm flex items-center gap-3">
          <AlertCircle className="h-5 w-5 shrink-0 text-red-400" />
          <span>{error}</span>
        </div>
      )}

      {/* Metadata Section */}
      <div className="p-6 rounded-xl bg-neutral-900/40 border border-neutral-800 space-y-4">
        <h2 className="text-sm font-semibold text-white uppercase tracking-wider">Problem Details</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1">Title *</label>
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Reverse a Linked List"
              className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1">
              Slug (optional, auto-generated)
            </label>
            <input
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="e.g. reverse-linked-list"
              className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1">Difficulty</label>
            <select
              value={difficulty}
              onChange={(e) => setDifficulty(e.target.value as DifficultyLevel)}
              className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 cursor-pointer"
            >
              <option value="beginner">Beginner</option>
              <option value="intermediate">Intermediate</option>
              <option value="advanced">Advanced</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1">Estimated Minutes</label>
            <input
              type="number"
              min={1}
              value={estimatedMinutes}
              onChange={(e) => setEstimatedMinutes(Number(e.target.value))}
              className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          {!isEditing && (
            <div>
              <label className="block text-xs font-medium text-neutral-400 mb-1">Initial Status</label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as ContentStatus)}
                className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 cursor-pointer"
              >
                <option value="draft">Draft</option>
                <option value="published">Published</option>
              </select>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1">Sort Order</label>
            <input
              type="number"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value))}
              className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>
      </div>

      {/* Description Editor (Markdown + Live Preview) */}
      <div className="p-6 rounded-xl bg-neutral-900/40 border border-neutral-800 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
            Problem Description (Markdown)
          </h2>
          <div className="flex items-center gap-1 bg-neutral-950 p-1 rounded-lg border border-neutral-800 text-xs">
            <button
              type="button"
              onClick={() => setPreviewTab('write')}
              className={`px-3 py-1 rounded-md transition cursor-pointer ${
                previewTab === 'write' ? 'bg-neutral-800 text-white font-semibold' : 'text-neutral-400'
              }`}
            >
              Write
            </button>
            <button
              type="button"
              onClick={() => setPreviewTab('preview')}
              className={`px-3 py-1 rounded-md transition cursor-pointer ${
                previewTab === 'preview' ? 'bg-neutral-800 text-white font-semibold' : 'text-neutral-400'
              }`}
            >
              Preview
            </button>
          </div>
        </div>

        {previewTab === 'write' ? (
          <textarea
            rows={10}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Write clear problem specifications, inputs, constraints, and algorithmic examples..."
            className="w-full bg-neutral-950 border border-neutral-800 rounded-lg p-3 text-xs text-white font-mono placeholder-neutral-500 focus:outline-none focus:border-indigo-500 resize-y"
          />
        ) : (
          <div className="p-4 bg-neutral-950 border border-neutral-800 rounded-lg min-h-[160px]">
            <MarkdownViewer content={description || '_No description provided yet._'} />
          </div>
        )}
      </div>

      {/* Starter and Solution Code */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Starter Code */}
        <div className="p-6 rounded-xl bg-neutral-900/40 border border-neutral-800 space-y-3">
          <div className="flex items-center gap-2 text-indigo-400 text-xs font-semibold uppercase tracking-wider">
            <Code2 className="h-4 w-4" />
            Starter Code Template
          </div>
          <p className="text-[11px] text-neutral-400">
            Initial skeleton provided to learners inside the workspace.
          </p>
          <textarea
            rows={12}
            value={starterCode}
            onChange={(e) => setStarterCode(e.target.value)}
            className="w-full bg-neutral-950 border border-neutral-800 rounded-lg p-3 text-xs text-white font-mono focus:outline-none focus:border-indigo-500 resize-none"
          />
        </div>

        {/* Reference Solution Code */}
        <div className="p-6 rounded-xl bg-neutral-900/40 border border-neutral-800 space-y-3">
          <div className="flex items-center gap-2 text-amber-400 text-xs font-semibold uppercase tracking-wider">
            <Sparkles className="h-4 w-4" />
            Reference Solution Code
          </div>
          <p className="text-[11px] text-neutral-400">
            Internal benchmark solution. Never disclosed to learners.
          </p>
          <textarea
            rows={12}
            value={solutionCode}
            onChange={(e) => setSolutionCode(e.target.value)}
            className="w-full bg-neutral-950 border border-neutral-800 rounded-lg p-3 text-xs text-white font-mono focus:outline-none focus:border-indigo-500 resize-none"
          />
        </div>
      </div>

      {/* Test Cases Section */}
      <div className="p-6 rounded-xl bg-neutral-900/40 border border-neutral-800 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
              Verification Test Cases ({testCases.length})
            </h2>
            <p className="text-[11px] text-neutral-400 mt-0.5">
              Visible sample test cases are shown in the learner workspace. Hidden test cases evaluate submissions.
            </p>
          </div>
          <button
            type="button"
            onClick={handleAddTestCase}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-neutral-800 hover:bg-neutral-700 text-white text-xs font-semibold rounded-lg transition cursor-pointer"
          >
            <Plus className="h-3.5 w-3.5" />
            Add Test Case
          </button>
        </div>

        <div className="space-y-4">
          {testCases.map((tc, idx) => (
            <div
              key={idx}
              className="p-4 rounded-lg bg-neutral-950 border border-neutral-800 space-y-3"
            >
              <div className="flex items-center justify-between gap-4">
                <input
                  type="text"
                  value={tc.title}
                  onChange={(e) => handleTestCaseChange(idx, 'title', e.target.value)}
                  placeholder={`Test Case ${idx + 1}`}
                  className="bg-neutral-900 border border-neutral-800 rounded px-2.5 py-1 text-xs text-white font-medium focus:outline-none focus:border-indigo-500 w-64"
                />

                <div className="flex items-center gap-4">
                  <label className="flex items-center gap-2 text-xs text-neutral-300 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={tc.hidden}
                      onChange={(e) => handleTestCaseChange(idx, 'hidden', e.target.checked)}
                      className="rounded bg-neutral-900 border-neutral-700 text-indigo-600 focus:ring-0 cursor-pointer"
                    />
                    <span>Hidden from learner</span>
                  </label>

                  <button
                    type="button"
                    onClick={() => handleRemoveTestCase(idx)}
                    className="p-1 text-neutral-500 hover:text-red-400 transition cursor-pointer"
                    title="Remove test case"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                  <label className="block text-[11px] font-medium text-neutral-400 mb-1">Input</label>
                  <textarea
                    rows={2}
                    value={tc.input}
                    onChange={(e) => handleTestCaseChange(idx, 'input', e.target.value)}
                    placeholder="e.g. nums = [2, 7, 11, 15], target = 9"
                    className="w-full bg-neutral-900 border border-neutral-800 rounded p-2 text-xs text-indigo-300 font-mono focus:outline-none focus:border-indigo-500 resize-none"
                  />
                </div>
                <div>
                  <label className="block text-[11px] font-medium text-neutral-400 mb-1">
                    Expected Output
                  </label>
                  <textarea
                    rows={2}
                    value={tc.expected_output}
                    onChange={(e) => handleTestCaseChange(idx, 'expected_output', e.target.value)}
                    placeholder="e.g. [0, 1]"
                    className="w-full bg-neutral-900 border border-neutral-800 rounded p-2 text-xs text-emerald-300 font-mono focus:outline-none focus:border-indigo-500 resize-none"
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </form>
  )
}
