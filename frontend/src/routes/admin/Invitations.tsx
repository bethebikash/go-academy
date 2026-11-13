import React, { useState, useEffect, useCallback } from 'react'
import {
  Mail,
  Plus,
  Copy,
  Check,
  Trash2,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { Invitation } from '../../types'
import { apiRequest, ApiError } from '../../lib/api'

export default function AdminInvitations() {
  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  // Creation modal state
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false)
  const [newEmail, setNewEmail] = useState<string>('')
  const [expiresInDays, setExpiresInDays] = useState<number>(7)
  const [submitting, setSubmitting] = useState<boolean>(false)
  const [modalError, setModalError] = useState<string | null>(null)
  const [createdInvite, setCreatedInvite] = useState<Invitation | null>(null)
  const [copied, setCopied] = useState<boolean>(false)

  const fetchInvitations = useCallback(async () => {
    try {
      setLoading(true)
      const data = await apiRequest<Invitation[]>('/admin/invitations?page=1&limit=50')
      setInvitations(data)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Failed to load invitations.')
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

  const handleCreateInvitation = async (e: React.FormEvent) => {
    e.preventDefault()
    setModalError(null)
    setSubmitting(true)
    try {
      const created = await apiRequest<Invitation>('/admin/invitations', {
        method: 'POST',
        body: JSON.stringify({
          email: newEmail,
          expires_in_days: expiresInDays,
        }),
      })
      setCreatedInvite(created)
      setNewEmail('')
      fetchInvitations()
    } catch (err) {
      if (err instanceof ApiError) {
        setModalError(err.message)
      } else {
        setModalError('Failed to generate invitation.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this invitation? The invitee will not be able to register.')) {
      return
    }
    try {
      await apiRequest(`/admin/invitations/${id}`, { method: 'DELETE' })
      fetchInvitations()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to revoke invitation')
    }
  }

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'accepted':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="h-3 w-3" />
            Accepted
          </span>
        )
      case 'revoked':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
            <XCircle className="h-3 w-3" />
            Revoked
          </span>
        )
      case 'expired':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-neutral-500/10 text-neutral-400 border border-neutral-500/20">
            <Clock className="h-3 w-3" />
            Expired
          </span>
        )
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Clock className="h-3 w-3" />
            Pending
          </span>
        )
    }
  }

  return (
    <div className="max-w-6xl space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-neutral-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Invitation Management</h1>
          <p className="text-sm text-neutral-400 mt-1">
            Generate and manage single-use invitation links for friend onboarding
          </p>
        </div>
        <button
          onClick={() => {
            setCreatedInvite(null)
            setModalError(null)
            setIsModalOpen(true)
          }}
          className="flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium py-2.5 px-4 shadow-lg shadow-indigo-600/20 transition cursor-pointer"
        >
          <Plus className="h-4 w-4" />
          Create Invitation
        </button>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="rounded-lg bg-red-950/40 border border-red-500/30 p-4 text-sm text-red-200 flex items-center gap-3">
          <AlertCircle className="h-5 w-5 text-red-400 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Invitations Table */}
      <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 overflow-hidden shadow-lg">
        {loading ? (
          <div className="py-16 flex flex-col items-center justify-center gap-3 text-neutral-400">
            <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
            <p className="text-sm">Loading invitations...</p>
          </div>
        ) : invitations.length === 0 ? (
          <div className="py-16 text-center space-y-3">
            <Mail className="h-10 w-10 text-neutral-600 mx-auto" />
            <h3 className="text-sm font-semibold text-neutral-300">No invitations created yet</h3>
            <p className="text-xs text-neutral-500 max-w-sm mx-auto">
              Create an invitation link to invite learners to join Go Academy.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-neutral-950/60 text-xs font-semibold uppercase tracking-wider text-neutral-400 border-b border-neutral-800">
                <tr>
                  <th className="px-6 py-4">Invited Email</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4">Expires</th>
                  <th className="px-6 py-4">Invited By</th>
                  <th className="px-6 py-4">Created</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-800/60">
                {invitations.map((inv) => (
                  <tr key={inv.id} className="hover:bg-neutral-800/30 transition">
                    <td className="px-6 py-4 font-medium text-white">{inv.email}</td>
                    <td className="px-6 py-4">{getStatusBadge(inv.status)}</td>
                    <td className="px-6 py-4 text-neutral-400 text-xs">
                      {new Date(inv.expires_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-neutral-400 text-xs">
                      {inv.invited_by_username || 'Admin'}
                    </td>
                    <td className="px-6 py-4 text-neutral-500 text-xs">
                      {new Date(inv.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {inv.status === 'pending' && (
                        <button
                          onClick={() => handleRevoke(inv.id)}
                          title="Revoke Invitation"
                          className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-neutral-800 rounded transition cursor-pointer"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="w-full max-w-md bg-neutral-900 border border-neutral-800 rounded-2xl p-6 shadow-2xl space-y-6">
            <div className="flex items-center justify-between border-b border-neutral-800 pb-3">
              <h3 className="text-lg font-bold text-white">Create New Invitation</h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-neutral-400 hover:text-white text-sm cursor-pointer"
              >
                ✕
              </button>
            </div>

            {createdInvite ? (
              <div className="space-y-4">
                <div className="rounded-lg bg-emerald-950/40 border border-emerald-500/30 p-4 text-xs text-emerald-300 space-y-1">
                  <p className="font-semibold text-emerald-200">Invitation Generated Successfully!</p>
                  <p>Share the link below with {createdInvite.email}.</p>
                </div>

                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-400 mb-2">
                    Registration Link
                  </label>
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      readOnly
                      value={createdInvite.invitation_url || ''}
                      className="w-full rounded-lg bg-neutral-950 border border-neutral-800 px-3 py-2 text-xs text-neutral-200 select-all font-mono"
                    />
                    <button
                      onClick={() => handleCopy(createdInvite.invitation_url || '')}
                      className="p-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white shrink-0 transition cursor-pointer"
                      title="Copy Link"
                    >
                      {copied ? <Check className="h-4 w-4 text-emerald-300" /> : <Copy className="h-4 w-4" />}
                    </button>
                  </div>
                </div>

                <div className="pt-2">
                  <button
                    onClick={() => setIsModalOpen(false)}
                    className="w-full rounded-lg bg-neutral-800 hover:bg-neutral-700 text-white text-sm font-medium py-2 transition cursor-pointer"
                  >
                    Done
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleCreateInvitation} className="space-y-4">
                {modalError && (
                  <div className="rounded-lg bg-red-950/40 border border-red-500/30 p-3 text-xs text-red-200 flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 text-red-400 shrink-0" />
                    <span>{modalError}</span>
                  </div>
                )}

                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-300 mb-1.5">
                    Invitee Email
                  </label>
                  <input
                    type="email"
                    required
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                    placeholder="friend@example.com"
                    className="w-full rounded-lg bg-neutral-950 border border-neutral-800 px-3.5 py-2 text-sm text-neutral-100 placeholder:text-neutral-600 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-300 mb-1.5">
                    Expiry Period
                  </label>
                  <select
                    value={expiresInDays}
                    onChange={(e) => setExpiresInDays(Number(e.target.value))}
                    className="w-full rounded-lg bg-neutral-950 border border-neutral-800 px-3 py-2 text-sm text-neutral-100 focus:outline-none focus:border-indigo-500"
                  >
                    <option value={1}>1 Day</option>
                    <option value={3}>3 Days</option>
                    <option value={7}>7 Days (Default)</option>
                    <option value={14}>14 Days</option>
                    <option value={30}>30 Days</option>
                  </select>
                </div>

                <div className="pt-3 flex items-center justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="rounded-lg px-4 py-2 text-sm text-neutral-400 hover:text-white transition cursor-pointer"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={submitting}
                    className="rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium py-2 px-4 shadow-lg shadow-indigo-600/20 transition disabled:opacity-50 cursor-pointer flex items-center gap-2"
                  >
                    {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
                    Create Link
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
