import type {
  User,
  Card,
  SearchResult,
  PriceHistory,
  SavedChart,
  Dashboard,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  SearchParams,
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

  async getCardPrices(id: string, range: string = '30d'): Promise<PriceHistory> {
    return this.request<PriceHistory>(`/api/cards/${id}/prices?range=${range}`)
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

  async saveChart(
    chart: Omit<SavedChart, 'id' | 'userId' | 'createdAt' | 'updatedAt'>,
  ): Promise<SavedChart> {
    return this.request<SavedChart>('/api/protected/user/charts', {
      method: 'POST',
      body: JSON.stringify(chart),
    })
  }

  async getSavedCharts(): Promise<SavedChart[]> {
    return this.request<SavedChart[]>('/api/protected/user/charts')
  }

  async deleteChart(id: string): Promise<void> {
    return this.request(`/api/protected/user/charts/${id}`, { method: 'DELETE' })
  }

  async updateChart(id: string, chart: Partial<SavedChart>): Promise<SavedChart> {
    return this.request<SavedChart>(`/api/protected/user/charts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(chart),
    })
  }
}

export const apiClient = new ApiClient(API_BASE_URL)

// Export ApiError type for use in components
export type { ApiError }
