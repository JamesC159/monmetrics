import type { ChartDrawing, DrawingType } from '@/types'

export interface Pt {
  x: number
  y: number
}

export const FIB_LEVELS = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1]
export const HIT_TOLERANCE = 6

export const DRAWING_TOOLS: { type: DrawingType; label: string }[] = [
  { type: 'trendline', label: 'Trend line' },
  { type: 'hline', label: 'Horizontal line' },
  { type: 'rect', label: 'Rectangle' },
  { type: 'text', label: 'Text note' },
  { type: 'freehand', label: 'Freehand' },
  { type: 'fib', label: 'Fibonacci retracement' },
]

/** Tools placed with a single click rather than a drag. */
export const SINGLE_POINT_TOOLS: DrawingType[] = ['hline', 'text']

export function distToSegment(p: Pt, a: Pt, b: Pt): number {
  const dx = b.x - a.x
  const dy = b.y - a.y
  const len2 = dx * dx + dy * dy
  const t = len2 === 0 ? 0 : Math.max(0, Math.min(1, ((p.x - a.x) * dx + (p.y - a.y) * dy) / len2))
  return Math.hypot(p.x - (a.x + t * dx), p.y - (a.y + t * dy))
}

/** Price levels for a retracement drawn from points[0] to points[1]. */
export function fibLevels(p1: number, p2: number): { ratio: number; price: number }[] {
  return FIB_LEVELS.map((ratio) => ({ ratio, price: p2 - (p2 - p1) * ratio }))
}

/** Ramer–Douglas–Peucker polyline simplification. */
export function simplifyPath<T extends Pt>(points: T[], epsilon: number): T[] {
  if (points.length < 3) return points
  let maxDist = 0
  let index = 0
  const first = points[0]
  const last = points[points.length - 1]
  for (let i = 1; i < points.length - 1; i++) {
    const d = distToSegment(points[i], first, last)
    if (d > maxDist) {
      maxDist = d
      index = i
    }
  }
  if (maxDist <= epsilon) return [first, last]
  const left = simplifyPath(points.slice(0, index + 1), epsilon)
  const right = simplifyPath(points.slice(index), epsilon)
  return [...left.slice(0, -1), ...right]
}

export interface HitResult {
  id: string
  /** Index of the grabbed anchor, or -1 for the whole drawing */
  handle: number
}

/**
 * Hit-tests a drawing given its anchors projected to pixels. `width` is the pane width,
 * used for full-width horizontal lines. `textWidth` approximates a text note's box.
 */
export function hitTestDrawing(
  d: ChartDrawing,
  px: Pt[],
  p: Pt,
  width: number,
  textWidth = 80,
): HitResult | null {
  const tol = HIT_TOLERANCE
  for (let i = 0; i < px.length; i++) {
    if (d.type !== 'freehand' && Math.hypot(p.x - px[i].x, p.y - px[i].y) <= tol + 2)
      return { id: d.id, handle: i }
  }
  const hit = { id: d.id, handle: -1 }
  switch (d.type) {
    case 'hline':
      return Math.abs(p.y - px[0].y) <= tol && p.x >= 0 && p.x <= width ? hit : null
    case 'trendline':
      return distToSegment(p, px[0], px[1]) <= tol ? hit : null
    case 'freehand':
      for (let i = 1; i < px.length; i++) if (distToSegment(p, px[i - 1], px[i]) <= tol) return hit
      return null
    case 'rect': {
      const [x1, x2] = [Math.min(px[0].x, px[1].x), Math.max(px[0].x, px[1].x)]
      const [y1, y2] = [Math.min(px[0].y, px[1].y), Math.max(px[0].y, px[1].y)]
      return p.x >= x1 - tol && p.x <= x2 + tol && p.y >= y1 - tol && p.y <= y2 + tol ? hit : null
    }
    case 'text':
      return p.x >= px[0].x - 4 &&
        p.x <= px[0].x + textWidth + 4 &&
        p.y >= px[0].y - 18 &&
        p.y <= px[0].y + 4
        ? hit
        : null
    case 'fib': {
      const [x1, x2] = [Math.min(px[0].x, px[1].x), Math.max(px[0].x, px[1].x)]
      if (p.x < x1 - tol || p.x > x2 + tol) return null
      for (const ratio of FIB_LEVELS) {
        const y = px[1].y - (px[1].y - px[0].y) * ratio
        if (Math.abs(p.y - y) <= tol) return hit
      }
      return null
    }
  }
  return null
}

export function newDrawingId(): string {
  const rand =
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : Math.random().toString(36).slice(2)
  return rand.replace(/[^A-Za-z0-9-]/g, '').slice(0, 36)
}
