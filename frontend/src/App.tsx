import { Routes, Route } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import LandingPage from './pages/LandingPage'
import Dashboard from './pages/Dashboard'
import CardDetail from './pages/CardDetail'
import Search from './pages/Search'
import Login from './pages/Login'
import Register from './pages/Register'
import NotFound from './pages/NotFound'
import PortfolioItemDetail from './pages/PortfolioItemDetail'
import MarketplaceCallback from './pages/MarketplaceCallback'
import { AuthProvider } from './context/AuthContext'
import { ToastProvider } from './context/ToastContext'
import { FavoritesProvider } from './context/FavoritesContext'

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <ToastProvider>
          <FavoritesProvider>
            <Routes>
              <Route path="/" element={<LandingPage />} />
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/search" element={<Search />} />
              <Route path="/card/:id" element={<CardDetail />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/portfolio/:id" element={<PortfolioItemDetail />} />
              <Route path="/marketplace/callback" element={<MarketplaceCallback />} />
              <Route path="*" element={<NotFound />} />
            </Routes>
          </FavoritesProvider>
        </ToastProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App