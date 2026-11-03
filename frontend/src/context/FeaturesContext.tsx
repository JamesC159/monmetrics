import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import type { FeatureFlags, FeatureKey } from '../types'
import { apiClient } from '../utils/api'

interface FeaturesContextValue {
  flags: FeatureFlags
  /** False until the server responds, so gated UI never flashes on then off. */
  loaded: boolean
  isEnabled: (key: FeatureKey) => boolean
  refresh: () => Promise<void>
}

const FeaturesContext = createContext<FeaturesContextValue | undefined>(undefined)

export function FeaturesProvider({ children }: { children: ReactNode }) {
  const [flags, setFlags] = useState<FeatureFlags>({})
  const [loaded, setLoaded] = useState(false)

  const refresh = useCallback(async () => {
    try {
      setFlags(await apiClient.getFeatures())
    } catch {
      // An unreachable flag endpoint leaves everything off, matching server defaults.
      setFlags({})
    } finally {
      setLoaded(true)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  const isEnabled = useCallback((key: FeatureKey) => flags[key] === true, [flags])

  return (
    <FeaturesContext.Provider value={{ flags, loaded, isEnabled, refresh }}>
      {children}
    </FeaturesContext.Provider>
  )
}

export function useFeatures() {
  const ctx = useContext(FeaturesContext)
  if (!ctx) throw new Error('useFeatures must be used within a FeaturesProvider')
  return ctx
}
