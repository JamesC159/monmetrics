import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate, useSearchParams } from 'react-router-dom'
import {
  ArrowLeft,
  TrendingUp,
  TrendingDown,
  BookmarkPlus,
  Share2,
  Tag,
  Store,
  LineChart as LineChartIcon,
} from 'lucide-react'
import { apiClient } from '../utils/api'
import { useAuth } from '../context/AuthContext'
import { useToast } from '../context/ToastContext'
import PriceChart from '../components/chart/PriceChart'
import IndicatorPanel from '../components/chart/IndicatorPanel'
import SaveChartModal from '../components/chart/SaveChartModal'
import AlertModal from '../components/chart/AlertModal'
import AlertList from '../components/alerts/AlertList'
import FavoriteButton from '../components/FavoriteButton'
import { useAlerts } from '../hooks/useAlerts'
import { readPendingAutosave, useAutosaveChart } from '../hooks/useAutosaveChart'
import type {
  Card,
  ChartDrawing,
  ChartIndicator,
  ChartSource,
  ChartTimeRange,
  ChartType,
  PriceAlert,
  PriceHistory,
  SavedChart,
  SavedChartRequest,
} from '../types'

const GUEST_MAX_INDICATORS = 1
const drawingsKey = (cardId: string) => `mm:drawings:${cardId}`

function readLocalDrawings(cardId: string): ChartDrawing[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(drawingsKey(cardId)) ?? '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export default function CardDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { addToast } = useToast()
  const [searchParams, setSearchParams] = useSearchParams()
  const chartParam = searchParams.get('chart')

  const [card, setCard] = useState<Card | null>(null)
  const [priceHistory, setPriceHistory] = useState<PriceHistory | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [timeRange, setTimeRange] = useState<ChartTimeRange>('30d')
  const [selectedSource, setSelectedSource] = useState<ChartSource>('all')
  const [indicators, setIndicators] = useState<ChartIndicator[]>([])
  const [showIndicators, setShowIndicators] = useState(false)
  const [loadedChart, setLoadedChart] = useState<SavedChart | null>(null)
  const [saveOpen, setSaveOpen] = useState(false)
  const [chartType, setChartType] = useState<ChartType>('line')
  const [showVolume, setShowVolume] = useState(false)
  const [drawings, setDrawings] = useState<ChartDrawing[]>([])
  const [alertModal, setAlertModal] = useState<{
    initialPrice?: number
    existing?: PriceAlert
  } | null>(null)
  const {
    alerts,
    upsert: upsertAlert,
    toggle: toggleAlert,
    move: moveAlert,
    remove: removeAlert,
  } = useAlerts(!!user, id)

  const maxIndicators = !user ? GUEST_MAX_INDICATORS : user.user_type === 'paid' ? 10 : 3

  // Unsaved drawings persist per card in localStorage; saved charts carry their own.
  useEffect(() => {
    if (id && !chartParam) setDrawings(readLocalDrawings(id))
  }, [id, chartParam])

  const handleDrawingsChange = (next: ChartDrawing[]) => {
    setDrawings(next)
    if (!id || loadedChart) return
    try {
      if (next.length) localStorage.setItem(drawingsKey(id), JSON.stringify(next))
      else localStorage.removeItem(drawingsKey(id))
    } catch {
      addToast('Could not store drawings locally', 'warning')
    }
  }

  const openAlertAt = (price?: number) => {
    if (!user) {
      navigate('/login')
      return
    }
    setAlertModal({ initialPrice: price })
  }

  // Load card data
  useEffect(() => {
    if (!id) {
      setError('Invalid card ID')
      setLoading(false)
      return
    }

    const loadCard = async () => {
      try {
        setLoading(true)
        setError(null)
        const cardData = await apiClient.getCard(id)
        setCard(cardData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load card')
        if (err instanceof Error && err.message.includes('not found')) {
          setTimeout(() => navigate('/search'), 3000)
        }
      } finally {
        setLoading(false)
      }
    }

    loadCard()
  }, [id, navigate])

  // Apply a saved chart's configuration when opened via ?chart=<id>
  useEffect(() => {
    if (!chartParam || !user) {
      setLoadedChart(null)
      return
    }
    let cancelled = false
    apiClient
      .getChart(chartParam)
      .then((chart) => {
        if (cancelled) return
        if (chart.card_id !== id) {
          navigate(`/card/${chart.card_id}?chart=${chart.id}`, { replace: true })
          return
        }
        setLoadedChart(chart)
        // A tab close or failed request mid-autosave can leave newer edits
        // in localStorage; prefer them over the server's stale copy.
        const pending = readPendingAutosave(chart.id)
        const applied = pending ?? chart
        setTimeRange(applied.time_range)
        setSelectedSource(applied.source ?? 'all')
        setIndicators(applied.indicators ?? [])
        setShowIndicators((applied.indicators ?? []).length > 0)
        setChartType(applied.chart_type ?? 'line')
        setShowVolume(!!applied.show_volume)
        setDrawings(applied.drawings ?? [])
      })
      .catch((err) => {
        if (cancelled) return
        addToast(err instanceof Error ? err.message : 'Could not load saved chart', 'error')
        setSearchParams({}, { replace: true })
      })
    return () => {
      cancelled = true
    }
  }, [chartParam, user, id, navigate, addToast, setSearchParams])

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const formatDateTime = (date: string) => {
    return new Date(date).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    })
  }

  const handleSaveChart = () => {
    if (!user) {
      navigate('/login')
      return
    }
    setSaveOpen(true)
  }

  const handleChartSaved = (chart: SavedChart) => {
    setLoadedChart(chart)
    setSearchParams({ chart: chart.id }, { replace: true })
    if (id) localStorage.removeItem(drawingsKey(id))
  }

  const chartPayload: SavedChartRequest | null = loadedChart
    ? {
        card_id: loadedChart.card_id,
        name: loadedChart.name,
        description: loadedChart.description ?? '',
        indicators,
        time_range: timeRange,
        source: selectedSource,
        chart_type: chartType,
        show_volume: showVolume,
        drawings,
      }
    : null

  const { status: autosaveStatus, retry: retryAutosave } = useAutosaveChart(
    loadedChart,
    chartPayload,
    setLoadedChart,
  )

  const handleShare = () => {
    if (navigator.share) {
      navigator.share({
        title: card?.name,
        text: `Check out the price analysis for ${card?.name}`,
        url: window.location.href,
      })
    } else {
      navigator.clipboard.writeText(window.location.href)
      alert('Link copied to clipboard!')
    }
  }

  if (loading) {
    return (
      <div className='min-h-screen bg-dark-900 flex items-center justify-center'>
        <div className='text-center'>
          <div className='inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4'></div>
          <h3 className='text-xl font-semibold text-white mb-2'>Loading Card...</h3>
          <p className='text-gray-400'>Getting the latest market data</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className='min-h-screen bg-dark-900 flex items-center justify-center'>
        <div className='text-center max-w-md'>
          <div className='text-red-400 text-6xl mb-4'>⚠️</div>
          <h3 className='text-xl font-semibold text-white mb-2'>Card Not Found</h3>
          <p className='text-gray-400 mb-6'>{error}</p>
          <Link to='/search' className='btn-primary'>
            Back to Search
          </Link>
        </div>
      </div>
    )
  }

  if (!card) return null

  return (
    <div className='min-h-screen bg-dark-900'>
      <SaveChartModal
        open={saveOpen}
        onClose={() => setSaveOpen(false)}
        card={card}
        timeRange={timeRange}
        source={selectedSource}
        indicators={indicators}
        chartType={chartType}
        showVolume={showVolume}
        drawings={drawings}
        existing={loadedChart}
        onSaved={handleChartSaved}
      />
      <AlertModal
        open={!!alertModal}
        onClose={() => setAlertModal(null)}
        cardId={card.id}
        cardName={card.name}
        currentPrice={card.current_price}
        initialPrice={alertModal?.initialPrice}
        existing={alertModal?.existing}
        source={selectedSource}
        onSaved={upsertAlert}
      />
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        {/* Header */}
        <div className='mb-8'>
          <Link
            to='/search'
            className='inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4'
          >
            <ArrowLeft className='w-4 h-4 mr-2' />
            Back to Search
          </Link>

          <div className='flex items-start justify-between'>
            <div>
              <h1 className='text-4xl font-bold text-white mb-2'>{card.name}</h1>
              <div className='flex items-center gap-4 text-gray-400'>
                <span className='flex items-center'>
                  <Tag className='w-4 h-4 mr-1' />
                  {card.game}
                </span>
                <span>{card.set}</span>
                {card.rarity && <span>• {card.rarity}</span>}
                {card.number && <span>• #{card.number}</span>}
              </div>
            </div>

            <div className='flex gap-2'>
              <FavoriteButton cardId={card.id} variant='button' />
              <button onClick={handleSaveChart} className='btn-ghost py-2 px-4 flex items-center'>
                <BookmarkPlus className='w-4 h-4 mr-2' />
                {loadedChart ? 'Update Chart' : 'Save Chart'}
              </button>
              <button onClick={handleShare} className='btn-ghost py-2 px-4 flex items-center'>
                <Share2 className='w-4 h-4 mr-2' />
                Share
              </button>
            </div>
          </div>
          {loadedChart && (
            <div className='mt-3 flex flex-wrap items-center gap-3'>
              <div className='inline-flex items-center gap-2 badge badge-cyan'>
                <LineChartIcon className='w-4 h-4' />
                Viewing saved chart: {loadedChart.name}
              </div>
              <span className='text-sm text-gray-400'>
                {autosaveStatus === 'saving' && 'Saving changes…'}
                {autosaveStatus === 'saved' && 'All changes saved'}
                {autosaveStatus === 'error' && (
                  <span className='text-error-400'>
                    Could not save changes.{' '}
                    <button
                      type='button'
                      onClick={retryAutosave}
                      className='underline text-primary-300'
                    >
                      Retry
                    </button>
                  </span>
                )}
              </span>
            </div>
          )}
        </div>

        <div className='grid grid-cols-1 lg:grid-cols-3 gap-8'>
          {/* Left Column - Card Image and Info */}
          <div className='lg:col-span-1 space-y-6'>
            {/* Card Image */}
            <div className='glass-effect rounded-2xl p-6'>
              <div className='aspect-[3/4] relative overflow-hidden rounded-lg mb-4'>
                <img
                  src={card.image_url}
                  alt={card.name}
                  className='w-full h-full object-cover'
                  onError={(e) => {
                    ;(e.target as HTMLImageElement).src =
                      'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'
                  }}
                />
                <div className='absolute top-3 left-3'>
                  <span className='px-3 py-1 bg-black/70 text-white text-sm rounded-full'>
                    {card.category === 'sealed' ? 'Sealed Product' : 'Trading Card'}
                  </span>
                </div>
              </div>

              {card.description && (
                <p className='text-gray-300 text-sm leading-relaxed'>{card.description}</p>
              )}
            </div>

            {/* Current Pricing */}
            <div className='glass-effect rounded-2xl p-6'>
              <h3 className='text-lg font-semibold text-white mb-4'>Current Market Price</h3>
              <div className='text-center mb-6'>
                <div className='text-3xl font-bold text-green-400 mb-2'>
                  {formatPrice(card.current_price)}
                </div>
                <div className='text-gray-400 text-sm'>
                  Last updated {formatDate(card.updated_at)}
                </div>
              </div>

              <div className='grid grid-cols-2 gap-4'>
                <div className='text-center p-4 bg-green-600/20 rounded-lg'>
                  <TrendingUp className='w-6 h-6 text-green-400 mx-auto mb-2' />
                  <div className='text-sm text-gray-400 mb-1'>All-Time High</div>
                  <div className='text-lg font-semibold text-green-400'>
                    {formatPrice(card.all_time_high)}
                  </div>
                  <div className='text-xs text-gray-500 mt-1'>{formatDate(card.ath_date)}</div>
                </div>

                <div className='text-center p-4 bg-red-600/20 rounded-lg'>
                  <TrendingDown className='w-6 h-6 text-red-400 mx-auto mb-2' />
                  <div className='text-sm text-gray-400 mb-1'>All-Time Low</div>
                  <div className='text-lg font-semibold text-red-400'>
                    {formatPrice(card.all_time_low)}
                  </div>
                  <div className='text-xs text-gray-500 mt-1'>{formatDate(card.atl_date)}</div>
                </div>
              </div>
            </div>

            {/* Tags */}
            {card.tags && card.tags.length > 0 && (
              <div className='glass-effect rounded-2xl p-6'>
                <h3 className='text-lg font-semibold text-white mb-4'>Tags</h3>
                <div className='flex flex-wrap gap-2'>
                  {card.tags.map((tag) => (
                    <span
                      key={tag}
                      className='px-3 py-1 bg-white/10 text-gray-300 rounded-full text-sm'
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Right Column - Charts and Listings */}
          <div className='lg:col-span-2 space-y-6'>
            {/* Price Chart */}
            <PriceChart
              cardId={card.id}
              timeRange={timeRange}
              onTimeRangeChange={setTimeRange}
              source={selectedSource}
              onSourceChange={setSelectedSource}
              indicators={indicators}
              onHistory={setPriceHistory}
              chartType={chartType}
              onChartTypeChange={setChartType}
              showVolume={showVolume}
              onShowVolumeChange={setShowVolume}
              drawings={drawings}
              onDrawingsChange={handleDrawingsChange}
              alerts={alerts}
              onAlertCreateAt={openAlertAt}
              onAlertMove={moveAlert}
              headerActions={
                <button
                  type='button'
                  onClick={() => setShowIndicators((v) => !v)}
                  aria-expanded={showIndicators}
                  className={`px-3 py-1 rounded text-sm transition-colors ${
                    showIndicators
                      ? 'bg-primary-500/30 text-primary-200'
                      : 'bg-white/10 text-gray-300 hover:bg-white/20'
                  }`}
                >
                  Indicators{indicators.length > 0 ? ` (${indicators.length})` : ''}
                </button>
              }
            >
              {showIndicators && (
                <IndicatorPanel
                  indicators={indicators}
                  onChange={setIndicators}
                  maxIndicators={maxIndicators}
                  timeRange={timeRange}
                  limitHint={
                    !user
                      ? 'Sign in to add more indicators and save charts.'
                      : user.user_type === 'paid'
                        ? undefined
                        : 'Free plan allows 3 indicators. Upgrade for up to 10.'
                  }
                />
              )}
              {user && alerts.length > 0 && (
                <div className='mt-4 p-4 rounded-xl bg-white/5 border border-white/10'>
                  <h4 className='text-sm font-semibold text-white mb-2'>
                    Your alerts on this card
                  </h4>
                  <AlertList
                    alerts={alerts}
                    onEdit={(a) => setAlertModal({ existing: a })}
                    onToggle={toggleAlert}
                    onDelete={removeAlert}
                  />
                </div>
              )}
            </PriceChart>

            {/* Current Listings */}
            <div className='glass-effect rounded-2xl p-6'>
              <div className='flex items-center justify-between mb-6'>
                <h3 className='text-lg font-semibold text-white'>Current Listings</h3>
                <span className='text-gray-400 text-sm'>
                  {priceHistory?.listings?.length || 0} active listings
                </span>
              </div>

              {priceHistory?.listings && priceHistory.listings.length > 0 ? (
                <div className='space-y-4'>
                  {priceHistory.listings.slice(0, 5).map((listing) => (
                    <div
                      key={listing.id}
                      className='flex items-center gap-4 p-4 bg-white/5 rounded-lg hover:bg-white/10 transition-colors'
                    >
                      {listing.image_url && (
                        <img
                          src={listing.image_url}
                          alt={listing.title}
                          className='w-16 h-20 object-cover rounded'
                          onError={(e) => {
                            ;(e.target as HTMLImageElement).style.display = 'none'
                          }}
                        />
                      )}

                      <div className='flex-1'>
                        <h4 className='font-medium text-white mb-1'>{listing.title}</h4>
                        <div className='flex items-center gap-4 text-sm text-gray-400'>
                          <span className='flex items-center'>
                            <Store className='w-4 h-4 mr-1' />
                            {listing.seller}
                          </span>
                          <span>Condition: {listing.condition}</span>
                          <span>Qty: {listing.quantity}</span>
                          <span className='capitalize'>{listing.source}</span>
                        </div>
                      </div>

                      <div className='text-right'>
                        <div className='text-xl font-bold text-green-400'>
                          {formatPrice(listing.price)}
                        </div>
                        <div className='text-xs text-gray-500'>
                          {formatDateTime(listing.created_at)}
                        </div>
                      </div>
                    </div>
                  ))}

                  {priceHistory.listings.length > 5 && (
                    <div className='text-center pt-4 border-t border-white/10'>
                      <span className='text-gray-400 text-sm'>
                        +{priceHistory.listings.length - 5} more listings available
                      </span>
                    </div>
                  )}
                </div>
              ) : (
                <div className='text-center py-8'>
                  <Store className='w-12 h-12 text-gray-600 mx-auto mb-3' />
                  <div className='text-gray-400 mb-2'>No Current Listings</div>
                  <div className='text-gray-500 text-sm'>
                    Check back later for marketplace activity
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
