import { Routes, Route } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import LandingPage from './pages/LandingPage'
import Dashboard from './pages/Dashboard'
import CardDetail from './pages/CardDetail'
import Search from './pages/Search'
import Login from './pages/Login'
import Register from './pages/Register'
import NotFound from './pages/NotFound'
import { AuthProvider } from './context/AuthContext'
import { ToastProvider } from './context/ToastContext'

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <ToastProvider>
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/search" element={<Search />} />
            <Route path="/card/:id" element={<CardDetail />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App