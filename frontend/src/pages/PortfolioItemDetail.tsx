import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, ExternalLink, Link2, Pencil, ShoppingCart, Trash2 } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import { apiClient } from '@/utils/api'
import { formatDate, formatDateTime, formatPercentage, formatPrice } from '@/utils/formatters'
import {
  CONDITION_LABELS,
  FINISH_LABELS,
  gainClass,
  gradeLabel,
  ITEM_TYPE_LABELS,
  PLACEHOLDER_IMAGE,
  signed,
  VALUE_SOURCE_LABELS,
} from '@/utils/portfolio'
import PriceChart from '@/components/chart/PriceChart'
import PortfolioItemForm from '@/components/portfolio/PortfolioItemForm'
import ListingComposer from '@/components/marketplace/ListingComposer'
import type {
  ChartSource,
  ChartTimeRange,
  LinkedAccountStatus,
  MarketplaceListing,
  MarketplaceProvider,
  PortfolioItemView,
} from '@/types'

const STATUS_BADGE: Record<MarketplaceListing['status'], string> = {
  published: 'badge-success',
  draft: 'badge-cyan',
  failed: 'badge-error',
  ended: 'bg-white/10 text-gray-400',
}

export default function PortfolioItemDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading } = useAuth()
  const { addToast } = useToast()

  const [item, setItem] = useState<PortfolioItemView | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [accounts, setAccounts] = useState<LinkedAccountStatus[]>([])
  const [listings, setListings] = useState<MarketplaceListing[]>([])
  const [editOpen, setEditOpen] = useState(false)
  const [sellProvider, setSellProvider] = useState<MarketplaceProvider | null>(null)
  const [timeRange, setTimeRange] = useState<ChartTimeRange>('90d')
  const [source, setSource] = useState<ChartSource>('all')

  const loadListings = useCallback(() => {
    if (!id) return
    apiClient
      .getListings(id)
      .then(setListings)
      .catch((err) => console.error('Failed to load listings:', err))
  }, [id])

  useEffect(() => {
    if (!isAuthenticated || !id) return
    apiClient
      .getPortfolioItem(id)
      .then(setItem)
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load item'))
    apiClient
      .getLinkedAccounts()
      .then(setAccounts)
      .catch((err) => console.error('Failed to load linked accounts:', err))
    loadListings()
  }, [isAuthenticated, id, loadListings])

  if (isLoading) {
    return (
      <div className='min-h-screen bg-dark-950 flex items-center justify-center'>
        <div className='loading-spinner' />
      </div>
    )
  }
  if (!isAuthenticated) return <Navigate to='/login' replace />

  if (error) {
    return (
      <div className='min-h-screen bg-dark-900 flex items-center justify-center'>
        <div className='text-center'>
          <h2 className='text-xl font-semibold text-white mb-2'>Item not available</h2>
          <p className='text-gray-400 mb-6'>{error}</p>
          <Link to='/dashboard?tab=portfolio' className='btn-primary'>
            Back to portfolio
          </Link>
        </div>
      </div>
    )
  }

  if (!item) {
    return (
      <div className='min-h-screen bg-dark-950 flex items-center justify-center'>
        <div className='loading-spinner' />
      </div>
    )
  }

  const remove = async () => {
    if (!window.confirm(`Remove "${item.display_name}" from your portfolio?`)) return
    try {
      await apiClient.deletePortfolioItem(item.id)
      addToast('Item removed', 'success')
      navigate('/dashboard?tab=portfolio')
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not remove item', 'error')
    }
  }

  const details: [string, string][] = [
    ['Type', ITEM_TYPE_LABELS[item.item_type]],
    ['Game', item.display_game],
    ['Set', item.display_set || '—'],
  ]
  if (item.card?.number) details.push(['Number', item.card.number])
  if (item.card?.rarity) details.push(['Rarity', item.card.rarity])
  if (item.item_type === 'raw_card' && item.condition)
    details.push(['Condition', `${item.condition} · ${CONDITION_LABELS[item.condition]}`])
  if (item.grading) {
    details.push(['Grade', gradeLabel(item)])
    if (item.grading.cert_number) details.push(['Cert #', item.grading.cert_number])
  }
  if (item.finish) details.push(['Finish', FINISH_LABELS[item.finish]])
  if (item.language) details.push(['Language', item.language])
  details.push(['Quantity', String(item.quantity)])
  if (item.purchase_date) details.push(['Purchased', formatDate(item.purchase_date)])
  if (item.purchase_source) details.push(['From', item.purchase_source])

  const valuationNote = (() => {
    switch (item.value_source) {
      case 'manual':
        return 'Using your manual value.'
      case 'graded_estimate':
        return `Estimated as ${item.multiplier}× the raw market price (${item.card ? formatPrice(item.card.current_price) : '—'}) for ${gradeLabel(item)}.`
      case 'market':
        return item.multiplier !== 1
          ? `Market price ${item.card ? formatPrice(item.card.current_price) : ''} × ${item.multiplier} for condition.`
          : 'Current market price.'
      default:
        return item.card_unavailable
          ? 'This catalog item is no longer available. Set a manual value to track it.'
          : 'No market data for custom items. Set a manual value to track it.'
    }
  })()

  return (
    <div className='min-h-screen bg-dark-900'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <Link
          to='/dashboard?tab=portfolio'
          className='inline-flex items-center text-gray-400 hover:text-white mb-4'
        >
          <ArrowLeft className='w-4 h-4 mr-2' /> Back to portfolio
        </Link>

        <div className='flex flex-wrap items-start justify-between gap-4 mb-8'>
          <div>
            <h1 className='text-3xl md:text-4xl font-bold text-white mb-2'>{item.display_name}</h1>
            <div className='flex flex-wrap items-center gap-2'>
              <span className='badge badge-neutral'>{ITEM_TYPE_LABELS[item.item_type]}</span>
              {item.grading && <span className='badge badge-gold'>{gradeLabel(item)}</span>}
              {!item.card_id && (
                <span className='badge bg-white/10 text-gray-300'>Custom item</span>
              )}
              {item.card_id && (
                <Link
                  to={`/card/${item.card_id}`}
                  className='text-sm text-secondary-300 hover:text-secondary-200 inline-flex items-center'
                >
                  View market page <ExternalLink className='w-3 h-3 ml-1' />
                </Link>
              )}
            </div>
          </div>
          <div className='flex gap-2'>
            <button
              type='button'
              onClick={() => setEditOpen(true)}
              className='btn-ghost py-2 px-4 flex items-center'
            >
              <Pencil className='w-4 h-4 mr-2' /> Edit
            </button>
            <button
              type='button'
              onClick={remove}
              className='btn-ghost py-2 px-4 flex items-center hover:text-error-400'
            >
              <Trash2 className='w-4 h-4 mr-2' /> Remove
            </button>
          </div>
        </div>

        <div className='grid grid-cols-1 lg:grid-cols-3 gap-8'>
          <div className='space-y-6'>
            <div className='glass-effect rounded-2xl p-6'>
              <img
                src={item.display_image_url || PLACEHOLDER_IMAGE}
                alt={item.display_name}
                className='w-full aspect-[3/4] object-cover rounded-lg mb-4'
                onError={(e) => ((e.target as HTMLImageElement).src = PLACEHOLDER_IMAGE)}
              />
              <dl className='grid grid-cols-2 gap-y-2 text-sm'>
                {details.map(([k, v]) => (
                  <div key={k} className='contents'>
                    <dt className='text-gray-400'>{k}</dt>
                    <dd className='text-white text-right'>{v}</dd>
                  </div>
                ))}
              </dl>
              {item.notes && (
                <p className='mt-4 pt-4 border-t border-white/10 text-sm text-gray-300 whitespace-pre-line'>
                  {item.notes}
                </p>
              )}
            </div>
          </div>

          <div className='lg:col-span-2 space-y-6'>
            <div className='glass-effect rounded-2xl p-6'>
              <h3 className='text-lg font-semibold text-white mb-4'>Valuation</h3>
              <div className='grid grid-cols-2 md:grid-cols-4 gap-4 mb-4'>
                <Metric
                  label='Value (each)'
                  value={item.value_source === 'none' ? '—' : formatPrice(item.unit_value)}
                />
                <Metric
                  label='Total value'
                  value={item.value_source === 'none' ? '—' : formatPrice(item.total_value)}
                />
                <Metric
                  label='Cost basis'
                  value={formatPrice(item.cost_basis)}
                  sub={`${formatPrice(item.purchase_price)} each`}
                />
                <Metric
                  label='Gain / loss'
                  value={
                    item.value_source === 'none'
                      ? '—'
                      : signed(formatPrice(item.gain_loss), item.gain_loss)
                  }
                  sub={
                    item.value_source === 'none'
                      ? undefined
                      : signed(formatPercentage(item.gain_loss_pct), item.gain_loss_pct)
                  }
                  className={
                    item.value_source === 'none' ? 'text-gray-400' : gainClass(item.gain_loss)
                  }
                />
              </div>
              <p className='text-sm text-gray-400'>
                <span className='badge badge-cyan text-xs mr-2'>
                  {VALUE_SOURCE_LABELS[item.value_source]}
                </span>
                {valuationNote}
              </p>
            </div>

            <div className='glass-effect rounded-2xl p-6'>
              <h3 className='text-lg font-semibold text-white mb-4 flex items-center'>
                <ShoppingCart className='w-5 h-5 mr-2 text-primary-400' /> Sell this item
              </h3>
              <div className='flex flex-wrap gap-3'>
                {accounts.map((a) =>
                  a.linked && a.account?.status !== 'revoked' ? (
                    <button
                      key={a.provider}
                      type='button'
                      onClick={() => setSellProvider(a.provider)}
                      className='btn-primary py-2 px-4'
                    >
                      Sell on {a.display_name}
                    </button>
                  ) : (
                    <Link
                      key={a.provider}
                      to='/dashboard?tab=accounts'
                      className='btn-ghost py-2 px-4 flex items-center'
                    >
                      <Link2 className='w-4 h-4 mr-2' /> Link {a.display_name} to sell
                    </Link>
                  ),
                )}
                {accounts.length === 0 && (
                  <p className='text-sm text-gray-400'>Loading marketplace accounts...</p>
                )}
              </div>

              {listings.length > 0 && (
                <div className='mt-6'>
                  <h4 className='text-sm font-semibold text-gray-300 mb-2'>Your listings</h4>
                  <ul className='divide-y divide-white/5'>
                    {listings.map((l) => (
                      <li key={l.id} className='py-3 flex flex-wrap items-center gap-3 text-sm'>
                        <span className={`badge text-xs capitalize ${STATUS_BADGE[l.status]}`}>
                          {l.status}
                        </span>
                        <span className='text-gray-400 w-20'>
                          {l.provider === 'ebay' ? 'eBay' : 'TCGPlayer'}
                        </span>
                        <span className='text-white flex-1 min-w-0 truncate'>{l.title}</span>
                        <span className='text-green-400'>{formatPrice(l.price)}</span>
                        <span className='text-gray-500 text-xs w-40 text-right'>
                          {l.external_listing_id ?? l.error_message ?? formatDateTime(l.created_at)}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>

            {item.card_id && item.card && (
              <PriceChart
                cardId={item.card_id}
                title='Market price history (raw)'
                timeRange={timeRange}
                onTimeRangeChange={setTimeRange}
                source={source}
                onSourceChange={setSource}
              />
            )}
          </div>
        </div>
      </div>

      <PortfolioItemForm
        open={editOpen}
        onClose={() => setEditOpen(false)}
        item={item}
        onSaved={(saved) => {
          setItem(saved)
          addToast('Item updated', 'success')
        }}
      />
      {sellProvider && (
        <ListingComposer
          open={!!sellProvider}
          onClose={() => setSellProvider(null)}
          item={item}
          provider={sellProvider}
          onCreated={() => loadListings()}
        />
      )}
    </div>
  )
}

function Metric({
  label,
  value,
  sub,
  className = 'text-white',
}: {
  label: string
  value: string
  sub?: string
  className?: string
}) {
  return (
    <div className='rounded-xl bg-white/5 p-3'>
      <div className='text-xs text-gray-400'>{label}</div>
      <div className={`text-lg font-bold ${className}`}>{value}</div>
      {sub && <div className='text-xs text-gray-400'>{sub}</div>}
    </div>
  )
}
