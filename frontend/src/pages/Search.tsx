import { useState, useEffect } from 'react'
import { ArrowLeft, Loader } from 'lucide-react'
import { Link } from 'react-router-dom'
import { apiClient } from '../utils/api'
import FeaturedCarousel from '../components/FeaturedCarousel'
import GameSection from '../components/GameSection'
import type { FeaturedContent, GameCardGroup } from '../types'

export default function Search() {
  const [featuredContent, setFeaturedContent] = useState<FeaturedContent[]>([])
  const [cardGroups, setCardGroups] = useState<GameCardGroup[]>([])
  const [sealedGroups, setSealedGroups] = useState<GameCardGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadSearchPageData()
  }, [])

  const loadSearchPageData = async () => {
    setLoading(true)
    setError(null)

    try {
      // Load all data in parallel
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
  }

  if (loading) {
    return (
      <div className='min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center'>
        <div className='text-center'>
          <Loader className='w-12 h-12 text-blue-500 animate-spin mx-auto mb-4' />
          <h3 className='text-xl font-semibold text-white mb-2'>Loading...</h3>
          <p className='text-gray-400'>Fetching the latest cards and deals</p>
        </div>
      </div>
    )
  }

  return (
    <div className='min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900'>
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
          <p className='text-gray-400'>
            Explore featured products, market movers, and popular cards across all games
          </p>
        </div>

        {error && (
          <div className='glass-effect rounded-lg p-6 mb-8 border-l-4 border-red-500'>
            <div className='text-red-400 font-medium'>Error Loading Content</div>
            <div className='text-gray-300'>{error}</div>
            <button onClick={loadSearchPageData} className='button-secondary mt-4'>
              Try Again
            </button>
          </div>
        )}

        {/* Featured Carousel */}
        {featuredContent.length > 0 && (
          <div className='mb-12'>
            <FeaturedCarousel items={featuredContent} autoPlay={true} interval={5000} />
          </div>
        )}

        {/* Cards Section */}
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

        {/* Sealed Products Section */}
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

        {/* Empty State */}
        {!loading &&
          !error &&
          featuredContent.length === 0 &&
          cardGroups.length === 0 &&
          sealedGroups.length === 0 && (
            <div className='glass-effect rounded-2xl p-16 text-center'>
              <h3 className='text-xl font-semibold text-gray-400 mb-2'>No Content Available</h3>
              <p className='text-gray-500 mb-6'>
                Check back soon for featured products and trending cards
              </p>
              <button onClick={loadSearchPageData} className='button-primary'>
                Refresh
              </button>
            </div>
          )}
      </div>
    </div>
  )
}
