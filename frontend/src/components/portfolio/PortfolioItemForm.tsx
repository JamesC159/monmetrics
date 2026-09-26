import { useEffect, useState } from 'react'
import { Search as SearchIcon, X } from 'lucide-react'
import Modal from '@/components/Modal'
import { apiClient } from '@/utils/api'
import { useDebounce } from '@/hooks/useDebounce'
import { formatPrice } from '@/utils/formatters'
import {
  CONDITION_LABELS,
  FINISH_LABELS,
  GRADING_COMPANIES,
  ITEM_TYPE_LABELS,
} from '@/utils/portfolio'
import type {
  Card,
  CardFinish,
  ConditionCode,
  GradingCompany,
  PortfolioItemRequest,
  PortfolioItemType,
  PortfolioItemView,
} from '@/types'

interface PortfolioItemFormProps {
  open: boolean
  onClose: () => void
  item?: PortfolioItemView | null
  onSaved: (item: PortfolioItemView) => void
}

const GAME_OPTIONS = ['Pokemon', 'Magic The Gathering', 'Yu-Gi-Oh', 'One Piece', 'Lorcana', 'Other']

interface FormState {
  mode: 'catalog' | 'custom'
  card: Card | null
  itemType: PortfolioItemType
  customName: string
  customGame: string
  customSet: string
  customImageUrl: string
  quantity: string
  condition: ConditionCode
  finish: CardFinish
  language: string
  company: GradingCompany
  grade: string
  certNumber: string
  purchasePrice: string
  purchaseDate: string
  purchaseSource: string
  useManualValue: boolean
  manualValue: string
  notes: string
}

const emptyForm = (): FormState => ({
  mode: 'catalog',
  card: null,
  itemType: 'raw_card',
  customName: '',
  customGame: 'Pokemon',
  customSet: '',
  customImageUrl: '',
  quantity: '1',
  condition: 'NM',
  finish: '',
  language: 'English',
  company: 'PSA',
  grade: '10',
  certNumber: '',
  purchasePrice: '',
  purchaseDate: '',
  purchaseSource: '',
  useManualValue: false,
  manualValue: '',
  notes: '',
})

function fromItem(item: PortfolioItemView): FormState {
  return {
    mode: item.card_id ? 'catalog' : 'custom',
    card: item.card ?? null,
    itemType: item.item_type,
    customName: item.custom_name ?? '',
    customGame: item.custom_game ?? 'Other',
    customSet: item.custom_set ?? '',
    customImageUrl: item.custom_image_url ?? '',
    quantity: String(item.quantity),
    condition: item.condition ?? 'NM',
    finish: item.finish ?? '',
    language: item.language ?? '',
    company: item.grading?.company ?? 'PSA',
    grade: item.grading ? String(item.grading.grade) : '10',
    certNumber: item.grading?.cert_number ?? '',
    purchasePrice: String(item.purchase_price ?? ''),
    purchaseDate: item.purchase_date ? item.purchase_date.slice(0, 10) : '',
    purchaseSource: item.purchase_source ?? '',
    useManualValue: item.manual_value !== undefined && item.manual_value !== null,
    manualValue:
      item.manual_value !== undefined && item.manual_value !== null
        ? String(item.manual_value)
        : '',
    notes: item.notes ?? '',
  }
}

export default function PortfolioItemForm({
  open,
  onClose,
  item,
  onSaved,
}: PortfolioItemFormProps) {
  const [form, setForm] = useState<FormState>(emptyForm)
  const [query, setQuery] = useState('')
  const debounced = useDebounce(query, 300)
  const [results, setResults] = useState<Card[]>([])
  const [searching, setSearching] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setForm(item ? fromItem(item) : emptyForm())
    setQuery('')
    setResults([])
    setError(null)
  }, [open, item])

  useEffect(() => {
    if (form.mode !== 'catalog' || debounced.trim().length < 2) {
      setResults([])
      return
    }
    let cancelled = false
    setSearching(true)
    apiClient
      .searchCards({ q: debounced.trim(), limit: 8 })
      .then((res) => !cancelled && setResults(res.cards ?? []))
      .catch(() => !cancelled && setResults([]))
      .finally(() => !cancelled && setSearching(false))
    return () => {
      cancelled = true
    }
  }, [debounced, form.mode])

  const set = <K extends keyof FormState>(key: K, value: FormState[K]) =>
    setForm((f) => ({ ...f, [key]: value }))

  const selectCard = (card: Card) => {
    setForm((f) => ({
      ...f,
      card,
      itemType:
        card.category === 'sealed' ? 'sealed' : f.itemType === 'sealed' ? 'raw_card' : f.itemType,
      purchasePrice: f.purchasePrice || String(card.current_price),
    }))
    setQuery('')
    setResults([])
  }

  const catalogSealed = form.mode === 'catalog' && form.card?.category === 'sealed'
  const catalogCard = form.mode === 'catalog' && form.card && form.card.category !== 'sealed'

  const buildRequest = (): PortfolioItemRequest | string => {
    const quantity = Number(form.quantity)
    const purchasePrice = form.purchasePrice === '' ? 0 : Number(form.purchasePrice)
    if (form.mode === 'catalog' && !form.card) return 'Select a product from the catalog'
    if (form.mode === 'custom' && !form.customName.trim()) return 'Custom items need a name'
    if (!Number.isInteger(quantity) || quantity < 1 || quantity > 10000)
      return 'Quantity must be 1-10,000'
    if (!Number.isFinite(purchasePrice) || purchasePrice < 0)
      return 'Purchase price must be 0 or more'
    let manual: number | null = null
    if (form.useManualValue) {
      manual = Number(form.manualValue)
      if (form.manualValue === '' || !Number.isFinite(manual) || manual < 0)
        return 'Manual value must be 0 or more'
    }
    const grade = Number(form.grade)
    if (form.itemType === 'graded_card') {
      const step = form.company === 'PSA' ? 1 : 0.5
      if (!Number.isFinite(grade) || grade < 1 || grade > 10 || (grade / step) % 1 !== 0) {
        return form.company === 'PSA'
          ? 'PSA grades are whole numbers 1-10'
          : 'Grades are 1-10 in 0.5 steps'
      }
    }
    return {
      card_id: form.mode === 'catalog' ? form.card!.id : '',
      item_type: form.itemType,
      custom_name: form.mode === 'custom' ? form.customName.trim() : '',
      custom_game: form.mode === 'custom' ? form.customGame : '',
      custom_set: form.mode === 'custom' ? form.customSet.trim() : '',
      custom_image_url: form.mode === 'custom' ? form.customImageUrl.trim() : '',
      quantity,
      condition: form.itemType === 'raw_card' ? form.condition : '',
      finish: form.itemType === 'sealed' ? '' : form.finish,
      language: form.language.trim(),
      grading:
        form.itemType === 'graded_card'
          ? { company: form.company, grade, cert_number: form.certNumber.trim() || undefined }
          : null,
      purchase_price: purchasePrice,
      purchase_date: form.purchaseDate,
      purchase_source: form.purchaseSource.trim(),
      manual_value: manual,
      notes: form.notes,
    }
  }

  const submit = async () => {
    const req = buildRequest()
    if (typeof req === 'string') {
      setError(req)
      return
    }
    setSaving(true)
    setError(null)
    try {
      const saved = item
        ? await apiClient.updatePortfolioItem(item.id, req)
        : await apiClient.createPortfolioItem(req)
      onSaved(saved)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save item')
    } finally {
      setSaving(false)
    }
  }

  const today = new Date().toISOString().slice(0, 10)

  return (
    <Modal
      open={open}
      onClose={onClose}
      size='lg'
      title={item ? 'Edit portfolio item' : 'Add to portfolio'}
      footer={
        <>
          <button type='button' onClick={onClose} className='btn-ghost py-2 px-4'>
            Cancel
          </button>
          <button
            type='button'
            onClick={submit}
            disabled={saving}
            className='btn-primary py-2 px-4'
          >
            {saving ? 'Saving...' : item ? 'Save changes' : 'Add item'}
          </button>
        </>
      }
    >
      <div className='space-y-5'>
        <div className='inline-flex rounded-lg bg-white/5 p-1' role='tablist'>
          {(['catalog', 'custom'] as const).map((m) => (
            <button
              key={m}
              type='button'
              role='tab'
              aria-selected={form.mode === m}
              onClick={() => set('mode', m)}
              className={`px-4 py-1.5 rounded-md text-sm ${form.mode === m ? 'bg-primary-500 text-dark-950 font-semibold' : 'text-gray-300'}`}
            >
              {m === 'catalog' ? 'From catalog' : 'Custom item'}
            </button>
          ))}
        </div>

        {form.mode === 'catalog' ? (
          <div>
            {form.card ? (
              <div className='flex items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10'>
                <img src={form.card.image_url} alt='' className='w-12 h-16 object-cover rounded' />
                <div className='flex-1 min-w-0'>
                  <div className='text-white font-medium truncate'>{form.card.name}</div>
                  <div className='text-xs text-gray-400'>
                    {form.card.game} · {form.card.set} · {formatPrice(form.card.current_price)}
                  </div>
                </div>
                <button
                  type='button'
                  onClick={() => set('card', null)}
                  className='p-1 text-gray-400 hover:text-white'
                  aria-label='Clear selection'
                >
                  <X className='w-4 h-4' />
                </button>
              </div>
            ) : (
              <div className='relative'>
                <label htmlFor='pf-search' className='block text-sm text-gray-300 mb-1'>
                  Search cards and sealed products
                </label>
                <div className='relative'>
                  <SearchIcon className='w-4 h-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2' />
                  <input
                    id='pf-search'
                    className='input pl-9'
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder='e.g. Charizard, Booster Box'
                    autoComplete='off'
                  />
                </div>
                {(results.length > 0 || searching) && (
                  <ul className='absolute z-10 mt-1 w-full max-h-64 overflow-y-auto rounded-lg bg-dark-800 border border-white/10 shadow-xl'>
                    {searching && <li className='px-3 py-2 text-sm text-gray-400'>Searching...</li>}
                    {results.map((c) => (
                      <li key={c.id}>
                        <button
                          type='button'
                          onClick={() => selectCard(c)}
                          className='w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-white/10'
                        >
                          <img src={c.image_url} alt='' className='w-8 h-10 object-cover rounded' />
                          <span className='flex-1 min-w-0'>
                            <span className='block text-sm text-white truncate'>{c.name}</span>
                            <span className='block text-xs text-gray-400 truncate'>
                              {c.game} · {c.set} · {c.category === 'sealed' ? 'Sealed' : 'Card'}
                            </span>
                          </span>
                          <span className='text-xs text-green-400'>
                            {formatPrice(c.current_price)}
                          </span>
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        ) : (
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
            <Field label='Name' id='pf-cname' className='md:col-span-2'>
              <input
                id='pf-cname'
                className='input'
                maxLength={120}
                value={form.customName}
                onChange={(e) => set('customName', e.target.value)}
              />
            </Field>
            <Field label='Game' id='pf-cgame'>
              <select
                id='pf-cgame'
                className='input'
                value={form.customGame}
                onChange={(e) => set('customGame', e.target.value)}
              >
                {GAME_OPTIONS.map((g) => (
                  <option key={g} value={g} className='bg-slate-800'>
                    {g}
                  </option>
                ))}
              </select>
            </Field>
            <Field label='Set (optional)' id='pf-cset'>
              <input
                id='pf-cset'
                className='input'
                maxLength={120}
                value={form.customSet}
                onChange={(e) => set('customSet', e.target.value)}
              />
            </Field>
            <Field label='Image URL (https, optional)' id='pf-cimg' className='md:col-span-2'>
              <input
                id='pf-cimg'
                type='url'
                className='input'
                value={form.customImageUrl}
                onChange={(e) => set('customImageUrl', e.target.value)}
                placeholder='https://...'
              />
            </Field>
          </div>
        )}

        <div className='grid grid-cols-1 md:grid-cols-3 gap-4'>
          <Field label='Item type' id='pf-type'>
            <select
              id='pf-type'
              className='input'
              value={form.itemType}
              disabled={catalogSealed}
              onChange={(e) => set('itemType', e.target.value as PortfolioItemType)}
            >
              {(Object.keys(ITEM_TYPE_LABELS) as PortfolioItemType[])
                .filter((t) => !(catalogCard && t === 'sealed'))
                .map((t) => (
                  <option key={t} value={t} className='bg-slate-800'>
                    {ITEM_TYPE_LABELS[t]}
                  </option>
                ))}
            </select>
          </Field>
          <Field label='Quantity' id='pf-qty'>
            <input
              id='pf-qty'
              type='number'
              min={1}
              max={10000}
              className='input'
              value={form.quantity}
              onChange={(e) => set('quantity', e.target.value)}
            />
          </Field>
          <Field label='Language' id='pf-lang'>
            <input
              id='pf-lang'
              className='input'
              maxLength={30}
              value={form.language}
              onChange={(e) => set('language', e.target.value)}
            />
          </Field>
        </div>

        {form.itemType === 'raw_card' && (
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
            <Field label='Condition' id='pf-cond'>
              <select
                id='pf-cond'
                className='input'
                value={form.condition}
                onChange={(e) => set('condition', e.target.value as ConditionCode)}
              >
                {(Object.keys(CONDITION_LABELS) as ConditionCode[]).map((c) => (
                  <option key={c} value={c} className='bg-slate-800'>
                    {c} - {CONDITION_LABELS[c]}
                  </option>
                ))}
              </select>
            </Field>
            <FinishField value={form.finish} onChange={(v) => set('finish', v)} />
          </div>
        )}

        {form.itemType === 'graded_card' && (
          <div className='grid grid-cols-1 md:grid-cols-4 gap-4'>
            <Field label='Grader' id='pf-grader'>
              <select
                id='pf-grader'
                className='input'
                value={form.company}
                onChange={(e) => set('company', e.target.value as GradingCompany)}
              >
                {GRADING_COMPANIES.map((c) => (
                  <option key={c} value={c} className='bg-slate-800'>
                    {c}
                  </option>
                ))}
              </select>
            </Field>
            <Field label='Grade' id='pf-grade'>
              <input
                id='pf-grade'
                type='number'
                min={1}
                max={10}
                step={form.company === 'PSA' ? 1 : 0.5}
                className='input'
                value={form.grade}
                onChange={(e) => set('grade', e.target.value)}
              />
            </Field>
            <Field label='Cert # (optional)' id='pf-cert'>
              <input
                id='pf-cert'
                className='input'
                maxLength={20}
                value={form.certNumber}
                onChange={(e) => set('certNumber', e.target.value)}
              />
            </Field>
            <FinishField value={form.finish} onChange={(v) => set('finish', v)} />
          </div>
        )}

        <div className='grid grid-cols-1 md:grid-cols-3 gap-4'>
          <Field label='Purchase price (each)' id='pf-price'>
            <input
              id='pf-price'
              type='number'
              min={0}
              step='0.01'
              className='input'
              value={form.purchasePrice}
              onChange={(e) => set('purchasePrice', e.target.value)}
            />
          </Field>
          <Field label='Purchase date' id='pf-date'>
            <input
              id='pf-date'
              type='date'
              max={today}
              className='input'
              value={form.purchaseDate}
              onChange={(e) => set('purchaseDate', e.target.value)}
            />
          </Field>
          <Field label='Purchased from' id='pf-src'>
            <input
              id='pf-src'
              className='input'
              maxLength={60}
              value={form.purchaseSource}
              onChange={(e) => set('purchaseSource', e.target.value)}
            />
          </Field>
        </div>

        <div>
          <label className='flex items-center gap-2 text-sm text-gray-300'>
            <input
              type='checkbox'
              checked={form.useManualValue}
              onChange={(e) => set('useManualValue', e.target.checked)}
            />
            Set my own value (overrides market price)
          </label>
          {form.useManualValue && (
            <input
              type='number'
              min={0}
              step='0.01'
              className='input mt-2'
              aria-label='Manual value (each)'
              placeholder='Value per unit'
              value={form.manualValue}
              onChange={(e) => set('manualValue', e.target.value)}
            />
          )}
        </div>

        <Field label='Notes' id='pf-notes'>
          <textarea
            id='pf-notes'
            className='input min-h-[70px]'
            maxLength={1000}
            value={form.notes}
            onChange={(e) => set('notes', e.target.value)}
          />
        </Field>

        {error && <p className='text-sm text-error-400'>{error}</p>}
      </div>
    </Modal>
  )
}

function Field({
  label,
  id,
  className,
  children,
}: {
  label: string
  id: string
  className?: string
  children: React.ReactNode
}) {
  return (
    <div className={className}>
      <label htmlFor={id} className='block text-sm text-gray-300 mb-1'>
        {label}
      </label>
      {children}
    </div>
  )
}

function FinishField({
  value,
  onChange,
}: {
  value: CardFinish
  onChange: (v: CardFinish) => void
}) {
  return (
    <Field label='Finish' id='pf-finish'>
      <select
        id='pf-finish'
        className='input'
        value={value}
        onChange={(e) => onChange(e.target.value as CardFinish)}
      >
        <option value='' className='bg-slate-800'>
          Not specified
        </option>
        {(Object.keys(FINISH_LABELS) as Exclude<CardFinish, ''>[]).map((f) => (
          <option key={f} value={f} className='bg-slate-800'>
            {FINISH_LABELS[f]}
          </option>
        ))}
      </select>
    </Field>
  )
}
