import { useEffect, useRef, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/utils/cn'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  footer?: ReactNode
  size?: 'md' | 'lg' | 'xl'
}

const sizes = { md: 'max-w-lg', lg: 'max-w-2xl', xl: 'max-w-4xl' }

export default function Modal({ open, onClose, title, children, footer, size = 'md' }: ModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose

  useEffect(() => {
    if (!open) return
    const previouslyFocused = document.activeElement as HTMLElement | null
    const dialog = dialogRef.current
    const focusable = () =>
      Array.from(
        dialog?.querySelectorAll<HTMLElement>(
          'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ) ?? [],
      )
    focusable()[0]?.focus()

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCloseRef.current()
        return
      }
      if (e.key !== 'Tab') return
      const items = focusable()
      if (items.length === 0) return
      const first = items[0]
      const last = items[items.length - 1]
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault()
        last.focus()
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault()
        first.focus()
      }
    }
    document.addEventListener('keydown', onKeyDown)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.body.style.overflow = prevOverflow
      previouslyFocused?.focus()
    }
  }, [open])

  if (!open) return null

  return (
    <div className='fixed inset-0 z-50 flex items-center justify-center p-4'>
      <div
        className='absolute inset-0 bg-black/70 backdrop-blur-sm'
        onClick={onClose}
        aria-hidden='true'
      />
      <div
        ref={dialogRef}
        role='dialog'
        aria-modal='true'
        aria-labelledby='modal-title'
        className={cn(
          'relative w-full glass-effect-strong bg-dark-900/95 rounded-2xl shadow-2xl flex flex-col max-h-[90vh]',
          sizes[size],
        )}
      >
        <div className='flex items-center justify-between px-6 py-4 border-b border-white/10'>
          <h2 id='modal-title' className='text-xl font-semibold text-white'>
            {title}
          </h2>
          <button
            type='button'
            onClick={onClose}
            className='p-1 rounded-lg text-gray-400 hover:text-white hover:bg-white/10'
            aria-label='Close'
          >
            <X className='w-5 h-5' />
          </button>
        </div>
        <div className='px-6 py-4 overflow-y-auto'>{children}</div>
        {footer && (
          <div className='px-6 py-4 border-t border-white/10 flex justify-end gap-3'>{footer}</div>
        )}
      </div>
    </div>
  )
}
