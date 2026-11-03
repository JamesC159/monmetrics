import { useEffect, useState } from 'react'
import { ArrowLeft, Loader } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'
import { apiClient } from '../utils/api'
import type { GameSummary, SetSummary } from '../types'

const PLACEHOLDER = 'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'

/** /browse — game index; /browse/:game — set index for that game. */
export default function Browse() {
  const { game } = useParams<{ game?: string }>()
  const [games, setGames] = useState<GameSummary[]>([])
  const [sets, setSets] = useState<SetSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    const request = game ? apiClient.getSets(game) : apiClient.getGames()
    request
      .then((res) => {
        if (cancelled) return
        if (game) setSets(res as SetSummary[])
        else setGames(res as GameSummary[])
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to load catalog')
      })
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [game])

  return (
    <div className='min-h-screen bg-dark-900'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        <Link
          to={game ? '/browse' : '/search'}
          className='inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4'
        >
          <ArrowLeft className='w-4 h-4 mr-2' />
          {game ? 'All games' : 'Back to discover'}
        </Link>

        <h1 className='text-4xl font-bold text-white mb-2'>{game ?? 'Browse by Game'}</h1>
        <p className='text-gray-400 mb-8'>
          {game
            ? 'Pick a set to see its products and eBay sales history'
            : 'Every game in the catalog, with how many products already have sales data'}
        </p>

        {error && (
          <div className='glass-effect rounded-lg p-6 mb-8 border-l-4 border-red-500'>
            <div className='text-red-400 font-medium'>Error Loading Catalog</div>
            <div className='text-gray-300'>{error}</div>
          </div>
        )}

        {loading ? (
          <div className='flex items-center justify-center py-20 text-gray-400'>
            <Loader className='w-6 h-6 animate-spin mr-3' />
            Loading catalog...
          </div>
        ) : game ? (
          <SetGrid game={game} sets={sets} />
        ) : (
          <GameGrid games={games} />
        )}
      </div>
    </div>
  )
}

function GameGrid({ games }: { games: GameSummary[] }) {
  if (games.length === 0) {
    return <EmptyCatalog />
  }
  return (
    <div className='grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4'>
      {games.map((g) => (
        <Link
          key={g.game}
          to={`/browse/${encodeURIComponent(g.game)}`}
          className='glass-effect rounded-xl overflow-hidden hover:bg-white/10 transition-colors'
        >
          <img
            src={g.image_url || PLACEHOLDER}
            alt={g.game}
            className='w-full h-32 object-cover'
            onError={(e) => {
              ;(e.target as HTMLImageElement).src = PLACEHOLDER
            }}
          />
          <div className='p-4'>
            <h2 className='text-white font-semibold mb-1'>{g.game}</h2>
            <p className='text-sm text-gray-400'>
              {g.card_count.toLocaleString()} cards · {g.sealed_count.toLocaleString()} sealed
            </p>
            <p className='text-xs text-gray-500 mt-1'>
              {g.priced_count.toLocaleString()} with sales data
            </p>
          </div>
        </Link>
      ))}
    </div>
  )
}

function SetGrid({ game, sets }: { game: string; sets: SetSummary[] }) {
  if (sets.length === 0) {
    return <EmptyCatalog />
  }
  return (
    <div className='grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4'>
      {sets.map((s) => (
        <Link
          key={`${s.game}-${s.set}`}
          to={`/search?game=${encodeURIComponent(game)}&set=${encodeURIComponent(s.set)}`}
          className='glass-effect rounded-xl overflow-hidden hover:bg-white/10 transition-colors'
        >
          <img
            src={s.image_url || PLACEHOLDER}
            alt={s.set}
            className='w-full h-32 object-cover'
            onError={(e) => {
              ;(e.target as HTMLImageElement).src = PLACEHOLDER
            }}
          />
          <div className='p-4'>
            <h2 className='text-white font-semibold mb-1 line-clamp-2'>{s.set}</h2>
            <p className='text-sm text-gray-400'>{s.product_count.toLocaleString()} products</p>
            <p className='text-xs text-gray-500 mt-1'>
              {s.priced_count > 0
                ? `${s.priced_count.toLocaleString()} with sales data`
                : 'No sales data yet'}
            </p>
          </div>
        </Link>
      ))}
    </div>
  )
}

function EmptyCatalog() {
  return (
    <div className='glass-effect rounded-2xl p-16 text-center'>
      <h3 className='text-xl font-semibold text-gray-400 mb-2'>Catalog is empty</h3>
      <p className='text-gray-500'>Run a catalog sync to ingest products.</p>
    </div>
  )
}
