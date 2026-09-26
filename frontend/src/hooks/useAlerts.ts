import { useCallback, useEffect, useState } from 'react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { alertToRequest } from '@/components/alerts/AlertList'
import type { PriceAlert } from '@/types'

/** Loads and mutates the signed-in user's alerts, optionally scoped to one card. */
export function useAlerts(enabled: boolean, cardId?: string) {
  const { addToast } = useToast()
  const [alerts, setAlerts] = useState<PriceAlert[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!enabled) {
      setAlerts([])
      return
    }
    let cancelled = false
    setLoading(true)
    apiClient
      .getAlerts(cardId)
      .then((list) => !cancelled && setAlerts(list))
      .catch(
        (err) =>
          !cancelled &&
          addToast(err instanceof Error ? err.message : 'Could not load alerts', 'error'),
      )
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [enabled, cardId, addToast])

  const upsert = useCallback((a: PriceAlert) => {
    setAlerts((list) =>
      list.some((x) => x.id === a.id) ? list.map((x) => (x.id === a.id ? a : x)) : [a, ...list],
    )
  }, [])

  const save = useCallback(
    async (a: PriceAlert, patch: Partial<PriceAlert>, success?: string) => {
      const optimistic = { ...a, ...patch }
      upsert(optimistic)
      try {
        const saved = await apiClient.updateAlert(a.id, alertToRequest(optimistic))
        upsert(saved)
        if (success) addToast(success, 'success')
      } catch (err) {
        upsert(a)
        addToast(err instanceof Error ? err.message : 'Could not update alert', 'error')
      }
    },
    [upsert, addToast],
  )

  const toggle = useCallback(
    (a: PriceAlert) => save(a, { active: !a.active }, a.active ? 'Alert paused' : 'Alert re-armed'),
    [save],
  )
  const move = useCallback(
    (a: PriceAlert, price: number) =>
      save(a, { target_price: price }, `Alert moved to $${price.toFixed(2)}`),
    [save],
  )

  const remove = useCallback(
    async (a: PriceAlert) => {
      setAlerts((list) => list.filter((x) => x.id !== a.id))
      try {
        await apiClient.deleteAlert(a.id)
      } catch (err) {
        upsert(a)
        addToast(err instanceof Error ? err.message : 'Could not delete alert', 'error')
      }
    },
    [upsert, addToast],
  )

  return { alerts, loading, upsert, toggle, move, remove }
}
