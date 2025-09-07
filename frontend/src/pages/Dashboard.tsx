import { useAuth } from '@/context/AuthContext'
import { Navigate, Link } from 'react-router-dom'
import { BarChart3, TrendingUp, Star, Plus } from 'lucide-react'

export default function Dashboard() {
  const { isAuthenticated, user, logout } = useAuth()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      <nav className="border-b border-white/10 bg-black/20 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link to="/" className="text-2xl font-bold gradient-text">
              MonMetrics
            </Link>

            <div className="flex items-center space-x-4">
              <span className="text-gray-300">Welcome, {user?.firstName}</span>
              <button onClick={logout} className="button-secondary">
                Sign Out
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">Dashboard</h1>
          <p className="text-gray-400">Manage your charts and track your favorite cards</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="glass-effect rounded-2xl p-6">
            <div className="flex items-center">
              <BarChart3 className="w-8 h-8 text-blue-400 mr-3" />
              <div>
                <h3 className="text-lg font-semibold text-white">Saved Charts</h3>
                <p className="text-gray-400">0 charts</p>
              </div>
            </div>
          </div>

          <div className="glass-effect rounded-2xl p-6">
            <div className="flex items-center">
              <TrendingUp className="w-8 h-8 text-green-400 mr-3" />
              <div>
                <h3 className="text-lg font-semibold text-white">Watchlist</h3>
                <p className="text-gray-400">0 cards</p>
              </div>
            </div>
          </div>

          <div className="glass-effect rounded-2xl p-6">
            <div className="flex items-center">
              <Star className="w-8 h-8 text-yellow-400 mr-3" />
              <div>
                <h3 className="text-lg font-semibold text-white">Plan</h3>
                <p className="text-gray-400 capitalize">{user?.userType || 'Free'}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="glass-effect rounded-2xl p-8">
          <div className="text-center py-16">
            <Plus className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-400 mb-2">No Charts Yet</h3>
            <p className="text-gray-500 mb-6">
              Start by searching for cards and creating your first chart
            </p>
            <Link to="/search" className="button-primary">
              Search Cards
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}