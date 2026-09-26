// src/types/index.ts

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  user_type: 'free' | 'paid'
  created_at: string
  updated_at: string
  is_active: boolean
  last_login_at?: string
}

export interface Card {
  id: string
  name: string
  set: string
  game: string
  category: 'card' | 'sealed'
  rarity?: string
  number?: string
  image_url: string
  description?: string
  created_at: string
  updated_at: string
  current_price: number
  all_time_high: number
  all_time_low: number
  ath_date: string
  atl_date: string
  search_terms: string[]
  tags?: string[]
  popularity_rank?: number // Based on 6-month popularity metrics
}

export interface PricePoint {
  id: string
  card_id: string
  price: number
  volume?: number
  source: 'ebay' | 'tcgplayer'
  timestamp: string
  created_at: string
}

export type ChartTimeRange = '1d' | '7d' | '30d' | '90d' | '1y' | '5y'
export type ChartSource = 'all' | 'ebay' | 'tcgplayer'
export type ChartType = 'line' | 'candle'

export type DrawingType = 'trendline' | 'hline' | 'rect' | 'text' | 'freehand' | 'fib'

export interface DrawingPoint {
  /** Unix seconds */
  time: number
  price: number
}

export interface ChartDrawing {
  id: string
  type: DrawingType
  points: DrawingPoint[]
  color: string
  text?: string
  line_width: number
}

export type AlertCondition = 'above' | 'below' | 'pct_change' | 'ema_cross'
export type AlertDirection = 'up' | 'down' | 'either'
export type AlertMode = 'once' | 'recurring'

export interface PriceAlertRequest {
  card_id: string
  source: ChartSource
  condition: AlertCondition
  target_price?: number
  pct?: number
  days?: number
  ema_period?: number
  direction?: AlertDirection
  mode: AlertMode
  cooldown_hours?: number
  notify_email: boolean
  note?: string
  active?: boolean
}

export interface PriceAlert extends Omit<PriceAlertRequest, 'active'> {
  id: string
  user_id: string
  active: boolean
  last_price?: number
  last_triggered_at?: string
  trigger_count: number
  created_at: string
  updated_at: string
  card_name?: string
  card_image_url?: string
}

export interface AppNotification {
  id: string
  alert_id?: string
  card_id?: string
  title: string
  message: string
  read: boolean
  created_at: string
}

export interface NotificationList {
  notifications: AppNotification[]
  unread_count: number
}

export interface SavedChart {
  id: string
  user_id: string
  card_id: string
  name: string
  description?: string
  indicators: ChartIndicator[]
  time_range: ChartTimeRange
  source: ChartSource
  chart_type?: ChartType
  show_volume?: boolean
  drawings?: ChartDrawing[]
  created_at: string
  updated_at: string
  card_name?: string
  card_image_url?: string
  card_game?: string
}

export interface SavedChartRequest {
  card_id: string
  name: string
  description: string
  indicators: ChartIndicator[]
  time_range: ChartTimeRange
  source: ChartSource
  chart_type: ChartType
  show_volume: boolean
  drawings: ChartDrawing[]
}

export interface ChartIndicator {
  type: string
  parameters: Record<string, any>
  color?: string
  visible: boolean
}

export interface MarketData {
  card_id: string
  date: string
  open_price: number
  close_price: number
  high_price: number
  low_price: number
  volume: number
  weighted_avg_price: number
}

export interface Listing {
  id: string
  card_id: string
  title: string
  price: number
  quantity: number
  condition: string
  seller: string
  source: 'ebay' | 'tcgplayer'
  image_url?: string
  created_at: string
  updated_at: string
}

export interface SearchResult {
  cards: Card[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PriceHistory {
  prices: PricePoint[]
  listings: Listing[]
  range?: string
  total?: number
  card_id?: string
  time_range?: string
  visible_from?: string
  market_data?: MarketData[]
  indicators?: Record<string, IndicatorPoint[]>
}

export interface IndicatorPoint {
  timestamp: string
  value: number
  label?: string
}

export interface Dashboard {
  user?: User
  saved_charts: SavedChart[]
  recently_viewed: Card[]
  user_stats: UserStats
  favorites_count: number
  portfolio_summary?: PortfolioSummary
}

// Favorites

export interface FavoriteView {
  id: string
  card_id: string
  created_at: string
  card?: Card
  change_7d: number
  change_7d_value: number
  change_30d: number
  change_30d_value: number
}

// Portfolio

export type PortfolioItemType = 'raw_card' | 'graded_card' | 'sealed'
export type ConditionCode = 'NM' | 'LP' | 'MP' | 'HP' | 'DMG'
export type GradingCompany = 'PSA' | 'BGS' | 'CGC' | 'SGC'
export type CardFinish = '' | 'normal' | 'holo' | 'reverse_holo' | 'foil' | '1st_edition'
export type ValueSource = 'market' | 'graded_estimate' | 'manual' | 'none'

export interface Grading {
  company: GradingCompany
  grade: number
  cert_number?: string
}

export interface PortfolioItem {
  id: string
  user_id: string
  card_id?: string
  item_type: PortfolioItemType
  custom_name?: string
  custom_game?: string
  custom_set?: string
  custom_image_url?: string
  quantity: number
  condition?: ConditionCode
  finish?: CardFinish
  language?: string
  grading?: Grading
  purchase_price: number
  purchase_date?: string
  purchase_source?: string
  manual_value?: number
  notes?: string
  created_at: string
  updated_at: string
}

export interface PortfolioItemView extends PortfolioItem {
  card?: Card
  card_unavailable?: boolean
  display_name: string
  display_game: string
  display_set: string
  display_image_url: string
  unit_value: number
  total_value: number
  cost_basis: number
  gain_loss: number
  gain_loss_pct: number
  value_source: ValueSource
  multiplier: number
}

export interface PortfolioItemRequest {
  card_id: string
  item_type: PortfolioItemType
  custom_name: string
  custom_game: string
  custom_set: string
  custom_image_url: string
  quantity: number
  condition: ConditionCode | ''
  finish: CardFinish
  language: string
  grading: Grading | null
  purchase_price: number
  purchase_date: string
  purchase_source: string
  manual_value: number | null
  notes: string
}

export interface AllocationSlice {
  key: string
  value: number
  count: number
}

export interface PortfolioMover {
  id: string
  display_name: string
  gain_loss: number
  gain_loss_pct: number
}

export interface PortfolioSummary {
  total_value: number
  cost_basis: number
  gain_loss: number
  gain_loss_pct: number
  item_count: number
  unit_count: number
  unvalued_count: number
  by_game: AllocationSlice[]
  by_item_type: AllocationSlice[]
  top_gainers: PortfolioMover[]
  top_losers: PortfolioMover[]
}

export interface PortfolioResponse {
  items: PortfolioItemView[]
  summary: PortfolioSummary
}

export interface PortfolioQuery {
  game?: string
  item_type?: PortfolioItemType | ''
  sort?: 'date' | 'value' | 'gain' | 'name'
}

// Marketplace

export type MarketplaceProvider = 'ebay' | 'tcgplayer'

export interface LinkedAccount {
  id: string
  provider: MarketplaceProvider
  external_user_id: string
  external_username: string
  scopes: string[]
  expires_at: string
  status: 'active' | 'expired' | 'revoked'
  linked_at: string
  updated_at: string
}

export interface LinkedAccountStatus {
  provider: MarketplaceProvider
  display_name: string
  linked: boolean
  account?: LinkedAccount
  mock: boolean
}

export interface SoldListing {
  title: string
  price: number
  shipping: number
  condition: string
  sold_at: string
  url?: string
  source: string
}

export interface CompsStats {
  count: number
  average: number
  median: number
  min: number
  max: number
  suggested_price: number
}

export interface CompsResponse {
  query: string
  sold: SoldListing[]
  stats: CompsStats
}

export interface ItemSpecific {
  name: string
  value: string
}

export interface ListingDraft {
  portfolio_item_id: string
  provider: MarketplaceProvider
  title: string
  description: string
  price: number
  quantity: number
  max_quantity?: number
  condition: string
  category_id: string
  format: 'fixed_price' | 'auction'
  duration: string
  shipping_price: number
  image_urls: string[]
  item_specifics: ItemSpecific[]
}

export interface PrefillResponse {
  draft: ListingDraft
  comps: CompsResponse
}

export interface MarketplaceListing {
  id: string
  portfolio_item_id: string
  card_id?: string
  provider: MarketplaceProvider
  external_listing_id?: string
  status: 'draft' | 'published' | 'failed' | 'ended'
  title: string
  description: string
  price: number
  quantity: number
  condition: string
  category_id?: string
  format: string
  duration?: string
  shipping_price: number
  image_urls: string[]
  item_specifics: ItemSpecific[]
  listing_url?: string
  error_message?: string
  created_at: string
  published_at?: string
  updated_at: string
}

export interface UserStats {
  charts_created: number
  indicators_used: number
  max_indicators: number
}

// Request/Response Types

export interface RegisterRequest {
  email: string
  password: string
  firstName: string
  lastName: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface SearchParams {
  q?: string
  game?: string
  category?: string
  page?: number
  limit?: number
}

export interface ErrorResponse {
  error: string
  success: false
  details?: Record<string, any>
}

export interface HealthResponse {
  status: string
  timestamp: string
  version: string
}

// API Error class
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public response?: any,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// Chart data types for Recharts
export interface ChartDataPoint {
  date: string
  fullDate: string
  price: number
  high: number
  low: number
  volume: number
}

// Available technical indicators
export type IndicatorType =
  | 'sma' // Simple Moving Average
  | 'ema' // Exponential Moving Average
  | 'bollinger' // Bollinger Bands
  | 'rsi' // Relative Strength Index
  | 'macd' // MACD
  | 'stochastic' // Stochastic Oscillator
  | 'williams_r' // Williams %R
  | 'cci' // Commodity Channel Index
  | 'atr' // Average True Range
  | 'volume_sma' // Volume Simple Moving Average

// Indicator configurations
export interface IndicatorConfig {
  type: IndicatorType
  name: string
  description: string
  parameters: {
    [key: string]: {
      label: string
      type: 'number' | 'select'
      default: any
      min?: number
      max?: number
      options?: { value: any; label: string }[]
    }
  }
  isPremium: boolean
}

// Available time ranges
export interface TimeRange {
  value: string
  label: string
  days: number
}

export const TIME_RANGES: TimeRange[] = [
  { value: '1d', label: '1 Day', days: 1 },
  { value: '7d', label: '7 Days', days: 7 },
  { value: '30d', label: '30 Days', days: 30 },
  { value: '90d', label: '90 Days', days: 90 },
  { value: '1y', label: '1 Year', days: 365 },
  { value: '5y', label: '5 Years', days: 1825 },
]

// Available games
export const GAMES = [
  'Pokemon',
  'Magic The Gathering',
  'Yu-Gi-Oh',
  'Dragon Ball Super',
  'One Piece',
  'Digimon',
  'Flesh and Blood',
] as const

export type GameType = (typeof GAMES)[number]

// Card conditions
export const CONDITIONS = [
  'Mint',
  'Near Mint',
  'Lightly Played',
  'Moderately Played',
  'Heavily Played',
  'Damaged',
] as const

export type ConditionType = (typeof CONDITIONS)[number]

// Marketplace sources
export const SOURCES = ['ebay', 'tcgplayer', 'cardmarket', 'tcgplayer_direct'] as const

export type SourceType = (typeof SOURCES)[number]

// Sort options for search
export interface SortOption {
  value: string
  label: string
  field: string
  direction: 'asc' | 'desc'
}

export const SORT_OPTIONS: SortOption[] = [
  { value: 'relevance', label: 'Relevance', field: 'score', direction: 'desc' },
  { value: 'price_high', label: 'Price: High to Low', field: 'current_price', direction: 'desc' },
  { value: 'price_low', label: 'Price: Low to High', field: 'current_price', direction: 'asc' },
  { value: 'name_asc', label: 'Name: A to Z', field: 'name', direction: 'asc' },
  { value: 'name_desc', label: 'Name: Z to A', field: 'name', direction: 'desc' },
  { value: 'newest', label: 'Newest First', field: 'created_at', direction: 'desc' },
  { value: 'oldest', label: 'Oldest First', field: 'created_at', direction: 'asc' },
]

// Filter options
export interface FilterOptions {
  games: string[]
  categories: string[]
  priceRanges: PriceRange[]
  rarities: string[]
}

export interface PriceRange {
  label: string
  min: number
  max?: number
}

export const PRICE_RANGES: PriceRange[] = [
  { label: 'Under $10', min: 0, max: 10 },
  { label: '$10 - $25', min: 10, max: 25 },
  { label: '$25 - $50', min: 25, max: 50 },
  { label: '$50 - $100', min: 50, max: 100 },
  { label: '$100 - $250', min: 100, max: 250 },
  { label: '$250 - $500', min: 250, max: 500 },
  { label: '$500 - $1,000', min: 500, max: 1000 },
  { label: '$1,000+', min: 1000 },
]

// Toast notification types
export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

// Form validation types
export interface ValidationError {
  field: string
  message: string
}

export interface FormState {
  isSubmitting: boolean
  errors: ValidationError[]
  touched: Record<string, boolean>
}

// Pagination interface
export interface Pagination {
  page: number
  per_page: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

// Generic API response wrapper
export interface ApiResponse<T> {
  data: T
  success: boolean
  message?: string
  errors?: ValidationError[]
  pagination?: Pagination
}

// WebSocket message types for real-time updates
export interface WebSocketMessage {
  type: 'price_update' | 'new_listing' | 'chart_update'
  data: any
  timestamp: string
}

// Local storage keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'monmetrics_auth_token',
  USER_PREFERENCES: 'monmetrics_user_preferences',
  RECENT_SEARCHES: 'monmetrics_recent_searches',
  CHART_SETTINGS: 'monmetrics_chart_settings',
} as const

// User preferences
export interface UserPreferences {
  theme: 'light' | 'dark' | 'system'
  currency: 'USD' | 'EUR' | 'GBP' | 'JPY'
  timeZone: string
  defaultTimeRange: string
  chartType: 'line' | 'candlestick' | 'area'
  notifications: {
    priceAlerts: boolean
    newListings: boolean
    marketUpdates: boolean
  }
}

// Featured Content for Carousel
export interface FeaturedContent {
  id: string
  type: 'product' | 'market_mover' | 'news' | 'pickup' | 'sponsored'
  title: string
  description?: string
  image_url: string
  card_id?: string
  link?: string
  priority: number // Higher = shown first
  active: boolean
  created_at: string
  expires_at?: string
  // Market mover specific
  price_change?: number // Percentage
  price_change_value?: number // Dollar amount
}

// Game-grouped cards for search page
export interface GameCardGroup {
  game: string
  category: 'card' | 'sealed'
  cards: Card[]
  total_count: number
}
