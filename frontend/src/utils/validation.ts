// OWASP-compliant validation utilities

export function validateEmail(email: string): boolean {
  // RFC 5322 simplified email regex
  const emailRegex =
    /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/
  return emailRegex.test(email) && email.length <= 254
}

export function validatePassword(password: string): { isValid: boolean; errors: string[] } {
  const errors: string[] = []

  // OWASP minimum length
  if (password.length < 8) {
    errors.push('Password must be at least 8 characters long')
  }

  // Maximum length to prevent DoS
  if (password.length > 128) {
    errors.push('Password must not exceed 128 characters')
  }

  // Require at least one lowercase letter
  if (!/(?=.*[a-z])/.test(password)) {
    errors.push('Password must contain at least one lowercase letter')
  }

  // Require at least one uppercase letter
  if (!/(?=.*[A-Z])/.test(password)) {
    errors.push('Password must contain at least one uppercase letter')
  }

  // Require at least one number
  if (!/(?=.*\d)/.test(password)) {
    errors.push('Password must contain at least one number')
  }

  // Require at least one special character
  if (!/(?=.*[!@#$%^&*(),.?":{}|<>_\-+=\[\]\\/'`;~])/.test(password)) {
    errors.push('Password must contain at least one special character')
  }

  return {
    isValid: errors.length === 0,
    errors,
  }
}

export function validateName(name: string): { isValid: boolean; error?: string } {
  const trimmed = name.trim()

  if (trimmed.length < 2) {
    return { isValid: false, error: 'Name must be at least 2 characters' }
  }

  if (trimmed.length > 50) {
    return { isValid: false, error: 'Name must not exceed 50 characters' }
  }

  // Check for at least one letter
  if (!/[a-zA-Z]/.test(trimmed)) {
    return { isValid: false, error: 'Name must contain at least one letter' }
  }

  return { isValid: true }
}

export function sanitizeName(name: string): string {
  // Trim whitespace
  let sanitized = name.trim()

  // Remove control characters
  sanitized = sanitized.replace(/[\x00-\x1F\x7F-\x9F]/g, '')

  // Normalize multiple spaces to single space
  sanitized = sanitized.replace(/\s+/g, ' ')

  return sanitized
}

export function normalizeEmail(email: string): string {
  return email.trim().toLowerCase()
}

export function getPasswordStrength(password: string): {
  score: number // 0-4
  label: string
  color: string
} {
  let score = 0

  // Length scoring
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  if (password.length >= 16) score++

  // Complexity scoring
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++
  if (/\d/.test(password)) score++
  if (/[!@#$%^&*(),.?":{}|<>_\-+=\[\]\\/'`;~]/.test(password)) score++

  // Normalize to 0-4 scale
  const normalizedScore = Math.min(4, Math.floor(score / 1.5))

  const labels = ['Very Weak', 'Weak', 'Fair', 'Strong', 'Very Strong']
  const colors = ['#ef4444', '#f59e0b', '#eab308', '#22c55e', '#10b981']

  return {
    score: normalizedScore,
    label: labels[normalizedScore],
    color: colors[normalizedScore],
  }
}
