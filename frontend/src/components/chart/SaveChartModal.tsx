import { useEffect, useState } from 'react'
import Modal from '@/components/Modal'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { indicatorLabel } from '@/utils/indicators'
import type {
  Card,
  ChartDrawing,
  ChartIndicator,
  ChartSource,
  ChartTimeRange,
  ChartType,
  SavedChart,
} from '@/types'

interface SaveChartModalProps {
  open: boolean
  onClose: () => void
  card: Card
  timeRange: ChartTimeRange
  source: ChartSource
  indicators: ChartIndicator[]
  chartType: ChartType
  showVolume: boolean
  drawings: ChartDrawing[]
  existing?: SavedChart | null
  onSaved: (chart: SavedChart) => void
}

function defaultName(card: Card, timeRange: string, indicators: ChartIndicator[]) {
  const parts = [card.name, timeRange.toUpperCase(), ...indicators.map(indicatorLabel)]
  return parts.join(' · ').slice(0, 100)
}

export default function SaveChartModal({
  open,
  onClose,
  card,
  timeRange,
  source,
  indicators,
  chartType,
  showVolume,
  drawings,
  existing,
  onSaved,
}: SaveChartModalProps) {
  const { addToast } = useToast()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setName(existing?.name ?? defaultName(card, timeRange, indicators))
    setDescription(existing?.description ?? '')
    setError(null)
  }, [open, existing, card, timeRange, indicators])

  const submit = async (asNew: boolean) => {
    const trimmed = name.trim()
    if (!trimmed || trimmed.length > 100) {
      setError('Name must be 1-100 characters')
      return
    }
    if (description.length > 500) {
      setError('Description must not exceed 500 characters')
      return
    }
    setSaving(true)
    setError(null)
    const payload = {
      card_id: card.id,
      name: trimmed,
      description,
      indicators,
      time_range: timeRange,
      source,
      chart_type: chartType,
      show_volume: showVolume,
      drawings,
    }
    try {
      const saved =
        existing && !asNew
          ? await apiClient.updateChart(existing.id, payload)
          : await apiClient.saveChart(payload)
      addToast(existing && !asNew ? 'Chart updated' : 'Chart saved', 'success')
      onSaved(saved)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save chart')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={existing ? 'Update saved chart' : 'Save chart'}
      footer={
        <>
          <button type='button' onClick={onClose} className='btn-ghost py-2 px-4'>
            Cancel
          </button>
          {existing && (
            <button
              type='button'
              disabled={saving}
              onClick={() => submit(true)}
              className='btn-ghost py-2 px-4'
            >
              Save as new
            </button>
          )}
          <button
            type='button'
            disabled={saving}
            onClick={() => submit(false)}
            className='btn-primary py-2 px-4'
          >
            {saving ? 'Saving...' : existing ? 'Update' : 'Save'}
          </button>
        </>
      }
    >
      <div className='space-y-4'>
        <div>
          <label htmlFor='chart-name' className='block text-sm text-gray-300 mb-1'>
            Name
          </label>
          <input
            id='chart-name'
            className='input'
            value={name}
            maxLength={100}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <div>
          <label htmlFor='chart-desc' className='block text-sm text-gray-300 mb-1'>
            Description (optional)
          </label>
          <textarea
            id='chart-desc'
            className='input min-h-[80px]'
            value={description}
            maxLength={500}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <div className='text-sm text-gray-400 space-y-1'>
          <div>
            Range: <span className='text-white'>{timeRange.toUpperCase()}</span> · Source:{' '}
            <span className='text-white capitalize'>{source}</span>
          </div>
          <div>
            Indicators:{' '}
            <span className='text-white'>
              {indicators.length ? indicators.map(indicatorLabel).join(', ') : 'None'}
            </span>
          </div>
        </div>
        {error && <p className='text-sm text-error-400'>{error}</p>}
      </div>
    </Modal>
  )
}
