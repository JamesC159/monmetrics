import { useEffect, useState } from 'react'
import { ArrowLeft, Loader, ShieldAlert } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { FeatureState } from '../types'
import { apiClient, type ApiError } from '../utils/api'
import { useAuth } from '../context/AuthContext'
import { useFeatures } from '../context/FeaturesContext'
import { useToast } from '../context/ToastContext'

export default function Admin() {
  const { isAuthenticated } = useAuth()
  const { refresh } = useFeatures()
  const { addToast } = useToast()
  const [states, setStates] = useState<FeatureState[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState<string | null>(null)
  const [denied, setDenied] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    apiClient
      .getAdminFeatures()
      .then((res) => !cancelled && setStates(res))
      .catch((err) => {
        if (cancelled) return
        const status = (err as ApiError)?.status
        if (status === 401 || status === 403) setDenied(true)
        else setError(err instanceof Error ? err.message : 'Failed to load feature flags')
      })
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [])

  const toggle = async (state: FeatureState) => {
    setSaving(state.key)
    try {
      setStates(await apiClient.updateAdminFeature(state.key, !state.enabled))
      // Keep the rest of the app in sync without a reload.
      await refresh()
      addToast(`${state.label} ${!state.enabled ? 'enabled' : 'disabled'}`, 'success')
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not update flag', 'error')
    } finally {
      setSaving(null)
    }
  }

  return (
    <div className='min-h-screen bg-dark-900'>
      <div className='max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <Link
          to='/dashboard'
          className='inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4'
        >
          <ArrowLeft className='w-4 h-4 mr-2' />
          Back to dashboard
        </Link>

        <h1 className='text-4xl font-bold text-white mb-2'>Feature Flags</h1>
        <p className='text-gray-400 mb-8'>
          Surfaces stay off until a real data source backs them. Toggles take effect immediately.
        </p>

        {!isAuthenticated || denied ? (
          <div className='glass-effect rounded-2xl p-10 text-center'>
            <ShieldAlert className='w-10 h-10 text-gray-600 mx-auto mb-3' />
            <h2 className='text-xl font-semibold text-gray-300 mb-2'>Admin access required</h2>
            <p className='text-gray-500 text-sm'>
              Your account is not on the <code className='text-gray-400'>ADMIN_EMAILS</code>{' '}
              allowlist.
            </p>
          </div>
        ) : loading ? (
          <div className='flex items-center justify-center py-20 text-gray-400'>
            <Loader className='w-6 h-6 animate-spin mr-3' />
            Loading flags...
          </div>
        ) : error ? (
          <div className='glass-effect rounded-lg p-6 border-l-4 border-red-500'>
            <div className='text-red-400 font-medium'>Error</div>
            <div className='text-gray-300'>{error}</div>
          </div>
        ) : (
          <div className='space-y-4'>
            {states.map((state) => (
              <div key={state.key} className='glass-effect rounded-xl p-5'>
                <div className='flex items-start justify-between gap-6'>
                  <div className='flex-1'>
                    <div className='flex items-center gap-3 mb-1'>
                      <h2 className='text-lg font-semibold text-white'>{state.label}</h2>
                      <code className='text-xs text-gray-500'>{state.key}</code>
                    </div>
                    <p className='text-sm text-gray-400'>{state.description}</p>

                    {state.blocker && !state.enabled && (
                      <p className='text-xs text-amber-400/80 mt-2'>Blocked: {state.blocker}</p>
                    )}

                    {state.updated_at && (
                      <p className='text-xs text-gray-600 mt-2'>
                        Changed {new Date(state.updated_at).toLocaleString()}
                        {state.updated_by && ` by ${state.updated_by}`}
                      </p>
                    )}
                  </div>

                  <button
                    type='button'
                    role='switch'
                    aria-checked={state.enabled}
                    aria-label={`${state.enabled ? 'Disable' : 'Enable'} ${state.label}`}
                    disabled={saving === state.key}
                    onClick={() => toggle(state)}
                    className={`shrink-0 w-14 h-8 rounded-full transition-colors relative disabled:opacity-50 ${
                      state.enabled ? 'bg-primary-500' : 'bg-white/15'
                    }`}
                  >
                    <span
                      className={`absolute top-1 w-6 h-6 rounded-full bg-white transition-transform ${
                        state.enabled ? 'translate-x-7' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
