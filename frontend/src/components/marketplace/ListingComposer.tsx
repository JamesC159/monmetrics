import { useEffect, useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import Modal from '@/components/Modal'
import { apiClient, type ApiError } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { formatDate, formatPrice } from '@/utils/formatters'
import type {
  CompsResponse,
  ListingDraft,
  MarketplaceListing,
  MarketplaceProvider,
  PortfolioItemView,
} from '@/types'

const PROVIDER_NAMES: Record<MarketplaceProvider, string> = { ebay: 'eBay', tcgplayer: 'TCGPlayer' }
const EBAY_DURATIONS = [
  { value: 'GTC', label: 'Good til cancelled' },
  { value: 'DAYS_3', label: '3 days' },
  { value: 'DAYS_5', label: '5 days' },
  { value: 'DAYS_7', label: '7 days' },
  { value: 'DAYS_10', label: '10 days' },
]

interface ListingComposerProps {
  open: boolean
  onClose: () => void
  item: PortfolioItemView
  provider: MarketplaceProvider
  onCreated: (listing: MarketplaceListing) => void
}

export default function ListingComposer({
  open,
  onClose,
  item,
  provider,
  onCreated,
}: ListingComposerProps) {
  const { addToast } = useToast()
  const [draft, setDraft] = useState<ListingDraft | null>(null)
  const [comps, setComps] = useState<CompsResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const titleMax = provider === 'ebay' ? 80 : 255
  const name = PROVIDER_NAMES[provider]

  useEffect(() => {
    if (!open) return
    let cancelled = false
    setLoading(true)
    setError(null)
    setDraft(null)
    apiClient
      .prefillListing(item.id, provider)
      .then((res) => {
        if (cancelled) return
        setDraft(res.draft)
        setComps(res.comps)
      })
      .catch(
        (err) =>
          !cancelled && setError(err instanceof Error ? err.message : 'Could not prepare listing'),
      )
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [open, item.id, provider])

  const update = <K extends keyof ListingDraft>(key: K, value: ListingDraft[K]) =>
    setDraft((d) => (d ? { ...d, [key]: value } : d))

  const validate = (d: ListingDraft): string | null => {
    if (!d.title.trim() || d.title.length > titleMax)
      return `Title must be 1-${titleMax} characters`
    if (!(d.price >= 0.01)) return 'Price must be at least $0.01'
    if (!Number.isInteger(d.quantity) || d.quantity < 1 || d.quantity > item.quantity)
      return `Quantity must be 1-${item.quantity}`
    if (d.shipping_price < 0) return 'Shipping cannot be negative'
    if (d.format === 'auction' && d.duration === 'GTC') return 'Auctions need a fixed duration'
    if (d.item_specifics.some((s) => !s.name.trim())) return 'Item specifics need a name'
    return null
  }

  const submit = async (publish: boolean) => {
    if (!draft) return
    const problem = validate(draft)
    if (problem) {
      setError(problem)
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const listing = await apiClient.createListing(draft, publish)
      addToast(
        publish
          ? `Listed on ${name}${listing.external_listing_id ? ` (${listing.external_listing_id})` : ''}`
          : 'Draft saved',
        'success',
      )
      onCreated(listing)
      onClose()
    } catch (err) {
      const apiErr = err as ApiError
      if (apiErr.details?.code === 'account_not_linked') {
        setError(`${apiErr.message}. Link it from Dashboard → Connected Accounts.`)
      } else {
        setError(apiErr.message || 'Could not create listing')
      }
      if (apiErr.details?.listing) onCreated(apiErr.details.listing as MarketplaceListing)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      size='xl'
      title={`Sell on ${name}`}
      footer={
        <>
          <button type='button' onClick={onClose} className='btn-ghost py-2 px-4'>
            Cancel
          </button>
          <button
            type='button'
            disabled={!draft || submitting}
            onClick={() => submit(false)}
            className='btn-ghost py-2 px-4'
          >
            Save draft
          </button>
          <button
            type='button'
            disabled={!draft || submitting}
            onClick={() => submit(true)}
            className='btn-primary py-2 px-4'
          >
            {submitting ? 'Publishing...' : `Publish to ${name}`}
          </button>
        </>
      }
    >
      {loading && (
        <p className='text-gray-400 py-8 text-center'>
          Searching recent sales and preparing your listing...
        </p>
      )}
      {draft && (
        <div className='grid grid-cols-1 lg:grid-cols-5 gap-6'>
          <div className='lg:col-span-3 space-y-4'>
            <div>
              <div className='flex justify-between text-sm text-gray-300 mb-1'>
                <label htmlFor='lc-title'>Title</label>
                <span
                  className={draft.title.length > titleMax ? 'text-error-400' : 'text-gray-500'}
                >
                  {draft.title.length}/{titleMax}
                </span>
              </div>
              <input
                id='lc-title'
                className='input'
                value={draft.title}
                onChange={(e) => update('title', e.target.value)}
              />
            </div>

            <div className='grid grid-cols-2 md:grid-cols-3 gap-3'>
              <NumberField
                id='lc-price'
                label='Price'
                value={draft.price}
                step='0.01'
                onChange={(v) => update('price', v)}
                hint={
                  comps && comps.stats.suggested_price > 0
                    ? `Suggested ${formatPrice(comps.stats.suggested_price)}`
                    : undefined
                }
              />
              <NumberField
                id='lc-qty'
                label={`Quantity (max ${item.quantity})`}
                value={draft.quantity}
                step='1'
                onChange={(v) => update('quantity', v)}
              />
              {provider === 'ebay' && (
                <NumberField
                  id='lc-ship'
                  label='Shipping'
                  value={draft.shipping_price}
                  step='0.01'
                  onChange={(v) => update('shipping_price', v)}
                />
              )}
            </div>

            <div className='grid grid-cols-1 md:grid-cols-3 gap-3'>
              <div>
                <label htmlFor='lc-cond' className='block text-sm text-gray-300 mb-1'>
                  Condition
                </label>
                <input
                  id='lc-cond'
                  className='input'
                  value={draft.condition}
                  onChange={(e) => update('condition', e.target.value)}
                />
              </div>
              {provider === 'ebay' && (
                <>
                  <div>
                    <label htmlFor='lc-format' className='block text-sm text-gray-300 mb-1'>
                      Format
                    </label>
                    <select
                      id='lc-format'
                      className='input'
                      value={draft.format}
                      onChange={(e) => {
                        const format = e.target.value as ListingDraft['format']
                        setDraft((d) =>
                          d
                            ? {
                                ...d,
                                format,
                                duration:
                                  format === 'auction' && d.duration === 'GTC'
                                    ? 'DAYS_7'
                                    : d.duration,
                              }
                            : d,
                        )
                      }}
                    >
                      <option value='fixed_price' className='bg-slate-800'>
                        Buy It Now
                      </option>
                      <option value='auction' className='bg-slate-800'>
                        Auction
                      </option>
                    </select>
                  </div>
                  <div>
                    <label htmlFor='lc-dur' className='block text-sm text-gray-300 mb-1'>
                      Duration
                    </label>
                    <select
                      id='lc-dur'
                      className='input'
                      value={draft.duration}
                      onChange={(e) => update('duration', e.target.value)}
                    >
                      {EBAY_DURATIONS.filter(
                        (d) => !(draft.format === 'auction' && d.value === 'GTC'),
                      ).map((d) => (
                        <option key={d.value} value={d.value} className='bg-slate-800'>
                          {d.label}
                        </option>
                      ))}
                    </select>
                  </div>
                </>
              )}
            </div>

            <div>
              <label htmlFor='lc-desc' className='block text-sm text-gray-300 mb-1'>
                Description
              </label>
              <textarea
                id='lc-desc'
                className='input min-h-[120px]'
                maxLength={5000}
                value={draft.description}
                onChange={(e) => update('description', e.target.value)}
              />
            </div>

            <div>
              <div className='flex items-center justify-between mb-2'>
                <span className='text-sm text-gray-300'>Item specifics</span>
                <button
                  type='button'
                  disabled={draft.item_specifics.length >= 30}
                  onClick={() =>
                    update('item_specifics', [...draft.item_specifics, { name: '', value: '' }])
                  }
                  className='text-xs flex items-center text-primary-300 hover:text-primary-200 disabled:opacity-40'
                >
                  <Plus className='w-3 h-3 mr-1' /> Add
                </button>
              </div>
              <div className='space-y-2'>
                {draft.item_specifics.map((s, i) => (
                  <div key={i} className='flex gap-2'>
                    <input
                      aria-label='Specific name'
                      className='input py-2 flex-1'
                      maxLength={65}
                      value={s.name}
                      onChange={(e) =>
                        update(
                          'item_specifics',
                          draft.item_specifics.map((x, j) =>
                            j === i ? { ...x, name: e.target.value } : x,
                          ),
                        )
                      }
                    />
                    <input
                      aria-label='Specific value'
                      className='input py-2 flex-1'
                      maxLength={100}
                      value={s.value}
                      onChange={(e) =>
                        update(
                          'item_specifics',
                          draft.item_specifics.map((x, j) =>
                            j === i ? { ...x, value: e.target.value } : x,
                          ),
                        )
                      }
                    />
                    <button
                      type='button'
                      onClick={() =>
                        update(
                          'item_specifics',
                          draft.item_specifics.filter((_, j) => j !== i),
                        )
                      }
                      className='p-2 text-gray-400 hover:text-error-400'
                      aria-label='Remove specific'
                    >
                      <Trash2 className='w-4 h-4' />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className='lg:col-span-2 space-y-4'>
            {draft.image_urls.length > 0 && (
              <div className='flex gap-2'>
                {draft.image_urls.map((src) => (
                  <img key={src} src={src} alt='' className='w-20 h-28 object-cover rounded' />
                ))}
              </div>
            )}
            <div className='p-4 rounded-xl bg-white/5 border border-white/10'>
              <h4 className='text-sm font-semibold text-white mb-2'>
                Recent sold ({comps?.stats.count ?? 0})
              </h4>
              {comps && comps.stats.count > 0 ? (
                <>
                  <div className='grid grid-cols-3 gap-2 text-center text-xs mb-3'>
                    <MiniStat label='Median' value={formatPrice(comps.stats.median)} />
                    <MiniStat label='Low' value={formatPrice(comps.stats.min)} />
                    <MiniStat label='High' value={formatPrice(comps.stats.max)} />
                  </div>
                  <ul className='max-h-64 overflow-y-auto divide-y divide-white/5 text-xs'>
                    {comps.sold.map((s, i) => (
                      <li key={i} className='py-2 flex justify-between gap-2'>
                        <span className='text-gray-300 min-w-0'>
                          <span className='block truncate'>{s.title}</span>
                          <span className='text-gray-500'>{formatDate(s.sold_at)}</span>
                        </span>
                        <span className='text-right whitespace-nowrap'>
                          <span className='block text-green-400'>{formatPrice(s.price)}</span>
                          <span className='text-gray-500'>
                            {s.shipping > 0 ? `+${formatPrice(s.shipping)} ship` : 'Free ship'}
                          </span>
                        </span>
                      </li>
                    ))}
                  </ul>
                </>
              ) : (
                <p className='text-xs text-gray-500'>
                  No recent sales found. Price is based on your item's current value.
                </p>
              )}
            </div>
          </div>
        </div>
      )}
      {error && <p className='text-sm text-error-400 mt-4'>{error}</p>}
    </Modal>
  )
}

function NumberField({
  id,
  label,
  value,
  step,
  onChange,
  hint,
}: {
  id: string
  label: string
  value: number
  step: string
  onChange: (v: number) => void
  hint?: string
}) {
  return (
    <div>
      <label htmlFor={id} className='block text-sm text-gray-300 mb-1'>
        {label}
      </label>
      <input
        id={id}
        type='number'
        min={0}
        step={step}
        className='input'
        value={Number.isNaN(value) ? '' : value}
        onChange={(e) => onChange(e.target.valueAsNumber)}
      />
      {hint && <p className='text-xs text-primary-300 mt-1'>{hint}</p>}
    </div>
  )
}

function MiniStat({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded bg-white/5 py-1'>
      <div className='text-gray-500'>{label}</div>
      <div className='text-white font-semibold'>{value}</div>
    </div>
  )
}
