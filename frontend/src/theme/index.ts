// MonMetrics Theme Configuration
// Inspired by modern gaming/TCG platforms with dark, vibrant aesthetics

export const theme = {
  // Color palette inspired by gaming/fantasy aesthetics
  colors: {
    // Primary brand colors - gold/amber for premium feel
    primary: {
      50: '#fffbeb',
      100: '#fef3c7',
      200: '#fde68a',
      300: '#fcd34d',
      400: '#fbbf24',
      500: '#f59e0b', // Main gold
      600: '#d97706',
      700: '#b45309',
      800: '#92400e',
      900: '#78350f',
      950: '#451a03',
    },
    // Secondary - cyan/teal for accents
    secondary: {
      50: '#ecfeff',
      100: '#cffafe',
      200: '#a5f3fc',
      300: '#67e8f9',
      400: '#22d3ee',
      500: '#06b6d4', // Main cyan
      600: '#0891b2',
      700: '#0e7490',
      800: '#155e75',
      900: '#164e63',
      950: '#083344',
    },
    // Accent - purple/magenta for highlights
    accent: {
      50: '#faf5ff',
      100: '#f3e8ff',
      200: '#e9d5ff',
      300: '#d8b4fe',
      400: '#c084fc',
      500: '#a855f7', // Main purple
      600: '#9333ea',
      700: '#7e22ce',
      800: '#6b21a8',
      900: '#581c87',
      950: '#3b0764',
    },
    // Success - emerald green
    success: {
      50: '#ecfdf5',
      100: '#d1fae5',
      200: '#a7f3d0',
      300: '#6ee7b7',
      400: '#34d399',
      500: '#10b981',
      600: '#059669',
      700: '#047857',
      800: '#065f46',
      900: '#064e3b',
    },
    // Warning - orange
    warning: {
      50: '#fff7ed',
      100: '#ffedd5',
      200: '#fed7aa',
      300: '#fdba74',
      400: '#fb923c',
      500: '#f97316',
      600: '#ea580c',
      700: '#c2410c',
      800: '#9a3412',
      900: '#7c2d12',
    },
    // Error - red
    error: {
      50: '#fef2f2',
      100: '#fee2e2',
      200: '#fecaca',
      300: '#fca5a5',
      400: '#f87171',
      500: '#ef4444',
      600: '#dc2626',
      700: '#b91c1c',
      800: '#991b1b',
      900: '#7f1d1d',
    },
    // Neutral - for backgrounds and text
    dark: {
      50: '#f8fafc',
      100: '#f1f5f9',
      200: '#e2e8f0',
      300: '#cbd5e1',
      400: '#94a3b8',
      500: '#64748b',
      600: '#475569',
      700: '#334155',
      800: '#1e293b',
      900: '#0f172a', // Main dark background
      950: '#020617', // Deepest black
    },
  },

  // Gradients for visual interest
  gradients: {
    gold: 'linear-gradient(135deg, #fbbf24 0%, #f59e0b 50%, #d97706 100%)',
    cyan: 'linear-gradient(135deg, #22d3ee 0%, #06b6d4 50%, #0891b2 100%)',
    purple: 'linear-gradient(135deg, #c084fc 0%, #a855f7 50%, #9333ea 100%)',
    sunset: 'linear-gradient(135deg, #f59e0b 0%, #f97316 50%, #ef4444 100%)',
    ocean: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 50%, #7e22ce 100%)',
    cosmic: 'linear-gradient(135deg, #a855f7 0%, #ec4899 50%, #f59e0b 100%)',
    treasure: 'linear-gradient(135deg, #fbbf24 0%, #22d3ee 50%, #a855f7 100%)',
  },

  // Typography scale
  typography: {
    fontFamily: {
      display: '"Inter", "Helvetica Neue", Arial, sans-serif', // For headings
      body: '"Inter", system-ui, -apple-system, sans-serif', // For body text
      mono: '"Fira Code", "Courier New", monospace', // For code/data
    },
    fontSize: {
      xs: ['0.75rem', { lineHeight: '1rem' }],
      sm: ['0.875rem', { lineHeight: '1.25rem' }],
      base: ['1rem', { lineHeight: '1.5rem' }],
      lg: ['1.125rem', { lineHeight: '1.75rem' }],
      xl: ['1.25rem', { lineHeight: '1.75rem' }],
      '2xl': ['1.5rem', { lineHeight: '2rem' }],
      '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
      '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
      '5xl': ['3rem', { lineHeight: '1' }],
      '6xl': ['3.75rem', { lineHeight: '1' }],
      '7xl': ['4.5rem', { lineHeight: '1' }],
      '8xl': ['6rem', { lineHeight: '1' }],
      '9xl': ['8rem', { lineHeight: '1' }],
    },
  },

  // Spacing for consistent layouts
  spacing: {
    section: '5rem', // 80px
    container: '7rem', // 112px
  },

  // Border radius for consistent curves
  borderRadius: {
    card: '0.75rem', // 12px
    button: '0.5rem', // 8px
    input: '0.5rem', // 8px
    badge: '9999px', // pill shape
  },

  // Shadows for depth
  shadows: {
    glow: {
      gold: '0 0 20px rgba(251, 191, 36, 0.3), 0 0 40px rgba(245, 158, 11, 0.2)',
      cyan: '0 0 20px rgba(34, 211, 238, 0.3), 0 0 40px rgba(6, 182, 212, 0.2)',
      purple: '0 0 20px rgba(192, 132, 252, 0.3), 0 0 40px rgba(168, 85, 247, 0.2)',
    },
    card: '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
    cardHover: '0 20px 25px -5px rgba(0, 0, 0, 0.4), 0 10px 10px -5px rgba(0, 0, 0, 0.2)',
  },

  // Animation durations
  animation: {
    fast: '150ms',
    base: '300ms',
    slow: '500ms',
  },

  // Z-index layers
  zIndex: {
    base: 0,
    dropdown: 1000,
    sticky: 1020,
    fixed: 1030,
    modalBackdrop: 1040,
    modal: 1050,
    popover: 1060,
    tooltip: 1070,
  },
} as const

// Utility function to get theme values
export const getThemeValue = (path: string) => {
  const keys = path.split('.')
  let value: any = theme

  for (const key of keys) {
    value = value[key]
    if (value === undefined) break
  }

  return value
}

export type Theme = typeof theme
