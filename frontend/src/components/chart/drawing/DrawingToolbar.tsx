import {
  Bell,
  Eraser,
  Minus,
  MousePointer2,
  PenLine,
  Pencil,
  Redo2,
  Square,
  Trash2,
  TrendingUp,
  Type,
  Undo2,
} from 'lucide-react'
import type { ComponentType } from 'react'
import type { DrawingType } from '@/types'
import { DRAWING_TOOLS } from './geometry'

export type Tool = 'cursor' | DrawingType

const ICONS: Record<DrawingType, ComponentType<{ className?: string }>> = {
  trendline: TrendingUp,
  hline: Minus,
  rect: Square,
  text: Type,
  freehand: Pencil,
  fib: PenLine,
}

export const DRAWING_COLORS = ['#f59e0b', '#22c55e', '#3b82f6', '#ef4444', '#a855f7', '#f9fafb']

interface DrawingToolbarProps {
  tool: Tool
  onToolChange: (tool: Tool) => void
  color: string
  onColorChange: (color: string) => void
  lineWidth: number
  onLineWidthChange: (width: number) => void
  hasSelection: boolean
  onDeleteSelected: () => void
  onClear: () => void
  canUndo: boolean
  onUndo: () => void
  canRedo: boolean
  onRedo: () => void
  drawingCount: number
  onAddAlert?: () => void
}

const btn =
  'p-1.5 rounded text-gray-300 hover:bg-white/15 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed'

export default function DrawingToolbar(props: DrawingToolbarProps) {
  const { tool, onToolChange, color, onColorChange, lineWidth, onLineWidthChange } = props
  return (
    <div
      className='flex flex-wrap items-center gap-1 mb-3 p-1.5 rounded-lg bg-white/5 border border-white/10'
      role='toolbar'
      aria-label='Drawing tools'
    >
      <button
        type='button'
        onClick={() => onToolChange('cursor')}
        aria-pressed={tool === 'cursor'}
        title='Select / move (Esc)'
        className={`${btn} ${tool === 'cursor' ? 'bg-primary-500/30 text-white' : ''}`}
      >
        <MousePointer2 className='w-4 h-4' />
      </button>
      {DRAWING_TOOLS.map(({ type, label }) => {
        const Icon = ICONS[type]
        return (
          <button
            key={type}
            type='button'
            onClick={() => onToolChange(tool === type ? 'cursor' : type)}
            aria-pressed={tool === type}
            title={label}
            aria-label={label}
            className={`${btn} ${tool === type ? 'bg-primary-500/30 text-white' : ''}`}
          >
            <Icon className='w-4 h-4' />
          </button>
        )
      })}

      <span className='w-px h-5 bg-white/15 mx-1' />

      {DRAWING_COLORS.map((c) => (
        <button
          key={c}
          type='button'
          onClick={() => onColorChange(c)}
          aria-label={`Color ${c}`}
          aria-pressed={color === c}
          className={`w-5 h-5 rounded-full border ${color === c ? 'border-white ring-2 ring-white/40' : 'border-white/20'}`}
          style={{ backgroundColor: c }}
        />
      ))}
      <select
        value={lineWidth}
        onChange={(e) => onLineWidthChange(Number(e.target.value))}
        className='ml-1 px-1 py-0.5 bg-dark-800/60 border border-white/10 rounded text-white text-xs'
        aria-label='Line width'
      >
        {[1, 2, 3, 4].map((w) => (
          <option key={w} value={w} className='bg-slate-800'>
            {w}px
          </option>
        ))}
      </select>

      <span className='w-px h-5 bg-white/15 mx-1' />

      <button
        type='button'
        onClick={props.onUndo}
        disabled={!props.canUndo}
        title='Undo (Ctrl+Z)'
        aria-label='Undo'
        className={btn}
      >
        <Undo2 className='w-4 h-4' />
      </button>
      <button
        type='button'
        onClick={props.onRedo}
        disabled={!props.canRedo}
        title='Redo (Ctrl+Y)'
        aria-label='Redo'
        className={btn}
      >
        <Redo2 className='w-4 h-4' />
      </button>
      <button
        type='button'
        onClick={props.onDeleteSelected}
        disabled={!props.hasSelection}
        title='Delete selected (Del)'
        aria-label='Delete selected drawing'
        className={btn}
      >
        <Eraser className='w-4 h-4' />
      </button>
      <button
        type='button'
        onClick={props.onClear}
        disabled={props.drawingCount === 0}
        title='Clear all drawings'
        aria-label='Clear all drawings'
        className={btn}
      >
        <Trash2 className='w-4 h-4' />
      </button>

      {props.onAddAlert && (
        <button
          type='button'
          onClick={props.onAddAlert}
          title='Add price alert (or right-click the chart)'
          className='ml-auto flex items-center gap-1 px-2 py-1 text-xs rounded bg-primary-500/20 text-primary-300 hover:bg-primary-500/30'
        >
          <Bell className='w-3.5 h-3.5' /> Alert
        </button>
      )}
    </div>
  )
}
