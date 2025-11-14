import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { Eye, EyeOff, Mail, Lock, ArrowLeft, Sparkles } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import { Layout } from '@/components/Layout'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const { login } = useAuth()
  const { addToast } = useToast()
  const navigate = useNavigate()
  const location = useLocation()

  const from = (location.state as any)?.from?.pathname || '/dashboard'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      await login(email, password)
      addToast('Successfully logged in!', 'success')
      navigate(from, { replace: true })
    } catch (error) {
      addToast(error instanceof Error ? error.message : 'Login failed', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Layout>
      <div className='min-h-screen bg-gradient-to-br from-dark-900 via-dark-950 to-black flex items-center justify-center px-4 py-12'>
        <div className='max-w-md w-full'>
          {/* Decorative elements */}
          <div className='absolute top-20 left-10 w-72 h-72 bg-primary-500/10 rounded-full blur-3xl'></div>
          <div className='absolute bottom-20 right-10 w-96 h-96 bg-secondary-500/10 rounded-full blur-3xl'></div>

          <div className='relative'>
            <div className='card border-primary-500/30 shadow-glow-gold animate-fade-in'>
              <div className='mb-8'>
                <Link
                  to='/'
                  className='inline-flex items-center text-gray-400 hover:text-primary-400 transition-colors mb-6 group'
                >
                  <ArrowLeft className='w-4 h-4 mr-2 group-hover:-translate-x-1 transition-transform' />
                  Back to Home
                </Link>

                <div className='flex items-center mb-4'>
                  <Sparkles className='w-8 h-8 text-primary-400 mr-3 animate-pulse' />
                  <h1 className='text-3xl font-display font-bold gradient-text-gold'>
                    Welcome Back
                  </h1>
                </div>
                <p className='text-gray-400'>Sign in to your MonMetrics account</p>
              </div>

              <form onSubmit={handleSubmit} className='space-y-6'>
                <div>
                  <label htmlFor='email' className='block text-sm font-medium text-gray-300 mb-2'>
                    Email Address
                  </label>
                  <div className='relative group'>
                    <Mail className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 group-focus-within:text-primary-400 w-5 h-5 transition-colors' />
                    <input
                      id='email'
                      type='email'
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className='input pl-10'
                      placeholder='Enter your email'
                      required
                    />
                  </div>
                </div>

                <div>
                  <label
                    htmlFor='password'
                    className='block text-sm font-medium text-gray-300 mb-2'
                  >
                    Password
                  </label>
                  <div className='relative group'>
                    <Lock className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 group-focus-within:text-primary-400 w-5 h-5 transition-colors' />
                    <input
                      id='password'
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className='input pl-10 pr-12'
                      placeholder='Enter your password'
                      required
                    />
                    <button
                      type='button'
                      onClick={() => setShowPassword(!showPassword)}
                      className='absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-primary-400 transition-colors'
                    >
                      {showPassword ? <EyeOff className='w-5 h-5' /> : <Eye className='w-5 h-5' />}
                    </button>
                  </div>
                </div>

                <div className='flex items-center justify-between'>
                  <label className='flex items-center cursor-pointer group'>
                    <input
                      type='checkbox'
                      className='rounded border-gray-600 text-primary-500 focus:ring-primary-500 bg-dark-800'
                    />
                    <span className='ml-2 text-sm text-gray-300 group-hover:text-white transition-colors'>
                      Remember me
                    </span>
                  </label>
                  <a
                    href='#'
                    className='text-sm text-secondary-400 hover:text-secondary-300 transition-colors'
                  >
                    Forgot password?
                  </a>
                </div>

                <button
                  type='submit'
                  disabled={isLoading}
                  className='w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed'
                >
                  {isLoading ? (
                    <div className='flex items-center justify-center'>
                      <div className='loading-spinner mr-2'></div>
                      Signing In...
                    </div>
                  ) : (
                    'Sign In'
                  )}
                </button>
              </form>

              <div className='mt-8 text-center'>
                <p className='text-gray-400'>
                  Don't have an account?{' '}
                  <Link
                    to='/register'
                    className='text-secondary-400 hover:text-secondary-300 font-medium transition-colors'
                  >
                    Sign up for free
                  </Link>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}
