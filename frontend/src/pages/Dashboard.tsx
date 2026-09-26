import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { Navigate, Link, useSearchParams } from 'react-router-dom'
import { BarChart3, Briefcase, Crown, Star } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { formatPercentage, formatPrice } from '@/utils/formatters'
import { gainClass, signed } from '@/utils/portfolio'
import PortfolioTab from '@/components/dashboard/PortfolioTab'
import WatchlistTab from '@/components/dashboard/WatchlistTab'
import SavedChartsTab from '@/components/dashboard/SavedChartsTab'
import ConnectedAccountsTab from '@/components/dashboard/ConnectedAccountsTab'
import AlertsTab from '@/components/dashboard/AlertsTab'
import type { Dashboard as DashboardData } from '@/types'

const TABS = [
  { id: 'portfolio', label: 'Portfolio' },
  { id: 'watchlist', label: 'Watchlist' },
  { id: 'charts', label: 'Saved Charts' },
  { id: 'alerts', label: 'Alerts' },
  { id: 'accounts', label: 'Connected Accounts' },
] as const

type TabId = (typeof TABS)[number]['id']

export default function Dashboard() {
  const { isAuthenticated, isLoading, user, logout } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const tabParam = searchParams.get('tab')
  const tab: TabId = TABS.some((t) => t.id === tabParam) ? (tabParam as TabId) : 'portfolio'
  const [data, setData] = useState<DashboardData | null>(null)

  const refresh = useCallback(() => {
    apiClient
      .getDashboard()
      .then(setData)
      .catch((err) => console.error('Failed to load dashboard:', err))
  }, [])

  useEffect(() => {
    if (isAuthenticated) refresh()
  }, [isAuthenticated, refresh])

  if (isLoading) {
    return (
      <div className='min-h-screen bg-dark-950 flex items-center justify-center'>
        <div className='loading-spinner' />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to='/login' replace />
  }

  const summary = data?.portfolio_summary
  const displayName = data?.user?.first_name || user?.first_name

  return (
    <div className='min-h-screen bg-dark-900'>
      <nav className='border-b border-white/10 bg-black/20 backdrop-blur-md'>
        <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
          <div className='flex justify-between items-center h-16'>
            <Link to='/' className='text-2xl font-bold text-primary-200'>
              MonMetrics
            </Link>

            <div className='flex items-center space-x-4'>
              <Link to='/search' className='nav-link'>
                Search
              </Link>
              {displayName && <span className='text-gray-300'>Welcome, {displayName}</span>}
              <button onClick={logout} className='btn-ghost py-2 px-4'>
                Sign Out
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <div className='mb-8'>
          <h1 className='text-4xl font-bold text-white mb-2'>Dashboard</h1>
          <p className='text-gray-400'>Track your collection, watchlist, and saved charts</p>
        </div>

        <div className='grid grid-cols-2 md:grid-cols-4 gap-4 mb-8'>
          <StatCard
            icon={<Briefcase className='w-7 h-7 text-primary-400' />}
            label='Portfolio'
            value={summary ? formatPrice(summary.total_value) : '—'}
            sub={
              summary ? (
                <span className={gainClass(summary.gain_loss)}>
                  {signed(formatPercentage(summary.gain_loss_pct), summary.gain_loss_pct)} all time
                </span>
              ) : undefined
            }
          />
          <StatCard
            icon={<BarChart3 className='w-7 h-7 text-secondary-400' />}
            label='Saved Charts'
            value={String(data?.user_stats.charts_created ?? 0)}
            sub={data ? `Max ${data.user_stats.max_indicators} indicators per chart` : undefined}
          />
          <StatCard
            icon={<Star className='w-7 h-7 text-yellow-400' />}
            label='Watchlist'
            value={`${data?.favorites_count ?? 0} items`}
          />
          <StatCard
            icon={<Crown className='w-7 h-7 text-accent-400' />}
            label='Plan'
            value={(data?.user?.user_type ?? user?.user_type ?? 'free').replace(/^./, (c) =>
              c.toUpperCase(),
            )}
          />
        </div>

        <div className='flex gap-1 mb-6 border-b border-white/10 overflow-x-auto' role='tablist'>
          {TABS.map((t) => (
            <button
              key={t.id}
              type='button'
              role='tab'
              aria-selected={tab === t.id}
              onClick={() => setSearchParams({ tab: t.id }, { replace: true })}
              className={`px-4 py-2 text-sm font-medium whitespace-nowrap border-b-2 -mb-px transition-colors ${
                tab === t.id
                  ? 'border-primary-500 text-primary-300'
                  : 'border-transparent text-gray-400 hover:text-white'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        <div role='tabpanel'>
          {tab === 'portfolio' && <PortfolioTab onChanged={refresh} />}
          {tab === 'watchlist' && <WatchlistTab onChanged={refresh} />}
          {tab === 'charts' && <SavedChartsTab onChanged={refresh} />}
          {tab === 'alerts' && <AlertsTab />}
          {tab === 'accounts' && <ConnectedAccountsTab />}
        </div>
      </div>
    </div>
  )
}

function StatCard({
  icon,
  label,
  value,
  sub,
}: {
  icon: React.ReactNode
  label: string
  value: string
  sub?: React.ReactNode
}) {
  return (
    <div className='glass-effect rounded-2xl p-5'>
      <div className='flex items-center gap-3'>
        {icon}
        <div className='min-w-0'>
          <h3 className='text-sm font-medium text-gray-400'>{label}</h3>
          <p className='text-xl font-bold text-white truncate'>{value}</p>
          {sub && <p className='text-xs text-gray-400 mt-0.5'>{sub}</p>}
        </div>
      </div>
    </div>
  )
}
