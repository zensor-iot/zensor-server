import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const base = '/ui/'
const designSystem = fileURLToPath(new URL('../docs/design-system', import.meta.url))
const logosDir = join(designSystem, 'logos')

const logoTypes = { '.svg': 'image/svg+xml', '.png': 'image/png', '.ico': 'image/x-icon' }

function designSystemLogos() {
  return {
    name: 'zensor-design-system-logos',
    configureServer(server) {
      server.middlewares.use(`${base}logos/`, (req, res, next) => {
        const file = decodeURIComponent(req.url.split('?')[0]).replace(/^\/+/, '')
        if (!readdirSync(logosDir).includes(file)) return next()
        res.setHeader('Content-Type', logoTypes[file.slice(file.lastIndexOf('.'))] || 'application/octet-stream')
        res.end(readFileSync(join(logosDir, file)))
      })
    },
    generateBundle() {
      for (const file of readdirSync(logosDir)) {
        this.emitFile({ type: 'asset', fileName: `logos/${file}`, source: readFileSync(join(logosDir, file)) })
      }
    },
  }
}

export default defineConfig({
  plugins: [react(), designSystemLogos()],
  base,
  resolve: {
    alias: {
      '@design-system': designSystem,
    },
  },
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
    fs: {
      allow: ['.', designSystem],
    },
    proxy: {
      '/v1': 'http://localhost:3000',
      '/auth': 'http://localhost:3000',
      '/healthz': 'http://localhost:3000',
      '/ws': {
        target: 'ws://localhost:3000',
        ws: true,
      },
    },
  },
})
