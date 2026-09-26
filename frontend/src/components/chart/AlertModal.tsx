import { useEffect, useState } from 'react'
import Modal from '@/components/Modal'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { SOURCE_OPTIONS } from './PriceChart'
import type {
  AlertCondition,
  AlertDirection,
  AlertMode,
  ChartSource,
  PriceAlert,
  PriceAlertRequest,
} from '@/types'

export const CONDITION_LABELS: Record<AlertCondition, string> = {
  above: 'Price crosses above',
  below: 'Price crosses below',
  pct_change: 'Price changes by %',
  ema_cross: 'Price crosses EMA',
}

interface AlertModalProps {
  open: boolean
  onClose: () => void
  cardId: string
  cardName: string
  currentPrice?: number
  /** Price picked on the chart for a new alert */
  initialPrice?: number
  source: ChartSource
  existing?: PriceAlert | null
  onSaved: (alert: PriceAlert) => void
}

export default function AlertModal({
  open,
  onClose,
  cardId,
  cardName,
  currentPrice,
  initialPrice,
  source,
  existing,
  onSaved,
}: AlertModalProps) {
  const { addToast } = useToast()
  const [condition, setCondition] = useState<AlertCondition>('above')
  const [target, setTarget] = useState('')
  const [pct, setPct] = useState('10')
  const [days, setDays] = useState('7')
  const [emaPeriod, setEmaPeriod] = useState('20')
  const [direction, setDirection] = useState<AlertDirection>('either')
  const [mode, setMode] = useState<AlertMode>('once')
  const [cooldown, setCooldown] = useState('24')
  const [alertSource, setAlertSource] = useState<ChartSource>(source)
  const [notifyEmail, setNotifyEmail] = useState(true)
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setError(null)
    if (existing) {
      setCondition(existing.condition)
      setTarget(existing.target_price ? String(existing.target_price) : '')
      setPct(String(existing.pct ?? 10))
      setDays(String(existing.days ?? 7))
      setEmaPeriod(String(existing.ema_period ?? 20))
      setDirection(existing.direction ?? 'either')
      setMode(existing.mode)
      setCooldown(String(existing.cooldown_hours ?? 24))
      setAlertSource(existing.source)
      setNotifyEmail(existing.notify_email)
      setNote(existing.note ?? '')
      return
    }
    const price = initialPrice ?? currentPrice
    setCondition(
      price !== undefined && currentPrice !== undefined && price < currentPrice ? 'below' : 'above',
    )
    setTarget(price !== undefined ? price.toFixed(2) : '')
    setPct('10')
    setDays('7')
    setEmaPeriod('20')
    setDirection('either')
    setMode('once')
    setCooldown('24')
    setAlertSource(source)
    setNotifyEmail(true)
    setNote('')
  }, [open, existing, initialPrice, currentPrice, source])

  const buildRequest = (): PriceAlertRequest | string => {
    const req: PriceAlertRequest = {
      card_id: cardId,
      source: alertSource,
      condition,
      mode,
      notify_email: notifyEmail,
      note: note.trim(),
    }
    if (condition === 'above' || condition === 'below') {
      const t = Number(target)
      if (!Number.isFinite(t) || t <= 0) return 'Enter a target price greater than 0'
      req.target_price = t
    } else if (condition === 'pct_change') {
      const p = Number(pct)
      const d = Number(days)
      if (!Number.isFinite(p) || p < 0.5 || p > 1000) return 'Percent must be between 0.5 and 1000'
      if (!Number.isInteger(d) || d < 1 || d > 365)
        return 'Days must be a whole number from 1 to 365'
      req.pct = p
      req.days = d
      req.direction = direction
    } else {
      const e = Number(emaPeriod)
      if (!Number.isInteger(e) || e < 2 || e > 200)
        return 'EMA period must be a whole number from 2 to 200'
      req.ema_period = e
      req.direction = direction
    }
    if (mode === 'recurring') {
      const c = Number(cooldown)
      if (!Number.isInteger(c) || c < 1 || c > 168) return 'Cooldown must be 1-168 hours'
      req.cooldown_hours = c
    }
    if (req.note && req.note.length > 200) return 'Note must not exceed 200 characters'
    if (existing) req.active = existing.active
    return req
  }

  const submit = async () => {
    const req = buildRequest()
    if (typeof req === 'string') {
      setError(req)
      return
    }
    setSaving(true)
    setError(null)
    try {
      const saved = existing
        ? await apiClient.updateAlert(existing.id, req)
        : await apiClient.createAlert(req)
      addToast(existing ? 'Alert updated' : 'Alert created', 'success')
      onSaved(saved)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save alert')
    } finally {
      setSaving(false)
    }
  }

  const field = 'input'
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={existing ? 'Edit price alert' : 'New price alert'}
      footer={
        <>
          <button type='button' onClick={onClose} className='btn-ghost py-2 px-4'>
            Cancel
          </button>
          <button
            type='button'
            disabled={saving}
            onClick={submit}
            className='btn-primary py-2 px-4'
          >
            {saving ? 'Saving...' : existing ? 'Update' : 'Create alert'}
          </button>
        </>
      }
    >
      <div className='space-y-4'>
        <p className='text-sm text-gray-400'>
          {cardName}
          {currentPrice !== undefined && <> · current ${currentPrice.toFixed(2)}</>}
        </p>

        <div className='grid grid-cols-2 gap-3'>
          <label className='block text-sm text-gray-300'>
            Condition
            <select
              className={`${field} mt-1`}
              value={condition}
              onChange={(e) => setCondition(e.target.value as AlertCondition)}
            >
              {(Object.keys(CONDITION_LABELS) as AlertCondition[]).map((c) => (
                <option key={c} value={c} className='bg-slate-800'>
                  {CONDITION_LABELS[c]}
                </option>
              ))}
            </select>
          </label>
          <label className='block text-sm text-gray-300'>
            Price source
            <select
              className={`${field} mt-1`}
              value={alertSource}
              onChange={(e) => setAlertSource(e.target.value as ChartSource)}
            >
              {SOURCE_OPTIONS.map((s) => (
                <option key={s.value} value={s.value} className='bg-slate-800'>
                  {s.label}
                </option>
              ))}
            </select>
          </label>
        </div>

        {(condition === 'above' || condition === 'below') && (
          <label className='block text-sm text-gray-300'>
            Target price ($)
            <input
              className={`${field} mt-1`}
              type='number'
              min={0.01}
              step={0.01}
              value={target}
              onChange={(e) => setTarget(e.target.value)}
            />
          </label>
        )}
        {condition === 'pct_change' && (
          <div className='grid grid-cols-2 gap-3'>
            <label className='block text-sm text-gray-300'>
              Change (%)
              <input
                className={`${field} mt-1`}
                type='number'
                min={0.5}
                max={1000}
                step={0.5}
                value={pct}
                onChange={(e) => setPct(e.target.value)}
              />
            </label>
            <label className='block text-sm text-gray-300'>
              Over (days)
              <input
                className={`${field} mt-1`}
                type='number'
                min={1}
                max={365}
                value={days}
                onChange={(e) => setDays(e.target.value)}
              />
            </label>
          </div>
        )}
        {condition === 'ema_cross' && (
          <label className='block text-sm text-gray-300'>
            EMA period (days)
            <input
              className={`${field} mt-1`}
              type='number'
              min={2}
              max={200}
              value={emaPeriod}
              onChange={(e) => setEmaPeriod(e.target.value)}
            />
          </label>
        )}
        {(condition === 'pct_change' || condition === 'ema_cross') && (
          <label className='block text-sm text-gray-300'>
            Direction
            <select
              className={`${field} mt-1`}
              value={direction}
              onChange={(e) => setDirection(e.target.value as AlertDirection)}
            >
              <option value='either' className='bg-slate-800'>
                Either way
              </option>
              <option value='up' className='bg-slate-800'>
                Up only
              </option>
              <option value='down' className='bg-slate-800'>
                Down only
              </option>
            </select>
          </label>
        )}

        <div className='grid grid-cols-2 gap-3'>
          <label className='block text-sm text-gray-300'>
            Trigger
            <select
              className={`${field} mt-1`}
              value={mode}
              onChange={(e) => setMode(e.target.value as AlertMode)}
            >
              <option value='once' className='bg-slate-800'>
                Once, then pause
              </option>
              <option value='recurring' className='bg-slate-800'>
                Every time (with cooldown)
              </option>
            </select>
          </label>
          {mode === 'recurring' && (
            <label className='block text-sm text-gray-300'>
              Cooldown (hours)
              <input
                className={`${field} mt-1`}
                type='number'
                min={1}
                max={168}
                value={cooldown}
                onChange={(e) => setCooldown(e.target.value)}
              />
            </label>
          )}
        </div>

        <label className='block text-sm text-gray-300'>
          Note (optional)
          <input
            className={`${field} mt-1`}
            maxLength={200}
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        </label>

        <label className='flex items-center gap-2 text-sm text-gray-300'>
          <input
            type='checkbox'
            checked={notifyEmail}
            onChange={(e) => setNotifyEmail(e.target.checked)}
          />
          Also email me when this alert fires
        </label>

        <p className='text-xs text-gray-500'>
          Alerts are checked periodically against the daily average price.
        </p>
        {error && <p className='text-sm text-error-400'>{error}</p>}
      </div>
    </Modal>
  )
}
