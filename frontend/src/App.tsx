import { Routes, Route, Navigate } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import LandingPage from './pages/LandingPage'
import { useAuth } from './context/AuthContext'
import Dashboard from './pages/Dashboard'
import CardDetail from './pages/CardDetail'
import Search from './pages/Search'
import Browse from './pages/Browse'
import Admin from './pages/Admin'
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import NotFound from './pages/NotFound'
import PortfolioItemDetail from './pages/PortfolioItemDetail'
import MarketplaceCallback from './pages/MarketplaceCallback'
import { AuthProvider } from './context/AuthContext'
import { ToastProvider } from './context/ToastContext'
import { FavoritesProvider } from './context/FavoritesContext'
import { FeaturesProvider } from './context/FeaturesContext'

// The dashboard is the home page for signed-in users, so the landing page is only shown to guests.
function HomeRoute() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) return null
  if (isAuthenticated) return <Navigate to='/dashboard' replace />
  return <LandingPage />
}

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <ToastProvider>
          <FavoritesProvider>
            <FeaturesProvider>
              <Routes>
                <Route path='/' element={<HomeRoute />} />
                <Route path='/login' element={<Login />} />
                <Route path='/register' element={<Register />} />
                <Route path='/forgot-password' element={<ForgotPassword />} />
                <Route path='/reset-password' element={<ResetPassword />} />
                <Route path='/search' element={<Search />} />
                <Route path='/browse' element={<Browse />} />
                <Route path='/browse/:game' element={<Browse />} />
                <Route path='/card/:id' element={<CardDetail />} />
                <Route path='/dashboard' element={<Dashboard />} />{' '}
                <Route path='/admin' element={<Admin />} />{' '}
                <Route path='/portfolio/:id' element={<PortfolioItemDetail />} />
                <Route path='/marketplace/callback' element={<MarketplaceCallback />} />
                <Route path='*' element={<NotFound />} />
              </Routes>
            </FeaturesProvider>
          </FavoritesProvider>
        </ToastProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App
