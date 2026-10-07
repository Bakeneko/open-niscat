import { mkdirSync, writeFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'

// Keeps web/dist/.gitkeep after Vite empties the folder: the Go embed needs the directory to exist.
function keepDistPlaceholder(): Plugin {
  return {
    name: 'keep-dist-placeholder',
    apply: 'build',
    closeBundle() {
      mkdirSync(fileURLToPath(new URL('./dist', import.meta.url)), { recursive: true })
      writeFileSync(fileURLToPath(new URL('./dist/.gitkeep', import.meta.url)), '')
    },
  }
}

export default defineConfig({
  plugins: [vue(), vuetify({ autoImport: true }), keepDistPlaceholder()],
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  // Vuetify makes one large chunk; the default 500 kB warning is only noise for a local tool.
  build: { chunkSizeWarningLimit: 2000 },
  server: { proxy: { '/api': 'http://127.0.0.1:8080', '/files': 'http://127.0.0.1:8080' } },
})
