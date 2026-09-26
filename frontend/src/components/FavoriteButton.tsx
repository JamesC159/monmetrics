import { Star } from 'lucide-react'
import { useFavorites } from '@/context/FavoritesContext'
import { cn } from '@/utils/cn'

interface FavoriteButtonProps {
  cardId: string
  variant?: 'icon' | 'button'
  className?: string
}

export default function FavoriteButton({
  cardId,
  variant = 'icon',
  className,
}: FavoriteButtonProps) {
  const { isFavorite, toggleFavorite } = useFavorites()
  const active = isFavorite(cardId)
  const label = active ? 'Remove from watchlist' : 'Add to watchlist'

  const onClick = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    toggleFavorite(cardId)
  }

  if (variant === 'button') {
    return (
      <button
        type='button'
        onClick={onClick}
        aria-pressed={active}
        className={cn('btn-ghost py-2 px-4 flex items-center', className)}
      >
        <Star className={cn('w-4 h-4 mr-2', active ? 'fill-yellow-400 text-yellow-400' : '')} />
        {active ? 'Watching' : 'Watch'}
      </button>
    )
  }

  return (
    <button
      type='button'
      onClick={onClick}
      aria-pressed={active}
      aria-label={label}
      title={label}
      className={cn(
        'p-1.5 rounded-full bg-black/60 hover:bg-black/80 transition-colors',
        className,
      )}
    >
      <Star className={cn('w-4 h-4', active ? 'fill-yellow-400 text-yellow-400' : 'text-white')} />
    </button>
  )
}
