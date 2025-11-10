import { Link } from 'react-router-dom'
import { TrendingUp, TrendingDown, ExternalLink } from 'lucide-react'
import type { Card } from '../types'

interface CardGridProps {
  cards: Card[]
  columns?: 3 | 4 | 6
  showAllDetails?: boolean
}

export default function CardGrid({ cards, columns = 6, showAllDetails = true }: CardGridProps) {
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

  const getGridClass = () => {
    switch (columns) {
      case 3:
        return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3'
      case 4:
        return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4'
      case 6:
        return 'grid-cols-2 md:grid-cols-3 lg:grid-cols-6'
      default:
        return 'grid-cols-2 md:grid-cols-3 lg:grid-cols-6'
    }
  }

  if (!cards || cards.length === 0) {
    return (
      <div className='text-center py-12'>
        <p className='text-gray-400'>No cards found</p>
      </div>
    )
  }

  return (
    <div className={`grid ${getGridClass()} gap-4`}>
      {cards.map((card) => (
        <Link
          key={card.id}
          to={`/card/${card.id}`}
          className='group glass-effect rounded-xl overflow-hidden hover:bg-white/10 transition-all duration-300 transform hover:-translate-y-1'
        >
          <div className='aspect-[3/4] relative overflow-hidden'>
            <img
              src={card.image_url}
              alt={card.name}
              className='w-full h-full object-cover group-hover:scale-105 transition-transform duration-300'
              onError={(e) => {
                ;(e.target as HTMLImageElement).src =
                  'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'
              }}
            />
            <div className='absolute top-2 left-2'>
              <span className='px-2 py-1 bg-black/70 text-white text-xs rounded-full'>
                {card.game}
              </span>
            </div>
            <div className='absolute top-2 right-2'>
              <span
                className={`px-2 py-1 text-xs rounded-full ${
                  card.category === 'sealed'
                    ? 'bg-purple-600/80 text-purple-100'
                    : 'bg-blue-600/80 text-blue-100'
                }`}
              >
                {card.category === 'sealed' ? 'Sealed' : 'Card'}
              </span>
            </div>
          </div>

          <div className='p-4'>
            <h3 className='text-sm font-bold text-white mb-2 line-clamp-2 group-hover:text-blue-300 transition-colors min-h-[2.5rem]'>
              {card.name}
            </h3>

            {showAllDetails && (
              <>
                <div className='space-y-1 mb-3 text-xs'>
                  <div className='flex items-center justify-between'>
                    <span className='text-gray-400'>Set:</span>
                    <span className='text-gray-300 truncate ml-2'>{card.set}</span>
                  </div>
                  {card.rarity && (
                    <div className='flex items-center justify-between'>
                      <span className='text-gray-400'>Rarity:</span>
                      <span className='text-gray-300'>{card.rarity}</span>
                    </div>
                  )}
                </div>
              </>
            )}

            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <span className='text-gray-400 text-xs'>Price:</span>
                <span className='text-base font-bold text-green-400'>
                  {formatPrice(card.current_price)}
                </span>
              </div>

              {showAllDetails && (
                <div className='grid grid-cols-2 gap-2 text-xs'>
                  <div className='flex items-center gap-1'>
                    <TrendingUp className='w-3 h-3 text-green-400' />
                    <div>
                      <div className='text-gray-400'>ATH</div>
                      <div className='text-green-400 font-medium'>
                        {formatPrice(card.all_time_high)}
                      </div>
                    </div>
                  </div>
                  <div className='flex items-center gap-1'>
                    <TrendingDown className='w-3 h-3 text-red-400' />
                    <div>
                      <div className='text-gray-400'>ATL</div>
                      <div className='text-red-400 font-medium'>
                        {formatPrice(card.all_time_low)}
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {showAllDetails && card.tags && card.tags.length > 0 && (
                <div className='flex flex-wrap gap-1 mt-2'>
                  {card.tags.slice(0, 2).map((tag) => (
                    <span
                      key={tag}
                      className='px-2 py-0.5 bg-white/10 text-gray-300 text-xs rounded'
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
            </div>

            {showAllDetails && (
              <div className='flex items-center justify-between mt-3 pt-3 border-t border-white/10'>
                <span className='text-gray-400 text-xs'>Updated {formatDate(card.updated_at)}</span>
                <ExternalLink className='w-3 h-3 text-gray-400 group-hover:text-blue-400 transition-colors' />
              </div>
            )}
          </div>
        </Link>
      ))}
    </div>
  )
}
