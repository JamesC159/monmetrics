import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Briefcase, Plus } from 'lucide-react'
import PortfolioSummary from '@/components/portfolio/PortfolioSummary'
import PortfolioItemForm from '@/components/portfolio/PortfolioItemForm'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { formatPercentage, formatPrice } from '@/utils/formatters'
import {
  gainClass,
  ITEM_TYPE_LABELS,
  itemSubtitle,
  PLACEHOLDER_IMAGE,
  signed,
  VALUE_SOURCE_LABELS,
} from '@/utils/portfolio'
import type { PortfolioItemType, PortfolioQuery, PortfolioResponse } from '@/types'

const GAME_FILTERS = ['', 'Pokemon', 'Magic The Gathering', 'Yu-Gi-Oh', 'Other']

export default function PortfolioTab({ onChanged }: { onChanged?: () => void }) {
  const navigate = useNavigate()
  const { addToast } = useToast()
  const [data, setData] = useState<PortfolioResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [query, setQuery] = useState<PortfolioQuery>({ sort: 'date' })
  const [formOpen, setFormOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setData(await apiClient.getPortfolio(query))
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not load portfolio', 'error')
    } finally {
      setLoading(false)
    }
  }, [query, addToast])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className='space-y-6'>
      {data && <PortfolioSummary summary={data.summary} />}

      <div className='glass-effect rounded-2xl p-6'>
        <div className='flex flex-wrap items-center gap-3 mb-4'>
          <h3 className='text-lg font-semibold text-white mr-auto'>Holdings</h3>
          <select
            aria-label='Filter by game'
            className='px-3 py-1.5 bg-white/10 border border-white/20 rounded text-white text-sm'
            value={query.game ?? ''}
            onChange={(e) => setQuery((q) => ({ ...q, game: e.target.value }))}
          >
            {GAME_FILTERS.map((g) => (
              <option key={g} value={g} className='bg-slate-800'>
                {g || 'All games'}
              </option>
            ))}
          </select>
          <select
            aria-label='Filter by type'
            className='px-3 py-1.5 bg-white/10 border border-white/20 rounded text-white text-sm'
            value={query.item_type ?? ''}
            onChange={(e) =>
              setQuery((q) => ({ ...q, item_type: e.target.value as PortfolioItemType | '' }))
            }
          >
            <option value='' className='bg-slate-800'>
              All types
            </option>
            {(Object.keys(ITEM_TYPE_LABELS) as PortfolioItemType[]).map((t) => (
              <option key={t} value={t} className='bg-slate-800'>
                {ITEM_TYPE_LABELS[t]}
              </option>
            ))}
          </select>
          <select
            aria-label='Sort'
            className='px-3 py-1.5 bg-white/10 border border-white/20 rounded text-white text-sm'
            value={query.sort ?? 'date'}
            onChange={(e) =>
              setQuery((q) => ({ ...q, sort: e.target.value as PortfolioQuery['sort'] }))
            }
          >
            <option value='date' className='bg-slate-800'>
              Newest
            </option>
            <option value='value' className='bg-slate-800'>
              Value
            </option>
            <option value='gain' className='bg-slate-800'>
              Gain / loss
            </option>
            <option value='name' className='bg-slate-800'>
              Name
            </option>
          </select>
          <button
            type='button'
            onClick={() => setFormOpen(true)}
            className='btn-primary py-2 px-4 flex items-center'
          >
            <Plus className='w-4 h-4 mr-1' /> Add item
          </button>
        </div>

        {loading && !data ? (
          <p className='text-gray-400 py-8 text-center'>Loading portfolio...</p>
        ) : !data || data.items.length === 0 ? (
          <div className='text-center py-12'>
            <Briefcase className='w-12 h-12 text-gray-600 mx-auto mb-3' />
            <p className='text-gray-400 mb-4'>
              No items match. Add cards, graded slabs, or sealed product you own.
            </p>
            <button
              type='button'
              onClick={() => setFormOpen(true)}
              className='btn-primary py-2 px-4'
            >
              Add your first item
            </button>
          </div>
        ) : (
          <div className='overflow-x-auto'>
            <table className='w-full text-sm'>
              <thead>
                <tr className='text-left text-gray-400 border-b border-white/10'>
                  <th className='py-2 pr-4 font-medium'>Item</th>
                  <th className='py-2 pr-4 font-medium'>Type</th>
                  <th className='py-2 pr-4 font-medium text-right'>Qty</th>
                  <th className='py-2 pr-4 font-medium text-right'>Value</th>
                  <th className='py-2 pr-4 font-medium text-right'>Cost</th>
                  <th className='py-2 font-medium text-right'>Gain / loss</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((item) => (
                  <tr
                    key={item.id}
                    tabIndex={0}
                    onClick={() => navigate(`/portfolio/${item.id}`)}
                    onKeyDown={(e) => e.key === 'Enter' && navigate(`/portfolio/${item.id}`)}
                    className='border-b border-white/5 hover:bg-white/5 cursor-pointer focus:outline-none focus:bg-white/10'
                  >
                    <td className='py-3 pr-4'>
                      <div className='flex items-center gap-3'>
                        <img
                          src={item.display_image_url || PLACEHOLDER_IMAGE}
                          alt=''
                          className='w-10 h-14 object-cover rounded'
                          onError={(e) => ((e.target as HTMLImageElement).src = PLACEHOLDER_IMAGE)}
                        />
                        <div className='min-w-0'>
                          <div className='text-white font-medium truncate'>{item.display_name}</div>
                          <div className='text-xs text-gray-400 truncate'>
                            {item.display_game}
                            {item.display_set ? ` · ${item.display_set}` : ''}
                            {!item.card_id && ' · Custom'}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className='py-3 pr-4 text-gray-300'>{itemSubtitle(item)}</td>
                    <td className='py-3 pr-4 text-right text-gray-300'>{item.quantity}</td>
                    <td className='py-3 pr-4 text-right'>
                      <div className='text-white'>
                        {item.value_source === 'none' ? '—' : formatPrice(item.total_value)}
                      </div>
                      <div className='text-xs text-gray-500'>
                        {VALUE_SOURCE_LABELS[item.value_source]}
                      </div>
                    </td>
                    <td className='py-3 pr-4 text-right text-gray-300'>
                      {formatPrice(item.cost_basis)}
                    </td>
                    <td className={`py-3 text-right ${gainClass(item.gain_loss)}`}>
                      {item.value_source === 'none' ? (
                        <span className='price-neutral'>—</span>
                      ) : (
                        <>
                          <div>{signed(formatPrice(item.gain_loss), item.gain_loss)}</div>
                          <div className='text-xs'>
                            {signed(formatPercentage(item.gain_loss_pct), item.gain_loss_pct)}
                          </div>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <PortfolioItemForm
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          addToast('Item added to portfolio', 'success')
          load()
          onChanged?.()
        }}
      />
    </div>
  )
}
