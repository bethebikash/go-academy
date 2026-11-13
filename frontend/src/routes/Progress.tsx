import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Flame,
  Award,
  BookOpen,
  Code2,
  FileText,
  Sparkles,
  CheckCircle2,
  Lock,
  Calendar,
  ExternalLink,
  ShieldCheck,
  Trophy,
} from 'lucide-react'
import { apiRequest } from '../lib/api'
import {
  UserSummaryStats,
  Badge,
  Certificate,
  ActivityCount,
  LearningPath,
} from '../types'

export default function Progress() {
  const [stats, setStats] = useState<UserSummaryStats | null>(null)
  const [activity, setActivity] = useState<ActivityCount[]>([])
  const [badges, setBadges] = useState<Badge[]>([])
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const [paths, setPaths] = useState<LearningPath[]>([])
  const [loading, setLoading] = useState(true)
  const [claiming, setClaiming] = useState<string | null>(null)
  const [selectedCert, setSelectedCert] = useState<Certificate | null>(null)
  const [activeTab, setActiveTab] = useState<'overview' | 'badges' | 'certificates'>('overview')

  useEffect(() => {
    loadAllData()
  }, [])

  const loadAllData = async () => {
    try {
      setLoading(true)
      const [statsData, actData, badgeData, certData, pathsData] = await Promise.all([
        apiRequest<UserSummaryStats>('/analytics/summary'),
        apiRequest<ActivityCount[]>('/analytics/activity?days=90'),
        apiRequest<Badge[]>('/analytics/badges'),
        apiRequest<Certificate[]>('/certificates/me'),
        apiRequest<LearningPath[]>('/learning-paths'),
      ])

      setStats(statsData)
      setActivity(actData || [])
      setBadges(badgeData || [])
      setCertificates(certData || [])
      setPaths(pathsData || [])
    } catch (err) {
      console.error('Failed to load progress data', err)
    } finally {
      setLoading(false)
    }
  }

  const handleClaimCertificate = async (pathId: string) => {
    try {
      setClaiming(pathId)
      const cert = await apiRequest<Certificate>('/certificates/claim', {
        method: 'POST',
        body: JSON.stringify({ learning_path_id: pathId }),
      })
      setCertificates((prev) => [cert, ...prev])
      setSelectedCert(cert)
      // Refresh stats and badges
      const [newStats, newBadges] = await Promise.all([
        apiRequest<UserSummaryStats>('/analytics/summary'),
        apiRequest<Badge[]>('/analytics/badges'),
      ])
      setStats(newStats)
      setBadges(newBadges)
    } catch (err: any) {
      alert('Unable to claim certificate: ' + (err?.message || 'Prerequisites unmet'))
    } finally {
      setClaiming(null)
    }
  }

  // Generate 90-day calendar matrix
  const activityMap = new Map<string, number>()
  activity.forEach((a) => activityMap.set(a.date, a.count))

  const today = new Date()
  const days: { dateStr: string; count: number }[] = []
  for (let i = 89; i >= 0; i--) {
    const d = new Date()
    d.setDate(today.getDate() - i)
    const dateStr = d.toISOString().split('T')[0]
    days.push({
      dateStr,
      count: activityMap.get(dateStr) || 0,
    })
  }

  const getActivityColor = (count: number) => {
    if (count === 0) return 'bg-neutral-900 border border-neutral-800'
    if (count <= 2) return 'bg-emerald-950 border border-emerald-800/80'
    if (count <= 4) return 'bg-emerald-700/80 border border-emerald-600'
    return 'bg-emerald-500 border border-emerald-400'
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="flex items-center gap-3 text-neutral-400 text-sm">
          <div className="h-5 w-5 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
          <span>Loading analytics & progress...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto space-y-8 pb-16">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-neutral-800 pb-6">
        <div>
          <div className="flex items-center gap-2 text-indigo-400 text-xs font-semibold tracking-wider uppercase mb-1">
            <Trophy className="h-3.5 w-3.5" />
            Learning Journey & Analytics
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Your Progress & Mastery</h1>
          <p className="text-sm text-neutral-400 mt-1 max-w-2xl">
            Track daily momentum, continuous learning streaks, verified skill badges, and official certificates.
          </p>
        </div>

        {/* Tab switcher */}
        <div className="flex items-center gap-1 bg-neutral-900/80 p-1 rounded-xl border border-neutral-800 text-xs font-medium self-start md:self-auto">
          <button
            onClick={() => setActiveTab('overview')}
            className={`px-3.5 py-1.5 rounded-lg transition cursor-pointer ${
              activeTab === 'overview'
                ? 'bg-neutral-800 text-white font-semibold shadow-sm'
                : 'text-neutral-400 hover:text-white'
            }`}
          >
            Overview
          </button>
          <button
            onClick={() => setActiveTab('badges')}
            className={`px-3.5 py-1.5 rounded-lg transition cursor-pointer ${
              activeTab === 'badges'
                ? 'bg-neutral-800 text-white font-semibold shadow-sm'
                : 'text-neutral-400 hover:text-white'
            }`}
          >
            Badges ({stats?.earned_badges_count || 0}/{stats?.total_badges_count || 5})
          </button>
          <button
            onClick={() => setActiveTab('certificates')}
            className={`px-3.5 py-1.5 rounded-lg transition cursor-pointer ${
              activeTab === 'certificates'
                ? 'bg-neutral-800 text-white font-semibold shadow-sm'
                : 'text-neutral-400 hover:text-white'
            }`}
          >
            Certificates ({certificates.length})
          </button>
        </div>
      </div>

      {/* Top Metric Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        {/* Streak Card */}
        <div className="col-span-2 sm:col-span-1 lg:col-span-2 p-5 rounded-xl bg-gradient-to-br from-amber-950/20 to-neutral-900/60 border border-amber-900/40 relative overflow-hidden flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-amber-400 uppercase tracking-wider">Active Streak</span>
            <div className="h-8 w-8 rounded-lg bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
              <Flame className="h-4 w-4 fill-amber-400 text-amber-400 animate-pulse" />
            </div>
          </div>
          <div className="my-2">
            <div className="text-3xl font-extrabold text-white flex items-baseline gap-1.5">
              <span>{stats?.streak.current_streak || 0}</span>
              <span className="text-sm font-medium text-neutral-400">days</span>
            </div>
            <div className="text-[11px] text-amber-300/80 mt-1">
              Best record: {stats?.streak.longest_streak || 0} consecutive days
            </div>
          </div>
        </div>

        {/* Lessons Completed */}
        <div className="p-4 rounded-xl bg-neutral-900/40 border border-neutral-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-neutral-400 text-xs">
            <span>Lessons</span>
            <BookOpen className="h-4 w-4 text-indigo-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">
            {stats?.completed_lessons || 0}
          </div>
          <span className="text-[10px] text-neutral-500">Completed</span>
        </div>

        {/* Challenges Solved */}
        <div className="p-4 rounded-xl bg-neutral-900/40 border border-neutral-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-neutral-400 text-xs">
            <span>Challenges</span>
            <Code2 className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">
            {stats?.solved_challenges || 0}
          </div>
          <span className="text-[10px] text-neutral-500">Solved</span>
        </div>

        {/* Study Notes */}
        <div className="p-4 rounded-xl bg-neutral-900/40 border border-neutral-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-neutral-400 text-xs">
            <span>Notes</span>
            <FileText className="h-4 w-4 text-sky-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">
            {stats?.total_notes || 0}
          </div>
          <span className="text-[10px] text-neutral-500">Recorded</span>
        </div>

        {/* Certificates */}
        <div className="p-4 rounded-xl bg-neutral-900/40 border border-neutral-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-neutral-400 text-xs">
            <span>Certificates</span>
            <Award className="h-4 w-4 text-purple-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">
            {stats?.total_certificates || 0}
          </div>
          <span className="text-[10px] text-neutral-500">Verified</span>
        </div>
      </div>

      {activeTab === 'overview' && (
        <div className="space-y-8">
          {/* Activity Heatmap Grid */}
          <div className="p-6 rounded-2xl bg-neutral-900/40 border border-neutral-800 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-indigo-400" />
                  90-Day Learning Activity
                </h2>
                <p className="text-xs text-neutral-400 mt-0.5">
                  Visual record of completed lessons, verified challenge runs, and study notes.
                </p>
              </div>

              <div className="flex items-center gap-2 text-[10px] text-neutral-400">
                <span>Less</span>
                <span className="h-2.5 w-2.5 rounded-sm bg-neutral-900 border border-neutral-800" />
                <span className="h-2.5 w-2.5 rounded-sm bg-emerald-950 border border-emerald-800" />
                <span className="h-2.5 w-2.5 rounded-sm bg-emerald-700" />
                <span className="h-2.5 w-2.5 rounded-sm bg-emerald-500" />
                <span>More</span>
              </div>
            </div>

            {/* Grid */}
            <div className="pt-2 overflow-x-auto">
              <div className="flex gap-1.5 min-w-[640px]">
                {Array.from({ length: 13 }, (_, weekIdx) => (
                  <div key={weekIdx} className="flex flex-col gap-1.5 flex-1">
                    {Array.from({ length: 7 }, (_, dayIdx) => {
                      const idx = weekIdx * 7 + dayIdx
                      if (idx >= days.length) return null
                      const d = days[idx]
                      return (
                        <div
                          key={d.dateStr}
                          title={`${d.dateStr}: ${d.count} activities`}
                          className={`h-3 rounded-sm transition cursor-pointer hover:ring-1 hover:ring-white ${getActivityColor(
                            d.count,
                          )}`}
                        />
                      )
                    })}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Path Completion & Certificates Summary */}
          <div className="p-6 rounded-2xl bg-neutral-900/40 border border-neutral-800 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
                  <ShieldCheck className="h-4 w-4 text-emerald-400" />
                  Learning Path Certificates
                </h2>
                <p className="text-xs text-neutral-400 mt-0.5">
                  Complete 100% of all curriculum lessons in a path to unlock verified credentials.
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {paths.map((path) => {
                const isClaimed = certificates.some((c) => c.learning_path_id === path.id)
                const isComplete = (path.progress_percent || 0) >= 100

                return (
                  <div
                    key={path.id}
                    className="p-5 rounded-xl bg-neutral-950/60 border border-neutral-800 flex flex-col justify-between space-y-4"
                  >
                    <div>
                      <div className="flex items-center justify-between text-xs mb-1.5">
                        <span className="font-semibold text-white truncate">{path.title}</span>
                        <span className="font-mono text-neutral-400">
                          {path.completed_lessons || 0}/{path.total_lessons || 0}
                        </span>
                      </div>
                      <div className="h-1.5 rounded-full bg-neutral-800 overflow-hidden">
                        <div
                          className="h-full bg-indigo-500 rounded-full transition-all duration-500"
                          style={{ width: `${path.progress_percent || 0}%` }}
                        />
                      </div>
                    </div>

                    <div className="flex items-center justify-between pt-2 border-t border-neutral-800/80">
                      {isClaimed ? (
                        <span className="inline-flex items-center gap-1.5 text-xs text-emerald-400 font-semibold">
                          <CheckCircle2 className="h-4 w-4" />
                          Certificate Claimed
                        </span>
                      ) : isComplete ? (
                        <button
                          onClick={() => handleClaimCertificate(path.id)}
                          disabled={claiming === path.id}
                          className="px-3.5 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-md shadow-emerald-600/30 transition cursor-pointer"
                        >
                          {claiming === path.id ? 'Claiming...' : 'Claim Certificate'}
                        </button>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 text-xs text-neutral-500">
                          <Lock className="h-3.5 w-3.5" />
                          Complete all lessons to unlock
                        </span>
                      )}

                      <Link
                        to={`/learning-paths/${path.slug}`}
                        className="text-xs text-indigo-400 hover:text-indigo-300 font-medium"
                      >
                        Go to Path →
                      </Link>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'badges' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {badges.map((badge) => (
              <div
                key={badge.key}
                className={`p-5 rounded-xl border flex flex-col justify-between transition ${
                  badge.awarded
                    ? 'bg-neutral-900/60 border-indigo-500/40 shadow-sm shadow-indigo-500/10'
                    : 'bg-neutral-900/20 border-neutral-800 opacity-60'
                }`}
              >
                <div className="flex items-start gap-4">
                  <div
                    className={`h-12 w-12 rounded-xl flex items-center justify-center shrink-0 ${
                      badge.awarded
                        ? 'bg-indigo-600/20 border border-indigo-500/30 text-indigo-400'
                        : 'bg-neutral-800 text-neutral-500 border border-neutral-700'
                    }`}
                  >
                    {badge.awarded ? (
                      <Sparkles className="h-6 w-6 text-amber-400" />
                    ) : (
                      <Lock className="h-5 w-5" />
                    )}
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-bold text-white">{badge.name}</span>
                      <span className="text-[10px] px-2 py-0.5 rounded-full bg-neutral-800 text-neutral-400 uppercase font-semibold">
                        {badge.category}
                      </span>
                    </div>
                    <p className="text-xs text-neutral-400 leading-relaxed">{badge.description}</p>
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-neutral-800/60 text-[11px] text-neutral-500">
                  {badge.awarded ? (
                    <span className="text-emerald-400 font-medium flex items-center gap-1.5">
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Unlocked on {new Date(badge.awarded_at!).toLocaleDateString()}
                    </span>
                  ) : (
                    <span>Locked</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'certificates' && (
        <div className="space-y-6">
          {certificates.length === 0 ? (
            <div className="py-16 text-center space-y-3 bg-neutral-900/30 rounded-2xl border border-neutral-800">
              <div className="h-12 w-12 rounded-xl bg-neutral-800 text-neutral-500 mx-auto flex items-center justify-center">
                <Award className="h-6 w-6" />
              </div>
              <h3 className="text-base font-semibold text-white">No certificates yet</h3>
              <p className="text-xs text-neutral-400 max-w-sm mx-auto">
                Finish all lessons in any Learning Path to earn an officially verifiable credential.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {certificates.map((cert) => (
                <div
                  key={cert.id}
                  className="p-6 rounded-2xl bg-gradient-to-br from-neutral-900 via-neutral-900/90 to-indigo-950/30 border border-neutral-800 hover:border-indigo-500/40 transition space-y-4 shadow-lg shadow-black/40"
                >
                  <div className="flex items-center justify-between">
                    <div className="h-10 w-10 rounded-xl bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
                      <Award className="h-5 w-5" />
                    </div>
                    <span className="font-mono text-xs text-neutral-400">{cert.certificate_code}</span>
                  </div>

                  <div>
                    <h3 className="text-lg font-bold text-white">{cert.learning_path_title}</h3>
                    <p className="text-xs text-neutral-400 mt-1">
                      Awarded to <span className="text-neutral-200 font-semibold">{cert.recipient_name}</span>
                    </p>
                    <p className="text-[11px] text-neutral-500 mt-0.5">
                      Issued on {new Date(cert.issued_at).toLocaleDateString()}
                    </p>
                  </div>

                  <div className="pt-4 border-t border-neutral-800 flex items-center justify-between">
                    <button
                      onClick={() => setSelectedCert(cert)}
                      className="text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition cursor-pointer"
                    >
                      View Certificate Card
                    </button>

                    <Link
                      to={`/certificates/verify/${cert.certificate_code}`}
                      target="_blank"
                      className="inline-flex items-center gap-1 text-xs text-neutral-400 hover:text-white transition"
                    >
                      <span>Public Verification</span>
                      <ExternalLink className="h-3 w-3" />
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Certificate Modal */}
      {selectedCert && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
          <div className="bg-neutral-950 border-2 border-amber-500/40 rounded-2xl p-8 max-w-xl w-full space-y-6 shadow-2xl relative text-center">
            <button
              onClick={() => setSelectedCert(null)}
              className="absolute top-4 right-4 text-neutral-400 hover:text-white text-lg font-bold cursor-pointer"
            >
              ×
            </button>

            <div className="space-y-2">
              <div className="flex justify-center mb-2">
                <div className="h-14 w-14 rounded-2xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
                  <Award className="h-8 w-8" />
                </div>
              </div>
              <div className="text-xs uppercase tracking-widest text-amber-400 font-bold">
                Certificate of Mastery
              </div>
              <h2 className="text-2xl font-black text-white tracking-tight">Go Academy</h2>
              <p className="text-xs text-neutral-400">This officially certifies that</p>
            </div>

            <div className="py-2 border-y border-neutral-800">
              <div className="text-xl font-bold text-amber-300">{selectedCert.recipient_name}</div>
              <p className="text-xs text-neutral-400 mt-1">has successfully mastered all curriculum requirements for</p>
              <div className="text-base font-semibold text-white mt-1">
                {selectedCert.learning_path_title}
              </div>
            </div>

            <div className="flex items-center justify-between text-xs text-neutral-400 pt-2 font-mono">
              <div>Issued: {new Date(selectedCert.issued_at).toLocaleDateString()}</div>
              <div>Code: {selectedCert.certificate_code}</div>
            </div>

            <div className="pt-2">
              <Link
                to={`/certificates/verify/${selectedCert.certificate_code}`}
                target="_blank"
                className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-neutral-900 border border-neutral-800 hover:border-neutral-700 text-neutral-200 text-xs font-semibold transition"
              >
                <span>Open Public Verification Page</span>
                <ExternalLink className="h-3.5 w-3.5" />
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
