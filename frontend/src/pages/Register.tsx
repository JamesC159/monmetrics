import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Eye, EyeOff, Mail, Lock, User, ArrowLeft, CheckCircle2, XCircle } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import {
  validateEmail,
  validatePassword,
  validateName,
  sanitizeName,
  normalizeEmail,
  getPasswordStrength,
} from '@/utils/validation'

export default function Register() {
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [passwordStrength, setPasswordStrength] = useState({ score: 0, label: '', color: '' })
  const [validationErrors, setValidationErrors] = useState<{
    email?: string
    password?: string[]
    firstName?: string
    lastName?: string
    confirmPassword?: string
  }>({})

  const { register } = useAuth()
  const { addToast } = useToast()
  const navigate = useNavigate()

  // Real-time password strength check
  useEffect(() => {
    if (formData.password) {
      setPasswordStrength(getPasswordStrength(formData.password))
    } else {
      setPasswordStrength({ score: 0, label: '', color: '' })
    }
  }, [formData.password])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }))

    // Clear validation error for this field
    setValidationErrors((prev) => ({
      ...prev,
      [name]: undefined,
    }))
  }

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    const errors: typeof validationErrors = {}

    // Validate on blur
    if (name === 'email' && value) {
      if (!validateEmail(value)) {
        errors.email = 'Please enter a valid email address'
      }
    }

    if (name === 'password' && value) {
      const validation = validatePassword(value)
      if (!validation.isValid) {
        errors.password = validation.errors
      }
    }

    if (name === 'firstName' && value) {
      const validation = validateName(value)
      if (!validation.isValid) {
        errors.firstName = validation.error
      }
    }

    if (name === 'lastName' && value) {
      const validation = validateName(value)
      if (!validation.isValid) {
        errors.lastName = validation.error
      }
    }

    if (name === 'confirmPassword' && value) {
      if (value !== formData.password) {
        errors.confirmPassword = 'Passwords do not match'
      }
    }

    setValidationErrors((prev) => ({ ...prev, ...errors }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    // Validate all fields
    const errors: typeof validationErrors = {}

    if (!validateEmail(formData.email)) {
      errors.email = 'Please enter a valid email address'
    }

    const passwordValidation = validatePassword(formData.password)
    if (!passwordValidation.isValid) {
      errors.password = passwordValidation.errors
    }

    const firstNameValidation = validateName(formData.firstName)
    if (!firstNameValidation.isValid) {
      errors.firstName = firstNameValidation.error
    }

    const lastNameValidation = validateName(formData.lastName)
    if (!lastNameValidation.isValid) {
      errors.lastName = lastNameValidation.error
    }

    if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = 'Passwords do not match'
    }

    if (Object.keys(errors).length > 0) {
      setValidationErrors(errors)
      addToast('Please fix the errors before submitting', 'error')
      return
    }

    setIsLoading(true)

    try {
      // Sanitize and normalize inputs before sending
      await register({
        firstName: sanitizeName(formData.firstName),
        lastName: sanitizeName(formData.lastName),
        email: normalizeEmail(formData.email),
        password: formData.password, // Don't modify password
      })
      addToast('Account created successfully!', 'success')
      navigate('/dashboard')
    } catch (error) {
      addToast(error instanceof Error ? error.message : 'Registration failed', 'error')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className='min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center px-4 py-8'>
      <div className='max-w-md w-full'>
        <div className='glass-effect rounded-2xl p-8'>
          <div className='mb-8'>
            <Link
              to='/'
              className='inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4'
            >
              <ArrowLeft className='w-4 h-4 mr-2' />
              Back to Home
            </Link>

            <h1 className='text-3xl font-bold text-white mb-2'>Create Account</h1>
            <p className='text-gray-400'>Start your trading card analysis journey</p>
          </div>

          <form onSubmit={handleSubmit} className='space-y-6'>
            <div className='grid grid-cols-2 gap-4'>
              <div>
                <label htmlFor='firstName' className='block text-sm font-medium text-gray-300 mb-2'>
                  First Name
                </label>
                <div className='relative'>
                  <User className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5' />
                  <input
                    id='firstName'
                    name='firstName'
                    type='text'
                    value={formData.firstName}
                    onChange={handleChange}
                    className='w-full pl-10 pr-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent'
                    placeholder='John'
                    required
                  />
                </div>
              </div>

              <div>
                <label htmlFor='lastName' className='block text-sm font-medium text-gray-300 mb-2'>
                  Last Name
                </label>
                <input
                  id='lastName'
                  name='lastName'
                  type='text'
                  value={formData.lastName}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  className={`w-full px-4 py-3 bg-white/10 border ${
                    validationErrors.lastName ? 'border-red-500' : 'border-white/20'
                  } rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent`}
                  placeholder='Doe'
                  required
                />
                {validationErrors.lastName && (
                  <p className='text-red-400 text-sm mt-1'>{validationErrors.lastName}</p>
                )}
              </div>
            </div>

            <div>
              <label htmlFor='email' className='block text-sm font-medium text-gray-300 mb-2'>
                Email Address
              </label>
              <div className='relative'>
                <Mail className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5' />
                <input
                  id='email'
                  name='email'
                  type='email'
                  value={formData.email}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  className={`w-full pl-10 pr-4 py-3 bg-white/10 border ${
                    validationErrors.email ? 'border-red-500' : 'border-white/20'
                  } rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent`}
                  placeholder='john@example.com'
                  required
                />
              </div>
              {validationErrors.email && (
                <p className='text-red-400 text-sm mt-1'>{validationErrors.email}</p>
              )}
            </div>

            <div>
              <label htmlFor='password' className='block text-sm font-medium text-gray-300 mb-2'>
                Password
              </label>
              <div className='relative'>
                <Lock className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5' />
                <input
                  id='password'
                  name='password'
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  className={`w-full pl-10 pr-12 py-3 bg-white/10 border ${
                    validationErrors.password ? 'border-red-500' : 'border-white/20'
                  } rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent`}
                  placeholder='Minimum 8 characters'
                  required
                />
                <button
                  type='button'
                  onClick={() => setShowPassword(!showPassword)}
                  className='absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white'
                >
                  {showPassword ? <EyeOff className='w-5 h-5' /> : <Eye className='w-5 h-5' />}
                </button>
              </div>

              {/* Password Strength Indicator */}
              {formData.password && (
                <div className='mt-2'>
                  <div className='flex items-center justify-between mb-1'>
                    <span className='text-xs text-gray-400'>Password Strength:</span>
                    <span className='text-xs font-medium' style={{ color: passwordStrength.color }}>
                      {passwordStrength.label}
                    </span>
                  </div>
                  <div className='w-full bg-gray-700 rounded-full h-1.5'>
                    <div
                      className='h-1.5 rounded-full transition-all duration-300'
                      style={{
                        width: `${(passwordStrength.score + 1) * 20}%`,
                        backgroundColor: passwordStrength.color,
                      }}
                    />
                  </div>
                </div>
              )}

              {validationErrors.password && (
                <div className='mt-2 space-y-1'>
                  {validationErrors.password.map((error, idx) => (
                    <p key={idx} className='text-red-400 text-xs flex items-center gap-1'>
                      <XCircle className='w-3 h-3' />
                      {error}
                    </p>
                  ))}
                </div>
              )}
            </div>

            <div>
              <label
                htmlFor='confirmPassword'
                className='block text-sm font-medium text-gray-300 mb-2'
              >
                Confirm Password
              </label>
              <div className='relative'>
                <Lock className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5' />
                <input
                  id='confirmPassword'
                  name='confirmPassword'
                  type={showConfirmPassword ? 'text' : 'password'}
                  value={formData.confirmPassword}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  className={`w-full pl-10 pr-12 py-3 bg-white/10 border ${
                    validationErrors.confirmPassword ? 'border-red-500' : 'border-white/20'
                  } rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent`}
                  placeholder='Confirm your password'
                  required
                />
                <button
                  type='button'
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className='absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white'
                >
                  {showConfirmPassword ? (
                    <EyeOff className='w-5 h-5' />
                  ) : (
                    <Eye className='w-5 h-5' />
                  )}
                </button>
              </div>
              {validationErrors.confirmPassword && (
                <p className='text-red-400 text-sm mt-1'>{validationErrors.confirmPassword}</p>
              )}
            </div>

            <div className='flex items-start'>
              <input
                type='checkbox'
                required
                className='mt-1 rounded border-gray-300 text-blue-600 focus:ring-blue-500'
              />
              <span className='ml-2 text-sm text-gray-300'>
                I agree to the{' '}
                <a href='#' className='text-blue-400 hover:text-blue-300'>
                  Terms of Service
                </a>{' '}
                and{' '}
                <a href='#' className='text-blue-400 hover:text-blue-300'>
                  Privacy Policy
                </a>
              </span>
            </div>

            <button
              type='submit'
              disabled={isLoading}
              className='w-full button-primary disabled:opacity-50 disabled:cursor-not-allowed'
            >
              {isLoading ? (
                <div className='flex items-center justify-center'>
                  <div className='loading-spinner mr-2'></div>
                  Creating Account...
                </div>
              ) : (
                'Create Account'
              )}
            </button>
          </form>

          <div className='mt-8 text-center'>
            <p className='text-gray-400'>
              Already have an account?{' '}
              <Link to='/login' className='text-blue-400 hover:text-blue-300 font-medium'>
                Sign in
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
