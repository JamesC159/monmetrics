import { useEffect, useState } from 'react'
import { Search as SearchIcon, X } from 'lucide-react'

interface SearchBarProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  /** Debounce before propagating keystrokes, in ms. */
  delay?: number
}

export default function SearchBar({
  value,
  onChange,
  placeholder = 'Search cards and sealed products...',
  delay = 350,
}: SearchBarProps) {
  const [draft, setDraft] = useState(value)

  // Keep in sync when the URL changes from outside (back button, filter links).
  useEffect(() => setDraft(value), [value])

  useEffect(() => {
    if (draft === value) return
    const timer = setTimeout(() => onChange(draft), delay)
    return () => clearTimeout(timer)
  }, [draft, value, delay, onChange])

  return (
    <div className='relative'>
      <SearchIcon className='absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500' />
      <input
        type='search'
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        placeholder={placeholder}
        aria-label='Search catalog'
        className='w-full pl-11 pr-10 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500'
      />
      {draft && (
        <button
          type='button'
          onClick={() => setDraft('')}
          aria-label='Clear search'
          className='absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white'
        >
          <X className='w-4 h-4' />
        </button>
      )}
    </div>
  )
}
