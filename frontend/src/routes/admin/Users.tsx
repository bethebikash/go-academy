import { useState, useEffect, useCallback } from 'react'
import {
  Users as UsersIcon,
  Shield,
  UserCheck,
  UserX,
  AlertCircle,
  Loader2,
  CheckCircle2,
} from 'lucide-react'
import { User } from '../../types'
import { apiRequest, ApiError } from '../../lib/api'
import { useAuth } from '../../features/auth/AuthContext'

export default function AdminUsers() {
  const { user: currentUser } = useAuth()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true)
      const data = await apiRequest<User[]>('/admin/users?page=1&limit=50')
      setUsers(data)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Failed to load user accounts.')
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const handleToggleStatus = async (user: User) => {
    if (user.id === currentUser?.id) {
      alert('You cannot deactivate your own administrative account.')
      return
    }

    const newStatus = user.status === 'active' ? 'disabled' : 'active'
    const confirmMsg =
      newStatus === 'disabled'
        ? `Are you sure you want to deactivate ${user.username}? They will no longer be able to log in.`
        : `Reactivate account for ${user.username}?`

    if (!confirm(confirmMsg)) return

    setActionLoading(user.id)
    try {
      await apiRequest(`/admin/users/${user.id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status: newStatus }),
      })
      fetchUsers()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update user status')
    } finally {
      setActionLoading(null)
    }
  }

  return (
    <div className="max-w-6xl space-y-6">
      {/* Header */}
      <div className="border-b border-neutral-800 pb-5">
        <h1 className="text-2xl font-bold text-white tracking-tight">User Management</h1>
        <p className="text-sm text-neutral-400 mt-1">
          Inspect registered platform users, roles, and manage account activation states
        </p>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="rounded-lg bg-red-950/40 border border-red-500/30 p-4 text-sm text-red-200 flex items-center gap-3">
          <AlertCircle className="h-5 w-5 text-red-400 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Users Table */}
      <div className="rounded-xl border border-neutral-800 bg-neutral-900/30 overflow-hidden shadow-lg">
        {loading ? (
          <div className="py-16 flex flex-col items-center justify-center gap-3 text-neutral-400">
            <Loader2 className="h-6 w-6 animate-spin text-indigo-500" />
            <p className="text-sm">Loading user accounts...</p>
          </div>
        ) : users.length === 0 ? (
          <div className="py-16 text-center space-y-3">
            <UsersIcon className="h-10 w-10 text-neutral-600 mx-auto" />
            <h3 className="text-sm font-semibold text-neutral-300">No users found</h3>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-neutral-950/60 text-xs font-semibold uppercase tracking-wider text-neutral-400 border-b border-neutral-800">
                <tr>
                  <th className="px-6 py-4">User</th>
                  <th className="px-6 py-4">Role</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4">Joined</th>
                  <th className="px-6 py-4">Last Login</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-800/60">
                {users.map((u) => (
                  <tr key={u.id} className="hover:bg-neutral-800/30 transition">
                    <td className="px-6 py-4">
                      <div className="font-medium text-white">{u.username}</div>
                      <div className="text-xs text-neutral-500">{u.email}</div>
                    </td>
                    <td className="px-6 py-4">
                      {u.role === 'admin' ? (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          <Shield className="h-3 w-3" />
                          Admin
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-neutral-800 text-neutral-300 border border-neutral-700">
                          Learner
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      {u.status === 'active' ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium text-emerald-400 bg-emerald-500/10">
                          <CheckCircle2 className="h-3 w-3" />
                          Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium text-neutral-400 bg-neutral-800">
                          Disabled
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-neutral-400 text-xs">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-neutral-500 text-xs">
                      {u.last_login_at ? new Date(u.last_login_at).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {u.id !== currentUser?.id && (
                        <button
                          disabled={actionLoading === u.id}
                          onClick={() => handleToggleStatus(u)}
                          className={`inline-flex items-center gap-1.5 text-xs font-medium py-1.5 px-3 rounded-lg border transition cursor-pointer ${
                            u.status === 'active'
                              ? 'text-red-400 border-red-500/20 hover:bg-red-950/30'
                              : 'text-emerald-400 border-emerald-500/20 hover:bg-emerald-950/30'
                          }`}
                        >
                          {actionLoading === u.id ? (
                            <Loader2 className="h-3 w-3 animate-spin" />
                          ) : u.status === 'active' ? (
                            <>
                              <UserX className="h-3.5 w-3.5" />
                              Deactivate
                            </>
                          ) : (
                            <>
                              <UserCheck className="h-3.5 w-3.5" />
                              Activate
                            </>
                          )}
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
    </div>
  )
}
