import { describe, expect, it } from 'vitest'
import { distToSegment, fibLevels, hitTestDrawing, simplifyPath } from './geometry'
import type { ChartDrawing } from '@/types'

const drawing = (type: ChartDrawing['type'], n: number, text?: string): ChartDrawing => ({
  id: 'd1',
  type,
  points: Array.from({ length: n }, (_, i) => ({
    time: 1_700_000_000 + i * 86_400,
    price: 10 + i,
  })),
  color: '#ffffff',
  line_width: 2,
  text,
})

describe('distToSegment', () => {
  it('measures perpendicular and endpoint distances', () => {
    expect(distToSegment({ x: 5, y: 5 }, { x: 0, y: 0 }, { x: 10, y: 0 })).toBe(5)
    expect(distToSegment({ x: 13, y: 4 }, { x: 0, y: 0 }, { x: 10, y: 0 })).toBe(5)
  })
})

describe('simplifyPath', () => {
  it('drops collinear points and keeps corners', () => {
    const pts = [
      { x: 0, y: 0 },
      { x: 5, y: 0.1 },
      { x: 10, y: 0 },
      { x: 10, y: 10 },
    ]
    expect(simplifyPath(pts, 1)).toEqual([pts[0], pts[2], pts[3]])
  })
})

describe('fibLevels', () => {
  it('retraces from the second point toward the first', () => {
    const levels = fibLevels(100, 200)
    expect(levels[0]).toEqual({ ratio: 0, price: 200 })
    expect(levels.find((l) => l.ratio === 0.5)?.price).toBe(150)
    expect(levels[levels.length - 1].price).toBe(100)
  })
})

describe('hitTestDrawing', () => {
  it('detects handles before bodies', () => {
    const d = drawing('trendline', 2)
    expect(
      hitTestDrawing(
        d,
        [
          { x: 0, y: 0 },
          { x: 100, y: 100 },
        ],
        { x: 1, y: 1 },
        500,
      ),
    ).toEqual({ id: 'd1', handle: 0 })
    expect(
      hitTestDrawing(
        d,
        [
          { x: 0, y: 0 },
          { x: 100, y: 100 },
        ],
        { x: 50, y: 52 },
        500,
      ),
    ).toEqual({ id: 'd1', handle: -1 })
    expect(
      hitTestDrawing(
        d,
        [
          { x: 0, y: 0 },
          { x: 100, y: 100 },
        ],
        { x: 50, y: 80 },
        500,
      ),
    ).toBeNull()
  })

  it('treats horizontal lines as full width', () => {
    const d = drawing('hline', 1)
    expect(hitTestDrawing(d, [{ x: 200, y: 40 }], { x: 10, y: 43 }, 500)?.handle).toBe(-1)
    expect(hitTestDrawing(d, [{ x: 200, y: 40 }], { x: 10, y: 60 }, 500)).toBeNull()
  })

  it('hits rectangles inside and fib levels only on their lines', () => {
    expect(
      hitTestDrawing(
        drawing('rect', 2),
        [
          { x: 0, y: 0 },
          { x: 100, y: 50 },
        ],
        { x: 50, y: 25 },
        500,
      ),
    ).not.toBeNull()
    const fib = drawing('fib', 2)
    const px = [
      { x: 0, y: 100 },
      { x: 100, y: 0 },
    ]
    expect(hitTestDrawing(fib, px, { x: 50, y: 50 }, 500)).not.toBeNull() // 50% level
    expect(hitTestDrawing(fib, px, { x: 50, y: 70 }, 500)).toBeNull()
  })
})
