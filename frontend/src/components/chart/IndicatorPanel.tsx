import { useState } from 'react'
import { Eye, EyeOff, Plus, Trash2 } from 'lucide-react'
import type { ChartIndicator } from '@/types'
import {
  DEFAULT_PARAMS,
  EMA_PRESETS,
  INDICATOR_COLORS,
  INDICATOR_TYPES,
  PARAM_LIMITS,
  indicatorLabel,
  validParams,
  type IndicatorType,
} from '@/utils/indicators'

interface IndicatorPanelProps {
  indicators: ChartIndicator[]
  onChange: (indicators: ChartIndicator[]) => void
  maxIndicators: number
  timeRange: string
  limitHint?: string
}

const PARAM_LABELS: Record<string, string> = {
  period: 'Period',
  stddev: 'Std dev',
  fast: 'Fast',
  slow: 'Slow',
  signal: 'Signal',
}

export default function IndicatorPanel({
  indicators,
  onChange,
  maxIndicators,
  timeRange,
  limitHint,
}: IndicatorPanelProps) {
  const [newType, setNewType] = useState<IndicatorType>('ema')
  const [draft, setDraft] = useState<Record<string, string>>(toStrings(DEFAULT_PARAMS.ema))
  const atLimit = indicators.length >= maxIndicators
  const draftParams = toNumbers(draft)
  const draftValid = validParams(newType, draftParams)

  const nextColor = () =>
    INDICATOR_COLORS.find((c) => !indicators.some((i) => i.color === c)) ?? INDICATOR_COLORS[0]

  const add = (type: IndicatorType, parameters: Record<string, number>) => {
    if (atLimit) return
    onChange([...indicators, { type, parameters, color: nextColor(), visible: true }])
  }

  const update = (index: number, patch: Partial<ChartIndicator>) =>
    onChange(indicators.map((ind, i) => (i === index ? { ...ind, ...patch } : ind)))

  const updateParam = (index: number, key: string, raw: string) => {
    const ind = indicators[index]
    const parameters = { ...ind.parameters, [key]: Number(raw) }
    if (validParams(ind.type, parameters)) update(index, { parameters })
  }

  const selectType = (type: IndicatorType) => {
    setNewType(type)
    setDraft(toStrings(DEFAULT_PARAMS[type]))
  }

  return (
    <div className='mt-4 p-4 rounded-xl bg-white/5 border border-white/10'>
      <div className='flex items-center justify-between mb-3'>
        <h4 className='text-sm font-semibold text-white'>Indicators</h4>
        <span className='text-xs text-gray-400'>
          {indicators.length}/{maxIndicators} used
        </span>
      </div>

      {timeRange === '1d' && (
        <p className='text-xs text-warning-400 mb-3'>
          Indicators use daily prices; try 7D or longer for meaningful lines.
        </p>
      )}

      {indicators.length > 0 && (
        <ul className='space-y-2 mb-3'>
          {indicators.map((ind, i) => {
            const limits = PARAM_LIMITS[ind.type as IndicatorType] ?? {}
            return (
              <li key={i} className='flex flex-wrap items-center gap-2 text-sm'>
                <input
                  type='color'
                  value={ind.color ?? INDICATOR_COLORS[0]}
                  onChange={(e) => update(i, { color: e.target.value })}
                  className='w-7 h-7 rounded bg-transparent border border-white/20 cursor-pointer'
                  aria-label={`Color for ${indicatorLabel(ind)}`}
                />
                <span className='text-gray-300 w-16'>
                  {INDICATOR_TYPES.find((t) => t.value === ind.type)?.label.split(' ')[0] ??
                    ind.type}
                </span>
                {Object.entries(limits).map(([key, { min, max, step }]) => (
                  <input
                    key={`${i}-${key}-${ind.parameters[key]}`}
                    type='number'
                    min={min}
                    max={max}
                    step={step ?? 1}
                    defaultValue={ind.parameters[key]}
                    onChange={(e) => updateParam(i, key, e.target.value)}
                    onBlur={(e) => updateParam(i, key, e.target.value)}
                    onKeyDown={(e) =>
                      e.key === 'Enter' && updateParam(i, key, (e.target as HTMLInputElement).value)
                    }
                    className='w-16 px-2 py-1 bg-dark-800/60 border border-white/10 rounded text-white'
                    aria-label={`${indicatorLabel(ind)} ${PARAM_LABELS[key] ?? key}`}
                    title={PARAM_LABELS[key] ?? key}
                  />
                ))}
                <button
                  type='button'
                  onClick={() => update(i, { visible: !ind.visible })}
                  className='p-1 text-gray-400 hover:text-white'
                  aria-label={ind.visible ? 'Hide indicator' : 'Show indicator'}
                >
                  {ind.visible ? <Eye className='w-4 h-4' /> : <EyeOff className='w-4 h-4' />}
                </button>
                <button
                  type='button'
                  onClick={() => onChange(indicators.filter((_, idx) => idx !== i))}
                  className='p-1 text-gray-400 hover:text-error-400'
                  aria-label='Remove indicator'
                >
                  <Trash2 className='w-4 h-4' />
                </button>
              </li>
            )
          })}
        </ul>
      )}

      <div className='flex flex-wrap items-center gap-2'>
        {EMA_PRESETS.map((p) => (
          <button
            key={p}
            type='button'
            disabled={atLimit}
            onClick={() => add('ema', { period: p })}
            className='px-2 py-1 text-xs rounded bg-white/10 text-gray-300 hover:bg-white/20 disabled:opacity-40 disabled:cursor-not-allowed'
          >
            EMA {p}
          </button>
        ))}
      </div>

      <div className='flex flex-wrap items-center gap-2 mt-3'>
        <select
          value={newType}
          onChange={(e) => selectType(e.target.value as IndicatorType)}
          className='px-2 py-1 text-sm bg-dark-800/60 border border-white/10 rounded text-white'
          aria-label='Indicator type'
        >
          {INDICATOR_TYPES.map((t) => (
            <option key={t.value} value={t.value} className='bg-slate-800'>
              {t.label}
            </option>
          ))}
        </select>
        {Object.entries(PARAM_LIMITS[newType]).map(([key, { min, max, step }]) => (
          <label key={key} className='flex items-center gap-1 text-xs text-gray-400'>
            {PARAM_LABELS[key] ?? key}
            <input
              type='number'
              min={min}
              max={max}
              step={step ?? 1}
              value={draft[key] ?? ''}
              onChange={(e) => setDraft({ ...draft, [key]: e.target.value })}
              className='w-16 px-2 py-1 text-sm bg-dark-800/60 border border-white/10 rounded text-white'
            />
          </label>
        ))}
        <button
          type='button'
          disabled={atLimit || !draftValid}
          onClick={() => add(newType, draftParams)}
          className='flex items-center px-2 py-1 text-xs rounded bg-primary-500/20 text-primary-300 hover:bg-primary-500/30 disabled:opacity-40 disabled:cursor-not-allowed'
        >
          <Plus className='w-3 h-3 mr-1' /> Add
        </button>
      </div>
      {!draftValid && (
        <p className='text-xs text-error-400 mt-2'>
          Check the parameter ranges{newType === 'macd' ? ' (fast must be less than slow)' : ''}.
        </p>
      )}
      {atLimit && limitHint && <p className='text-xs text-gray-400 mt-2'>{limitHint}</p>}
    </div>
  )
}

function toStrings(p: Record<string, number>): Record<string, string> {
  return Object.fromEntries(Object.entries(p).map(([k, v]) => [k, String(v)]))
}

function toNumbers(p: Record<string, string>): Record<string, number> {
  return Object.fromEntries(Object.entries(p).map(([k, v]) => [k, Number(v)]))
}
