/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Primary - Gold/Amber for premium gaming feel
        primary: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },
        // Secondary - Cyan/Teal for accents
        secondary: {
          50: '#ecfeff',
          100: '#cffafe',
          200: '#a5f3fc',
          300: '#67e8f9',
          400: '#22d3ee',
          500: '#06b6d4',
          600: '#0891b2',
          700: '#0e7490',
          800: '#155e75',
          900: '#164e63',
          950: '#083344',
        },
        // Accent - Purple/Magenta for highlights
        accent: {
          50: '#faf5ff',
          100: '#f3e8ff',
          200: '#e9d5ff',
          300: '#d8b4fe',
          400: '#c084fc',
          500: '#a855f7',
          600: '#9333ea',
          700: '#7e22ce',
          800: '#6b21a8',
          900: '#581c87',
          950: '#3b0764',
        },
        // Dark backgrounds
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
          900: '#0f172a',
          950: '#020617',
        },
        // Success
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
        // Warning
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
        // Error
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
      },
      fontFamily: {
        display: ['"Inter"', '"Helvetica Neue"', 'Arial', 'sans-serif'],
        body: ['"Inter"', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"Fira Code"', '"Courier New"', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'fade-in-up': 'fadeInUp 0.6s ease-out',
        'slide-up': 'slideUp 0.5s ease-out',
        'slide-down': 'slideDown 0.4s ease-out',
        'pulse-glow': 'pulseGlow 2s ease-in-out infinite',
        'pulse-glow-gold': 'pulseGlowGold 2s ease-in-out infinite',
        'pulse-glow-cyan': 'pulseGlowCyan 2s ease-in-out infinite',
        float: 'float 3s ease-in-out infinite',
        shimmer: 'shimmer 2s linear infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeInUp: {
          '0%': { opacity: '0', transform: 'translateY(20px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        pulseGlow: {
          '0%, 100%': { boxShadow: '0 0 20px rgba(59, 130, 246, 0.5)' },
          '50%': { boxShadow: '0 0 30px rgba(59, 130, 246, 0.8)' },
        },
        pulseGlowGold: {
          '0%, 100%': {
            boxShadow: '0 0 20px rgba(251, 191, 36, 0.3), 0 0 40px rgba(245, 158, 11, 0.2)',
          },
          '50%': {
            boxShadow: '0 0 30px rgba(251, 191, 36, 0.5), 0 0 60px rgba(245, 158, 11, 0.3)',
          },
        },
        pulseGlowCyan: {
          '0%, 100%': {
            boxShadow: '0 0 20px rgba(34, 211, 238, 0.3), 0 0 40px rgba(6, 182, 212, 0.2)',
          },
          '50%': { boxShadow: '0 0 30px rgba(34, 211, 238, 0.5), 0 0 60px rgba(6, 182, 212, 0.3)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-10px)' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
      },
      backgroundImage: {
        'gradient-gold': 'linear-gradient(135deg, #fbbf24 0%, #f59e0b 50%, #d97706 100%)',
        'gradient-cyan': 'linear-gradient(135deg, #22d3ee 0%, #06b6d4 50%, #0891b2 100%)',
        'gradient-purple': 'linear-gradient(135deg, #c084fc 0%, #a855f7 50%, #9333ea 100%)',
        'gradient-sunset': 'linear-gradient(135deg, #f59e0b 0%, #f97316 50%, #ef4444 100%)',
        'gradient-ocean': 'linear-gradient(135deg, #06b6d4 0%, #0891b2 50%, #7e22ce 100%)',
        'gradient-cosmic': 'linear-gradient(135deg, #a855f7 0%, #ec4899 50%, #f59e0b 100%)',
        'gradient-treasure': 'linear-gradient(135deg, #fbbf24 0%, #22d3ee 50%, #a855f7 100%)',
        'shimmer-gradient':
          'linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent)',
      },
      boxShadow: {
        'glow-gold': '0 0 20px rgba(251, 191, 36, 0.3), 0 0 40px rgba(245, 158, 11, 0.2)',
        'glow-cyan': '0 0 20px rgba(34, 211, 238, 0.3), 0 0 40px rgba(6, 182, 212, 0.2)',
        'glow-purple': '0 0 20px rgba(192, 132, 252, 0.3), 0 0 40px rgba(168, 85, 247, 0.2)',
        card: '0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
        'card-hover': '0 20px 25px -5px rgba(0, 0, 0, 0.4), 0 10px 10px -5px rgba(0, 0, 0, 0.2)',
      },
    },
  },
  plugins: [],
}
