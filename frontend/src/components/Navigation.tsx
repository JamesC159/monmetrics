import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Menu, X, Search, User, LogOut, Home, BarChart3, Sparkles } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'

export const Navigation = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const { isAuthenticated, user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  const isActive = (path: string) => location.pathname === path

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <nav className='fixed top-0 w-full z-50 bg-dark-950/95 backdrop-blur-md border-b border-primary-500/20'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
        <div className='flex justify-between items-center h-16'>
          {/* Logo */}
          <Link to='/' className='flex items-center group'>
            <Sparkles className='w-6 h-6 text-primary-400 mr-2 group-hover:animate-pulse' />
            <h1 className='text-2xl font-display font-bold gradient-text-gold'>MonMetrics</h1>
          </Link>

          {/* Desktop Navigation */}
          <div className='hidden md:flex items-center space-x-1'>
            <Link
              to='/'
              className={`flex items-center px-4 py-2 rounded-lg transition-all duration-200 ${
                isActive('/')
                  ? 'bg-primary-500/20 text-primary-400'
                  : 'text-gray-300 hover:text-white hover:bg-white/5'
              }`}
            >
              <Home className='w-4 h-4 mr-2' />
              Home
            </Link>

            <Link
              to='/search'
              className={`flex items-center px-4 py-2 rounded-lg transition-all duration-200 ${
                isActive('/search')
                  ? 'bg-secondary-500/20 text-secondary-400'
                  : 'text-gray-300 hover:text-white hover:bg-white/5'
              }`}
            >
              <Search className='w-4 h-4 mr-2' />
              Search
            </Link>

            {isAuthenticated && (
              <Link
                to='/dashboard'
                className={`flex items-center px-4 py-2 rounded-lg transition-all duration-200 ${
                  isActive('/dashboard')
                    ? 'bg-accent-500/20 text-accent-400'
                    : 'text-gray-300 hover:text-white hover:bg-white/5'
                }`}
              >
                <BarChart3 className='w-4 h-4 mr-2' />
                Dashboard
              </Link>
            )}
          </div>

          {/* Auth Buttons */}
          <div className='hidden md:flex items-center space-x-4'>
            {isAuthenticated ? (
              <div className='flex items-center space-x-4'>
                <div className='flex items-center px-4 py-2 bg-white/5 rounded-lg border border-primary-500/30'>
                  <User className='w-4 h-4 text-primary-400 mr-2' />
                  <span className='text-sm text-white font-medium'>
                    {user?.first_name || 'User'}
                  </span>
                </div>
                <button
                  onClick={handleLogout}
                  className='flex items-center px-4 py-2 bg-error-500/20 hover:bg-error-500/30 text-error-400 rounded-lg border border-error-500/40 transition-all duration-200'
                >
                  <LogOut className='w-4 h-4 mr-2' />
                  Logout
                </button>
              </div>
            ) : (
              <div className='flex items-center space-x-4'>
                <Link
                  to='/login'
                  className='px-4 py-2 text-gray-300 hover:text-white transition-colors'
                >
                  Sign In
                </Link>
                <Link to='/register' className='btn-primary'>
                  Get Started
                </Link>
              </div>
            )}
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className='md:hidden text-gray-300 hover:text-white p-2'
          >
            {isMenuOpen ? <X className='w-6 h-6' /> : <Menu className='w-6 h-6' />}
          </button>
        </div>
      </div>

      {/* Mobile Menu */}
      {isMenuOpen && (
        <div className='md:hidden bg-dark-900/95 backdrop-blur-lg border-t border-white/10 animate-slide-down'>
          <div className='px-4 py-4 space-y-2'>
            <Link
              to='/'
              onClick={() => setIsMenuOpen(false)}
              className={`flex items-center px-4 py-3 rounded-lg transition-all ${
                isActive('/')
                  ? 'bg-primary-500/20 text-primary-400'
                  : 'text-gray-300 hover:bg-white/5'
              }`}
            >
              <Home className='w-4 h-4 mr-3' />
              Home
            </Link>

            <Link
              to='/search'
              onClick={() => setIsMenuOpen(false)}
              className={`flex items-center px-4 py-3 rounded-lg transition-all ${
                isActive('/search')
                  ? 'bg-secondary-500/20 text-secondary-400'
                  : 'text-gray-300 hover:bg-white/5'
              }`}
            >
              <Search className='w-4 h-4 mr-3' />
              Search
            </Link>

            {isAuthenticated && (
              <Link
                to='/dashboard'
                onClick={() => setIsMenuOpen(false)}
                className={`flex items-center px-4 py-3 rounded-lg transition-all ${
                  isActive('/dashboard')
                    ? 'bg-accent-500/20 text-accent-400'
                    : 'text-gray-300 hover:bg-white/5'
                }`}
              >
                <BarChart3 className='w-4 h-4 mr-3' />
                Dashboard
              </Link>
            )}

            <div className='pt-4 border-t border-white/10 space-y-2'>
              {isAuthenticated ? (
                <>
                  <div className='px-4 py-3 bg-white/5 rounded-lg border border-primary-500/30'>
                    <div className='flex items-center'>
                      <User className='w-4 h-4 text-primary-400 mr-2' />
                      <span className='text-sm text-white font-medium'>
                        {user?.first_name || 'User'}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => {
                      handleLogout()
                      setIsMenuOpen(false)
                    }}
                    className='w-full flex items-center px-4 py-3 bg-error-500/20 text-error-400 rounded-lg border border-error-500/40'
                  >
                    <LogOut className='w-4 h-4 mr-3' />
                    Logout
                  </button>
                </>
              ) : (
                <>
                  <Link
                    to='/login'
                    onClick={() => setIsMenuOpen(false)}
                    className='block w-full px-4 py-3 text-center text-gray-300 hover:bg-white/5 rounded-lg transition-all'
                  >
                    Sign In
                  </Link>
                  <Link
                    to='/register'
                    onClick={() => setIsMenuOpen(false)}
                    className='block w-full btn-primary text-center'
                  >
                    Get Started
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </nav>
  )
}
