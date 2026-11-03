import { useEffect, useState } from 'react'
import type { CardSearchSort, GameSummary, SetSummary } from '@/types'
import { apiClient } from '@/utils/api'

export interface CatalogFilterValues {
  game: string
  category: string
  set: string
  product_type: string
  rarity: string
  priced: boolean
  sort: CardSearchSort
}

interface CatalogFiltersProps {
  values: CatalogFilterValues
  onChange: (patch: Partial<CatalogFilterValues>) => void
  onReset: () => void
}

const SORTS: { value: CardSearchSort; label: string }[] = [
  { value: 'relevance', label: 'Most relevant' },
  { value: 'price_desc', label: 'Price: high to low' },
  { value: 'price_asc', label: 'Price: low to high' },
  { value: 'name', label: 'Name A-Z' },
  { value: 'recent', label: 'Recently updated' },
]

// Mirrors the product_type values produced by the catalog normalizer.
const PRODUCT_TYPES = [
  'single',
  'booster_box',
  'booster_pack',
  'elite_trainer_box',
  'bundle',
  'collection_box',
  'build_and_battle',
  'blister',
  'deck',
  'tin',
  'case',
  'box',
  'sealed_other',
]

const humanize = (value: string) =>
  value.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())

const selectClass =
  'px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500'

export default function CatalogFilters({ values, onChange, onReset }: CatalogFiltersProps) {
  const [games, setGames] = useState<GameSummary[]>([])
  const [sets, setSets] = useState<SetSummary[]>([])

  useEffect(() => {
    apiClient
      .getGames()
      .then(setGames)
      .catch(() => setGames([]))
  }, [])

  useEffect(() => {
    if (!values.game) {
      setSets([])
      return
    }
    let cancelled = false
    apiClient
      .getSets(values.game, values.category || undefined)
      .then((res) => !cancelled && setSets(res))
      .catch(() => !cancelled && setSets([]))
    return () => {
      cancelled = true
    }
  }, [values.game, values.category])

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <select
        value={values.game}
        onChange={(e) => onChange({ game: e.target.value, set: '' })}
        aria-label='Game'
        className={selectClass}
      >
        <option value='' className='bg-slate-800'>
          All games
        </option>
        {games.map((g) => (
          <option key={g.game} value={g.game} className='bg-slate-800'>
            {g.game}
          </option>
        ))}
      </select>

      <select
        value={values.category}
        onChange={(e) => onChange({ category: e.target.value })}
        aria-label='Category'
        className={selectClass}
      >
        <option value='' className='bg-slate-800'>
          Cards &amp; sealed
        </option>
        <option value='card' className='bg-slate-800'>
          Cards
        </option>
        <option value='sealed' className='bg-slate-800'>
          Sealed
        </option>
      </select>

      <select
        value={values.set}
        onChange={(e) => onChange({ set: e.target.value })}
        aria-label='Set'
        disabled={!values.game}
        className={`${selectClass} disabled:opacity-40`}
      >
        <option value='' className='bg-slate-800'>
          All sets
        </option>
        {sets.map((s) => (
          <option key={`${s.game}-${s.set}`} value={s.set} className='bg-slate-800'>
            {s.set} ({s.product_count})
          </option>
        ))}
      </select>

      <select
        value={values.product_type}
        onChange={(e) => onChange({ product_type: e.target.value })}
        aria-label='Product type'
        className={selectClass}
      >
        <option value='' className='bg-slate-800'>
          Any product type
        </option>
        {PRODUCT_TYPES.map((t) => (
          <option key={t} value={t} className='bg-slate-800'>
            {humanize(t)}
          </option>
        ))}
      </select>

      <input
        type='text'
        value={values.rarity}
        onChange={(e) => onChange({ rarity: e.target.value })}
        placeholder='Rarity'
        aria-label='Rarity'
        className={`${selectClass} w-32 placeholder-gray-500`}
      />

      <select
        value={values.sort}
        onChange={(e) => onChange({ sort: e.target.value as CardSearchSort })}
        aria-label='Sort order'
        className={selectClass}
      >
        {SORTS.map((s) => (
          <option key={s.value} value={s.value} className='bg-slate-800'>
            {s.label}
          </option>
        ))}
      </select>

      <label className='flex items-center gap-2 text-sm text-gray-300'>
        <input
          type='checkbox'
          checked={values.priced}
          onChange={(e) => onChange({ priced: e.target.checked })}
        />
        Has sales data
      </label>

      <button type='button' onClick={onReset} className='btn-ghost text-sm'>
        Reset
      </button>
    </div>
  )
}
