import type {
  Card,
  SearchResult,
  PriceHistory,
  SavedChart,
  SavedChartRequest,
  Dashboard,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  SearchParams,
  FavoriteView,
  PortfolioResponse,
  PortfolioQuery,
  PortfolioItemView,
  PortfolioItemRequest,
  LinkedAccountStatus,
  MarketplaceProvider,
  CompsResponse,
  PrefillResponse,
  ListingDraft,
  MarketplaceListing,
  PriceAlert,
  PriceAlertRequest,
  NotificationList,
} from '@/types'

// Safe environment variable access with fallback
const getEnvVar = (key: string, fallback: string): string => {
  if (typeof import.meta !== 'undefined' && import.meta.env) {
    return import.meta.env[key] || fallback
  }
  return fallback
}

const API_BASE_URL = getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080')

interface ApiError extends Error {
  status?: number
  details?: Record<string, any>
}

class ApiClient {
  private baseUrl: string
  private token: string | null = null

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
    // Only access localStorage on client side
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('authToken')
    }
  }

  setToken(token: string) {
    this.token = token
    if (typeof window !== 'undefined') {
      localStorage.setItem('authToken', token)
    }
  }

  clearToken() {
    this.token = null
    if (typeof window !== 'undefined') {
      localStorage.removeItem('authToken')
    }
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`

    // Create headers using Headers constructor for proper typing
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')

    // Add authorization header if token exists
    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    // Add any additional headers from options
    if (options.headers) {
      if (options.headers instanceof Headers) {
        options.headers.forEach((value, key) => {
          headers.set(key, value)
        })
      } else if (Array.isArray(options.headers)) {
        options.headers.forEach(([key, value]) => {
          headers.set(key, value)
        })
      } else {
        Object.entries(options.headers).forEach(([key, value]) => {
          if (typeof value === 'string') {
            headers.set(key, value)
          }
        })
      }
    }

    const config: RequestInit = {
      ...options,
      headers,
    }

    try {
      const response = await fetch(url, config)

      if (!response.ok) {
        let errorData: any = {}
        try {
          errorData = await response.json()
        } catch {
          // If response is not JSON, use status text
          errorData = { error: response.statusText }
        }

        const error: ApiError = new Error(
          errorData.error || `HTTP ${response.status}: ${response.statusText}`,
        )
        error.status = response.status
        error.details = errorData.details
        throw error
      }

      // Handle empty responses (like 204 No Content)
      if (response.status === 204 || response.headers.get('content-length') === '0') {
        return {} as T
      }

      return await response.json()
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Network request failed')
    }
  }

  // Health check
  async health() {
    return this.request<{ status: string; timestamp: string; version: string }>('/health')
  }

  // Auth methods
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async logout(): Promise<void> {
    return this.request('/api/auth/logout', { method: 'POST' })
  }

  // Card methods
  async searchCards(params: SearchParams = {}): Promise<SearchResult> {
    const searchParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString())
      }
    })

    const queryString = searchParams.toString()
    const endpoint = `/api/cards/search${queryString ? `?${queryString}` : ''}`

    return this.request<SearchResult>(endpoint)
  }

  async getCard(id: string): Promise<Card> {
    return this.request<Card>(`/api/cards/${id}`)
  }

  async getCardPrices(id: string, range: string = '30d', warmupDays = 0): Promise<PriceHistory> {
    const params = new URLSearchParams({ range })
    if (warmupDays > 0) params.set('warmup_days', String(warmupDays))
    return this.request<PriceHistory>(`/api/cards/${encodeURIComponent(id)}/prices?${params}`)
  }

  // Featured content and organized search
  async getFeaturedContent(): Promise<import('@/types').FeaturedContent[]> {
    return this.request<import('@/types').FeaturedContent[]>('/api/featured-content')
  }

  async getCardsByGame(): Promise<import('@/types').GameCardGroup[]> {
    return this.request<import('@/types').GameCardGroup[]>('/api/cards/by-game')
  }

  async getSealedByGame(): Promise<import('@/types').GameCardGroup[]> {
    return this.request<import('@/types').GameCardGroup[]>('/api/sealed/by-game')
  }

  // Protected methods (require authentication)
  async getDashboard(): Promise<Dashboard> {
    return this.request<Dashboard>('/api/protected/user/dashboard')
  }

  async saveChart(chart: SavedChartRequest): Promise<SavedChart> {
    return this.request<SavedChart>('/api/protected/user/charts', {
      method: 'POST',
      body: JSON.stringify(chart),
    })
  }

  async getSavedCharts(): Promise<SavedChart[]> {
    return this.request<SavedChart[]>('/api/protected/user/charts')
  }

  async getChart(id: string): Promise<SavedChart> {
    return this.request<SavedChart>(`/api/protected/user/charts/${encodeURIComponent(id)}`)
  }

  async deleteChart(id: string): Promise<void> {
    return this.request(`/api/protected/user/charts/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    })
  }

  async updateChart(id: string, chart: SavedChartRequest): Promise<SavedChart> {
    return this.request<SavedChart>(`/api/protected/user/charts/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(chart),
    })
  }

  // Price alerts
  async getAlerts(cardId?: string): Promise<PriceAlert[]> {
    const qs = cardId ? `?card_id=${encodeURIComponent(cardId)}` : ''
    return this.request<PriceAlert[]>(`/api/protected/user/alerts${qs}`)
  }

  async createAlert(alert: PriceAlertRequest): Promise<PriceAlert> {
    return this.request<PriceAlert>('/api/protected/user/alerts', {
      method: 'POST',
      body: JSON.stringify(alert),
    })
  }

  async updateAlert(id: string, alert: PriceAlertRequest): Promise<PriceAlert> {
    return this.request<PriceAlert>(`/api/protected/user/alerts/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(alert),
    })
  }

  async deleteAlert(id: string): Promise<void> {
    await this.request(`/api/protected/user/alerts/${encodeURIComponent(id)}`, { method: 'DELETE' })
  }

  // Notifications
  async getNotifications(unreadOnly = false, limit = 20): Promise<NotificationList> {
    const params = new URLSearchParams({ limit: String(limit) })
    if (unreadOnly) params.set('unread', '1')
    return this.request<NotificationList>(`/api/protected/user/notifications?${params}`)
  }

  async markNotificationRead(id: string): Promise<void> {
    await this.request(`/api/protected/user/notifications/${encodeURIComponent(id)}/read`, {
      method: 'POST',
    })
  }

  async markAllNotificationsRead(): Promise<void> {
    await this.request('/api/protected/user/notifications/read-all', { method: 'POST' })
  }

  // Favorites
  async getFavorites(): Promise<FavoriteView[]> {
    return this.request<FavoriteView[]>('/api/protected/favorites')
  }

  async addFavorite(cardId: string): Promise<void> {
    await this.request('/api/protected/favorites', {
      method: 'POST',
      body: JSON.stringify({ card_id: cardId }),
    })
  }

  async removeFavorite(cardId: string): Promise<void> {
    await this.request(`/api/protected/favorites/${encodeURIComponent(cardId)}`, {
      method: 'DELETE',
    })
  }

  // Portfolio
  async getPortfolio(query: PortfolioQuery = {}): Promise<PortfolioResponse> {
    const params = new URLSearchParams()
    Object.entries(query).forEach(([k, v]) => {
      if (v) params.set(k, String(v))
    })
    const qs = params.toString()
    return this.request<PortfolioResponse>(`/api/protected/portfolio${qs ? `?${qs}` : ''}`)
  }

  async getPortfolioItem(id: string): Promise<PortfolioItemView> {
    return this.request<PortfolioItemView>(`/api/protected/portfolio/${encodeURIComponent(id)}`)
  }

  async createPortfolioItem(item: PortfolioItemRequest): Promise<PortfolioItemView> {
    return this.request<PortfolioItemView>('/api/protected/portfolio', {
      method: 'POST',
      body: JSON.stringify(item),
    })
  }

  async updatePortfolioItem(id: string, item: PortfolioItemRequest): Promise<PortfolioItemView> {
    return this.request<PortfolioItemView>(`/api/protected/portfolio/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(item),
    })
  }

  async deletePortfolioItem(id: string): Promise<void> {
    await this.request(`/api/protected/portfolio/${encodeURIComponent(id)}`, { method: 'DELETE' })
  }

  // Marketplace
  async getLinkedAccounts(): Promise<LinkedAccountStatus[]> {
    return this.request<LinkedAccountStatus[]>('/api/protected/marketplace/accounts')
  }

  async connectMarketplace(provider: MarketplaceProvider): Promise<{ auth_url: string }> {
    return this.request<{ auth_url: string }>(`/api/protected/marketplace/${provider}/connect`)
  }

  async completeMarketplaceLink(
    provider: MarketplaceProvider,
    code: string,
    state: string,
  ): Promise<{ provider: string; username: string }> {
    return this.request(`/api/protected/marketplace/${provider}/callback`, {
      method: 'POST',
      body: JSON.stringify({ code, state }),
    })
  }

  async disconnectMarketplace(provider: MarketplaceProvider): Promise<void> {
    await this.request(`/api/protected/marketplace/${provider}`, { method: 'DELETE' })
  }

  async getComps(portfolioItemId: string, provider: MarketplaceProvider): Promise<CompsResponse> {
    const params = new URLSearchParams({ portfolio_item_id: portfolioItemId, provider })
    return this.request<CompsResponse>(`/api/protected/marketplace/comps?${params}`)
  }

  async prefillListing(
    portfolioItemId: string,
    provider: MarketplaceProvider,
  ): Promise<PrefillResponse> {
    return this.request<PrefillResponse>('/api/protected/marketplace/listings/prefill', {
      method: 'POST',
      body: JSON.stringify({ portfolio_item_id: portfolioItemId, provider }),
    })
  }

  async createListing(draft: ListingDraft, publish: boolean): Promise<MarketplaceListing> {
    return this.request<MarketplaceListing>('/api/protected/marketplace/listings', {
      method: 'POST',
      body: JSON.stringify({ draft, publish }),
    })
  }

  async getListings(portfolioItemId?: string): Promise<MarketplaceListing[]> {
    const qs = portfolioItemId ? `?portfolio_item_id=${encodeURIComponent(portfolioItemId)}` : ''
    return this.request<MarketplaceListing[]>(`/api/protected/marketplace/listings${qs}`)
  }
}

export const apiClient = new ApiClient(API_BASE_URL)

// Export ApiError type for use in components
export type { ApiError }
