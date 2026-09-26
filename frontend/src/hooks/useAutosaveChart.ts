import { useCallback, useEffect, useRef, useState } from 'react'
import { apiClient } from '@/utils/api'
import type { SavedChart, SavedChartRequest } from '@/types'

export type AutosaveStatus = 'idle' | 'saving' | 'saved' | 'error'

const pendingKey = (chartId: string) => `mm:chart-autosave:${chartId}`
const DEBOUNCE_MS = 800

function toRequest(chart: SavedChart): SavedChartRequest {
  return {
    card_id: chart.card_id,
    name: chart.name,
    description: chart.description ?? '',
    indicators: chart.indicators ?? [],
    time_range: chart.time_range,
    source: chart.source,
    chart_type: chart.chart_type ?? 'line',
    show_volume: !!chart.show_volume,
    drawings: chart.drawings ?? [],
  }
}

/** Reads a chart edit that never made it to the server (e.g. tab closed mid-save). */
export function readPendingAutosave(chartId: string): SavedChartRequest | null {
  try {
    const raw = localStorage.getItem(pendingKey(chartId))
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

/**
 * Debounces and serializes saves of an already-persisted chart's live settings.
 * Keeps unsent edits in localStorage so a crash or reload can recover them.
 */
export function useAutosaveChart(
  chart: SavedChart | null,
  payload: SavedChartRequest | null,
  onSaved: (chart: SavedChart) => void,
) {
  const [status, setStatus] = useState<AutosaveStatus>('idle')
  const chartIdRef = useRef<string | null>(null)
  const savedJsonRef = useRef('')
  const latestRef = useRef<SavedChartRequest | null>(null)
  const savingRef = useRef(false)
  const dirtyRef = useRef(false)
  const timerRef = useRef<ReturnType<typeof setTimeout>>()

  // New chart loaded: baseline is the server's copy, not local state, so any
  // recovered pending edit is correctly seen as dirty.
  useEffect(() => {
    if (chart?.id === chartIdRef.current) return
    chartIdRef.current = chart?.id ?? null
    savedJsonRef.current = chart ? JSON.stringify(toRequest(chart)) : ''
    latestRef.current = null
    dirtyRef.current = false
    setStatus('idle')
  }, [chart])

  const runSave = useCallback(async () => {
    const chartId = chartIdRef.current
    const toSave = latestRef.current
    if (!chartId || !toSave || savingRef.current) return
    savingRef.current = true
    dirtyRef.current = false
    setStatus('saving')
    try {
      const saved = await apiClient.updateChart(chartId, toSave)
      savedJsonRef.current = JSON.stringify(toSave)
      try {
        localStorage.removeItem(pendingKey(chartId))
      } catch {
        /* best effort */
      }
      setStatus('saved')
      onSaved(saved)
    } catch {
      setStatus('error')
    } finally {
      savingRef.current = false
      if (dirtyRef.current) runSave()
    }
  }, [onSaved])

  useEffect(() => {
    if (!chart?.id || !payload) return
    const json = JSON.stringify(payload)
    if (json === savedJsonRef.current) return
    latestRef.current = payload
    dirtyRef.current = true
    try {
      localStorage.setItem(pendingKey(chart.id), json)
    } catch {
      /* best effort */
    }
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(runSave, DEBOUNCE_MS)
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [chart?.id, payload, runSave])

  // Flush pending edits instead of losing them to navigation/unmount.
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      if (dirtyRef.current) runSave()
    }
  }, [runSave])

  const retry = useCallback(() => {
    if (latestRef.current) runSave()
  }, [runSave])

  return { status, retry }
}
