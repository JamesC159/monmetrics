import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// inotify does not fire for Windows-mounted paths under WSL, so fall back to polling there
const usePolling = process.env.VITE_USE_POLLING === 'true' || __dirname.startsWith('/mnt/')

export default defineConfig({
  plugins: [react()],
  server: {
    port: Number(process.env.PORT) || 3000,
    strictPort: true,
    watch: {
      usePolling,
      interval: 300,
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      input: {
        main: path.resolve(__dirname, 'index.html'),
      },
    },
  },
  ssr: {
    noExternal: ['react-router-dom'],
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
  },
})
