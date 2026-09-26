// MonMetrics Theme Configuration
// Neutral graphite palette for a focused market-analysis experience

export const theme = {
  // Color palette for the graphite visual system
  colors: {
    // Primary brand color
    primary: {
      50: '#fafafa',
      100: '#f4f4f5',
      200: '#e4e4e7',
      300: '#d4d4d8',
      400: '#a1a1aa',
      500: '#71717a',
      600: '#52525b',
      700: '#3f3f46',
      800: '#27272a',
      900: '#18181b',
      950: '#09090b',
    },
    // Secondary supporting contrast
    secondary: {
      50: '#f8fafc',
      100: '#f1f5f9',
      200: '#e2e8f0',
      300: '#cbd5e1',
      400: '#94a3b8',
      500: '#64748b',
      600: '#475569',
      700: '#334155',
      800: '#1e293b',
      900: '#0f172a',
      950: '#020617',
    },
    // Accent selected state
    accent: {
      50: '#ffffff',
      100: '#fafafa',
      200: '#f4f4f5',
      300: '#e4e4e7',
      400: '#d4d4d8',
      500: '#a1a1aa',
      600: '#71717a',
      700: '#52525b',
      800: '#3f3f46',
      900: '#27272a',
      950: '#18181b',
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
      gold: '0 0 20px rgba(228, 228, 231, 0.24), 0 0 40px rgba(161, 161, 170, 0.16)',
      cyan: '0 0 20px rgba(203, 213, 225, 0.22), 0 0 40px rgba(100, 116, 139, 0.16)',
      neutral: '0 0 20px rgba(244, 244, 245, 0.24), 0 0 40px rgba(113, 113, 122, 0.16)',
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
