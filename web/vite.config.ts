import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return undefined
          }

          if (id.includes('@ffmpeg/ffmpeg') || id.includes('@ffmpeg/util')) {
            return 'vendor-ffmpeg'
          }

          if (id.includes('@element-plus/icons-vue')) {
            return 'vendor-ep-icons'
          }

          if (id.includes('element-plus') || id.includes('@element-plus')) {
            return 'vendor-element-plus'
          }

          if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router') || id.includes('vue-i18n')) {
            return 'vendor-vue'
          }

          if (id.includes('lodash-es') || id.includes('dayjs') || id.includes('cropperjs') || id.includes('lucide-vue-next')) {
            return 'vendor-utils'
          }

          return 'vendor-misc'
        }
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3012,
    proxy: {
      '/api': {
        target: 'http://localhost:5678',
        changeOrigin: true
      },
      '/static': {
        target: 'http://localhost:5678',
        changeOrigin: true
      }
    }
  }
})
