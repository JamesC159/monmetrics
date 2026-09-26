import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
} from 'react'
import type {
  IChartApi,
  IPriceLine,
  ISeriesApi,
  SeriesType,
  UTCTimestamp,
} from 'lightweight-charts'
import type {
  ChartDrawing,
  ChartIndicator,
  ChartSource,
  ChartTimeRange,
  ChartType,
  DrawingPoint,
  PriceAlert,
  PriceHistory,
} from '@/types'
import { apiClient } from '@/utils/api'
import { formatPrice } from '@/utils/formatters'
import {
  buildChartSeries,
  indicatorLabel,
  seriesKey,
  warmupDaysFor,
  type ChartPoint,
} from '@/utils/indicators'
import { DrawingPrimitive } from './drawing/DrawingPrimitive'
import DrawingToolbar, { DRAWING_COLORS, type Tool } from './drawing/DrawingToolbar'
import {
  newDrawingId,
  simplifyPath,
  SINGLE_POINT_TOOLS,
  type HitResult,
  type Pt,
} from './drawing/geometry'

export const TIME_RANGE_OPTIONS: { value: ChartTimeRange; label: string }[] = [
  { value: '1d', label: '1D' },
  { value: '7d', label: '7D' },
  { value: '30d', label: '30D' },
  { value: '90d', label: '90D' },
  { value: '1y', label: '1Y' },
  { value: '5y', label: '5Y' },
]

export const SOURCE_OPTIONS: { value: ChartSource; label: string }[] = [
  { value: 'all', label: 'All Sources' },
  { value: 'ebay', label: 'eBay' },
  { value: 'tcgplayer', label: 'TCGPlayer' },
]

interface PriceChartProps {
  cardId: string
  timeRange: ChartTimeRange
  onTimeRangeChange: (range: ChartTimeRange) => void
  source: ChartSource
  onSourceChange: (source: ChartSource) => void
  indicators?: ChartIndicator[]
  title?: string
  headerActions?: ReactNode
  children?: ReactNode
  onHistory?: (history: PriceHistory | null) => void
  showStats?: boolean
  chartType?: ChartType
  onChartTypeChange?: (type: ChartType) => void
  showVolume?: boolean
  onShowVolumeChange?: (show: boolean) => void
  /** Providing onDrawingsChange enables the drawing toolbar. */
  drawings?: ChartDrawing[]
  onDrawingsChange?: (drawings: ChartDrawing[]) => void
  alerts?: PriceAlert[]
  onAlertCreateAt?: (price: number) => void
  onAlertMove?: (alert: PriceAlert, price: number) => void
}

type Drag =
  | { mode: 'create'; drawing: ChartDrawing; pixels: (Pt & DrawingPoint)[] }
  | {
      mode: 'move'
      hit: HitResult
      start: DrawingPoint
      original: ChartDrawing
      working: ChartDrawing[]
    }
  | { mode: 'alert'; alert: PriceAlert; line: IPriceLine; price: number }

const toTime = (ms: number) => Math.floor(ms / 1000) as UTCTimestamp
const round2 = (v: number) => Math.round(v * 100) / 100
const MACD_SIGNAL_COLOR = '#f97316'
const NO_DRAWINGS: ChartDrawing[] = []

export default function PriceChart({
  cardId,
  timeRange,
  onTimeRangeChange,
  source,
  onSourceChange,
  indicators = [],
  title = 'Price History',
  headerActions,
  children,
  onHistory,
  showStats = true,
  chartType = 'line',
  onChartTypeChange,
  showVolume = false,
  onShowVolumeChange,
  drawings,
  onDrawingsChange,
  alerts,
  onAlertCreateAt,
  onAlertMove,
}: PriceChartProps) {
  const [history, setHistory] = useState<PriceHistory | null>(null)
  const [loading, setLoading] = useState(false)
  const warmup = warmupDaysFor(indicators)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiClient
      .getCardPrices(cardId, timeRange, warmup)
      .then((data) => {
        if (cancelled) return
        setHistory(data)
        onHistory?.(data)
      })
      .catch((err) => {
        if (cancelled) return
        console.error('Failed to load price history:', err)
        setHistory(null)
        onHistory?.(null)
      })
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
    // onHistory is intentionally excluded so parents need not memoize it
  }, [cardId, timeRange, warmup])

  const data = useMemo(
    () => buildChartSeries(history?.prices ?? [], source, indicators, history?.visible_from),
    [history, source, indicators],
  )

  const drawingEnabled = !!onDrawingsChange
  const drawingList = drawings ?? NO_DRAWINGS
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const libRef = useRef<typeof import('lightweight-charts') | null>(null)
  const mainSeriesRef = useRef<ISeriesApi<SeriesType> | null>(null)
  const extraSeriesRef = useRef<ISeriesApi<SeriesType>[]>([])
  const primitiveRef = useRef<DrawingPrimitive | null>(null)
  const alertLinesRef = useRef<Map<string, IPriceLine>>(new Map())
  const dragRef = useRef<Drag | null>(null)
  const undoRef = useRef<ChartDrawing[][]>([])
  const redoRef = useRef<ChartDrawing[][]>([])
  const [libReady, setLibReady] = useState(false)
  const [seriesVersion, setSeriesVersion] = useState(0)
  const [tool, setTool] = useState<Tool>('cursor')
  const [color, setColor] = useState(DRAWING_COLORS[0])
  const [lineWidth, setLineWidth] = useState(2)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [preview, setPreview] = useState<ChartDrawing | null>(null)
  const [pendingText, setPendingText] = useState<{
    x: number
    y: number
    point: DrawingPoint
  } | null>(null)
  const [hoverTime, setHoverTime] = useState<number | null>(null)
  const [, setHistoryVersion] = useState(0)

  const byTime = useMemo(() => new Map(data.map((p) => [toTime(p.timestamp) as number, p])), [data])

  // Handlers read the latest values through a ref so listeners are attached only once.
  const latest = useRef({
    tool,
    color,
    lineWidth,
    drawings: drawingList,
    alerts,
    onDrawingsChange,
    onAlertMove,
    onAlertCreateAt,
    selectedId,
  })
  latest.current = {
    tool,
    color,
    lineWidth,
    drawings: drawingList,
    alerts,
    onDrawingsChange,
    onAlertMove,
    onAlertCreateAt,
    selectedId,
  }

  const commit = useCallback((next: ChartDrawing[]) => {
    const { drawings: current, onDrawingsChange: change } = latest.current
    if (!change) return
    undoRef.current = [...undoRef.current.slice(-49), current]
    redoRef.current = []
    setHistoryVersion((v) => v + 1)
    change(next)
  }, [])

  const undo = useCallback(() => {
    const prev = undoRef.current.pop()
    if (!prev) return
    redoRef.current.push(latest.current.drawings)
    setHistoryVersion((v) => v + 1)
    setSelectedId(null)
    latest.current.onDrawingsChange?.(prev)
  }, [])

  const redo = useCallback(() => {
    const next = redoRef.current.pop()
    if (!next) return
    undoRef.current.push(latest.current.drawings)
    setHistoryVersion((v) => v + 1)
    latest.current.onDrawingsChange?.(next)
  }, [])

  // Chart instance; the library touches the DOM so it is only loaded in the browser.
  useEffect(() => {
    let disposed = false
    import('lightweight-charts').then((lib) => {
      if (disposed || !containerRef.current) return
      libRef.current = lib
      chartRef.current = lib.createChart(containerRef.current, {
        autoSize: true,
        layout: {
          background: { type: lib.ColorType.Solid, color: 'transparent' },
          textColor: '#9CA3AF',
          fontSize: 12,
          attributionLogo: false,
        },
        grid: {
          vertLines: { color: 'rgba(55,65,81,0.5)' },
          horzLines: { color: 'rgba(55,65,81,0.5)' },
        },
        rightPriceScale: { borderColor: '#374151' },
        timeScale: { borderColor: '#374151', rightOffset: 4 },
        crosshair: { mode: lib.CrosshairMode.Normal },
      })
      chartRef.current.subscribeCrosshairMove((param) =>
        setHoverTime(param.time === undefined ? null : Number(param.time)),
      )
      setLibReady(true)
    })
    return () => {
      disposed = true
      chartRef.current?.remove()
      chartRef.current = null
      mainSeriesRef.current = null
      extraSeriesRef.current = []
    }
  }, [])

  // (Re)build all series whenever the data or display options change.
  useEffect(() => {
    const chart = chartRef.current
    const lib = libRef.current
    if (!chart || !lib) return

    if (mainSeriesRef.current && primitiveRef.current)
      mainSeriesRef.current.detachPrimitive(primitiveRef.current)
    for (const s of [mainSeriesRef.current, ...extraSeriesRef.current]) if (s) chart.removeSeries(s)
    while (chart.panes().length > 1) chart.removePane(chart.panes().length - 1)
    extraSeriesRef.current = []
    alertLinesRef.current.clear()
    mainSeriesRef.current = null
    if (data.length === 0) return

    const priceFormat = {
      type: 'custom' as const,
      minMove: 0.01,
      formatter: (p: number) => formatPrice(p),
    }
    const main: ISeriesApi<SeriesType> =
      chartType === 'candle'
        ? chart.addSeries(lib.CandlestickSeries, {
            priceFormat,
            upColor: '#22c55e',
            downColor: '#ef4444',
            borderVisible: false,
            wickUpColor: '#22c55e',
            wickDownColor: '#ef4444',
          })
        : chart.addSeries(lib.LineSeries, { priceFormat, color: '#3B82F6', lineWidth: 2 })
    main.setData(
      data.map((p) =>
        chartType === 'candle'
          ? { time: toTime(p.timestamp), open: p.open, high: p.high, low: p.low, close: p.close }
          : { time: toTime(p.timestamp), value: p.price },
      ),
    )
    mainSeriesRef.current = main

    const extras: ISeriesApi<SeriesType>[] = []
    const lineData = (key: string) =>
      data
        .filter((p) => typeof p[key] === 'number')
        .map((p) => ({ time: toTime(p.timestamp), value: p[key] as number }))
    const overlay = {
      lastValueVisible: false,
      priceLineVisible: false,
      crosshairMarkerVisible: false,
      priceFormat,
    }

    if (showVolume) {
      const vol = chart.addSeries(lib.HistogramSeries, {
        priceScaleId: 'volume',
        color: 'rgba(148,163,184,0.35)',
        priceFormat: { type: 'volume' },
        lastValueVisible: false,
        priceLineVisible: false,
      })
      vol.setData(data.map((p) => ({ time: toTime(p.timestamp), value: p.volume })))
      chart.priceScale('volume').applyOptions({ scaleMargins: { top: 0.82, bottom: 0 } })
      extras.push(vol)
    }

    let pane = 1
    indicators.forEach((ind, i) => {
      if (!ind.visible) return
      const color = ind.color ?? '#22c55e'
      const title = indicatorLabel(ind)
      switch (ind.type) {
        case 'ema':
        case 'sma': {
          const s = chart.addSeries(lib.LineSeries, { ...overlay, color, lineWidth: 2, title })
          s.setData(lineData(seriesKey(i, 'value')))
          extras.push(s)
          break
        }
        case 'bollinger': {
          const mid = chart.addSeries(lib.LineSeries, { ...overlay, color, lineWidth: 1, title })
          mid.setData(lineData(seriesKey(i, 'value')))
          for (const key of ['upper', 'lower']) {
            const band = chart.addSeries(lib.LineSeries, {
              ...overlay,
              color,
              lineWidth: 1,
              lineStyle: lib.LineStyle.Dashed,
            })
            band.setData(lineData(seriesKey(i, key)))
            extras.push(band)
          }
          extras.push(mid)
          break
        }
        case 'rsi': {
          const s = chart.addSeries(
            lib.LineSeries,
            {
              color,
              lineWidth: 2,
              title,
              priceFormat: { type: 'price', precision: 1, minMove: 0.1 },
            },
            pane,
          )
          s.setData(lineData(seriesKey(i, 'value')))
          for (const level of [70, 30])
            s.createPriceLine({
              price: level,
              color: '#6b7280',
              lineStyle: lib.LineStyle.Dashed,
              lineWidth: 1,
              axisLabelVisible: false,
            })
          extras.push(s)
          pane++
          break
        }
        case 'macd': {
          const macdFormat = { type: 'price' as const, precision: 2, minMove: 0.01 }
          const hist = chart.addSeries(
            lib.HistogramSeries,
            { priceFormat: macdFormat, lastValueVisible: false, priceLineVisible: false },
            pane,
          )
          hist.setData(
            data
              .filter((p) => typeof p[seriesKey(i, 'hist')] === 'number')
              .map((p) => {
                const v = p[seriesKey(i, 'hist')] as number
                return {
                  time: toTime(p.timestamp),
                  value: v,
                  color: v >= 0 ? 'rgba(34,197,94,0.5)' : 'rgba(239,68,68,0.5)',
                }
              }),
          )
          const m = chart.addSeries(
            lib.LineSeries,
            { color, lineWidth: 2, title, priceFormat: macdFormat },
            pane,
          )
          m.setData(lineData(seriesKey(i, 'value')))
          const sig = chart.addSeries(
            lib.LineSeries,
            {
              color: MACD_SIGNAL_COLOR,
              lineWidth: 1,
              priceFormat: macdFormat,
              lastValueVisible: false,
            },
            pane,
          )
          sig.setData(lineData(seriesKey(i, 'signal')))
          extras.push(hist, m, sig)
          pane++
          break
        }
      }
    })
    extraSeriesRef.current = extras
    chart.panes().forEach((p, idx) => p.setStretchFactor(idx === 0 ? 3 : 1))

    if (!primitiveRef.current) primitiveRef.current = new DrawingPrimitive()
    main.attachPrimitive(primitiveRef.current)
    primitiveRef.current.setTimes(data.map((p) => toTime(p.timestamp) as number))
    chart.timeScale().fitContent()
    setSeriesVersion((v) => v + 1)
  }, [libReady, data, chartType, showVolume, indicators])

  // Alert price lines on the main series
  useEffect(() => {
    const main = mainSeriesRef.current
    const lib = libRef.current
    if (!main || !lib) return
    for (const line of alertLinesRef.current.values()) main.removePriceLine(line)
    alertLinesRef.current.clear()
    for (const a of alerts ?? []) {
      if ((a.condition !== 'above' && a.condition !== 'below') || !a.target_price) continue
      const line = main.createPriceLine({
        price: a.target_price,
        color: a.active ? '#f59e0b' : '#6b7280',
        lineWidth: 1,
        lineStyle: lib.LineStyle.Dashed,
        axisLabelVisible: true,
        title: `🔔 ${a.condition === 'above' ? '↑' : '↓'}`,
      })
      alertLinesRef.current.set(a.id, line)
    }
  }, [alerts, seriesVersion])

  useEffect(() => {
    primitiveRef.current?.setState(drawingList, preview, selectedId)
  }, [drawingList, preview, selectedId, seriesVersion])

  // Pointer interaction for drawing, moving drawings and dragging alert lines
  useEffect(() => {
    const el = containerRef.current
    if (!el || !libReady || (!drawingEnabled && !onAlertMove && !onAlertCreateAt)) return
    let capturing = false

    const local = (e: { clientX: number; clientY: number }): Pt | null => {
      const chart = chartRef.current
      if (!chart) return null
      const rect = el.getBoundingClientRect()
      const p = { x: e.clientX - rect.left, y: e.clientY - rect.top }
      const paneHeight = chart.panes()[0]?.getHeight() ?? 0
      if (p.x < 0 || p.x > chart.timeScale().width() || p.y < 0 || p.y > paneHeight) return null
      return p
    }
    const toData = (p: Pt): DrawingPoint | null => {
      const prim = primitiveRef.current
      const time = prim?.xToTime(p.x)
      const price = prim?.yToPrice(p.y)
      if (time == null || price == null || price <= 0) return null
      return { time, price: round2(price) }
    }
    const alertAt = (p: Pt) => {
      const main = mainSeriesRef.current
      for (const a of latest.current.alerts ?? []) {
        const line = alertLinesRef.current.get(a.id)
        const y = line ? main?.priceToCoordinate(line.options().price) : null
        if (line && y != null && Math.abs(y - p.y) <= 5) return { alert: a, line }
      }
      return null
    }

    const onPointerDown = (e: PointerEvent) => {
      if (e.button !== 0) return
      const p = local(e)
      if (!p) return
      const { tool: t, color: c, lineWidth: w, drawings: list } = latest.current
      const point = toData(p)
      el.focus({ preventScroll: true })

      if (drawingEnabled && t !== 'cursor') {
        if (!point) return
        capturing = true
        e.preventDefault()
        e.stopPropagation()
        if (t === 'text') {
          setPendingText({ x: p.x, y: p.y, point })
          return
        }
        if (SINGLE_POINT_TOOLS.includes(t)) {
          commit([
            ...list,
            { id: newDrawingId(), type: t, points: [point], color: c, line_width: w },
          ])
          setTool('cursor')
          return
        }
        const drawing: ChartDrawing = {
          id: newDrawingId(),
          type: t,
          points: t === 'freehand' ? [point] : [point, point],
          color: c,
          line_width: w,
        }
        dragRef.current = { mode: 'create', drawing, pixels: [{ ...p, ...point }] }
        setPreview(drawing)
        el.setPointerCapture(e.pointerId)
        return
      }

      const hit = drawingEnabled ? primitiveRef.current?.hit(p) : null
      if (hit && point) {
        capturing = true
        e.preventDefault()
        e.stopPropagation()
        setSelectedId(hit.id)
        const original = list.find((d) => d.id === hit.id)!
        dragRef.current = { mode: 'move', hit, start: point, original, working: list }
        el.setPointerCapture(e.pointerId)
        return
      }
      const alertHit = latest.current.onAlertMove ? alertAt(p) : null
      if (alertHit) {
        capturing = true
        e.preventDefault()
        e.stopPropagation()
        dragRef.current = { mode: 'alert', ...alertHit, price: alertHit.line.options().price }
        el.setPointerCapture(e.pointerId)
        return
      }
      setSelectedId(null)
    }

    const onPointerMove = (e: PointerEvent) => {
      const drag = dragRef.current
      const rect = el.getBoundingClientRect()
      const p = { x: e.clientX - rect.left, y: e.clientY - rect.top }
      if (!drag) {
        const inPane = local(e)
        const { tool: t } = latest.current
        let cursor = ''
        if (inPane && drawingEnabled && t !== 'cursor') cursor = 'crosshair'
        else if (inPane && drawingEnabled && primitiveRef.current?.hit(inPane)) cursor = 'move'
        else if (inPane && latest.current.onAlertMove && alertAt(inPane)) cursor = 'ns-resize'
        el.style.cursor = cursor
        return
      }
      const point = toData(p)
      if (!point) return
      if (drag.mode === 'create') {
        if (drag.drawing.type === 'freehand') {
          const last = drag.pixels[drag.pixels.length - 1]
          if (Math.hypot(p.x - last.x, p.y - last.y) < 2) return
          drag.pixels.push({ ...p, ...point })
          drag.drawing = {
            ...drag.drawing,
            points: drag.pixels.map(({ time, price }) => ({ time, price })),
          }
        } else {
          drag.drawing = { ...drag.drawing, points: [drag.drawing.points[0], point] }
        }
        setPreview(drag.drawing)
      } else if (drag.mode === 'move') {
        const dt = point.time - drag.start.time
        const dp = point.price - drag.start.price
        const moved: ChartDrawing = {
          ...drag.original,
          points: drag.original.points.map((pt, i) =>
            drag.hit.handle === -1
              ? { time: pt.time + dt, price: Math.max(0.01, round2(pt.price + dp)) }
              : i === drag.hit.handle
                ? point
                : pt,
          ),
        }
        drag.working = latest.current.drawings.map((d) => (d.id === moved.id ? moved : d))
        primitiveRef.current?.setState(drag.working, null, moved.id)
      } else {
        drag.price = point.price
        drag.line.applyOptions({ price: point.price })
      }
    }

    const onPointerUp = () => {
      const drag = dragRef.current
      dragRef.current = null
      capturing = false
      if (!drag) return
      if (drag.mode === 'create') {
        setPreview(null)
        let d = drag.drawing
        if (d.type === 'freehand') {
          const simplified = simplifyPath(drag.pixels, 1.5).slice(0, 500)
          if (simplified.length < 2) return
          d = { ...d, points: simplified.map(({ time, price }) => ({ time, price })) }
        } else {
          const a = primitiveRef.current?.project(d)
          if (!a || Math.hypot(a[1].x - a[0].x, a[1].y - a[0].y) < 4) return
        }
        commit([...latest.current.drawings, d])
        setSelectedId(d.id)
        setTool('cursor')
      } else if (drag.mode === 'move') {
        if (drag.working !== latest.current.drawings) commit(drag.working)
      } else if (drag.price !== drag.alert.target_price) {
        latest.current.onAlertMove?.(drag.alert, round2(drag.price))
      }
    }

    // Keep the chart from panning while we own the gesture.
    const block = (e: Event) => {
      if (capturing) e.stopPropagation()
    }
    const onContextMenu = (e: MouseEvent) => {
      const create = latest.current.onAlertCreateAt
      const p = local(e)
      if (!create || !p) return
      e.preventDefault()
      const price = primitiveRef.current?.yToPrice(p.y)
      if (price != null && price > 0) create(round2(price))
    }

    el.addEventListener('pointerdown', onPointerDown, true)
    el.addEventListener('pointermove', onPointerMove, true)
    el.addEventListener('pointerup', onPointerUp, true)
    el.addEventListener('pointercancel', onPointerUp, true)
    el.addEventListener('mousedown', block, true)
    el.addEventListener('touchstart', block, true)
    el.addEventListener('contextmenu', onContextMenu)
    return () => {
      el.removeEventListener('pointerdown', onPointerDown, true)
      el.removeEventListener('pointermove', onPointerMove, true)
      el.removeEventListener('pointerup', onPointerUp, true)
      el.removeEventListener('pointercancel', onPointerUp, true)
      el.removeEventListener('mousedown', block, true)
      el.removeEventListener('touchstart', block, true)
      el.removeEventListener('contextmenu', onContextMenu)
    }
  }, [libReady, drawingEnabled, commit, !!onAlertMove, !!onAlertCreateAt])

  useEffect(() => {
    if (containerRef.current)
      containerRef.current.style.touchAction = drawingEnabled && tool !== 'cursor' ? 'none' : ''
  }, [tool, drawingEnabled])

  const deleteSelected = () => {
    if (!selectedId) return
    commit(drawingList.filter((d) => d.id !== selectedId))
    setSelectedId(null)
  }

  const onKeyDown = (e: ReactKeyboardEvent<HTMLDivElement>) => {
    if (!drawingEnabled || e.target !== e.currentTarget) return
    const mod = e.ctrlKey || e.metaKey
    if (e.key === 'Escape') {
      setTool('cursor')
      setSelectedId(null)
      setPreview(null)
      dragRef.current = null
    } else if ((e.key === 'Delete' || e.key === 'Backspace') && selectedId) {
      e.preventDefault()
      deleteSelected()
    } else if (mod && e.key.toLowerCase() === 'z' && !e.shiftKey) {
      e.preventDefault()
      undo()
    } else if (
      mod &&
      (e.key.toLowerCase() === 'y' || (e.key.toLowerCase() === 'z' && e.shiftKey))
    ) {
      e.preventDefault()
      redo()
    }
  }

  const submitText = (text: string) => {
    const trimmed = text.trim().slice(0, 200)
    if (pendingText && trimmed) {
      commit([
        ...drawingList,
        {
          id: newDrawingId(),
          type: 'text',
          points: [pendingText.point],
          color,
          line_width: lineWidth,
          text: trimmed,
        },
      ])
    }
    setPendingText(null)
    setTool('cursor')
  }

  const hovered: ChartPoint | undefined =
    (hoverTime !== null ? byTime.get(hoverTime) : undefined) ?? data[data.length - 1]

  return (
    <div className='glass-effect rounded-2xl p-6'>
      <div className='flex flex-wrap items-center justify-between gap-3 mb-4'>
        <h3 className='text-lg font-semibold text-white'>{title}</h3>
        <div className='flex flex-wrap items-center gap-2'>
          {onChartTypeChange && (
            <div
              className='flex rounded overflow-hidden border border-white/20 text-sm'
              role='group'
              aria-label='Chart type'
            >
              {(['line', 'candle'] as ChartType[]).map((t) => (
                <button
                  key={t}
                  type='button'
                  onClick={() => onChartTypeChange(t)}
                  aria-pressed={chartType === t}
                  className={`px-3 py-1 ${chartType === t ? 'bg-primary-500 text-dark-950 font-semibold' : 'bg-white/10 text-gray-300 hover:bg-white/20'}`}
                >
                  {t === 'line' ? 'Line' : 'Candles'}
                </button>
              ))}
            </div>
          )}
          {onShowVolumeChange && (
            <label className='flex items-center gap-1 text-sm text-gray-300'>
              <input
                type='checkbox'
                checked={showVolume}
                onChange={(e) => onShowVolumeChange(e.target.checked)}
              />
              Volume
            </label>
          )}
          <select
            value={source}
            onChange={(e) => onSourceChange(e.target.value as ChartSource)}
            className='px-3 py-1 bg-white/10 border border-white/20 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500'
            aria-label='Price source'
          >
            {SOURCE_OPTIONS.map((s) => (
              <option key={s.value} value={s.value} className='bg-slate-800'>
                {s.label}
              </option>
            ))}
          </select>
          {headerActions}
        </div>
      </div>

      <div className='flex flex-wrap gap-2 mb-4'>
        {TIME_RANGE_OPTIONS.map((r) => (
          <button
            key={r.value}
            type='button'
            onClick={() => onTimeRangeChange(r.value)}
            aria-pressed={timeRange === r.value}
            className={`px-3 py-1 rounded text-sm transition-colors ${
              timeRange === r.value
                ? 'bg-primary-500 text-dark-950 font-semibold'
                : 'bg-white/10 text-gray-400 hover:bg-white/20'
            }`}
          >
            {r.label}
          </button>
        ))}
      </div>

      {drawingEnabled && (
        <DrawingToolbar
          tool={tool}
          onToolChange={(t) => {
            setTool(t)
            setSelectedId(null)
          }}
          color={color}
          onColorChange={setColor}
          lineWidth={lineWidth}
          onLineWidthChange={setLineWidth}
          hasSelection={!!selectedId}
          onDeleteSelected={deleteSelected}
          onClear={() => {
            commit([])
            setSelectedId(null)
          }}
          canUndo={undoRef.current.length > 0}
          onUndo={undo}
          canRedo={redoRef.current.length > 0}
          onRedo={redo}
          drawingCount={drawingList.length}
          onAddAlert={
            onAlertCreateAt && hovered
              ? () => onAlertCreateAt(data[data.length - 1].price)
              : undefined
          }
        />
      )}

      {hovered && (
        <div className='flex flex-wrap gap-x-4 gap-y-1 mb-2 text-xs text-gray-400' aria-live='off'>
          <span>{hovered.date}</span>
          <span className='text-blue-400'>
            {chartType === 'candle'
              ? `O ${formatPrice(hovered.open)} H ${formatPrice(hovered.high)} L ${formatPrice(hovered.low)} C ${formatPrice(hovered.close)}`
              : `Price ${formatPrice(hovered.price)}`}
          </span>
          {indicators.map((ind, i) =>
            ind.visible && typeof hovered[seriesKey(i, 'value')] === 'number' ? (
              <span key={i} style={{ color: ind.color }}>
                {indicatorLabel(ind)} {(hovered[seriesKey(i, 'value')] as number).toFixed(2)}
              </span>
            ) : null,
          )}
        </div>
      )}

      <div
        className='relative'
        style={{
          height:
            320 +
            indicators.filter((i) => i.visible && (i.type === 'rsi' || i.type === 'macd')).length *
              110,
        }}
      >
        <div
          ref={containerRef}
          tabIndex={drawingEnabled ? 0 : -1}
          onKeyDown={onKeyDown}
          className='absolute inset-0 outline-none focus-visible:ring-2 focus-visible:ring-primary-500/50 rounded'
          aria-label={
            drawingEnabled
              ? 'Price chart. Use the toolbar to draw; Delete removes the selected drawing.'
              : 'Price chart'
          }
        />
        {pendingText && (
          <input
            autoFocus
            maxLength={200}
            placeholder='Note text, Enter to place'
            className='absolute z-10 px-2 py-1 text-xs bg-dark-800 border border-primary-500 rounded text-white w-48'
            style={{ left: pendingText.x, top: Math.max(0, pendingText.y - 14) }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') submitText(e.currentTarget.value)
              if (e.key === 'Escape') submitText('')
            }}
            onBlur={(e) => submitText(e.currentTarget.value)}
            aria-label='Text note'
          />
        )}
        {loading && !history ? (
          <div className='absolute inset-0 flex items-center justify-center text-gray-400'>
            Loading chart data...
          </div>
        ) : data.length === 0 ? (
          <div className='absolute inset-0 flex items-center justify-center text-center'>
            <div>
              <div className='text-gray-400 mb-2'>No price data available</div>
              <div className='text-gray-500 text-sm'>Try selecting a different time range</div>
            </div>
          </div>
        ) : null}
      </div>

      {children}

      {showStats && data.length > 0 && (
        <div className='grid grid-cols-2 md:grid-cols-4 gap-4 mt-6 pt-6 border-t border-white/10'>
          <Stat label='Data Points' value={String(data.length)} />
          <Stat
            label='Avg Price'
            value={formatPrice(data.reduce((s, d) => s + d.price, 0) / data.length)}
          />
          <Stat
            label='Period High'
            value={formatPrice(Math.max(...data.map((d) => d.high)))}
            className='text-green-400'
          />
          <Stat
            label='Period Low'
            value={formatPrice(Math.min(...data.map((d) => d.low)))}
            className='text-red-400'
          />
        </div>
      )}
    </div>
  )
}

function Stat({
  label,
  value,
  className = 'text-white',
}: {
  label: string
  value: string
  className?: string
}) {
  return (
    <div className='text-center'>
      <div className='text-sm text-gray-400'>{label}</div>
      <div className={`text-lg font-semibold ${className}`}>{value}</div>
    </div>
  )
}
