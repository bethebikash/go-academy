import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Code2,
  Clock,
  Sparkles,
  Search,
  ArrowRight,
  AlertCircle,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import { ChallengeSummary, DifficultyLevel } from '../types'

export default function Challenges() {
  const [challenges, setChallenges] = useState<ChallengeSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedDifficulty, setSelectedDifficulty] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    fetchChallenges()
  }, [selectedDifficulty])

  const fetchChallenges = async () => {
    try {
      setLoading(true)
      setError(null)
      const query = selectedDifficulty !== 'all' ? `?difficulty=${selectedDifficulty}` : ''
      const data = await apiRequest<ChallengeSummary[]>(`/challenges${query}`)
      setChallenges(data || [])
    } catch (err: any) {
      setError(err?.message || 'Failed to load coding challenges')
    } finally {
      setLoading(false)
    }
  }

  const filteredChallenges = challenges.filter((c) => {
    if (!searchQuery.trim()) return true
    const q = searchQuery.toLowerCase()
    return (
      c.title.toLowerCase().includes(q) ||
      c.description.toLowerCase().includes(q) ||
      c.slug.toLowerCase().includes(q)
    )
  })

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

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-neutral-800 pb-6">
        <div>
          <div className="flex items-center gap-2 text-indigo-400 text-xs font-semibold tracking-wider uppercase mb-1">
            <Sparkles className="h-3.5 w-3.5" />
            Hands-on Practice
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Interactive Go Challenges</h1>
          <p className="text-sm text-neutral-400 mt-1 max-w-2xl">
            Test and sharpen your Go fundamentals, concurrency primitives, and idiomatic patterns with verified test suites.
          </p>
        </div>
      </div>

      {/* Filters and Search Bar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 bg-neutral-900/60 p-4 rounded-xl border border-neutral-800">
        {/* Difficulty Tabs */}
        <div className="flex items-center gap-1.5 p-1 bg-neutral-950 rounded-lg border border-neutral-800 text-xs font-medium">
          {['all', 'beginner', 'intermediate', 'advanced'].map((lvl) => (
            <button
              key={lvl}
              onClick={() => setSelectedDifficulty(lvl)}
              className={`px-3 py-1.5 rounded-md capitalize transition cursor-pointer ${
                selectedDifficulty === lvl
                  ? 'bg-neutral-800 text-white font-semibold shadow-sm'
                  : 'text-neutral-400 hover:text-white'
              }`}
            >
              {lvl}
            </button>
          ))}
        </div>

        {/* Search Input */}
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-500" />
          <input
            type="text"
            placeholder="Search challenges..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-neutral-950 border border-neutral-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-indigo-500 transition"
          />
        </div>
      </div>

      {/* Error state */}
      {error && (
        <div className="p-4 rounded-xl bg-red-950/40 border border-red-800/60 text-red-300 text-sm flex items-center gap-3">
          <AlertCircle className="h-5 w-5 shrink-0 text-red-400" />
          <span>{error}</span>
          <button
            onClick={fetchChallenges}
            className="ml-auto text-xs underline hover:text-white cursor-pointer"
          >
            Retry
          </button>
        </div>
      )}

      {/* Loading Skeletons */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="h-44 rounded-xl bg-neutral-900/40 border border-neutral-800 animate-pulse p-6 space-y-4"
            >
              <div className="h-4 bg-neutral-800 rounded w-1/3"></div>
              <div className="h-4 bg-neutral-800 rounded w-3/4"></div>
              <div className="h-10 bg-neutral-800 rounded w-full"></div>
            </div>
          ))}
        </div>
      ) : filteredChallenges.length === 0 ? (
        /* Empty State */
        <div className="py-16 text-center space-y-3 bg-neutral-900/30 rounded-2xl border border-neutral-800/80">
          <div className="h-12 w-12 rounded-xl bg-neutral-800 text-neutral-400 mx-auto flex items-center justify-center">
            <Code2 className="h-6 w-6" />
          </div>
          <h3 className="text-base font-semibold text-white">No challenges found</h3>
          <p className="text-xs text-neutral-400 max-w-sm mx-auto">
            {searchQuery
              ? `No challenges matched your search "${searchQuery}".`
              : 'There are no challenges available for the selected difficulty level.'}
          </p>
        </div>
      ) : (
        /* Challenges Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredChallenges.map((challenge) => (
            <div
              key={challenge.id}
              className="group flex flex-col justify-between p-5 rounded-xl bg-neutral-900/40 hover:bg-neutral-900/70 border border-neutral-800 hover:border-neutral-700 transition"
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-medium border capitalize ${getDifficultyColor(
                      challenge.difficulty,
                    )}`}
                  >
                    {challenge.difficulty}
                  </span>
                  <div className="flex items-center gap-1.5 text-xs text-neutral-400">
                    <Clock className="h-3.5 w-3.5" />
                    <span>{challenge.estimated_minutes} mins</span>
                  </div>
                </div>

                <div>
                  <h3 className="text-base font-semibold text-white group-hover:text-indigo-400 transition">
                    {challenge.title}
                  </h3>
                  <p className="text-xs text-neutral-400 mt-1 line-clamp-2 leading-relaxed">
                    {challenge.description.replace(/[#*`]/g, '')}
                  </p>
                </div>
              </div>

              <div className="pt-4 mt-4 border-t border-neutral-800/60 flex items-center justify-between">
                <span className="text-[11px] text-neutral-500 font-mono">
                  slug: {challenge.slug}
                </span>
                <Link
                  to={`/challenges/${challenge.slug}`}
                  className="inline-flex items-center gap-1.5 text-xs font-semibold text-indigo-400 hover:text-indigo-300 group-hover:translate-x-0.5 transition"
                >
                  Solve Challenge
                  <ArrowRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
