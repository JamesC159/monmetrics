import { useState } from 'react'
import { ChevronRight, Package, CreditCard } from 'lucide-react'
import { Link } from 'react-router-dom'
import CardGrid from './CardGrid'
import type { GameCardGroup } from '../types'

interface GameSectionProps {
  group: GameCardGroup
}

export default function GameSection({ group }: GameSectionProps) {
  const [expanded, setExpanded] = useState(false)

  // Show only 2 rows (12 items) by default
  const displayedCards = expanded ? group.cards : group.cards.slice(0, 12)
  const hasMore = group.total_count > 12

  const getIcon = () => {
    return group.category === 'sealed' ? (
      <Package className='w-5 h-5 mr-2' />
    ) : (
      <CreditCard className='w-5 h-5 mr-2' />
    )
  }

  return (
    <div className='glass-effect rounded-2xl p-6'>
      <div className='flex items-center justify-between mb-6'>
        <div className='flex items-center'>
          {getIcon()}
          <h3 className='text-2xl font-bold text-white'>{group.game}</h3>
          <span className='ml-3 px-3 py-1 bg-white/10 text-gray-300 text-sm rounded-full'>
            {group.total_count} {group.category === 'sealed' ? 'products' : 'cards'}
          </span>
        </div>

        {hasMore && !expanded && (
          <Link
            to={`/search?game=${encodeURIComponent(group.game)}&category=${group.category}`}
            className='flex items-center text-blue-400 hover:text-blue-300 transition-colors text-sm font-medium'
          >
            View All
            <ChevronRight className='w-4 h-4 ml-1' />
          </Link>
        )}
      </div>

      <CardGrid cards={displayedCards} columns={6} showAllDetails={false} />

      {hasMore && !expanded && (
        <div className='mt-6 text-center'>
          <button
            onClick={() => setExpanded(true)}
            className='button-secondary inline-flex items-center'
          >
            Show More from {group.game}
            <ChevronRight className='w-4 h-4 ml-2' />
          </button>
        </div>
      )}

      {expanded && (
        <div className='mt-6 flex gap-3 justify-center'>
          <button onClick={() => setExpanded(false)} className='button-secondary'>
            Show Less
          </button>
          <Link
            to={`/search?game=${encodeURIComponent(group.game)}&category=${group.category}`}
            className='button-primary inline-flex items-center'
          >
            View All in Search
            <ChevronRight className='w-4 h-4 ml-2' />
          </Link>
        </div>
      )}
    </div>
  )
}
