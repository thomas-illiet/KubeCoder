import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    // Swagger UI is intentionally isolated in a lazy-loaded documentation chunk.
    chunkSizeWarningLimit: 1800,
  },
})
