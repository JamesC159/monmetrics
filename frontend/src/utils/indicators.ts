import type { ChartIndicator, PricePoint } from '@/types'

export interface ChartPoint {
  date: string
  fullDate: string
  timestamp: number
  price: number
  open: number
  close: number
  high: number
  low: number
  volume: number
  [indicatorKey: string]: number | string | null
}

export type IndicatorType = 'ema' | 'sma' | 'bollinger' | 'rsi' | 'macd'

export const INDICATOR_TYPES: {
  value: IndicatorType
  label: string
  pane: 'overlay' | 'separate'
}[] = [
  { value: 'ema', label: 'EMA', pane: 'overlay' },
  { value: 'sma', label: 'SMA', pane: 'overlay' },
  { value: 'bollinger', label: 'Bollinger Bands', pane: 'overlay' },
  { value: 'rsi', label: 'RSI', pane: 'separate' },
  { value: 'macd', label: 'MACD', pane: 'separate' },
]

export const DEFAULT_PARAMS: Record<IndicatorType, Record<string, number>> = {
  ema: { period: 20 },
  sma: { period: 20 },
  bollinger: { period: 20, stddev: 2 },
  rsi: { period: 14 },
  macd: { fast: 12, slow: 26, signal: 9 },
}

/** Parameter bounds, mirrored by backend validateIndicatorParams. */
export const PARAM_LIMITS: Record<
  IndicatorType,
  Record<string, { min: number; max: number; step?: number }>
> = {
  ema: { period: { min: 2, max: 200 } },
  sma: { period: { min: 2, max: 200 } },
  bollinger: { period: { min: 2, max: 200 }, stddev: { min: 0.5, max: 4, step: 0.5 } },
  rsi: { period: { min: 2, max: 100 } },
  macd: { fast: { min: 2, max: 199 }, slow: { min: 3, max: 200 }, signal: { min: 2, max: 50 } },
}

export function validParams(type: string, params: Record<string, unknown>): boolean {
  const limits = PARAM_LIMITS[type as IndicatorType]
  if (!limits) return false
  for (const [key, { min, max, step }] of Object.entries(limits)) {
    const v = Number(params[key])
    if (!Number.isFinite(v) || v < min || v > max) return false
    if (!step && !Number.isInteger(v)) return false
  }
  if (type === 'macd' && Number(params.fast) >= Number(params.slow)) return false
  return true
}

export const EMA_PERIOD_MIN = 2
export const EMA_PERIOD_MAX = 200
export const EMA_PRESETS = [9, 20, 50, 200]
export const INDICATOR_COLORS = [
  '#22c55e',
  '#f59e0b',
  '#a855f7',
  '#06b6d4',
  '#f43f5e',
  '#eab308',
  '#3b82f6',
  '#ec4899',
  '#14b8a6',
  '#f97316',
]

/**
 * Exponential moving average seeded with the simple average of the first `period` values.
 * Entries before the seed are null.
 */
export function calculateEMA(values: number[], period: number): (number | null)[] {
  const out: (number | null)[] = new Array(values.length).fill(null)
  if (!Number.isInteger(period) || period < 1 || values.length < period) return out

  const alpha = 2 / (period + 1)
  let sum = 0
  for (let i = 0; i < period; i++) sum += values[i]
  let prev = sum / period
  out[period - 1] = prev
  for (let i = period; i < values.length; i++) {
    prev = values[i] * alpha + prev * (1 - alpha)
    out[i] = prev
  }
  return out
}

export function calculateSMA(values: number[], period: number): (number | null)[] {
  const out: (number | null)[] = new Array(values.length).fill(null)
  if (!Number.isInteger(period) || period < 1) return out
  let sum = 0
  for (let i = 0; i < values.length; i++) {
    sum += values[i]
    if (i >= period) sum -= values[i - period]
    if (i >= period - 1) out[i] = sum / period
  }
  return out
}

/** Bollinger Bands using the population standard deviation over the window. */
export function calculateBollinger(values: number[], period: number, stddev: number) {
  const middle = calculateSMA(values, period)
  const upper: (number | null)[] = new Array(values.length).fill(null)
  const lower: (number | null)[] = new Array(values.length).fill(null)
  for (let i = period - 1; i < values.length; i++) {
    const m = middle[i]
    if (m === null) continue
    let sq = 0
    for (let j = i - period + 1; j <= i; j++) sq += (values[j] - m) ** 2
    const sd = Math.sqrt(sq / period)
    upper[i] = m + stddev * sd
    lower[i] = m - stddev * sd
  }
  return { middle, upper, lower }
}

/** RSI with Wilder smoothing; the first value appears at index `period`. */
export function calculateRSI(values: number[], period: number): (number | null)[] {
  const out: (number | null)[] = new Array(values.length).fill(null)
  if (!Number.isInteger(period) || period < 1 || values.length <= period) return out
  let gain = 0
  let loss = 0
  for (let i = 1; i <= period; i++) {
    const d = values[i] - values[i - 1]
    if (d > 0) gain += d
    else loss -= d
  }
  gain /= period
  loss /= period
  const rsi = () => (loss === 0 ? (gain === 0 ? 50 : 100) : 100 - 100 / (1 + gain / loss))
  out[period] = rsi()
  for (let i = period + 1; i < values.length; i++) {
    const d = values[i] - values[i - 1]
    gain = (gain * (period - 1) + Math.max(d, 0)) / period
    loss = (loss * (period - 1) + Math.max(-d, 0)) / period
    out[i] = rsi()
  }
  return out
}

export function calculateMACD(values: number[], fast: number, slow: number, signal: number) {
  const fastEma = calculateEMA(values, fast)
  const slowEma = calculateEMA(values, slow)
  const macd = values.map((_, i) =>
    fastEma[i] !== null && slowEma[i] !== null ? fastEma[i]! - slowEma[i]! : null,
  )
  const start = macd.findIndex((v) => v !== null)
  const signalLine: (number | null)[] = new Array(values.length).fill(null)
  if (start >= 0) {
    const sig = calculateEMA(macd.slice(start) as number[], signal)
    sig.forEach((v, i) => (signalLine[start + i] = v))
  }
  const hist = macd.map((v, i) =>
    v !== null && signalLine[i] !== null ? v - signalLine[i]! : null,
  )
  return { macd, signal: signalLine, hist }
}

/** Output series per indicator; the `value` key maps to indicatorKey(i), others to `${indicatorKey(i)}_${key}`. */
export function computeIndicator(
  ind: ChartIndicator,
  closes: number[],
): Record<string, (number | null)[]> {
  const p = ind.parameters ?? {}
  switch (ind.type) {
    case 'ema':
      return { value: calculateEMA(closes, Number(p.period)) }
    case 'sma':
      return { value: calculateSMA(closes, Number(p.period)) }
    case 'bollinger': {
      const b = calculateBollinger(closes, Number(p.period), Number(p.stddev))
      return { value: b.middle, upper: b.upper, lower: b.lower }
    }
    case 'rsi':
      return { value: calculateRSI(closes, Number(p.period)) }
    case 'macd': {
      const m = calculateMACD(closes, Number(p.fast), Number(p.slow), Number(p.signal))
      return { value: m.macd, signal: m.signal, hist: m.hist }
    }
  }
  return {}
}

export const indicatorKey = (index: number) => `ind_${index}`
export const seriesKey = (index: number, key: string) =>
  key === 'value' ? indicatorKey(index) : `${indicatorKey(index)}_${key}`

export function indicatorLabel(ind: ChartIndicator): string {
  const p = ind.parameters ?? {}
  switch (ind.type) {
    case 'sma':
      return `SMA ${p.period ?? ''}`.trim()
    case 'bollinger':
      return `BB ${p.period},${p.stddev}`
    case 'rsi':
      return `RSI ${p.period ?? ''}`.trim()
    case 'macd':
      return `MACD ${p.fast},${p.slow},${p.signal}`
  }
  return `EMA ${p.period ?? ''}`.trim()
}

/** Warm-up history needed so the slowest indicator has converged at the start of the visible range. */
export function warmupDaysFor(indicators: ChartIndicator[]): number {
  const lengths = indicators.map((i) => {
    const p = i.parameters ?? {}
    if (i.type === 'macd') return (Number(p.slow) || 0) + (Number(p.signal) || 0)
    return Number(p.period) || 0
  })
  if (lengths.length === 0) return 0
  return Math.min(400, Math.max(...lengths) * 3)
}

const DAY_MS = 86_400_000

/** Groups raw price points into one point per UTC day with OHLC, optionally filtered by source. */
export function aggregateDaily(prices: PricePoint[], source: string): ChartPoint[] {
  const filtered = (source === 'all' ? prices : prices.filter((p) => p.source === source))
    .map((p) => ({ price: p.price, ts: new Date(p.timestamp).getTime() }))
    .sort((a, b) => a.ts - b.ts)
  const byDay = new Map<number, number[]>()
  for (const p of filtered) {
    const day = Math.floor(p.ts / DAY_MS) * DAY_MS
    const bucket = byDay.get(day)
    if (bucket) bucket.push(p.price)
    else byDay.set(day, [p.price])
  }
  return Array.from(byDay.entries())
    .map(([day, dayPrices]) => {
      const avg = dayPrices.reduce((s, v) => s + v, 0) / dayPrices.length
      return {
        date: new Date(day).toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
          timeZone: 'UTC',
        }),
        fullDate: new Date(day).toISOString(),
        timestamp: day,
        price: round2(avg),
        open: round2(dayPrices[0]),
        close: round2(dayPrices[dayPrices.length - 1]),
        high: round2(Math.max(...dayPrices)),
        low: round2(Math.min(...dayPrices)),
        volume: dayPrices.length,
      }
    })
    .sort((a, b) => a.timestamp - b.timestamp)
}

/**
 * Computes indicator series over the full (warm-up inclusive) series, then trims to the
 * visible range so early visible points already have converged values.
 */
export function buildChartSeries(
  prices: PricePoint[],
  source: string,
  indicators: ChartIndicator[],
  visibleFrom?: string,
): ChartPoint[] {
  const series = aggregateDaily(prices, source)
  const closes = series.map((p) => p.price)
  indicators.forEach((ind, i) => {
    if (!ind.visible) return
    for (const [key, values] of Object.entries(computeIndicator(ind, closes))) {
      series.forEach((point, idx) => {
        const v = values[idx]
        point[seriesKey(i, key)] = v === null ? null : round2(v)
      })
    }
  })
  if (!visibleFrom) return series
  const cutoff = Math.floor(new Date(visibleFrom).getTime() / DAY_MS) * DAY_MS
  return series.filter((p) => p.timestamp >= cutoff)
}

function round2(v: number): number {
  return Math.round(v * 100) / 100
}
