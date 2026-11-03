import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Eye, EyeOff, Lock, ArrowLeft, Sparkles } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { Layout } from '@/components/Layout'

export default function ResetPassword() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''

  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const { addToast } = useToast()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (password !== confirmPassword) {
      addToast('Passwords do not match', 'error')
      return
    }

    setIsLoading(true)
    try {
      const { message } = await apiClient.resetPassword(token, password)
      addToast(message, 'success')
      navigate('/login', { replace: true })
    } catch (error) {
      addToast(error instanceof Error ? error.message : 'Could not reset password', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Layout>
      <div className='min-h-screen bg-dark-900 flex items-center justify-center px-4 py-12'>
        <div className='max-w-md w-full'>
          <div className='card border-primary-500/30 shadow-glow-gold animate-fade-in'>
            <div className='mb-8'>
              <Link
                to='/login'
                className='inline-flex items-center text-gray-400 hover:text-primary-400 transition-colors mb-6 group'
              >
                <ArrowLeft className='w-4 h-4 mr-2 group-hover:-translate-x-1 transition-transform' />
                Back to Sign In
              </Link>

              <div className='flex items-center mb-4'>
                <Sparkles className='w-8 h-8 text-primary-400 mr-3' />
                <h1 className='text-3xl font-display font-bold text-primary-200'>New Password</h1>
              </div>
              <p className='text-gray-400'>Choose a new password for your account.</p>
            </div>

            {!token ? (
              <p className='text-gray-300'>
                This reset link is missing its token. Request a new one from the{' '}
                <Link to='/forgot-password' className='text-secondary-400 hover:text-secondary-300'>
                  forgot password
                </Link>{' '}
                page.
              </p>
            ) : (
              <form onSubmit={handleSubmit} className='space-y-6'>
                <div>
                  <label htmlFor='password' className='block text-sm font-medium text-gray-300 mb-2'>
                    New Password
                  </label>
                  <div className='relative group'>
                    <Lock className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 group-focus-within:text-primary-400 w-5 h-5 transition-colors' />
                    <input
                      id='password'
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className='input pl-10 pr-12'
                      placeholder='At least 8 characters'
                      minLength={8}
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

                <div>
                  <label
                    htmlFor='confirmPassword'
                    className='block text-sm font-medium text-gray-300 mb-2'
                  >
                    Confirm Password
                  </label>
                  <div className='relative group'>
                    <Lock className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 group-focus-within:text-primary-400 w-5 h-5 transition-colors' />
                    <input
                      id='confirmPassword'
                      type={showPassword ? 'text' : 'password'}
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className='input pl-10'
                      placeholder='Re-enter your new password'
                      required
                    />
                  </div>
                </div>

                <button
                  type='submit'
                  disabled={isLoading}
                  className='w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed'
                >
                  {isLoading ? 'Saving...' : 'Set New Password'}
                </button>
              </form>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
}
