import { useCallback, useEffect, useState } from 'react'
import { Link2, Unlink } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { formatDate } from '@/utils/formatters'
import type { LinkedAccountStatus, MarketplaceProvider } from '@/types'

export default function ConnectedAccountsTab() {
  const { addToast } = useToast()
  const [accounts, setAccounts] = useState<LinkedAccountStatus[] | null>(null)
  const [busy, setBusy] = useState<MarketplaceProvider | null>(null)

  const load = useCallback(() => {
    apiClient
      .getLinkedAccounts()
      .then(setAccounts)
      .catch((err) => {
        addToast(err instanceof Error ? err.message : 'Could not load accounts', 'error')
        setAccounts([])
      })
  }, [addToast])

  useEffect(load, [load])

  const connect = async (provider: MarketplaceProvider) => {
    setBusy(provider)
    try {
      const { auth_url } = await apiClient.connectMarketplace(provider)
      const url = new URL(auth_url)
      if (url.protocol !== 'https:' && url.protocol !== 'http:')
        throw new Error('Unexpected authorization URL')
      window.location.assign(url.toString())
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not start linking', 'error')
      setBusy(null)
    }
  }

  const disconnect = async (provider: MarketplaceProvider, name: string) => {
    if (!window.confirm(`Disconnect your ${name} account?`)) return
    setBusy(provider)
    try {
      await apiClient.disconnectMarketplace(provider)
      addToast(`${name} disconnected`, 'success')
      load()
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not disconnect', 'error')
    } finally {
      setBusy(null)
    }
  }

  if (!accounts) return <p className='text-gray-400 py-8 text-center'>Loading accounts...</p>

  return (
    <div className='space-y-4'>
      {accounts.some((a) => a.mock) && (
        <div className='badge badge-warning'>
          Mock mode: linking and listings are simulated, no real marketplace calls are made.
        </div>
      )}
      <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
        {accounts.map((a) => (
          <div key={a.provider} className='glass-effect rounded-2xl p-6'>
            <div className='flex items-center justify-between mb-3'>
              <h3 className='text-lg font-semibold text-white'>{a.display_name}</h3>
              <span className={`badge ${a.linked ? 'badge-success' : 'bg-white/10 text-gray-400'}`}>
                {a.linked ? (a.account?.status === 'expired' ? 'Expired' : 'Linked') : 'Not linked'}
              </span>
            </div>
            {a.linked && a.account ? (
              <div className='text-sm text-gray-300 space-y-1 mb-4'>
                <div>
                  Account: <span className='text-white'>{a.account.external_username}</span>
                </div>
                <div>Linked {formatDate(a.account.linked_at)}</div>
                <div className='text-xs text-gray-500'>Scopes: {a.account.scopes.join(', ')}</div>
              </div>
            ) : (
              <p className='text-sm text-gray-400 mb-4'>
                Link your {a.display_name} seller account to create listings from your portfolio in
                one click.
              </p>
            )}
            {a.linked ? (
              <div className='flex gap-2'>
                {a.account?.status === 'expired' && (
                  <button
                    type='button'
                    disabled={busy === a.provider}
                    onClick={() => connect(a.provider)}
                    className='btn-primary py-2 px-4'
                  >
                    Relink
                  </button>
                )}
                <button
                  type='button'
                  disabled={busy === a.provider}
                  onClick={() => disconnect(a.provider, a.display_name)}
                  className='btn-ghost py-2 px-4 flex items-center'
                >
                  <Unlink className='w-4 h-4 mr-2' /> Disconnect
                </button>
              </div>
            ) : (
              <button
                type='button'
                disabled={busy === a.provider}
                onClick={() => connect(a.provider)}
                className='btn-primary py-2 px-4 flex items-center'
              >
                <Link2 className='w-4 h-4 mr-2' /> Connect {a.display_name}
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
