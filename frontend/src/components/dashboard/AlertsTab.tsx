import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Bell } from 'lucide-react'
import AlertList from '@/components/alerts/AlertList'
import AlertModal from '@/components/chart/AlertModal'
import { useAlerts } from '@/hooks/useAlerts'
import { useAuth } from '@/context/AuthContext'
import type { PriceAlert } from '@/types'

export default function AlertsTab() {
  const { user } = useAuth()
  const { alerts, loading, upsert, toggle, remove } = useAlerts(true)
  const [editing, setEditing] = useState<PriceAlert | null>(null)
  const limit = user?.user_type === 'paid' ? 25 : 3
  const active = alerts.filter((a) => a.active).length

  if (loading && alerts.length === 0)
    return <p className='text-gray-400 py-8 text-center'>Loading alerts...</p>

  if (alerts.length === 0) {
    return (
      <div className='glass-effect rounded-2xl text-center py-16'>
        <Bell className='w-12 h-12 text-gray-600 mx-auto mb-3' />
        <p className='text-gray-400 mb-4'>
          No price alerts yet. Open a card and right-click its chart, or use the Alert button.
        </p>
        <Link to='/search' className='btn-primary py-2 px-4'>
          Search cards
        </Link>
      </div>
    )
  }

  return (
    <div className='glass-effect rounded-2xl p-6'>
      <div className='flex items-center justify-between mb-3'>
        <h3 className='text-lg font-semibold text-white'>Price alerts</h3>
        <span className='text-sm text-gray-400'>
          {active}/{limit} active
        </span>
      </div>
      <AlertList alerts={alerts} showCard onEdit={setEditing} onToggle={toggle} onDelete={remove} />
      {editing && (
        <AlertModal
          open
          onClose={() => setEditing(null)}
          cardId={editing.card_id}
          cardName={editing.card_name ?? 'Card'}
          source={editing.source}
          existing={editing}
          onSaved={upsert}
        />
      )}
    </div>
  )
}
