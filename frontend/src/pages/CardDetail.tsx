import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  TrendingUp,
  TrendingDown,
  Star,
  BookmarkPlus,
  Share2,
  ExternalLink,
  Calendar,
  Tag,
  Store,
  Users
} from 'lucide-react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { apiClient } from '../utils/api'
import { useAuth } from '../context/AuthContext'
import type { Card, PricePoint, PriceHistory, Listing } from '../types'

export default function CardDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()

  const [card, setCard] = useState<Card | null>(null)
  const [priceHistory, setPriceHistory] = useState<PriceHistory | null>(null)
  const [loading, setLoading] = useState(true)
  const [priceLoading, setPriceLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [timeRange, setTimeRange] = useState('30d')
  const [selectedSource, setSelectedSource] = useState('all')
  const [showIndicators, setShowIndicators] = useState(false)

  const timeRanges = [
    { value: '1d', label: '1D' },
    { value: '7d', label: '7D' },
    { value: '30d', label: '30D' },
    { value: '90d', label: '90D' },
    { value: '1y', label: '1Y' },
    { value: '5y', label: '5Y' }
  ]

  const sources = [
    { value: 'all', label: 'All Sources' },
    { value: 'ebay', label: 'eBay' },
    { value: 'tcgplayer', label: 'TCGPlayer' }
  ]

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

  // Load price history
  useEffect(() => {
    if (!id) return

    const loadPriceHistory = async () => {
      try {
        setPriceLoading(true)
        const historyData = await apiClient.getCardPrices(id, timeRange)
        setPriceHistory(historyData)
      } catch (err) {
        console.error('Failed to load price history:', err)
        // Don't show error for price history failure, just log it
      } finally {
        setPriceLoading(false)
      }
    }

    loadPriceHistory()
  }, [id, timeRange])

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(price)
  }

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  const formatDateTime = (date: string) => {
    return new Date(date).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    })
  }

  // Process price data for chart
  const processChartData = () => {
    if (!priceHistory?.prices) return []

    const filteredPrices = selectedSource === 'all'
      ? priceHistory.prices
      : priceHistory.prices.filter(p => p.source === selectedSource)

    // Group by date and average prices
    const priceMap = new Map()

    filteredPrices.forEach(price => {
      const dateKey = new Date(price.timestamp).toDateString()
      if (!priceMap.has(dateKey)) {
        priceMap.set(dateKey, { prices: [], timestamp: price.timestamp })
      }
      priceMap.get(dateKey).prices.push(price.price)
    })

    return Array.from(priceMap.entries()).map(([date, data]) => {
      const prices = data.prices
      const avgPrice = prices.reduce((sum: number, p: number) => sum + p, 0) / prices.length
      const maxPrice = Math.max(...prices)
      const minPrice = Math.min(...prices)

      return {
        date: new Date(data.timestamp).toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric'
        }),
        fullDate: data.timestamp,
        price: Number(avgPrice.toFixed(2)),
        high: Number(maxPrice.toFixed(2)),
        low: Number(minPrice.toFixed(2)),
        volume: prices.length
      }
    }).sort((a, b) => new Date(a.fullDate).getTime() - new Date(b.fullDate).getTime())
  }

  const chartData = processChartData()

  const handleSaveChart = () => {
    if (!user) {
      navigate('/login')
      return
    }
    // TODO: Implement save chart functionality
    alert('Save chart functionality coming soon!')
  }

  const handleShare = () => {
    if (navigator.share) {
      navigator.share({
        title: card?.name,
        text: `Check out the price analysis for ${card?.name}`,
        url: window.location.href
      })
    } else {
      navigator.clipboard.writeText(window.location.href)
      alert('Link copied to clipboard!')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
          <h3 className="text-xl font-semibold text-white mb-2">Loading Card...</h3>
          <p className="text-gray-400">Getting the latest market data</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center">
        <div className="text-center max-w-md">
          <div className="text-red-400 text-6xl mb-4">⚠️</div>
          <h3 className="text-xl font-semibold text-white mb-2">Card Not Found</h3>
          <p className="text-gray-400 mb-6">{error}</p>
          <Link to="/search" className="button-primary">
            Back to Search
          </Link>
        </div>
      </div>
    )
  }

  if (!card) return null

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <Link
            to="/search"
            className="inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Search
          </Link>

          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-4xl font-bold text-white mb-2">{card.name}</h1>
              <div className="flex items-center gap-4 text-gray-400">
                <span className="flex items-center">
                  <Tag className="w-4 h-4 mr-1" />
                  {card.game}
                </span>
                <span>{card.set}</span>
                {card.rarity && <span>• {card.rarity}</span>}
                {card.number && <span>• #{card.number}</span>}
              </div>
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleSaveChart}
                className="button-secondary flex items-center"
              >
                <BookmarkPlus className="w-4 h-4 mr-2" />
                Save
              </button>
              <button
                onClick={handleShare}
                className="button-secondary flex items-center"
              >
                <Share2 className="w-4 h-4 mr-2" />
                Share
              </button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Card Image and Info */}
          <div className="lg:col-span-1 space-y-6">
            {/* Card Image */}
            <div className="glass-effect rounded-2xl p-6">
              <div className="aspect-[3/4] relative overflow-hidden rounded-lg mb-4">
                <img
                  src={card.image_url}
                  alt={card.name}
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    (e.target as HTMLImageElement).src = 'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'
                  }}
                />
                <div className="absolute top-3 left-3">
                  <span className="px-3 py-1 bg-black/70 text-white text-sm rounded-full">
                    {card.category === 'sealed' ? 'Sealed Product' : 'Trading Card'}
                  </span>
                </div>
              </div>

              {card.description && (
                <p className="text-gray-300 text-sm leading-relaxed">
                  {card.description}
                </p>
              )}
            </div>

            {/* Current Pricing */}
            <div className="glass-effect rounded-2xl p-6">
              <h3 className="text-lg font-semibold text-white mb-4">Current Market Price</h3>
              <div className="text-center mb-6">
                <div className="text-3xl font-bold text-green-400 mb-2">
                  {formatPrice(card.current_price)}
                </div>
                <div className="text-gray-400 text-sm">
                  Last updated {formatDate(card.updated_at)}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-green-600/20 rounded-lg">
                  <TrendingUp className="w-6 h-6 text-green-400 mx-auto mb-2" />
                  <div className="text-sm text-gray-400 mb-1">All-Time High</div>
                  <div className="text-lg font-semibold text-green-400">
                    {formatPrice(card.all_time_high)}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    {formatDate(card.ath_date)}
                  </div>
                </div>

                <div className="text-center p-4 bg-red-600/20 rounded-lg">
                  <TrendingDown className="w-6 h-6 text-red-400 mx-auto mb-2" />
                  <div className="text-sm text-gray-400 mb-1">All-Time Low</div>
                  <div className="text-lg font-semibold text-red-400">
                    {formatPrice(card.all_time_low)}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    {formatDate(card.atl_date)}
                  </div>
                </div>
              </div>
            </div>

            {/* Tags */}
            {card.tags && card.tags.length > 0 && (
              <div className="glass-effect rounded-2xl p-6">
                <h3 className="text-lg font-semibold text-white mb-4">Tags</h3>
                <div className="flex flex-wrap gap-2">
                  {card.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-3 py-1 bg-white/10 text-gray-300 rounded-full text-sm"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Right Column - Charts and Listings */}
          <div className="lg:col-span-2 space-y-6">
            {/* Price Chart */}
            <div className="glass-effect rounded-2xl p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white">Price History</h3>
                <div className="flex gap-2">
                  <select
                    value={selectedSource}
                    onChange={(e) => setSelectedSource(e.target.value)}
                    className="px-3 py-1 bg-white/10 border border-white/20 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {sources.map((source) => (
                      <option key={source.value} value={source.value} className="bg-slate-800">
                        {source.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Time Range Selector */}
              <div className="flex gap-2 mb-6">
                {timeRanges.map((range) => (
                  <button
                    key={range.value}
                    onClick={() => setTimeRange(range.value)}
                    className={`px-3 py-1 rounded text-sm transition-colors ${
                      timeRange === range.value
                        ? 'bg-blue-600 text-white'
                        : 'bg-white/10 text-gray-400 hover:bg-white/20'
                    }`}
                  >
                    {range.label}
                  </button>
                ))}
              </div>

              {/* Chart */}
              <div className="h-80">
                {priceLoading ? (
                  <div className="flex items-center justify-center h-full">
                    <div className="text-gray-400">Loading chart data...</div>
                  </div>
                ) : chartData.length > 0 ? (
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis
                        dataKey="date"
                        stroke="#9CA3AF"
                        fontSize={12}
                      />
                      <YAxis
                        stroke="#9CA3AF"
                        fontSize={12}
                        tickFormatter={(value) => `$${value}`}
                      />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: '#1F2937',
                          border: '1px solid #374151',
                          borderRadius: '8px',
                          color: '#F9FAFB'
                        }}
                        formatter={(value: number) => [formatPrice(value), 'Price']}
                      />
                      <Line
                        type="monotone"
                        dataKey="price"
                        stroke="#3B82F6"
                        strokeWidth={2}
                        dot={false}
                        activeDot={{ r: 4, fill: '#3B82F6' }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <div className="text-center">
                      <div className="text-gray-400 mb-2">No price data available</div>
                      <div className="text-gray-500 text-sm">
                        Try selecting a different time range
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Chart Stats */}
              {chartData.length > 0 && (
                <div className="grid grid-cols-4 gap-4 mt-6 pt-6 border-t border-white/10">
                  <div className="text-center">
                    <div className="text-sm text-gray-400">Data Points</div>
                    <div className="text-lg font-semibold text-white">{chartData.length}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-sm text-gray-400">Avg Price</div>
                    <div className="text-lg font-semibold text-white">
                      {formatPrice(chartData.reduce((sum, d) => sum + d.price, 0) / chartData.length)}
                    </div>
                  </div>
                  <div className="text-center">
                    <div className="text-sm text-gray-400">Period High</div>
                    <div className="text-lg font-semibold text-green-400">
                      {formatPrice(Math.max(...chartData.map(d => d.high)))}
                    </div>
                  </div>
                  <div className="text-center">
                    <div className="text-sm text-gray-400">Period Low</div>
                    <div className="text-lg font-semibold text-red-400">
                      {formatPrice(Math.min(...chartData.map(d => d.low)))}
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Current Listings */}
            <div className="glass-effect rounded-2xl p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white">Current Listings</h3>
                <span className="text-gray-400 text-sm">
                  {priceHistory?.listings?.length || 0} active listings
                </span>
              </div>

              {priceHistory?.listings && priceHistory.listings.length > 0 ? (
                <div className="space-y-4">
                  {priceHistory.listings.slice(0, 5).map((listing) => (
                    <div
                      key={listing.id}
                      className="flex items-center gap-4 p-4 bg-white/5 rounded-lg hover:bg-white/10 transition-colors"
                    >
                      {listing.image_url && (
                        <img
                          src={listing.image_url}
                          alt={listing.title}
                          className="w-16 h-20 object-cover rounded"
                          onError={(e) => {
                            (e.target as HTMLImageElement).style.display = 'none'
                          }}
                        />
                      )}

                      <div className="flex-1">
                        <h4 className="font-medium text-white mb-1">{listing.title}</h4>
                        <div className="flex items-center gap-4 text-sm text-gray-400">
                          <span className="flex items-center">
                            <Store className="w-4 h-4 mr-1" />
                            {listing.seller}
                          </span>
                          <span>Condition: {listing.condition}</span>
                          <span>Qty: {listing.quantity}</span>
                          <span className="capitalize">
                            {listing.source}
                          </span>
                        </div>
                      </div>

                      <div className="text-right">
                        <div className="text-xl font-bold text-green-400">
                          {formatPrice(listing.price)}
                        </div>
                        <div className="text-xs text-gray-500">
                          {formatDateTime(listing.created_at)}
                        </div>
                      </div>
                    </div>
                  ))}

                  {priceHistory.listings.length > 5 && (
                    <div className="text-center pt-4 border-t border-white/10">
                      <span className="text-gray-400 text-sm">
                        +{priceHistory.listings.length - 5} more listings available
                      </span>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Store className="w-12 h-12 text-gray-600 mx-auto mb-3" />
                  <div className="text-gray-400 mb-2">No Current Listings</div>
                  <div className="text-gray-500 text-sm">
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