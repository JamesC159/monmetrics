import { Bell, BellOff, Pencil, Trash2 } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { PriceAlert, PriceAlertRequest } from '@/types'
import { formatDateTime, formatPrice } from '@/utils/formatters'

export function alertToRequest(a: PriceAlert): PriceAlertRequest {
  return {
    card_id: a.card_id,
    source: a.source,
    condition: a.condition,
    target_price: a.target_price,
    pct: a.pct,
    days: a.days,
    ema_period: a.ema_period,
    direction: a.direction,
    mode: a.mode,
    cooldown_hours: a.cooldown_hours,
    notify_email: a.notify_email,
    note: a.note,
    active: a.active,
  }
}

export function describeAlert(a: PriceAlert): string {
  const dir = a.direction === 'up' ? ' up' : a.direction === 'down' ? ' down' : ''
  switch (a.condition) {
    case 'above':
      return `Crosses above ${formatPrice(a.target_price ?? 0)}`
    case 'below':
      return `Crosses below ${formatPrice(a.target_price ?? 0)}`
    case 'pct_change':
      return `Moves${dir} ${a.pct}% over ${a.days}d`
    case 'ema_cross':
      return `Crosses${dir} EMA ${a.ema_period}`
  }
}

interface AlertListProps {
  alerts: PriceAlert[]
  showCard?: boolean
  onEdit: (alert: PriceAlert) => void
  onToggle: (alert: PriceAlert) => void
  onDelete: (alert: PriceAlert) => void
}

export default function AlertList({
  alerts,
  showCard,
  onEdit,
  onToggle,
  onDelete,
}: AlertListProps) {
  return (
    <ul className='divide-y divide-white/10'>
      {alerts.map((a) => (
        <li key={a.id} className='flex flex-wrap items-center gap-3 py-2 text-sm'>
          {showCard && (
            <Link
              to={`/card/${a.card_id}`}
              className='flex items-center gap-2 min-w-0 w-48 text-white hover:text-primary-300'
            >
              {a.card_image_url && (
                <img src={a.card_image_url} alt='' className='w-6 h-8 object-cover rounded' />
              )}
              <span className='truncate'>{a.card_name ?? 'Card'}</span>
            </Link>
          )}
          <span className={`flex-1 min-w-[10rem] ${a.active ? 'text-gray-200' : 'text-gray-500'}`}>
            {describeAlert(a)}
            <span className='text-xs text-gray-500'>
              {' '}
              · {a.source === 'all' ? 'all sources' : a.source} ·{' '}
              {a.mode === 'once' ? 'once' : `every ${a.cooldown_hours}h max`}
              {a.notify_email ? ' · email' : ''}
            </span>
            {a.note && <span className='block text-xs text-gray-500 truncate'>{a.note}</span>}
          </span>
          <span className='text-xs text-gray-500 w-36'>
            {a.last_triggered_at ? `Fired ${formatDateTime(a.last_triggered_at)}` : 'Not fired yet'}
          </span>
          <span className='flex items-center gap-1'>
            <button
              type='button'
              onClick={() => onToggle(a)}
              className='p-1 text-gray-400 hover:text-white'
              title={a.active ? 'Pause alert' : 'Re-arm alert'}
              aria-label={a.active ? 'Pause alert' : 'Re-arm alert'}
            >
              {a.active ? (
                <Bell className='w-4 h-4 text-warning-400' />
              ) : (
                <BellOff className='w-4 h-4' />
              )}
            </button>
            <button
              type='button'
              onClick={() => onEdit(a)}
              className='p-1 text-gray-400 hover:text-white'
              aria-label='Edit alert'
            >
              <Pencil className='w-4 h-4' />
            </button>
            <button
              type='button'
              onClick={() => onDelete(a)}
              className='p-1 text-gray-400 hover:text-error-400'
              aria-label='Delete alert'
            >
              <Trash2 className='w-4 h-4' />
            </button>
          </span>
        </li>
      ))}
    </ul>
  )
}
