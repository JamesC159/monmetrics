import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Mail, ArrowLeft, Sparkles } from 'lucide-react'
import { apiClient } from '@/utils/api'
import { useToast } from '@/context/ToastContext'
import { Layout } from '@/components/Layout'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [sent, setSent] = useState(false)

  const { addToast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      const { message } = await apiClient.forgotPassword(email)
      setSent(true)
      addToast(message, 'success')
    } catch (error) {
      addToast(error instanceof Error ? error.message : 'Could not send reset link', 'error')
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
                <h1 className='text-3xl font-display font-bold text-primary-200'>Reset Password</h1>
              </div>
              <p className='text-gray-400'>
                Enter your account email and we'll send you a link to choose a new password.
              </p>
            </div>

            {sent ? (
              <p className='text-gray-300'>
                If an account exists for <span className='text-white'>{email}</span>, check your
                inbox for the reset link. It expires in one hour.
              </p>
            ) : (
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

                <button
                  type='submit'
                  disabled={isLoading}
                  className='w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed'
                >
                  {isLoading ? 'Sending...' : 'Send Reset Link'}
                </button>
              </form>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
}
