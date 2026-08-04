import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const base = '/ui/'

export default defineConfig({
  plugins: [react()],
  base,
  build: {
    outDir: '../internal/infra/httpserver/web/dist',
    emptyOutDir: true,
  },
  test: {
    env: {
      BASE_URL: base,
    },
  },
  server: {
    proxy: {
      '/v1': 'http://localhost:3000',
      '/ws': {
        target: 'ws://localhost:3000',
        ws: true,
      },
    },
  },
})
