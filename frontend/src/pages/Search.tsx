import { useState, useEffect, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Search as SearchIcon, Filter, ArrowLeft, Star, TrendingUp, TrendingDown, ExternalLink } from 'lucide-react'
import { apiClient } from '../utils/api'
import type { Card, SearchResult } from '../types'

export default function Search() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get('q') || '')
  const [selectedGame, setSelectedGame] = useState(searchParams.get('game') || '')
  const [selectedCategory, setSelectedCategory] = useState(searchParams.get('category') || '')
  const [results, setResults] = useState<SearchResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showFilters, setShowFilters] = useState(false)

  const games = ['Pokemon', 'Magic The Gathering', 'Yu-Gi-Oh']
  const categories = [
    { value: '', label: 'All' },
    { value: 'card', label: 'Cards' },
    { value: 'sealed', label: 'Sealed Products' }
  ]

  const searchCards = useCallback(async (page = 1) => {
    setLoading(true)
    setError(null)

    try {
      const params: any = { page, limit: 20 }

      if (query.trim()) params.q = query.trim()
      if (selectedGame) params.game = selectedGame
      if (selectedCategory) params.category = selectedCategory

      const result = await apiClient.searchCards(params)
      setResults(result)

      // Update URL params
      const newParams = new URLSearchParams()
      if (query.trim()) newParams.set('q', query.trim())
      if (selectedGame) newParams.set('game', selectedGame)
      if (selectedCategory) newParams.set('category', selectedCategory)
      if (page > 1) newParams.set('page', page.toString())

      setSearchParams(newParams)

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed')
      setResults(null)
    } finally {
      setLoading(false)
    }
  }, [query, selectedGame, selectedCategory, setSearchParams])

  // Load initial results if query params exist
  useEffect(() => {
    if (query || selectedGame || selectedCategory) {
      searchCards()
    }
  }, []) // Only run on mount

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    searchCards()
  }

  const handlePageChange = (page: number) => {
    searchCards(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const clearFilters = () => {
    setQuery('')
    setSelectedGame('')
    setSelectedCategory('')
    setResults(null)
    setSearchParams({})
  }

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

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link
            to="/"
            className="inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Home
          </Link>

          <h1 className="text-4xl font-bold text-white mb-2">Search Trading Cards</h1>
          <p className="text-gray-400">Find and analyze your favorite trading cards</p>
        </div>

        <div className="glass-effect rounded-2xl p-8 mb-8">
          <form onSubmit={handleSearch} className="space-y-6">
            <div className="flex gap-4">
              <div className="flex-1 relative">
                <SearchIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                <input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="w-full pl-12 pr-4 py-4 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-lg"
                  placeholder="Search for Pokemon, Magic, Yu-Gi-Oh cards..."
                />
              </div>
              <button
                type="button"
                onClick={() => setShowFilters(!showFilters)}
                className={`button-secondary flex items-center ${showFilters ? 'bg-blue-600' : ''}`}
              >
                <Filter className="w-5 h-5 mr-2" />
                Filters
              </button>
              <button
                type="submit"
                disabled={loading}
                className="button-primary px-8"
              >
                {loading ? 'Searching...' : 'Search'}
              </button>
            </div>

            {showFilters && (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-6 bg-white/5 rounded-lg border border-white/10">
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">Game</label>
                  <select
                    value={selectedGame}
                    onChange={(e) => setSelectedGame(e.target.value)}
                    className="w-full p-3 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">All Games</option>
                    {games.map((game) => (
                      <option key={game} value={game} className="bg-slate-800">
                        {game}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">Category</label>
                  <select
                    value={selectedCategory}
                    onChange={(e) => setSelectedCategory(e.target.value)}
                    className="w-full p-3 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {categories.map((category) => (
                      <option key={category.value} value={category.value} className="bg-slate-800">
                        {category.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex items-end">
                  <button
                    type="button"
                    onClick={clearFilters}
                    className="w-full p-3 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded-lg transition-colors"
                  >
                    Clear Filters
                  </button>
                </div>
              </div>
            )}
          </form>
        </div>

        {error && (
          <div className="glass-effect rounded-lg p-6 mb-8 border-l-4 border-red-500">
            <div className="text-red-400 font-medium">Search Error</div>
            <div className="text-gray-300">{error}</div>
          </div>
        )}

        {loading && (
          <div className="glass-effect rounded-2xl p-16 text-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
            <h3 className="text-xl font-semibold text-white mb-2">Searching...</h3>
            <p className="text-gray-400">Finding the best cards for you</p>
          </div>
        )}

        {!loading && !results && !error && (
          <div className="glass-effect rounded-2xl p-16 text-center">
            <SearchIcon className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-400 mb-2">Start Your Search</h3>
            <p className="text-gray-500 mb-6">
              Enter a card name, set, or game to find detailed price analysis and charts
            </p>
            <div className="flex flex-wrap gap-2 justify-center">
              {['Charizard', 'Black Lotus', 'Blue-Eyes White Dragon', 'Pikachu'].map((suggestion) => (
                <button
                  key={suggestion}
                  onClick={() => {
                    setQuery(suggestion)
                    searchCards()
                  }}
                  className="px-4 py-2 bg-white/10 hover:bg-white/20 text-gray-300 rounded-lg text-sm transition-colors"
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        )}

        {results && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div className="text-white">
                <span className="text-2xl font-bold">{results.total.toLocaleString()}</span>
                <span className="text-gray-400 ml-2">
                  {results.total === 1 ? 'result' : 'results'} found
                </span>
              </div>
              <div className="text-gray-400 text-sm">
                Page {results.page} of {results.total_pages}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {results.cards.map((card) => (
                <Link
                  key={card.id}
                  to={`/card/${card.id}`}
                  className="group glass-effect rounded-xl overflow-hidden hover:bg-white/10 transition-all duration-300 transform hover:-translate-y-1"
                >
                  <div className="aspect-[3/4] relative overflow-hidden">
                    <img
                      src={card.image_url}
                      alt={card.name}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      onError={(e) => {
                        (e.target as HTMLImageElement).src = 'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'
                      }}
                    />
                    <div className="absolute top-3 left-3">
                      <span className="px-2 py-1 bg-black/70 text-white text-xs rounded-full">
                        {card.game}
                      </span>
                    </div>
                    <div className="absolute top-3 right-3">
                      <span className={`px-2 py-1 text-xs rounded-full ${
                        card.category === 'sealed'
                          ? 'bg-purple-600/80 text-purple-100'
                          : 'bg-blue-600/80 text-blue-100'
                      }`}>
                        {card.category === 'sealed' ? 'Sealed' : 'Card'}
                      </span>
                    </div>
                  </div>

                  <div className="p-6">
                    <h3 className="text-lg font-bold text-white mb-2 line-clamp-2 group-hover:text-blue-300 transition-colors">
                      {card.name}
                    </h3>

                    <div className="space-y-2 mb-4">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-400">Set:</span>
                        <span className="text-gray-300">{card.set}</span>
                      </div>
                      {card.rarity && (
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-gray-400">Rarity:</span>
                          <span className="text-gray-300">{card.rarity}</span>
                        </div>
                      )}
                    </div>

                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-gray-400 text-sm">Current Price:</span>
                        <span className="text-xl font-bold text-green-400">
                          {formatPrice(card.current_price)}
                        </span>
                      </div>

                      <div className="grid grid-cols-2 gap-3 text-sm">
                        <div className="flex items-center">
                          <TrendingUp className="w-4 h-4 text-green-400 mr-1" />
                          <div>
                            <div className="text-gray-400">ATH</div>
                            <div className="text-green-400 font-medium">
                              {formatPrice(card.all_time_high)}
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center">
                          <TrendingDown className="w-4 h-4 text-red-400 mr-1" />
                          <div>
                            <div className="text-gray-400">ATL</div>
                            <div className="text-red-400 font-medium">
                              {formatPrice(card.all_time_low)}
                            </div>
                          </div>
                        </div>
                      </div>

                      {card.tags && card.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-3">
                          {card.tags.slice(0, 3).map((tag) => (
                            <span
                              key={tag}
                              className="px-2 py-1 bg-white/10 text-gray-300 text-xs rounded"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
                      <span className="text-gray-400 text-sm">
                        Updated {formatDate(card.updated_at)}
                      </span>
                      <ExternalLink className="w-4 h-4 text-gray-400 group-hover:text-blue-400 transition-colors" />
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Pagination */}
            {results.total_pages > 1 && (
              <div className="flex items-center justify-center space-x-2 mt-8">
                <button
                  onClick={() => handlePageChange(results.page - 1)}
                  disabled={results.page === 1}
                  className="px-4 py-2 bg-white/10 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors"
                >
                  Previous
                </button>

                {/* Page numbers */}
                {Array.from({ length: Math.min(5, results.total_pages) }, (_, i) => {
                  const pageNum = Math.max(1, Math.min(results.total_pages - 4, results.page - 2)) + i
                  if (pageNum > results.total_pages) return null

                  return (
                    <button
                      key={pageNum}
                      onClick={() => handlePageChange(pageNum)}
                      className={`px-4 py-2 rounded-lg transition-colors ${
                        pageNum === results.page
                          ? 'bg-blue-600 text-white'
                          : 'bg-white/10 text-white hover:bg-white/20'
                      }`}
                    >
                      {pageNum}
                    </button>
                  )
                })}

                <button
                  onClick={() => handlePageChange(results.page + 1)}
                  disabled={results.page === results.total_pages}
                  className="px-4 py-2 bg-white/10 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}