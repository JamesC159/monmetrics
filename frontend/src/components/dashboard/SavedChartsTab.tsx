import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { BarChart3, Trash2 } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { formatDate } from '@/utils/formatters'
import { indicatorLabel } from '@/utils/indicators'
import { PLACEHOLDER_IMAGE } from '@/utils/portfolio'
import type { SavedChart } from '@/types'

export default function SavedChartsTab({ onChanged }: { onChanged?: () => void }) {
  const navigate = useNavigate()
  const { addToast } = useToast()
  const [charts, setCharts] = useState<SavedChart[] | null>(null)

  useEffect(() => {
    apiClient
      .getSavedCharts()
      .then(setCharts)
      .catch((err) => {
        addToast(err instanceof Error ? err.message : 'Could not load charts', 'error')
        setCharts([])
      })
  }, [addToast])

  const remove = async (chart: SavedChart) => {
    if (!window.confirm(`Delete saved chart "${chart.name}"?`)) return
    try {
      await apiClient.deleteChart(chart.id)
      setCharts((prev) => prev?.filter((c) => c.id !== chart.id) ?? null)
      addToast('Chart deleted', 'success')
      onChanged?.()
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not delete chart', 'error')
    }
  }

  if (!charts) return <p className='text-gray-400 py-8 text-center'>Loading charts...</p>

  if (charts.length === 0) {
    return (
      <div className='glass-effect rounded-2xl text-center py-16'>
        <BarChart3 className='w-12 h-12 text-gray-600 mx-auto mb-3' />
        <p className='text-gray-400 mb-4'>
          No saved charts yet. Open any card, add indicators, and click Save Chart.
        </p>
        <Link to='/search' className='btn-primary py-2 px-4'>
          Search cards
        </Link>
      </div>
    )
  }

  return (
    <div className='grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4'>
      {charts.map((chart) => {
        const open = () => navigate(`/card/${chart.card_id}?chart=${chart.id}`)
        return (
          <div
            key={chart.id}
            role='link'
            tabIndex={0}
            onClick={open}
            onKeyDown={(e) => e.key === 'Enter' && open()}
            className='glass-effect rounded-xl p-4 flex gap-4 cursor-pointer hover:bg-white/10 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500'
          >
            <img
              src={chart.card_image_url || PLACEHOLDER_IMAGE}
              alt=''
              className='w-14 h-20 object-cover rounded'
              onError={(e) => ((e.target as HTMLImageElement).src = PLACEHOLDER_IMAGE)}
            />
            <div className='flex-1 min-w-0'>
              <div className='flex items-start justify-between gap-2'>
                <h4 className='text-white font-semibold truncate'>{chart.name}</h4>
                <button
                  type='button'
                  onClick={(e) => {
                    e.stopPropagation()
                    remove(chart)
                  }}
                  className='p-1 text-gray-400 hover:text-error-400'
                  aria-label={`Delete ${chart.name}`}
                >
                  <Trash2 className='w-4 h-4' />
                </button>
              </div>
              <div className='text-xs text-gray-400 truncate'>
                {chart.card_name ?? 'Unavailable card'}
              </div>
              <div className='flex flex-wrap gap-1 mt-2'>
                <span className='badge badge-gold text-xs'>{chart.time_range.toUpperCase()}</span>
                {chart.source !== 'all' && (
                  <span className='badge badge-cyan text-xs capitalize'>{chart.source}</span>
                )}
                {chart.indicators.map((ind, i) => (
                  <span
                    key={i}
                    className='badge text-xs bg-white/10 text-gray-200'
                    style={{ borderLeft: `3px solid ${ind.color ?? '#22c55e'}` }}
                  >
                    {indicatorLabel(ind)}
                    {!ind.visible && ' (hidden)'}
                  </span>
                ))}
              </div>
              <div className='text-xs text-gray-500 mt-2'>
                Updated {formatDate(chart.updated_at)}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
