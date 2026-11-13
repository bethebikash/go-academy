import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Award, AlertCircle, ArrowLeft, Terminal, ShieldCheck } from 'lucide-react'
import { apiRequest } from '../lib/api'
import { Certificate } from '../types'

export default function VerifyCertificate() {
  const { code } = useParams<{ code: string }>()
  const [cert, setCert] = useState<Certificate | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (code) {
      verifyCode(code)
    }
  }, [code])

  const verifyCode = async (certCode: string) => {
    try {
      setLoading(true)
      setError(null)
      const data = await apiRequest<Certificate>(`/certificates/verify/${certCode}`)
      setCert(data)
    } catch (err: any) {
      setError(err?.message || 'Certificate not found or invalid')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-100 flex flex-col justify-between p-6">
      <header className="max-w-4xl mx-auto w-full flex items-center justify-between pb-8 border-b border-neutral-800">
        <Link to="/" className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-600 text-white font-bold">
            <Terminal className="h-5 w-5" />
          </div>
          <div>
            <span className="font-semibold tracking-tight text-white block">Go Academy</span>
            <span className="text-xs text-neutral-500">Credential Verification</span>
          </div>
        </Link>

        <Link
          to="/"
          className="inline-flex items-center gap-1.5 text-xs text-neutral-400 hover:text-white transition"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Go to Platform
        </Link>
      </header>

      <main className="max-w-xl mx-auto w-full my-auto py-12">
        {loading ? (
          <div className="text-center py-16 space-y-3">
            <div className="h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin mx-auto" />
            <p className="text-xs text-neutral-400">Verifying credential against registry...</p>
          </div>
        ) : error || !cert ? (
          <div className="p-8 rounded-2xl bg-neutral-900/60 border border-red-900/50 text-center space-y-4">
            <div className="h-12 w-12 rounded-xl bg-red-950/40 border border-red-800 text-red-400 mx-auto flex items-center justify-center">
              <AlertCircle className="h-6 w-6" />
            </div>
            <h2 className="text-lg font-bold text-white">Invalid or Unrecognized Certificate</h2>
            <p className="text-xs text-neutral-400 max-w-sm mx-auto">
              No matching record was found for certificate code <span className="font-mono text-red-300">{code}</span>.
            </p>
          </div>
        ) : (
          <div className="p-8 rounded-2xl bg-neutral-900/40 border-2 border-amber-500/40 shadow-2xl space-y-6 text-center relative overflow-hidden">
            <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-amber-500 via-indigo-500 to-emerald-500" />

            <div className="space-y-2">
              <div className="h-16 w-16 rounded-2xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 mx-auto">
                <Award className="h-9 w-9" />
              </div>
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold">
                <ShieldCheck className="h-3.5 w-3.5" />
                <span>Verified Official Credential</span>
              </div>
              <h1 className="text-2xl font-black text-white tracking-tight mt-2">Certificate of Mastery</h1>
              <p className="text-xs text-neutral-400">Issued by Go Academy Interactive Platform</p>
            </div>

            <div className="py-4 border-y border-neutral-800 space-y-2">
              <div className="text-xs text-neutral-400">Awarded to</div>
              <div className="text-2xl font-bold text-amber-300">{cert.recipient_name}</div>
              <div className="text-xs text-neutral-400">for successfully mastering</div>
              <div className="text-lg font-semibold text-white">{cert.learning_path_title}</div>
            </div>

            <div className="grid grid-cols-2 gap-4 text-left font-mono text-xs bg-neutral-950 p-4 rounded-xl border border-neutral-800">
              <div>
                <span className="text-[10px] text-neutral-500 uppercase font-bold block">Issue Date</span>
                <span className="text-neutral-300">{new Date(cert.issued_at).toLocaleDateString()}</span>
              </div>
              <div>
                <span className="text-[10px] text-neutral-500 uppercase font-bold block">Credential Code</span>
                <span className="text-indigo-400 font-semibold">{cert.certificate_code}</span>
              </div>
            </div>
          </div>
        )}
      </main>

      <footer className="text-center text-xs text-neutral-500 py-6">
        © {new Date().getFullYear()} Go Academy. All credentials cryptographically verified.
      </footer>
    </div>
  )
}
