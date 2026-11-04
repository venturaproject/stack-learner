import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

const backendTarget = process.env.VITE_BACKEND_URL ?? 'http://localhost:8000'

export default defineConfig({
  plugins: [tailwindcss(), react()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
      '@/config': path.resolve(import.meta.dirname, './src/config'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': { target: backendTarget, changeOrigin: true },
    },
  },
  build: {
    outDir: path.resolve(import.meta.dirname, './dist'),
    emptyOutDir: true,
  },
})
