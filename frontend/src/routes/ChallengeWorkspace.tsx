import { useEffect, useState, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  ArrowLeft,
  Play,
  RotateCcw,
  CheckCircle2,
  XCircle,
  Copy,
  Check,
  FileCode,
  FileText,
  History,
  Clock,
  Sparkles,
  AlertCircle,
  Send,
  Save,
  Tag as TagIcon,
  Terminal,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import {
  ChallengeDetail,
  DifficultyLevel,
  Note,
  ExecutionResult,
  SubmissionDetail,
  SubmissionSummary,
  TestCaseOutcome,
} from '../types'
import MarkdownViewer from '../components/content/MarkdownViewer'

export default function ChallengeWorkspace() {
  const { slug } = useParams<{ slug: string }>()
  const [challenge, setChallenge] = useState<ChallengeDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Code editor state
  const [code, setCode] = useState<string>('')
  const [activeTab, setActiveTab] = useState<'problem' | 'notes' | 'history'>('problem')
  const [activeTestIndex, setActiveTestIndex] = useState<number>(0)
  const [copiedCode, setCopiedCode] = useState(false)

  // Execution state
  const [isRunning, setIsRunning] = useState(false)
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null)
  const [lastSubmission, setLastSubmission] = useState<SubmissionDetail | null>(null)
  const [submissions, setSubmissions] = useState<SubmissionSummary[]>([])
  const [loadingSubmissions, setLoadingSubmissions] = useState(false)

  // Notes state
  const [noteTitle, setNoteTitle] = useState('')
  const [noteContent, setNoteContent] = useState('')
  const [noteTags, setNoteTags] = useState('challenges, algorithms')
  const [noteSaved, setNoteSaved] = useState(false)
  const [savingNote, setSavingNote] = useState(false)

  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (slug) {
      loadChallenge(slug)
      loadSubmissions(slug)
    }
  }, [slug])

  const loadChallenge = async (challengeSlug: string) => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiRequest<ChallengeDetail>(`/challenges/${challengeSlug}`)
      setChallenge(data)

      const savedCode = localStorage.getItem(`code_${challengeSlug}`)
      if (savedCode) {
        setCode(savedCode)
      } else {
        setCode(data.starter_code || '')
      }

      setNoteTitle(`Notes on: ${data.title}`)
    } catch (err: any) {
      setError(err?.message || 'Failed to load challenge')
    } finally {
      setLoading(false)
    }
  }

  const loadSubmissions = async (challengeSlug: string) => {
    try {
      setLoadingSubmissions(true)
      const data = await apiRequest<SubmissionSummary[]>(`/challenges/${challengeSlug}/submissions`)
      setSubmissions(data || [])
    } catch {
      // Ignore initial history failure if none exist
    } finally {
      setLoadingSubmissions(false)
    }
  }

  const handleCodeChange = (newCode: string) => {
    setCode(newCode)
    if (slug) {
      localStorage.setItem(`code_${slug}`, newCode)
    }
  }

  const handleResetCode = () => {
    if (challenge && window.confirm('Reset code to the original starter template? Your edits will be lost.')) {
      setCode(challenge.starter_code || '')
      if (slug) {
        localStorage.removeItem(`code_${slug}`)
      }
    }
  }

  const handleCopyCode = () => {
    navigator.clipboard.writeText(code)
    setCopiedCode(true)
    setTimeout(() => setCopiedCode(false), 2000)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Tab') {
      e.preventDefault()
      const textarea = textareaRef.current
      if (!textarea) return
      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const newCode = code.substring(0, start) + '\t' + code.substring(end)
      handleCodeChange(newCode)
      setTimeout(() => {
        textarea.selectionStart = textarea.selectionEnd = start + 1
      }, 0)
    }
  }

  // Run sample tests
  const handleRunTests = async () => {
    if (!slug) return
    try {
      setIsRunning(true)
      setExecutionResult(null)
      setActiveTestIndex(0)

      const res = await apiRequest<ExecutionResult>(`/challenges/${slug}/run`, {
        method: 'POST',
        body: JSON.stringify({ code }),
      })
      setExecutionResult(res)
    } catch (err: any) {
      setExecutionResult({
        status: 'error',
        total_tests: 0,
        passed_tests: 0,
        execution_time_ms: 0,
        compiler_output: err?.message || 'Sandbox execution request failed',
        test_outcomes: [],
      })
    } finally {
      setIsRunning(false)
    }
  }

  // Submit solution
  const handleSubmit = async () => {
    if (!slug) return
    try {
      setIsRunning(true)
      setExecutionResult(null)
      setLastSubmission(null)
      setActiveTestIndex(0)

      const sub = await apiRequest<SubmissionDetail>(`/challenges/${slug}/submit`, {
        method: 'POST',
        body: JSON.stringify({ code }),
      })

      setLastSubmission(sub)
      setExecutionResult({
        status: sub.status as any,
        total_tests: sub.total_test_cases,
        passed_tests: sub.passed_test_cases,
        execution_time_ms: sub.execution_time_ms,
        compiler_output: sub.error_output,
        test_outcomes: sub.test_outcomes,
      })

      loadSubmissions(slug)
    } catch (err: any) {
      setExecutionResult({
        status: 'error',
        total_tests: 0,
        passed_tests: 0,
        execution_time_ms: 0,
        compiler_output: err?.message || 'Submission evaluation request failed',
        test_outcomes: [],
      })
    } finally {
      setIsRunning(false)
    }
  }

  // Save Note
  const handleSaveNote = async () => {
    if (!noteTitle.trim()) return
    try {
      setSavingNote(true)
      const tagsArray = noteTags
        .split(',')
        .map((t) => t.trim().toLowerCase())
        .filter((t) => t.length > 0)

      await apiRequest<Note>('/notes', {
        method: 'POST',
        body: JSON.stringify({
          title: noteTitle.trim(),
          content: noteContent,
          pinned: false,
          tags: tagsArray,
        }),
      })

      setNoteSaved(true)
      setTimeout(() => setNoteSaved(false), 3000)
    } catch (err: any) {
      alert('Failed to save note: ' + (err?.message || 'Unknown error'))
    } finally {
      setSavingNote(false)
    }
  }

  const loadPastCode = (pastCode: string) => {
    if (window.confirm('Load this submission code into your editor? Your current unsaved edits will be replaced.')) {
      handleCodeChange(pastCode)
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

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="flex items-center gap-3 text-neutral-400 text-sm">
          <div className="h-5 w-5 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
          <span>Loading workspace...</span>
        </div>
      </div>
    )
  }

  if (error || !challenge) {
    return (
      <div className="max-w-md mx-auto py-16 text-center space-y-4">
        <div className="h-12 w-12 rounded-xl bg-red-950/40 border border-red-800 text-red-400 mx-auto flex items-center justify-center">
          <AlertCircle className="h-6 w-6" />
        </div>
        <h2 className="text-lg font-bold text-white">Challenge Not Found</h2>
        <p className="text-xs text-neutral-400">{error || 'The requested challenge does not exist.'}</p>
        <Link
          to="/challenges"
          className="inline-flex items-center gap-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300"
        >
          <ArrowLeft className="h-4 w-4" />
          Return to Challenges
        </Link>
      </div>
    )
  }

  const lineCount = Math.max(code.split('\n').length, 18)
  const lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1)
  const outcomes: TestCaseOutcome[] = executionResult?.test_outcomes || []
  const activeOutcome = outcomes[activeTestIndex]

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] -m-8 overflow-hidden bg-neutral-950">
      {/* Top Header */}
      <div className="h-14 border-b border-neutral-800/80 bg-neutral-900/60 px-6 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-3">
          <Link
            to="/challenges"
            className="flex items-center gap-1.5 text-xs text-neutral-400 hover:text-white transition px-2.5 py-1 rounded-md hover:bg-neutral-800"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            Challenges
          </Link>
          <div className="h-4 w-px bg-neutral-800" />
          <h1 className="text-sm font-bold text-white truncate max-w-md">{challenge.title}</h1>
          <span
            className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium border capitalize ${getDifficultyColor(
              challenge.difficulty,
            )}`}
          >
            {challenge.difficulty}
          </span>
        </div>

        <div className="flex items-center gap-3 text-xs text-neutral-400">
          <div className="flex items-center gap-1">
            <Clock className="h-3.5 w-3.5" />
            <span>{challenge.estimated_minutes} mins</span>
          </div>
        </div>
      </div>

      {/* Split Pane Container */}
      <div className="flex-1 flex min-h-0">
        {/* Left Pane */}
        <div className="w-1/2 border-r border-neutral-800/80 flex flex-col bg-neutral-900/20">
          {/* Tabs */}
          <div className="h-11 border-b border-neutral-800 bg-neutral-900/40 px-4 flex items-center gap-2 shrink-0">
            <button
              onClick={() => setActiveTab('problem')}
              className={`flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer ${
                activeTab === 'problem'
                  ? 'bg-neutral-800 text-white font-semibold'
                  : 'text-neutral-400 hover:text-white hover:bg-neutral-800/50'
              }`}
            >
              <FileCode className="h-3.5 w-3.5 text-indigo-400" />
              Problem
            </button>
            <button
              onClick={() => setActiveTab('notes')}
              className={`flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer ${
                activeTab === 'notes'
                  ? 'bg-neutral-800 text-white font-semibold'
                  : 'text-neutral-400 hover:text-white hover:bg-neutral-800/50'
              }`}
            >
              <FileText className="h-3.5 w-3.5 text-amber-400" />
              Notes
            </button>
            <button
              onClick={() => setActiveTab('history')}
              className={`flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer ${
                activeTab === 'history'
                  ? 'bg-neutral-800 text-white font-semibold'
                  : 'text-neutral-400 hover:text-white hover:bg-neutral-800/50'
              }`}
            >
              <History className="h-3.5 w-3.5 text-sky-400" />
              History ({submissions.length})
            </button>
          </div>

          {/* Left Tab Content */}
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {activeTab === 'problem' ? (
              <div className="space-y-6">
                <div>
                  <h2 className="text-xl font-bold text-white mb-2">{challenge.title}</h2>
                  <div className="text-xs text-neutral-400">
                    <MarkdownViewer content={challenge.description} />
                  </div>
                </div>

                {/* Visible Sample Test Cases */}
                {challenge.sample_test_cases.length > 0 && (
                  <div className="space-y-4 pt-4 border-t border-neutral-800">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-neutral-400 flex items-center gap-1.5">
                      <Sparkles className="h-3.5 w-3.5 text-indigo-400" />
                      Sample Test Cases
                    </h3>
                    <div className="space-y-3">
                      {challenge.sample_test_cases.map((tc, idx) => (
                        <div
                          key={tc.id}
                          className="rounded-lg border border-neutral-800 bg-neutral-900/80 p-3.5 space-y-2 text-xs"
                        >
                          <div className="font-semibold text-neutral-300">
                            {tc.title || `Sample ${idx + 1}`}
                          </div>
                          <div className="space-y-1">
                            <span className="text-[11px] text-neutral-500 font-medium">Input:</span>
                            <div className="font-mono bg-neutral-950 p-2 rounded border border-neutral-800 text-indigo-300 overflow-x-auto">
                              {tc.input}
                            </div>
                          </div>
                          <div className="space-y-1">
                            <span className="text-[11px] text-neutral-500 font-medium">Expected Output:</span>
                            <div className="font-mono bg-neutral-950 p-2 rounded border border-neutral-800 text-emerald-300 overflow-x-auto">
                              {tc.expected_output}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : activeTab === 'notes' ? (
              /* Notes Tab */
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-xs font-bold uppercase tracking-wider text-neutral-400">
                    Challenge Notes
                  </h3>
                  <button
                    onClick={handleSaveNote}
                    disabled={savingNote}
                    className="flex items-center gap-1.5 px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition cursor-pointer disabled:opacity-50"
                  >
                    {noteSaved ? (
                      <>
                        <Check className="h-3.5 w-3.5 text-emerald-300" />
                        <span>Saved!</span>
                      </>
                    ) : (
                      <>
                        <Save className="h-3.5 w-3.5" />
                        <span>Save Note</span>
                      </>
                    )}
                  </button>
                </div>

                <div className="space-y-3">
                  <div>
                    <label className="block text-[11px] font-medium text-neutral-400 mb-1">Title</label>
                    <input
                      type="text"
                      value={noteTitle}
                      onChange={(e) => setNoteTitle(e.target.value)}
                      className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-neutral-400 mb-1 flex items-center gap-1">
                      <TagIcon className="h-3 w-3" />
                      Tags
                    </label>
                    <input
                      type="text"
                      value={noteTags}
                      onChange={(e) => setNoteTags(e.target.value)}
                      placeholder="e.g. go, algorithms, channels"
                      className="w-full bg-neutral-950 border border-neutral-800 rounded-lg px-3 py-2 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-neutral-400 mb-1">Notes</label>
                    <textarea
                      rows={12}
                      value={noteContent}
                      onChange={(e) => setNoteContent(e.target.value)}
                      placeholder="Record algorithmic complexity, tradeoffs, or key observations..."
                      className="w-full bg-neutral-950 border border-neutral-800 rounded-lg p-3 text-xs text-neutral-200 font-mono placeholder-neutral-600 focus:outline-none focus:border-indigo-500 resize-none"
                    />
                  </div>
                </div>
              </div>
            ) : (
              /* Submissions History Tab */
              <div className="space-y-4">
                <h3 className="text-xs font-bold uppercase tracking-wider text-neutral-400">
                  Past Submissions
                </h3>
                {loadingSubmissions ? (
                  <div className="text-xs text-neutral-400 py-6 text-center">Loading history...</div>
                ) : submissions.length === 0 ? (
                  <div className="text-xs text-neutral-500 py-8 text-center bg-neutral-900/40 rounded-xl border border-neutral-800">
                    No submissions recorded yet. Submit your solution to track progress!
                  </div>
                ) : (
                  <div className="space-y-3">
                    {submissions.map((sub) => (
                      <div
                        key={sub.id}
                        className="rounded-lg border border-neutral-800 bg-neutral-900/60 p-3.5 flex items-center justify-between text-xs"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold uppercase ${
                                sub.status === 'passed'
                                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                  : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                              }`}
                            >
                              {sub.status}
                            </span>
                            <span className="font-mono text-neutral-300">
                              {sub.passed_test_cases}/{sub.total_test_cases} passed
                            </span>
                          </div>
                          <div className="text-[11px] text-neutral-500">
                            {new Date(sub.created_at).toLocaleString()} • {sub.execution_time_ms}ms
                          </div>
                        </div>

                        <button
                          onClick={async () => {
                            try {
                              const detail = await apiRequest<SubmissionDetail>(
                                `/challenges/${slug}/submissions/${sub.id}`,
                              )
                              loadPastCode(detail.code)
                            } catch (e: any) {
                              alert('Failed to load past submission code: ' + e?.message)
                            }
                          }}
                          className="px-2.5 py-1 text-[11px] font-medium bg-neutral-800 hover:bg-neutral-700 text-neutral-200 rounded transition cursor-pointer"
                        >
                          Restore Code
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Right Pane: Code Editor & Execution Output */}
        <div className="w-1/2 flex flex-col bg-neutral-950">
          {/* Editor Header Toolbar */}
          <div className="h-11 border-b border-neutral-800 bg-neutral-900/50 px-4 flex items-center justify-between shrink-0">
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs font-semibold text-neutral-300">solution.go</span>
              <span className="text-[10px] text-neutral-500 font-mono">Go 1.23</span>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={handleCopyCode}
                title="Copy code"
                className="p-1.5 text-neutral-400 hover:text-white rounded hover:bg-neutral-800 transition cursor-pointer"
              >
                {copiedCode ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
              </button>
              <button
                onClick={handleResetCode}
                title="Reset to starter code"
                className="flex items-center gap-1 px-2.5 py-1 text-xs text-neutral-400 hover:text-amber-400 rounded hover:bg-neutral-800 transition cursor-pointer"
              >
                <RotateCcw className="h-3 w-3" />
                <span>Reset</span>
              </button>
            </div>
          </div>

          {/* Editor Area with Line Numbers */}
          <div className="flex-1 relative flex overflow-hidden font-mono text-xs bg-neutral-950">
            {/* Line numbers column */}
            <div className="w-12 select-none py-4 text-right pr-3 text-neutral-600 bg-neutral-950 border-r border-neutral-900 shrink-0 font-mono text-xs leading-5">
              {lineNumbers.map((n) => (
                <div key={n}>{n}</div>
              ))}
            </div>

            {/* Code Textarea */}
            <textarea
              ref={textareaRef}
              value={code}
              onChange={(e) => handleCodeChange(e.target.value)}
              onKeyDown={handleKeyDown}
              spellCheck={false}
              autoCapitalize="off"
              autoComplete="off"
              className="flex-1 p-4 bg-transparent text-neutral-100 font-mono text-xs leading-5 resize-none focus:outline-none scrollbar-thin scrollbar-thumb-neutral-800"
            />
          </div>

          {/* Execution Drawer / Results Area */}
          {executionResult && (
            <div className="h-52 border-t border-neutral-800 bg-neutral-900/90 flex flex-col shrink-0">
              {/* Status Header */}
              <div className="h-9 border-b border-neutral-800 px-4 flex items-center justify-between bg-neutral-950/70">
                <div className="flex items-center gap-2 overflow-x-auto">
                  {executionResult.status === 'compile_error' ? (
                    <div className="flex items-center gap-1.5 text-rose-400 font-mono text-xs">
                      <XCircle className="h-3.5 w-3.5" />
                      <span>Compilation Error</span>
                    </div>
                  ) : executionResult.status === 'timed_out' ? (
                    <div className="flex items-center gap-1.5 text-amber-400 font-mono text-xs">
                      <Clock className="h-3.5 w-3.5" />
                      <span>Execution Timed Out (5s limit)</span>
                    </div>
                  ) : (
                    outcomes.map((o, idx) => (
                      <button
                        key={o.test_case_id || idx}
                        onClick={() => setActiveTestIndex(idx)}
                        className={`flex items-center gap-1.5 px-2.5 py-1 rounded text-[11px] font-medium transition cursor-pointer shrink-0 ${
                          activeTestIndex === idx
                            ? 'bg-neutral-800 text-white font-semibold'
                            : 'text-neutral-400 hover:text-white'
                        }`}
                      >
                        {o.passed ? (
                          <CheckCircle2 className="h-3 w-3 text-emerald-400" />
                        ) : (
                          <XCircle className="h-3 w-3 text-rose-400" />
                        )}
                        <span>{o.hidden ? `Hidden ${idx + 1}` : `Case ${idx + 1}`}</span>
                      </button>
                    ))
                  )}
                </div>

                <div className="flex items-center gap-3 text-[11px] font-mono text-neutral-400 shrink-0">
                  <span>{executionResult.execution_time_ms}ms</span>
                  <span>
                    {executionResult.passed_tests}/{executionResult.total_tests} passed
                  </span>
                </div>
              </div>

              {/* Drawer Body */}
              <div className="flex-1 p-3 overflow-y-auto font-mono text-xs space-y-2">
                {executionResult.compiler_output ? (
                  <div className="p-3 bg-neutral-950 rounded border border-rose-900/40 text-rose-300 font-mono text-[11px] whitespace-pre-wrap leading-relaxed">
                    <div className="flex items-center gap-2 text-neutral-400 mb-2 font-semibold">
                      <Terminal className="h-3.5 w-3.5" />
                      Compiler Logs
                    </div>
                    {executionResult.compiler_output}
                  </div>
                ) : activeOutcome ? (
                  <div className="space-y-2">
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <div className="text-[10px] text-neutral-500 uppercase font-semibold">Input</div>
                        <div className="bg-neutral-950 p-2 rounded text-indigo-300 text-[11px] border border-neutral-800 overflow-x-auto">
                          {activeOutcome.input || 'None'}
                        </div>
                      </div>
                      <div>
                        <div className="text-[10px] text-neutral-500 uppercase font-semibold">
                          Expected Output
                        </div>
                        <div className="bg-neutral-950 p-2 rounded text-emerald-400 text-[11px] border border-neutral-800 overflow-x-auto">
                          {activeOutcome.expected_output || 'None'}
                        </div>
                      </div>
                    </div>

                    <div>
                      <div className="text-[10px] text-neutral-500 uppercase font-semibold">
                        Actual Output
                      </div>
                      <div
                        className={`p-2 rounded text-[11px] border font-mono ${
                          activeOutcome.passed
                            ? 'bg-emerald-950/20 text-emerald-300 border-emerald-900/40'
                            : 'bg-rose-950/20 text-rose-300 border-rose-900/40'
                        }`}
                      >
                        {activeOutcome.actual_output || 'No output produced'}
                      </div>
                    </div>

                    {activeOutcome.error_message && (
                      <div className="text-[11px] text-rose-400 bg-rose-950/30 p-2 rounded border border-rose-900/30">
                        {activeOutcome.error_message}
                      </div>
                    )}
                  </div>
                ) : null}
              </div>
            </div>
          )}

          {/* Success Banner */}
          {lastSubmission && lastSubmission.status === 'passed' && (
            <div className="px-4 py-2 bg-emerald-950/50 border-t border-emerald-800/50 text-xs text-emerald-300 flex items-center justify-between">
              <div className="flex items-center gap-2 font-medium">
                <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                <span>All test cases passed! Solution verified in {lastSubmission.execution_time_ms}ms.</span>
              </div>
            </div>
          )}

          {/* Bottom Execution Bar */}
          <div className="h-14 border-t border-neutral-800 bg-neutral-900/90 px-4 flex items-center justify-between shrink-0">
            <div className="text-xs text-neutral-400 font-mono">
              Docker Sandbox Boundary: Active
            </div>

            <div className="flex items-center gap-3">
              <button
                onClick={handleRunTests}
                disabled={isRunning}
                className="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 text-white text-xs font-semibold transition cursor-pointer disabled:opacity-50"
              >
                <Play className="h-3.5 w-3.5 text-emerald-400 fill-emerald-400" />
                <span>{isRunning ? 'Evaluating...' : 'Run Sample Tests'}</span>
              </button>

              <button
                onClick={handleSubmit}
                disabled={isRunning}
                className="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/30 transition cursor-pointer disabled:opacity-50"
              >
                <Send className="h-3.5 w-3.5" />
                <span>Submit Solution</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
