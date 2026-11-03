import type { CardPrinting } from '@/types'
import { cn } from '@/utils/cn'
import { formatPrice } from '@/utils/formatters'
import { gradeLabel } from '@/utils/grades'

interface PrintingListProps {
  printings: CardPrinting[]
  /** Show at most this many rows, then a "+N more" line. */
  limit?: number
  /** The chart's current series, highlighted when it is one of these printings. */
  selected?: string
  /** Makes rows buttons that pick a chart series. Leave unset inside a link. */
  onSelect?: (series: string) => void
  /** Which printings have a chart to switch to. Defaults to all of them. */
  canSelect?: (series: string) => boolean
  className?: string
}

export default function PrintingList({
  printings,
  limit,
  selected,
  onSelect,
  canSelect = () => true,
  className,
}: PrintingListProps) {
  const shown = limit ? printings.slice(0, limit) : printings
  const hidden = printings.length - shown.length

  return (
    <ul className={cn('space-y-1', className)} aria-label='Printings'>
      {shown.map((printing) => {
        const content = (
          <>
            <span className='truncate'>{gradeLabel(printing.series)}</span>
            <span
              className={cn(
                'shrink-0 tabular-nums font-semibold',
                printing.market_price ? 'text-green-400' : 'text-gray-500',
              )}
            >
              {printing.market_price ? formatPrice(printing.market_price) : '—'}
            </span>
          </>
        )
        const active = printing.series === selected
        return (
          <li key={printing.series}>
            {onSelect && canSelect(printing.series) ? (
              <button
                type='button'
                onClick={() => onSelect(printing.series)}
                aria-pressed={active}
                className={cn(
                  'w-full flex items-center justify-between gap-2 px-3 py-2 rounded-lg text-sm transition-colors',
                  active
                    ? 'bg-primary-500/20 text-white'
                    : 'text-gray-300 hover:bg-white/10 hover:text-white',
                )}
              >
                {content}
              </button>
            ) : (
              <div
                className={cn(
                  'flex items-center justify-between gap-2',
                  onSelect ? 'px-3 py-2 text-sm text-gray-400' : 'text-xs text-gray-400',
                )}
              >
                {content}
              </div>
            )}
          </li>
        )
      })}
      {hidden > 0 && <li className='text-xs text-gray-500'>+{hidden} more</li>}
    </ul>
  )
}
