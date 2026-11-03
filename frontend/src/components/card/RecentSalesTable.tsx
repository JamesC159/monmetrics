import { useEffect, useState } from 'react'
import { ExternalLink } from 'lucide-react'
import type { SoldSale } from '@/types'
import { apiClient } from '@/utils/api'
import { formatPrice } from '@/utils/formatters'
import { gradeLabel } from '@/utils/grades'

interface RecentSalesTableProps {
  cardId: string
  grade: string
  source?: string
  pageSize?: number
}

const formatSoldAt = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

/** Individual scraped eBay sales behind the price history. */
export default function RecentSalesTable({
  cardId,
  grade,
  source = 'ebay',
  pageSize = 25,
}: RecentSalesTableProps) {
  const [sales, setSales] = useState<SoldSale[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setPage(1)
  }, [cardId, grade, source])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    apiClient
      .getCardSales(cardId, { grade, source, page, limit: pageSize })
      .then((res) => {
        if (cancelled) return
        setTotal(res.total)
        setSales((prev) => (page === 1 ? res.sales : [...prev, ...res.sales]))
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to load sales')
      })
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [cardId, grade, source, page, pageSize])

  return (
    <div className='card'>
      <div className='flex items-baseline justify-between mb-4'>
        <h2 className='text-xl font-semibold text-white'>Recent eBay Sales</h2>
        <span className='text-sm text-gray-400'>
          {gradeLabel(grade)}
          {total > 0 && ` · ${total.toLocaleString()} sold`}
        </span>
      </div>

      {error && <div className='text-red-400 text-sm mb-3'>{error}</div>}

      {sales.length === 0 && !loading && !error ? (
        <p className='text-gray-400 text-sm'>No completed sales collected for this grade yet.</p>
      ) : (
        <div className='overflow-x-auto'>
          <table className='w-full text-sm'>
            <thead>
              <tr className='text-left text-gray-400 border-b border-white/10'>
                <th className='py-2 pr-4 font-medium'>Date</th>
                <th className='py-2 pr-4 font-medium'>Listing</th>
                <th className='py-2 pr-4 font-medium text-right'>Price</th>
                <th className='py-2 font-medium text-right'>Shipping</th>
              </tr>
            </thead>
            <tbody>
              {sales.map((sale) => (
                <tr key={sale.id} className='border-b border-white/5 last:border-0'>
                  <td className='py-2 pr-4 text-gray-300 whitespace-nowrap'>
                    {formatSoldAt(sale.sold_at)}
                  </td>
                  <td className='py-2 pr-4 text-gray-200'>
                    <div className='flex items-center gap-2'>
                      {sale.url ? (
                        <a
                          href={sale.url}
                          target='_blank'
                          rel='noopener noreferrer'
                          className='hover:text-primary-400 transition-colors line-clamp-1'
                        >
                          {sale.title}
                        </a>
                      ) : (
                        <span className='line-clamp-1'>{sale.title}</span>
                      )}
                      {sale.url && <ExternalLink className='w-3 h-3 text-gray-500 shrink-0' />}
                      {sale.best_offer && (
                        <span className='shrink-0 px-1.5 py-0.5 rounded bg-white/10 text-gray-400 text-xs'>
                          Best offer
                        </span>
                      )}
                    </div>
                  </td>
                  <td className='py-2 pr-4 text-right text-white font-medium whitespace-nowrap'>
                    {formatPrice(sale.price)}
                  </td>
                  <td className='py-2 text-right text-gray-400 whitespace-nowrap'>
                    {sale.shipping > 0 ? formatPrice(sale.shipping) : 'Free'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {loading && <div className='text-gray-400 text-sm mt-3'>Loading sales...</div>}

      {sales.length < total && !loading && (
        <button type='button' className='btn-ghost mt-4' onClick={() => setPage((p) => p + 1)}>
          Load more
        </button>
      )}
    </div>
  )
}
