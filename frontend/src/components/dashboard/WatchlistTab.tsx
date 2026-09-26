import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Star, Trash2 } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useFavorites } from '@/context/FavoritesContext'
import { useToast } from '@/context/ToastContext'
import { formatPercentage, formatPrice } from '@/utils/formatters'
import { gainClass, PLACEHOLDER_IMAGE, signed } from '@/utils/portfolio'
import type { FavoriteView } from '@/types'

export default function WatchlistTab({ onChanged }: { onChanged?: () => void }) {
  const { toggleFavorite, favoriteIds } = useFavorites()
  const { addToast } = useToast()
  const [items, setItems] = useState<FavoriteView[] | null>(null)

  useEffect(() => {
    apiClient
      .getFavorites()
      .then(setItems)
      .catch((err) => {
        addToast(err instanceof Error ? err.message : 'Could not load watchlist', 'error')
        setItems([])
      })
  }, [addToast])

  // Keep the list in sync when a favorite is removed elsewhere (context is the source of truth)
  const visible = items?.filter((f) => favoriteIds.has(f.card_id)) ?? null

  const remove = async (cardId: string) => {
    await toggleFavorite(cardId)
    onChanged?.()
  }

  if (!visible) return <p className='text-gray-400 py-8 text-center'>Loading watchlist...</p>

  if (visible.length === 0) {
    return (
      <div className='glass-effect rounded-2xl text-center py-16'>
        <Star className='w-12 h-12 text-gray-600 mx-auto mb-3' />
        <p className='text-gray-400 mb-4'>
          Your watchlist is empty. Star cards and sealed products to track them here.
        </p>
        <Link to='/search' className='btn-primary py-2 px-4'>
          Browse cards
        </Link>
      </div>
    )
  }

  return (
    <div className='glass-effect rounded-2xl p-6'>
      <h3 className='text-lg font-semibold text-white mb-4'>Watchlist</h3>
      <ul className='divide-y divide-white/5'>
        {visible.map((f) => (
          <li key={f.id} className='flex items-center gap-4 py-3'>
            <Link
              to={`/card/${f.card_id}`}
              className='flex items-center gap-4 flex-1 min-w-0 group'
            >
              <img
                src={f.card?.image_url || PLACEHOLDER_IMAGE}
                alt=''
                className='w-10 h-14 object-cover rounded'
                onError={(e) => ((e.target as HTMLImageElement).src = PLACEHOLDER_IMAGE)}
              />
              <div className='min-w-0'>
                <div className='text-white font-medium truncate group-hover:text-primary-300'>
                  {f.card?.name ?? 'Unavailable item'}
                </div>
                <div className='text-xs text-gray-400 truncate'>
                  {f.card
                    ? `${f.card.game} · ${f.card.set} · ${f.card.category === 'sealed' ? 'Sealed' : 'Card'}`
                    : ''}
                </div>
              </div>
            </Link>
            <div className='text-right w-28'>
              <div className='text-white font-semibold'>
                {f.card ? formatPrice(f.card.current_price) : '—'}
              </div>
            </div>
            <Change label='7D' pct={f.change_7d} value={f.change_7d_value} />
            <Change label='30D' pct={f.change_30d} value={f.change_30d_value} />
            <button
              type='button'
              onClick={() => remove(f.card_id)}
              className='p-2 text-gray-400 hover:text-error-400'
              aria-label={`Remove ${f.card?.name ?? 'item'} from watchlist`}
            >
              <Trash2 className='w-4 h-4' />
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}

function Change({ label, pct, value }: { label: string; pct: number; value: number }) {
  return (
    <div className='text-right w-24 hidden sm:block'>
      <div className='text-xs text-gray-500'>{label}</div>
      <div className={`text-sm ${gainClass(pct)}`}>{signed(formatPercentage(pct), pct)}</div>
      <div className={`text-xs ${gainClass(value)}`}>{signed(formatPrice(value), value)}</div>
    </div>
  )
}
