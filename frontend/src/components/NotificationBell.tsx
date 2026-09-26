import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Bell } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { formatDateTime } from '@/utils/formatters'
import type { AppNotification } from '@/types'

const POLL_MS = 60_000

export default function NotificationBell() {
  const navigate = useNavigate()
  const { addToast } = useToast()
  const [items, setItems] = useState<AppNotification[]>([])
  const [unread, setUnread] = useState(0)
  const [open, setOpen] = useState(false)
  const seen = useRef<Set<string> | null>(null)
  const rootRef = useRef<HTMLDivElement>(null)

  const load = useCallback(async () => {
    try {
      const res = await apiClient.getNotifications(false, 20)
      // Toast only notifications that arrived after the first load.
      if (seen.current) {
        for (const n of res.notifications)
          if (!n.read && !seen.current.has(n.id)) addToast(`${n.title}: ${n.message}`, 'info', 8000)
      }
      seen.current = new Set(res.notifications.map((n) => n.id))
      setItems(res.notifications)
      setUnread(res.unread_count)
    } catch {
      // Polling failures are transient; the next tick retries.
    }
  }, [addToast])

  useEffect(() => {
    load()
    const timer = window.setInterval(
      () => document.visibilityState === 'visible' && load(),
      POLL_MS,
    )
    const onVisible = () => document.visibilityState === 'visible' && load()
    document.addEventListener('visibilitychange', onVisible)
    return () => {
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [load])

  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => !rootRef.current?.contains(e.target as Node) && setOpen(false)
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  const openItem = async (n: AppNotification) => {
    setOpen(false)
    if (!n.read) {
      setItems((list) => list.map((x) => (x.id === n.id ? { ...x, read: true } : x)))
      setUnread((u) => Math.max(0, u - 1))
      apiClient.markNotificationRead(n.id).catch(() => undefined)
    }
    if (n.card_id) navigate(`/card/${n.card_id}`)
  }

  const markAll = async () => {
    setItems((list) => list.map((x) => ({ ...x, read: true })))
    setUnread(0)
    try {
      await apiClient.markAllNotificationsRead()
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Could not update notifications', 'error')
      load()
    }
  }

  return (
    <div ref={rootRef} className='relative'>
      <button
        type='button'
        onClick={() => setOpen((v) => !v)}
        aria-haspopup='true'
        aria-expanded={open}
        aria-label={unread ? `Notifications, ${unread} unread` : 'Notifications'}
        className='relative p-2 text-gray-300 hover:text-white'
      >
        <Bell className='w-5 h-5' />
        {unread > 0 && (
          <span className='absolute top-0.5 right-0.5 min-w-[1.1rem] h-[1.1rem] px-1 rounded-full bg-error-500 text-white text-[10px] leading-[1.1rem] text-center'>
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>
      {open && (
        <div className='absolute right-0 mt-2 w-80 max-h-96 overflow-y-auto rounded-xl bg-dark-900 border border-white/10 shadow-xl z-50'>
          <div className='flex items-center justify-between px-4 py-2 border-b border-white/10'>
            <span className='text-sm font-semibold text-white'>Notifications</span>
            {unread > 0 && (
              <button
                type='button'
                onClick={markAll}
                className='text-xs text-primary-300 hover:text-primary-200'
              >
                Mark all read
              </button>
            )}
          </div>
          {items.length === 0 ? (
            <p className='px-4 py-6 text-sm text-gray-500 text-center'>No notifications yet</p>
          ) : (
            <ul>
              {items.map((n) => (
                <li key={n.id}>
                  <button
                    type='button'
                    onClick={() => openItem(n)}
                    className={`w-full text-left px-4 py-3 border-b border-white/5 hover:bg-white/5 ${n.read ? 'opacity-60' : ''}`}
                  >
                    <div className='flex items-start gap-2'>
                      {!n.read && (
                        <span className='mt-1.5 w-2 h-2 rounded-full bg-primary-400 shrink-0' />
                      )}
                      <div className='min-w-0'>
                        <div className='text-sm text-white truncate'>{n.title}</div>
                        <div className='text-xs text-gray-400'>{n.message}</div>
                        <div className='text-[11px] text-gray-500 mt-1'>
                          {formatDateTime(n.created_at)}
                        </div>
                      </div>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
