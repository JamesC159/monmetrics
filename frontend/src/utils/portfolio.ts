import type {
  CardFinish,
  ConditionCode,
  GradingCompany,
  PortfolioItemType,
  PortfolioItemView,
  ValueSource,
} from '@/types'

export const ITEM_TYPE_LABELS: Record<PortfolioItemType, string> = {
  raw_card: 'Raw Card',
  graded_card: 'Graded Card',
  sealed: 'Sealed',
}

export const CONDITION_LABELS: Record<ConditionCode, string> = {
  NM: 'Near Mint',
  LP: 'Lightly Played',
  MP: 'Moderately Played',
  HP: 'Heavily Played',
  DMG: 'Damaged',
}

export const FINISH_LABELS: Record<Exclude<CardFinish, ''>, string> = {
  normal: 'Normal',
  holo: 'Holo',
  reverse_holo: 'Reverse Holo',
  foil: 'Foil',
  '1st_edition': '1st Edition',
}

export const VALUE_SOURCE_LABELS: Record<ValueSource, string> = {
  market: 'Market price',
  graded_estimate: 'Graded estimate',
  manual: 'Manual value',
  none: 'No value',
}

export const GRADING_COMPANIES: GradingCompany[] = ['PSA', 'BGS', 'CGC', 'SGC']

export const PLACEHOLDER_IMAGE = 'https://via.placeholder.com/300x400/374151/9CA3AF?text=No+Image'

export function gradeLabel(item: Pick<PortfolioItemView, 'grading'>): string {
  if (!item.grading) return ''
  return `${item.grading.company} ${item.grading.grade}`
}

export function itemSubtitle(item: PortfolioItemView): string {
  if (item.item_type === 'graded_card') return gradeLabel(item)
  if (item.item_type === 'raw_card' && item.condition) return CONDITION_LABELS[item.condition]
  return ITEM_TYPE_LABELS[item.item_type]
}

export function gainClass(value: number): string {
  if (value > 0) return 'price-up'
  if (value < 0) return 'price-down'
  return 'price-neutral'
}

export function signed(value: string, raw: number): string {
  return raw > 0 ? `+${value}` : value
}
