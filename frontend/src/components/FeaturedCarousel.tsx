import { useState, useEffect } from 'react'
import { ChevronLeft, ChevronRight, TrendingUp, TrendingDown, ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { FeaturedContent } from '../types'

interface FeaturedCarouselProps {
  items: FeaturedContent[]
  autoPlay?: boolean
  interval?: number
}

export default function FeaturedCarousel({
  items,
  autoPlay = true,
  interval = 5000,
}: FeaturedCarouselProps) {
  const [currentIndex, setCurrentIndex] = useState(0)

  useEffect(() => {
    if (!autoPlay || items.length <= 1) return

    const timer = setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % items.length)
    }, interval)

    return () => clearInterval(timer)
  }, [autoPlay, interval, items.length])

  if (!items || items.length === 0) {
    return null
  }

  const goToPrevious = () => {
    setCurrentIndex((prev) => (prev - 1 + items.length) % items.length)
  }

  const goToNext = () => {
    setCurrentIndex((prev) => (prev + 1) % items.length)
  }

  const currentItem = items[currentIndex]

  const getBadgeColor = (type: FeaturedContent['type']) => {
    switch (type) {
      case 'market_mover':
        return 'bg-yellow-600/80 text-yellow-100'
      case 'product':
        return 'bg-blue-600/80 text-blue-100'
      case 'news':
        return 'bg-purple-600/80 text-purple-100'
      case 'pickup':
        return 'bg-green-600/80 text-green-100'
      case 'sponsored':
        return 'bg-orange-600/80 text-orange-100'
      default:
        return 'bg-gray-600/80 text-gray-100'
    }
  }

  const formatTypeLabel = (type: FeaturedContent['type']) => {
    switch (type) {
      case 'market_mover':
        return 'Market Mover'
      case 'product':
        return 'Featured Product'
      case 'news':
        return 'News'
      case 'pickup':
        return 'Popular Pickup'
      case 'sponsored':
        return 'Sponsored'
      default:
        return type
    }
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  return (
    <div className='relative glass-effect rounded-2xl overflow-hidden'>
      <div className='relative h-96 md:h-[450px]'>
        <img
          src={currentItem.image_url}
          alt={currentItem.title}
          className='w-full h-full object-cover'
          onError={(e) => {
            ;(e.target as HTMLImageElement).src =
              'https://via.placeholder.com/1200x450/374151/9CA3AF?text=Featured+Content'
          }}
        />

        {/* Gradient overlay */}
        <div className='absolute inset-0 bg-gradient-to-t from-black/90 via-black/50 to-transparent' />

        {/* Content */}
        <div className='absolute inset-0 flex flex-col justify-end p-8'>
          <div className='mb-3'>
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium ${getBadgeColor(
                currentItem.type,
              )}`}
            >
              {formatTypeLabel(currentItem.type)}
            </span>
          </div>

          <h2 className='text-3xl md:text-4xl font-bold text-white mb-3 max-w-3xl'>
            {currentItem.title}
          </h2>

          {currentItem.description && (
            <p className='text-lg text-gray-300 mb-4 max-w-2xl line-clamp-2'>
              {currentItem.description}
            </p>
          )}

          {/* Market mover specific info */}
          {currentItem.type === 'market_mover' && currentItem.price_change !== undefined && (
            <div className='flex items-center gap-4 mb-4'>
              <div
                className={`flex items-center gap-2 px-4 py-2 rounded-lg ${
                  currentItem.price_change >= 0
                    ? 'bg-green-600/30 text-green-300'
                    : 'bg-red-600/30 text-red-300'
                }`}
              >
                {currentItem.price_change >= 0 ? (
                  <TrendingUp className='w-5 h-5' />
                ) : (
                  <TrendingDown className='w-5 h-5' />
                )}
                <span className='font-bold text-lg'>
                  {currentItem.price_change >= 0 ? '+' : ''}
                  {currentItem.price_change.toFixed(1)}%
                </span>
              </div>
              {currentItem.price_change_value !== undefined && (
                <div className='text-white'>
                  <span className='text-2xl font-bold'>
                    {formatPrice(Math.abs(currentItem.price_change_value))}
                  </span>
                  <span className='text-gray-400 ml-2'>change</span>
                </div>
              )}
            </div>
          )}

          {/* CTA Button */}
          <div className='flex gap-3'>
            {currentItem.card_id ? (
              <Link
                to={`/card/${currentItem.card_id}`}
                className='button-primary inline-flex items-center'
              >
                View Details
                <ExternalLink className='w-4 h-4 ml-2' />
              </Link>
            ) : currentItem.link ? (
              <a
                href={currentItem.link}
                target='_blank'
                rel='noopener noreferrer'
                className='button-primary inline-flex items-center'
              >
                Learn More
                <ExternalLink className='w-4 h-4 ml-2' />
              </a>
            ) : null}
          </div>
        </div>

        {/* Navigation arrows */}
        {items.length > 1 && (
          <>
            <button
              onClick={goToPrevious}
              className='absolute left-4 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white rounded-full p-3 transition-all'
              aria-label='Previous slide'
            >
              <ChevronLeft className='w-6 h-6' />
            </button>
            <button
              onClick={goToNext}
              className='absolute right-4 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white rounded-full p-3 transition-all'
              aria-label='Next slide'
            >
              <ChevronRight className='w-6 h-6' />
            </button>
          </>
        )}
      </div>

      {/* Indicators */}
      {items.length > 1 && (
        <div className='absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2 z-10'>
          {items.map((_, index) => (
            <button
              key={index}
              onClick={() => setCurrentIndex(index)}
              className={`h-2 rounded-full transition-all ${
                index === currentIndex ? 'w-8 bg-white' : 'w-2 bg-white/50 hover:bg-white/70'
              }`}
              aria-label={`Go to slide ${index + 1}`}
            />
          ))}
        </div>
      )}
    </div>
  )
}
