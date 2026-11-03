import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft, Loader } from 'lucide-react'
import { Link, useSearchParams } from 'react-router-dom'
import { apiClient } from '../utils/api'
import FeaturedCarousel from '../components/FeaturedCarousel'
import GameSection from '../components/GameSection'
import CardGrid from '../components/CardGrid'
import SearchBar from '../components/catalog/SearchBar'
import CatalogFilters, { type CatalogFilterValues } from '../components/catalog/CatalogFilters'
import type { CardSearchSort, FeaturedContent, GameCardGroup, SearchResult } from '../types'

// Any of these in the URL switches the page from discovery mode to results mode.
const FILTER_KEYS = ['q', 'game', 'category', 'set', 'product_type', 'rarity', 'priced'] as const
const RESULTS_PER_PAGE = 24

export default function Search() {
  const [searchParams, setSearchParams] = useSearchParams()

  const filters: CatalogFilterValues = useMemo(
    () => ({
      game: searchParams.get('game') ?? '',
      category: searchParams.get('category') ?? '',
      set: searchParams.get('set') ?? '',
      product_type: searchParams.get('product_type') ?? '',
      rarity: searchParams.get('rarity') ?? '',
      priced: searchParams.get('priced') === 'true',
      sort: (searchParams.get('sort') as CardSearchSort) || 'relevance',
    }),
    [searchParams],
  )
  const query = searchParams.get('q') ?? ''
  const page = Number(searchParams.get('page')) || 1
  const isSearching = FILTER_KEYS.some((key) => searchParams.get(key))

  const patchParams = useCallback(
    (patch: Record<string, string | boolean | number | undefined>) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          Object.entries(patch).forEach(([key, value]) => {
            if (value === undefined || value === '' || value === false) next.delete(key)
            else next.set(key, String(value))
          })
          // Any filter change invalidates the current page offset.
          if (!('page' in patch)) next.delete('page')
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )

  return (
    <div className='min-h-screen bg-dark-900'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <div className='mb-8'>
          <Link
            to='/'
            className='inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4'
          >
            <ArrowLeft className='w-4 h-4 mr-2' />
            Back to Home
          </Link>

          <h1 className='text-4xl font-bold text-white mb-2'>Discover Trading Cards</h1>
          <p className='text-gray-400 mb-6'>
            Search the catalog or explore featured products across all games
          </p>

          <div className='space-y-4'>
            <SearchBar value={query} onChange={(q) => patchParams({ q })} />
            <CatalogFilters
              values={filters}
              onChange={patchParams}
              onReset={() => setSearchParams({}, { replace: true })}
            />
            <Link
              to='/browse'
              className='inline-block text-sm text-primary-400 hover:text-primary-300'
            >
              Browse by game and set &rarr;
            </Link>
          </div>
        </div>

        {isSearching ? (
          <SearchResults
            query={query}
            filters={filters}
            page={page}
            onPageChange={(p) => patchParams({ page: p > 1 ? p : undefined })}
          />
        ) : (
          <DiscoverSections />
        )}
      </div>
    </div>
  )
}

interface SearchResultsProps {
  query: string
  filters: CatalogFilterValues
  page: number
  onPageChange: (page: number) => void
}

function SearchResults({ query, filters, page, onPageChange }: SearchResultsProps) {
  const [result, setResult] = useState<SearchResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    apiClient
      .searchCards({
        q: query || undefined,
        game: filters.game || undefined,
        category: filters.category || undefined,
        set: filters.set || undefined,
        product_type: filters.product_type || undefined,
        rarity: filters.rarity || undefined,
        priced: filters.priced || undefined,
        sort: filters.sort,
        page,
        limit: RESULTS_PER_PAGE,
      })
      .then((res) => !cancelled && setResult(res))
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Search failed')
        setResult(null)
      })
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [query, filters, page])

  if (loading && !result) {
    return (
      <div className='flex items-center justify-center py-20 text-gray-400'>
        <Loader className='w-6 h-6 animate-spin mr-3' />
        Searching catalog...
      </div>
    )
  }

  if (error) {
    return (
      <div className='glass-effect rounded-lg p-6 border-l-4 border-red-500'>
        <div className='text-red-400 font-medium'>Search Error</div>
        <div className='text-gray-300'>{error}</div>
      </div>
    )
  }

  if (!result) return null

  return (
    <div>
      <p className='text-gray-400 text-sm mb-4'>
        {result.total.toLocaleString()} result{result.total === 1 ? '' : 's'}
        {result.total_pages > 1 && ` · page ${result.page} of ${result.total_pages}`}
      </p>

      <CardGrid cards={result.cards} columns={6} />

      {result.total_pages > 1 && (
        <div className='flex items-center justify-center gap-3 mt-8'>
          <button
            type='button'
            className='btn-ghost'
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
          >
            Previous
          </button>
          <span className='text-gray-400 text-sm'>
            {result.page} / {result.total_pages}
          </span>
          <button
            type='button'
            className='btn-ghost'
            disabled={page >= result.total_pages}
            onClick={() => onPageChange(page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}

function DiscoverSections() {
  const [featuredContent, setFeaturedContent] = useState<FeaturedContent[]>([])
  const [cardGroups, setCardGroups] = useState<GameCardGroup[]>([])
  const [sealedGroups, setSealedGroups] = useState<GameCardGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [featured, cards, sealed] = await Promise.all([
        apiClient.getFeaturedContent().catch(() => []),
        apiClient.getCardsByGame().catch(() => []),
        apiClient.getSealedByGame().catch(() => []),
      ])
      setFeaturedContent(featured)
      setCardGroups(cards)
      setSealedGroups(sealed)
    } catch (err) {
      console.error('Error loading search page:', err)
      setError(err instanceof Error ? err.message : 'Failed to load content')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  if (loading) {
    return (
      <div className='flex items-center justify-center py-20 text-gray-400'>
        <Loader className='w-6 h-6 animate-spin mr-3' />
        Fetching the latest cards and deals
      </div>
    )
  }

  return (
    <>
      {error && (
        <div className='glass-effect rounded-lg p-6 mb-8 border-l-4 border-red-500'>
          <div className='text-red-400 font-medium'>Error Loading Content</div>
          <div className='text-gray-300'>{error}</div>
          <button onClick={load} className='btn-ghost mt-4'>
            Try Again
          </button>
        </div>
      )}

      {featuredContent.length > 0 && (
        <div className='mb-12'>
          <FeaturedCarousel items={featuredContent} autoPlay={true} interval={5000} />
        </div>
      )}

      {cardGroups.length > 0 && (
        <div className='mb-12'>
          <div className='mb-6'>
            <h2 className='text-3xl font-bold text-white mb-2'>Cards</h2>
            <p className='text-gray-400'>Browse popular cards organized by game</p>
          </div>
          <div className='space-y-8'>
            {cardGroups.map((group) => (
              <GameSection key={`${group.game}-${group.category}`} group={group} />
            ))}
          </div>
        </div>
      )}

      {sealedGroups.length > 0 && (
        <div className='mb-12'>
          <div className='mb-6'>
            <h2 className='text-3xl font-bold text-white mb-2'>Sealed Products</h2>
            <p className='text-gray-400'>Browse popular sealed products organized by game</p>
          </div>
          <div className='space-y-8'>
            {sealedGroups.map((group) => (
              <GameSection key={`${group.game}-${group.category}`} group={group} />
            ))}
          </div>
        </div>
      )}

      {!error &&
        featuredContent.length === 0 &&
        cardGroups.length === 0 &&
        sealedGroups.length === 0 && (
          <div className='glass-effect rounded-2xl p-16 text-center'>
            <h3 className='text-xl font-semibold text-gray-400 mb-2'>No Content Available</h3>
            <p className='text-gray-500 mb-6'>
              Run the catalog sync and scraper to populate cards and price history
            </p>
            <button onClick={load} className='btn-primary'>
              Refresh
            </button>
          </div>
        )}
    </>
  )
}
