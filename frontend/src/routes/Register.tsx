import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Terminal, Lock, User as UserIcon, Mail, AlertCircle, ArrowRight, Loader2, CheckCircle2 } from 'lucide-react'
import { useAuth } from '../features/auth/AuthContext'
import { apiRequest, ApiError } from '../lib/api'

const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, 'Username must be at least 3 characters')
      .max(30, 'Username cannot exceed 30 characters')
      .regex(/^[a-zA-Z0-9_-]+$/, 'Username can only contain letters, numbers, hyphens, and underscores'),
    password: z
      .string()
      .min(10, 'Password must be at least 10 characters')
      .refine(
        (val) => /[a-zA-Z]/.test(val) && /[^a-zA-Z]/.test(val),
        'Password must contain both letters and numbers/symbols',
      ),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

type RegisterFormValues = z.infer<typeof registerSchema>

interface TokenValidationResult {
  email: string
  expires_at: string
  valid: boolean
}

export default function Register() {
  const { register: registerUser } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')

  const [validating, setValidating] = useState<boolean>(true)
  const [invitedEmail, setInvitedEmail] = useState<string | null>(null)
  const [tokenError, setTokenError] = useState<string | null>(null)
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: '',
      password: '',
      confirmPassword: '',
    },
  })

  useEffect(() => {
    if (!token) {
      setTokenError('No invitation token was provided in the URL. Go Academy registration is strictly invite-only.')
      setValidating(false)
      return
    }

    async function validateToken() {
      try {
        const data = await apiRequest<TokenValidationResult>(`/auth/invitations/validate?token=${token}`)
        setInvitedEmail(data.email)
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.code === 'INVITATION_EXPIRED') {
            setTokenError('This invitation link has expired. Please request a new invitation from an administrator.')
          } else if (err.code === 'INVITATION_REVOKED') {
            setTokenError('This invitation has been revoked by an administrator.')
          } else if (err.code === 'INVITATION_ALREADY_USED') {
            setTokenError('This invitation has already been accepted. If this is your account, please sign in.')
          } else {
            setTokenError(err.message || 'The provided invitation link is invalid.')
          }
        } else {
          setTokenError('Failed to validate invitation token. Please check your network connection.')
        }
      } finally {
        setValidating(false)
      }
    }

    validateToken()
  }, [token])

  const onSubmit = async (values: RegisterFormValues) => {
    if (!token) return
    setServerError(null)
    try {
      await registerUser(token, values.username, values.password)
      navigate('/dashboard', { replace: true })
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message)
      } else {
        setServerError('An unexpected error occurred during registration. Please try again.')
      }
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-950 text-neutral-100 p-4">
      <div className="w-full max-w-md space-y-8 bg-neutral-900/60 p-8 rounded-2xl border border-neutral-800 shadow-2xl backdrop-blur-sm">
        {/* Logo and Brand */}
        <div className="text-center space-y-2">
          <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-600 text-white shadow-lg shadow-indigo-600/30">
            <Terminal className="h-6 w-6" />
          </div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Join Go Academy</h2>
          <p className="text-sm text-neutral-400">Complete your registration using your invitation</p>
        </div>

        {validating ? (
          <div className="py-12 flex flex-col items-center justify-center gap-3 text-neutral-400">
            <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
            <p className="text-sm">Verifying invitation link...</p>
          </div>
        ) : tokenError ? (
          <div className="space-y-6">
            <div className="rounded-lg bg-red-950/40 border border-red-500/30 p-5 text-sm text-red-200 flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-red-400 shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="font-semibold text-red-300">Invitation Unavailable</p>
                <p className="text-xs leading-relaxed text-red-300/80">{tokenError}</p>
              </div>
            </div>

            <div className="text-center">
              <Link
                to="/login"
                className="inline-flex items-center gap-2 text-sm text-indigo-400 hover:text-indigo-300 font-medium transition"
              >
                Return to Sign In
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>
        ) : (
          <>
            {/* Valid Invitation Badge */}
            <div className="rounded-lg bg-emerald-950/40 border border-emerald-500/30 p-3 text-xs text-emerald-300 flex items-center gap-2.5">
              <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0" />
              <span>
                Verified invitation for <strong className="font-semibold text-white">{invitedEmail}</strong>
              </span>
            </div>

            {/* Server Error Alert */}
            {serverError && (
              <div className="rounded-lg bg-red-950/50 border border-red-500/30 p-4 text-sm text-red-200 flex items-start gap-3">
                <AlertCircle className="h-5 w-5 text-red-400 shrink-0 mt-0.5" />
                <span>{serverError}</span>
              </div>
            )}

            {/* Registration Form */}
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-400 mb-2">
                  Invited Email
                </label>
                <div className="relative">
                  <Mail className="absolute left-3.5 top-3 h-4 w-4 text-neutral-600" />
                  <input
                    type="email"
                    value={invitedEmail || ''}
                    disabled
                    className="w-full rounded-lg bg-neutral-950/50 border border-neutral-800/80 pl-10 pr-4 py-2.5 text-sm text-neutral-400 cursor-not-allowed select-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-300 mb-2">
                  Choose Username
                </label>
                <div className="relative">
                  <UserIcon className="absolute left-3.5 top-3 h-4 w-4 text-neutral-500" />
                  <input
                    type="text"
                    {...register('username')}
                    placeholder="gopher_dev"
                    className={`w-full rounded-lg bg-neutral-950 border pl-10 pr-4 py-2.5 text-sm text-neutral-100 placeholder:text-neutral-600 focus:outline-none focus:ring-2 transition ${
                      errors.username
                        ? 'border-red-500/50 focus:ring-red-500/30'
                        : 'border-neutral-800 focus:border-indigo-500 focus:ring-indigo-500/20'
                    }`}
                  />
                </div>
                {errors.username && <p className="mt-1.5 text-xs text-red-400">{errors.username.message}</p>}
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-300 mb-2">
                  Set Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-3.5 top-3 h-4 w-4 text-neutral-500" />
                  <input
                    type="password"
                    {...register('password')}
                    placeholder="Minimum 10 characters..."
                    className={`w-full rounded-lg bg-neutral-950 border pl-10 pr-4 py-2.5 text-sm text-neutral-100 placeholder:text-neutral-600 focus:outline-none focus:ring-2 transition ${
                      errors.password
                        ? 'border-red-500/50 focus:ring-red-500/30'
                        : 'border-neutral-800 focus:border-indigo-500 focus:ring-indigo-500/20'
                    }`}
                  />
                </div>
                {errors.password && <p className="mt-1.5 text-xs text-red-400">{errors.password.message}</p>}
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-300 mb-2">
                  Confirm Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-3.5 top-3 h-4 w-4 text-neutral-500" />
                  <input
                    type="password"
                    {...register('confirmPassword')}
                    placeholder="Re-enter password..."
                    className={`w-full rounded-lg bg-neutral-950 border pl-10 pr-4 py-2.5 text-sm text-neutral-100 placeholder:text-neutral-600 focus:outline-none focus:ring-2 transition ${
                      errors.confirmPassword
                        ? 'border-red-500/50 focus:ring-red-500/30'
                        : 'border-neutral-800 focus:border-indigo-500 focus:ring-indigo-500/20'
                    }`}
                  />
                </div>
                {errors.confirmPassword && (
                  <p className="mt-1.5 text-xs text-red-400">{errors.confirmPassword.message}</p>
                )}
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-2.5 px-4 text-sm flex items-center justify-center gap-2 shadow-lg shadow-indigo-600/20 transition disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Creating Account...
                  </>
                ) : (
                  <>
                    Complete Registration
                    <ArrowRight className="h-4 w-4" />
                  </>
                )}
              </button>
            </form>

            <div className="text-center pt-2">
              <span className="text-xs text-neutral-500">Already registered? </span>
              <Link to="/login" className="text-xs text-indigo-400 hover:text-indigo-300 font-medium">
                Sign in here
              </Link>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
