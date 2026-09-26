import type {
  IChartApi,
  ISeriesApi,
  ISeriesPrimitive,
  IPrimitivePaneRenderer,
  IPrimitivePaneView,
  Logical,
  SeriesAttachedParameter,
  SeriesType,
  Time,
} from 'lightweight-charts'
import type { CanvasRenderingTarget2D } from 'fancy-canvas'
import type { ChartDrawing } from '@/types'
import { fibLevels, hitTestDrawing, type HitResult, type Pt } from './geometry'

const DAY_SECONDS = 86_400

/** Renders user drawings on the main price pane, anchored in (time, price) space. */
export class DrawingPrimitive implements ISeriesPrimitive<Time> {
  private chart: IChartApi | null = null
  private series: ISeriesApi<SeriesType> | null = null
  private requestUpdate: (() => void) | null = null
  private drawings: ChartDrawing[] = []
  private preview: ChartDrawing | null = null
  private selectedId: string | null = null
  private times: number[] = []
  private readonly view: IPrimitivePaneView

  constructor() {
    const renderer: IPrimitivePaneRenderer = { draw: (target) => this.draw(target) }
    this.view = { renderer: () => renderer, zOrder: () => 'top' }
  }

  attached(param: SeriesAttachedParameter<Time>) {
    this.chart = param.chart as IChartApi
    this.series = param.series
    this.requestUpdate = param.requestUpdate
  }

  detached() {
    this.chart = null
    this.series = null
    this.requestUpdate = null
  }

  paneViews() {
    return [this.view]
  }

  setTimes(times: number[]) {
    this.times = times
    this.requestUpdate?.()
  }

  setState(drawings: ChartDrawing[], preview: ChartDrawing | null, selectedId: string | null) {
    this.drawings = drawings
    this.preview = preview
    this.selectedId = selectedId
    this.requestUpdate?.()
  }

  /** Maps a time to an x coordinate, interpolating between daily bars and extrapolating past the ends. */
  timeToX(t: number): number | null {
    const ts = this.times
    if (!this.chart || ts.length === 0) return null
    let logical: number
    if (t <= ts[0]) logical = (t - ts[0]) / DAY_SECONDS
    else if (t >= ts[ts.length - 1]) logical = ts.length - 1 + (t - ts[ts.length - 1]) / DAY_SECONDS
    else {
      let lo = 0
      let hi = ts.length - 1
      while (hi - lo > 1) {
        const mid = (lo + hi) >> 1
        if (ts[mid] <= t) lo = mid
        else hi = mid
      }
      logical = lo + (t - ts[lo]) / (ts[hi] - ts[lo])
    }
    return this.chart.timeScale().logicalToCoordinate(logical as Logical)
  }

  xToTime(x: number): number | null {
    const ts = this.times
    if (!this.chart || ts.length === 0) return null
    const logical = this.chart.timeScale().coordinateToLogical(x)
    if (logical === null) return null
    const l = Number(logical)
    if (l <= 0) return Math.round(ts[0] + l * DAY_SECONDS)
    if (l >= ts.length - 1)
      return Math.round(ts[ts.length - 1] + (l - (ts.length - 1)) * DAY_SECONDS)
    const i = Math.floor(l)
    return Math.round(ts[i] + (l - i) * (ts[i + 1] - ts[i]))
  }

  priceToY(price: number): number | null {
    return this.series?.priceToCoordinate(price) ?? null
  }

  yToPrice(y: number): number | null {
    const p = this.series?.coordinateToPrice(y)
    return p === null || p === undefined ? null : Number(p)
  }

  project(d: ChartDrawing): Pt[] | null {
    const out: Pt[] = []
    for (const p of d.points) {
      const x = this.timeToX(p.time)
      const y = this.priceToY(p.price)
      if (x === null || y === null) return null
      out.push({ x, y })
    }
    return out
  }

  /** Topmost drawing under the point, if any. */
  hit(p: Pt): HitResult | null {
    const width = this.chart?.timeScale().width() ?? 0
    for (let i = this.drawings.length - 1; i >= 0; i--) {
      const d = this.drawings[i]
      const px = this.project(d)
      if (!px) continue
      const r = hitTestDrawing(d, px, p, width, textWidth(d.text ?? ''))
      if (r) return r
    }
    return null
  }

  private draw(target: CanvasRenderingTarget2D) {
    target.useMediaCoordinateSpace(({ context: ctx, mediaSize }) => {
      const all = this.preview ? [...this.drawings, this.preview] : this.drawings
      for (const d of all) {
        const px = this.project(d)
        if (!px) continue
        ctx.save()
        ctx.strokeStyle = d.color
        ctx.fillStyle = d.color
        ctx.lineWidth = d.line_width || 2
        ctx.lineCap = 'round'
        ctx.lineJoin = 'round'
        renderDrawing(ctx, d, px, mediaSize.width)
        if (d.id === this.selectedId) drawHandles(ctx, d, px)
        ctx.restore()
      }
    })
  }
}

function textWidth(text: string): number {
  return Math.max(20, text.length * 7 + 8)
}

function renderDrawing(ctx: CanvasRenderingContext2D, d: ChartDrawing, px: Pt[], width: number) {
  switch (d.type) {
    case 'trendline':
      line(ctx, px[0], px[1])
      break
    case 'hline': {
      line(ctx, { x: 0, y: px[0].y }, { x: width, y: px[0].y })
      label(
        ctx,
        `${formatPrice(d.points[0].price)}${d.text ? ` · ${d.text}` : ''}`,
        6,
        px[0].y - 6,
        d.color,
      )
      break
    }
    case 'rect': {
      const x = Math.min(px[0].x, px[1].x)
      const y = Math.min(px[0].y, px[1].y)
      const w = Math.abs(px[1].x - px[0].x)
      const h = Math.abs(px[1].y - px[0].y)
      ctx.globalAlpha = 0.15
      ctx.fillRect(x, y, w, h)
      ctx.globalAlpha = 1
      ctx.strokeRect(x, y, w, h)
      break
    }
    case 'text': {
      ctx.font = '12px system-ui, sans-serif'
      const text = d.text ?? ''
      const w = textWidth(text)
      ctx.globalAlpha = 0.85
      ctx.fillStyle = '#111827'
      ctx.fillRect(px[0].x - 4, px[0].y - 16, w, 20)
      ctx.globalAlpha = 1
      ctx.strokeRect(px[0].x - 4, px[0].y - 16, w, 20)
      ctx.fillStyle = d.color
      ctx.fillText(text, px[0].x, px[0].y - 2)
      break
    }
    case 'freehand': {
      ctx.beginPath()
      ctx.moveTo(px[0].x, px[0].y)
      for (let i = 1; i < px.length; i++) ctx.lineTo(px[i].x, px[i].y)
      ctx.stroke()
      break
    }
    case 'fib': {
      const x1 = Math.min(px[0].x, px[1].x)
      const x2 = Math.max(px[0].x, px[1].x)
      ctx.setLineDash([4, 4])
      line(ctx, px[0], px[1])
      ctx.setLineDash([])
      for (const { ratio, price } of fibLevels(d.points[0].price, d.points[1].price)) {
        const y = px[1].y - (px[1].y - px[0].y) * ratio
        ctx.lineWidth = 1
        line(ctx, { x: x1, y }, { x: x2, y })
        label(ctx, `${(ratio * 100).toFixed(1)}% ${formatPrice(price)}`, x1 + 4, y - 3, d.color)
      }
      break
    }
  }
}

function drawHandles(ctx: CanvasRenderingContext2D, d: ChartDrawing, px: Pt[]) {
  const handles = d.type === 'freehand' ? [px[0], px[px.length - 1]] : px
  ctx.fillStyle = '#ffffff'
  ctx.lineWidth = 1.5
  for (const p of handles) {
    ctx.beginPath()
    ctx.arc(p.x, p.y, 4, 0, Math.PI * 2)
    ctx.fill()
    ctx.stroke()
  }
}

function line(ctx: CanvasRenderingContext2D, a: Pt, b: Pt) {
  ctx.beginPath()
  ctx.moveTo(a.x, a.y)
  ctx.lineTo(b.x, b.y)
  ctx.stroke()
}

function label(ctx: CanvasRenderingContext2D, text: string, x: number, y: number, color: string) {
  ctx.font = '11px system-ui, sans-serif'
  ctx.fillStyle = color
  ctx.fillText(text, x, y)
}

function formatPrice(v: number): string {
  return `$${v.toFixed(2)}`
}
