import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import { apiClient } from '@/utils/api'
import type { MarketplaceProvider } from '@/types'

const PROVIDERS: MarketplaceProvider[] = ['ebay', 'tcgplayer']

export default function MarketplaceCallback() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading } = useAuth()
  const { addToast } = useToast()
  const [error, setError] = useState<string | null>(null)
  const started = useRef(false)

  useEffect(() => {
    if (isLoading || started.current) return
    if (!isAuthenticated) {
      setError('Please sign in again to finish linking your account.')
      return
    }
    const provider = params.get('provider') as MarketplaceProvider | null
    const code = params.get('code')
    const state = params.get('state')
    if (params.get('error')) {
      setError('Authorization was cancelled.')
      return
    }
    if (!provider || !PROVIDERS.includes(provider) || !code || !state) {
      setError('Invalid authorization response.')
      return
    }
    started.current = true
    apiClient
      .completeMarketplaceLink(provider, code, state)
      .then((res) => {
        addToast(
          `Linked ${provider === 'ebay' ? 'eBay' : 'TCGPlayer'} account ${res.username}`,
          'success',
        )
        navigate('/dashboard?tab=accounts', { replace: true })
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not link account'))
  }, [isLoading, isAuthenticated, params, navigate, addToast])

  return (
    <div className='min-h-screen bg-dark-900 flex items-center justify-center'>
      <div className='glass-effect rounded-2xl p-8 text-center max-w-md'>
        {error ? (
          <>
            <h2 className='text-xl font-semibold text-white mb-2'>Account linking failed</h2>
            <p className='text-gray-400 mb-6'>{error}</p>
            <Link to='/dashboard?tab=accounts' className='btn-primary'>
              Back to connected accounts
            </Link>
          </>
        ) : (
          <>
            <div className='loading-spinner mx-auto mb-4' />
            <p className='text-gray-300'>Finishing account link...</p>
          </>
        )}
      </div>
    </div>
  )
}
