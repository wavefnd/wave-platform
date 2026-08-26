import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    lib: {
      entry: {
        core: fileURLToPath(new URL('./core/index.ts', import.meta.url)),
        dom: fileURLToPath(new URL('./dom/index.ts', import.meta.url)),
        vue: fileURLToPath(new URL('./vue/index.ts', import.meta.url)),
        style: fileURLToPath(new URL('./style.ts', import.meta.url)),
      },
      formats: ['es'],
      fileName: (_format, entryName) => `${entryName}.js`,
      cssFileName: 'style',
    },
    rollupOptions: { external: ['vue'] },
    sourcemap: true,
  },
})
