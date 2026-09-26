import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '@/utils/api'
import { useAuth } from './AuthContext'
import { useToast } from './ToastContext'

interface FavoritesContextType {
  favoriteIds: Set<string>
  isFavorite: (cardId: string) => boolean
  toggleFavorite: (cardId: string) => Promise<void>
  refreshFavorites: () => Promise<void>
}

const FavoritesContext = createContext<FavoritesContextType | undefined>(undefined)

export function useFavorites() {
  const context = useContext(FavoritesContext)
  if (context === undefined) {
    throw new Error('useFavorites must be used within a FavoritesProvider')
  }
  return context
}

export function FavoritesProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const { addToast } = useToast()
  const navigate = useNavigate()
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set())

  const refreshFavorites = useCallback(async () => {
    try {
      const favs = await apiClient.getFavorites()
      setFavoriteIds(new Set(favs.map((f) => f.card_id)))
    } catch (err) {
      console.error('Failed to load favorites:', err)
    }
  }, [])

  useEffect(() => {
    if (isAuthenticated) refreshFavorites()
    else setFavoriteIds(new Set())
  }, [isAuthenticated, refreshFavorites])

  const toggleFavorite = useCallback(
    async (cardId: string) => {
      if (!isAuthenticated) {
        navigate('/login')
        return
      }
      const wasFavorite = favoriteIds.has(cardId)
      const apply = (add: boolean) =>
        setFavoriteIds((prev) => {
          const next = new Set(prev)
          if (add) next.add(cardId)
          else next.delete(cardId)
          return next
        })

      apply(!wasFavorite)
      try {
        if (wasFavorite) await apiClient.removeFavorite(cardId)
        else await apiClient.addFavorite(cardId)
      } catch (err) {
        apply(wasFavorite)
        addToast(err instanceof Error ? err.message : 'Could not update favorites', 'error')
      }
    },
    [isAuthenticated, favoriteIds, navigate, addToast],
  )

  const value = useMemo(
    () => ({
      favoriteIds,
      isFavorite: (id: string) => favoriteIds.has(id),
      toggleFavorite,
      refreshFavorites,
    }),
    [favoriteIds, toggleFavorite, refreshFavorites],
  )

  return <FavoritesContext.Provider value={value}>{children}</FavoritesContext.Provider>
}
