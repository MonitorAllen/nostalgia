import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'

const normalizeModuleId = (id: string) => id.replaceAll('\\', '/')

const manualChunks = (id: string) => {
  const normalizedId = normalizeModuleId(id)

  if (
    normalizedId.includes('/node_modules/ckeditor5/') ||
    normalizedId.includes('/node_modules/@ckeditor/ckeditor5-vue/')
  ) {
    return 'ckeditor'
  }

  if (
    normalizedId.includes('/node_modules/prismjs/') ||
    normalizedId.includes('/node_modules/dompurify/') ||
    normalizedId.includes('/src/util/sanitizeHtml.ts')
  ) {
    return 'content-rendering'
  }

  if (
    normalizedId.includes('/src/views/admin/AdminArticleEditorView.vue') ||
    normalizedId.includes('/src/admin/editor/')
  ) {
    return 'admin-editor'
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 3000,
    strictPort: true,
    proxy: {
      '/api': 'http://localhost:8080',
      '/v1': {
        target: 'http://localhost:9091',
        changeOrigin: true,
      },
      '/resources': 'http://localhost:8080'
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks
      }
    }
  }
})
