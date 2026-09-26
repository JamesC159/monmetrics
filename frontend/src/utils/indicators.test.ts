import { describe, expect, it } from 'vitest'
import {
  aggregateDaily,
  buildChartSeries,
  calculateBollinger,
  calculateEMA,
  calculateMACD,
  calculateRSI,
  calculateSMA,
  indicatorKey,
  seriesKey,
  validParams,
  warmupDaysFor,
} from './indicators'
import type { PricePoint } from '@/types'

describe('calculateSMA', () => {
  it('averages a sliding window', () => {
    expect(calculateSMA([1, 2, 3, 4, 5], 3)).toEqual([null, null, 2, 3, 4])
  })
})

describe('calculateBollinger', () => {
  it('uses population standard deviation around the SMA', () => {
    const { middle, upper, lower } = calculateBollinger([2, 4, 4, 4, 5, 5, 7, 9], 8, 2)
    // mean 5, population sd 2
    expect(middle[7]).toBe(5)
    expect(upper[7]).toBeCloseTo(9)
    expect(lower[7]).toBeCloseTo(1)
    expect(upper[6]).toBeNull()
  })
})

describe('calculateRSI', () => {
  it('is 100 for a strictly rising series and 0 for a falling one', () => {
    expect(calculateRSI([1, 2, 3, 4, 5], 3)[3]).toBe(100)
    expect(calculateRSI([5, 4, 3, 2, 1], 3)[4]).toBe(0)
  })

  it('matches a hand-computed Wilder value', () => {
    // changes: +1, -1, +2, then +1
    const rsi = calculateRSI([10, 11, 10, 12, 13], 3)
    // seed avg gain 1, avg loss 1/3 -> RSI 75
    expect(rsi[3]).toBeCloseTo(75)
    // next: gain (1*2+1)/3 = 1, loss (1/3*2)/3 = 2/9 -> RS 4.5 -> 81.818
    expect(rsi[4]).toBeCloseTo(81.818, 2)
    expect(rsi.slice(0, 3)).toEqual([null, null, null])
  })
})

describe('calculateMACD', () => {
  it('derives signal and histogram from the EMA difference', () => {
    const values = Array.from({ length: 40 }, (_, i) => 100 + Math.sin(i / 3) * 10)
    const { macd, signal, hist } = calculateMACD(values, 3, 6, 4)
    const fast = calculateEMA(values, 3)
    const slow = calculateEMA(values, 6)
    expect(macd[4]).toBeNull()
    expect(macd[5]).toBeCloseTo(fast[5]! - slow[5]!)
    expect(signal[7]).toBeNull()
    expect(signal[8]).toBeCloseTo((macd[5]! + macd[6]! + macd[7]! + macd[8]!) / 4)
    expect(hist[20]).toBeCloseTo(macd[20]! - signal[20]!)
  })
})

describe('validParams', () => {
  it('enforces per-type bounds', () => {
    expect(validParams('bollinger', { period: 20, stddev: 2.5 })).toBe(true)
    expect(validParams('bollinger', { period: 20, stddev: 5 })).toBe(false)
    expect(validParams('rsi', { period: 14.5 })).toBe(false)
    expect(validParams('macd', { fast: 26, slow: 12, signal: 9 })).toBe(false)
    expect(validParams('vwap', {})).toBe(false)
  })
})

describe('aggregateDaily', () => {
  it('produces OHLC per UTC day in chronological order', () => {
    const at = (h: number, price: number): PricePoint => ({
      id: String(h),
      card_id: 'c',
      price,
      source: 'ebay',
      timestamp: new Date(Date.UTC(2025, 0, 1, h)).toISOString(),
      created_at: '',
    })
    const [day] = aggregateDaily([at(20, 30), at(2, 10), at(12, 40)], 'all')
    expect(day).toMatchObject({ open: 10, close: 30, high: 40, low: 10, price: 26.67, volume: 3 })
    expect(day.timestamp).toBe(Date.UTC(2025, 0, 1))
  })
})

describe('calculateEMA', () => {
  it('seeds with SMA and applies smoothing', () => {
    const ema = calculateEMA([1, 2, 3, 4, 5, 6], 3)
    // alpha = 0.5; seed = (1+2+3)/3 = 2; 4 -> 3; 5 -> 4; 6 -> 5
    expect(ema).toEqual([null, null, 2, 3, 4, 5])
  })

  it('returns all nulls when series is shorter than the period', () => {
    expect(calculateEMA([1, 2], 5)).toEqual([null, null])
  })

  it('rejects non-integer periods', () => {
    expect(calculateEMA([1, 2, 3], 2.5)).toEqual([null, null, null])
  })

  it('matches a hand-computed 10-period value', () => {
    const values = Array.from({ length: 12 }, (_, i) => 10 + i)
    const ema = calculateEMA(values, 10)
    const alpha = 2 / 11
    const seed = values.slice(0, 10).reduce((a, b) => a + b, 0) / 10
    const e11 = values[10] * alpha + seed * (1 - alpha)
    expect(ema[9]).toBeCloseTo(seed)
    expect(ema[10]).toBeCloseTo(e11)
  })
})

describe('warmupDaysFor', () => {
  it('uses 3x the longest period capped at 400', () => {
    expect(warmupDaysFor([])).toBe(0)
    expect(
      warmupDaysFor([
        { type: 'ema', parameters: { period: 20 }, visible: true },
        { type: 'ema', parameters: { period: 50 }, visible: true },
      ]),
    ).toBe(150)
    expect(warmupDaysFor([{ type: 'ema', parameters: { period: 200 }, visible: true }])).toBe(400)
    expect(
      warmupDaysFor([
        { type: 'macd', parameters: { fast: 12, slow: 26, signal: 9 }, visible: true },
      ]),
    ).toBe(105)
  })
})

describe('buildChartSeries', () => {
  const day = (n: number) => new Date(Date.UTC(2025, 0, 1 + n, 12)).toISOString()
  const prices: PricePoint[] = Array.from({ length: 10 }, (_, i) => ({
    id: String(i),
    card_id: 'c',
    price: i + 1,
    source: 'ebay',
    timestamp: day(i),
    created_at: day(i),
  }))

  it('computes EMA on warm-up data then trims to the visible range', () => {
    const series = buildChartSeries(
      prices,
      'all',
      [{ type: 'ema', parameters: { period: 3 }, visible: true }],
      day(5),
    )
    expect(series).toHaveLength(5)
    expect(series[0].price).toBe(6)
    expect(series[0][indicatorKey(0)]).toBe(5)
  })

  it('skips hidden indicators and filters by source', () => {
    const series = buildChartSeries(prices, 'tcgplayer', [
      { type: 'ema', parameters: { period: 3 }, visible: false },
    ])
    expect(series).toHaveLength(0)
    const all = buildChartSeries(prices, 'all', [
      { type: 'ema', parameters: { period: 3 }, visible: false },
    ])
    expect(all[5][indicatorKey(0)]).toBeUndefined()
  })

  it('writes multi-output indicators under suffixed keys', () => {
    const series = buildChartSeries(prices, 'all', [
      { type: 'bollinger', parameters: { period: 3, stddev: 2 }, visible: true },
    ])
    expect(series[2][seriesKey(0, 'value')]).toBe(2)
    expect(series[2][seriesKey(0, 'upper')]).toBeCloseTo(2 + 2 * Math.sqrt(2 / 3), 2)
    expect(series[2][seriesKey(0, 'lower')]).toBeCloseTo(2 - 2 * Math.sqrt(2 / 3), 2)
  })
})
