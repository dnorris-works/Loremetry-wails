import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import {fileURLToPath, URL} from 'node:url'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Relative URLs so the shipped WebKit asset server can load JS/CSS.
  base: './',
  // Keep Wails / Go logs visible in the same terminal.
  clearScreen: false,
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: 'localhost',
    port: 5173,
    strictPort: true,
    hmr: {
      host: 'localhost',
      protocol: 'ws',
    },
  },
})
