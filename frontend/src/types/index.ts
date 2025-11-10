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

export interface SavedChart {
  id: string
  user_id: string
  card_id: string
  name: string
  description?: string
  indicators: ChartIndicator[]
  time_range: '1d' | '7d' | '30d' | '90d' | '1y' | '5y'
  created_at: string
  updated_at: string
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
  range: string
  total: number
  market_data?: MarketData[]
  indicators?: Record<string, IndicatorPoint[]>
}

export interface IndicatorPoint {
  timestamp: string
  value: number
  label?: string
}

export interface Dashboard {
  saved_charts: SavedChart[]
  recently_viewed: Card[]
  user_stats: UserStats
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
  first_name: string
  last_name: string
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
  constructor(message: string, public status: number, public response?: any) {
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
