import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { apiClient } from '@/utils/api'
import { User, AuthResponse } from '@/types'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (data: RegisterData) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
}

interface RegisterData {
  email: string
  password: string
  firstName: string
  lastName: string
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Check for existing auth token on mount
    const token = localStorage.getItem('authToken')
    if (token) {
      apiClient.setToken(token)
      // Validate token by fetching user data
      validateToken()
    } else {
      setIsLoading(false)
    }
  }, [])

  const validateToken = async () => {
    try {
      await apiClient.getDashboard()
      // Extract user from dashboard or create a minimal user object
      // For now, we'll just mark token as valid
      // In production, you'd want a dedicated /api/auth/me endpoint
      setUser({
        id: '',
        email: '',
        first_name: '',
        last_name: '',
        user_type: 'free',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        is_active: true,
      })
      setIsLoading(false)
    } catch (error) {
      // Token is invalid or endpoint doesn't exist, clear it
      console.log('Token validation failed:', error)
      apiClient.clearToken()
      setUser(null)
      setIsLoading(false)
    }
  }

  const login = async (email: string, password: string) => {
    try {
      const response: AuthResponse = await apiClient.login({ email, password })
      apiClient.setToken(response.token)
      setUser(response.user)
    } catch (error) {
      throw error
    }
  }

  const register = async (data: RegisterData) => {
    try {
      const response: AuthResponse = await apiClient.register({
        email: data.email,
        password: data.password,
        first_name: data.firstName,
        last_name: data.lastName,
      })
      apiClient.setToken(response.token)
      setUser(response.user)
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    apiClient.clearToken()
    setUser(null)
    // Optionally call backend logout endpoint
    apiClient.logout().catch(console.error)
  }

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    register,
    logout,
    isAuthenticated: !!user,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
