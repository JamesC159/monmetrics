import { useEffect, useState } from 'react'
import type { ChartTimeRange, PriceSummary } from '@/types'
import { apiClient } from '@/utils/api'
import { formatPrice } from '@/utils/formatters'

interface PriceSummaryStatsProps {
  cardId: string
  grade: string
  source?: string
}

const WINDOWS: { key: ChartTimeRange; label: string }[] = [
  { key: '1d', label: '24H' },
  { key: '7d', label: '7D' },
  { key: '30d', label: '30D' },
  { key: '90d', label: '90D' },
  { key: '1y', label: '1Y' },
  { key: '5y', label: '5Y' },
]

/** Server-computed rolling window stats over the scraped sold listings. */
export default function PriceSummaryStats({
  cardId,
  grade,
  source = 'ebay',
}: PriceSummaryStatsProps) {
  const [summary, setSummary] = useState<PriceSummary | null>(null)

  useEffect(() => {
    let cancelled = false
    apiClient
      .getPriceSummary(cardId, grade, source)
      .then((res) => !cancelled && setSummary(res[0] ?? null))
      .catch(() => !cancelled && setSummary(null))
    return () => {
      cancelled = true
    }
  }, [cardId, grade, source])

  if (!summary) return null

  const populated = WINDOWS.filter((w) => (summary.windows[w.key]?.count ?? 0) > 0)
  if (populated.length === 0) return null

  return (
    <div className='grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3'>
      {populated.map((w) => {
        const stats = summary.windows[w.key]
        const up = stats.change_pct >= 0
        return (
          <div key={w.key} className='bg-white/5 rounded-lg p-3'>
            <div className='text-xs text-gray-400 mb-1'>{w.label}</div>
            <div className='text-white font-semibold'>{formatPrice(stats.median)}</div>
            <div className={`text-xs ${up ? 'text-green-400' : 'text-red-400'}`}>
              {up ? '+' : ''}
              {stats.change_pct.toFixed(1)}%
            </div>
            <div className='text-xs text-gray-500 mt-1'>
              {formatPrice(stats.min)}–{formatPrice(stats.max)} · {stats.count} sold
            </div>
          </div>
        )
      })}
    </div>
  )
}
