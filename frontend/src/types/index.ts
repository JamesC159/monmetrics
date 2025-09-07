export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  userType: 'free' | 'paid'
  createdAt: string
  updatedAt: string
  isActive: boolean
  lastLoginAt?: string
}

export interface Card {
  id: string
  name: string
  set: string
  game: string
  category: 'card' | 'sealed'
  rarity?: string
  number?: string
  imageUrl: string
  description?: string
  currentPrice: number
  allTimeHigh: number
  allTimeLow: number
  athDate: string
  atlDate: string
  searchTerms: string[]
  tags?: string[]
  createdAt: string
  updatedAt: string
}

export interface PricePoint {
  id: string
  cardId: string
  price: number
  volume?: number
  source: 'ebay' | 'tcgplayer'
  timestamp: string
  createdAt: string
}

export interface MarketData {
  cardId: string
  date: string
  openPrice: number
  closePrice: number
  highPrice: number
  lowPrice: number
  volume: number
  weightedAvgPrice: number
}

export interface ChartIndicator {
  type: string
  parameters: Record<string, any>
  color?: string
  visible: boolean
}

export interface SavedChart {
  id: string
  userId: string
  cardId: string
  name: string
  description?: string
  indicators: ChartIndicator[]
  timeRange: '1d' | '7d' | '30d' | '90d' | '1y' | '5y'
  createdAt: string
  updatedAt: string
}

export interface SearchResult {
  cards: Card[]
  total: number
  page: number
  perPage: number
  totalPages: number
}

export interface PriceHistory {
  prices: PricePoint[]
  marketData: MarketData[]
  indicators?: Record<string, IndicatorPoint[]>
}

export interface IndicatorPoint {
  timestamp: string
  value: number
  label?: string
}

export interface Dashboard {
  savedCharts: SavedChart[]
  recentlyViewed: Card[]
  userStats: UserStats
}

export interface UserStats {
  chartsCreated: number
  indicatorsUsed: number
  maxIndicators: number
  daysActive: number
  lastActiveDate: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ErrorResponse {
  error: string
  details?: Record<string, any>
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  firstName: string
  lastName: string
}

export interface SearchParams {
  q?: string
  game?: string
  category?: string
  page?: number
  limit?: number
}