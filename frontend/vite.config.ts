import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', '@vueuse/core', { '@unhead/vue': ['useHead'] }],
      dirs: ['src/composables']
    }),
    Components({
      dirs: ['src/components', 'src/layouts']
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 8080
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@use "@/design/index.scss" as *;',
        silenceDeprecations: ['import', 'global-builtin']
      }
    }
  }
})
