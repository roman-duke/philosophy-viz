import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // Dev-only reverse proxy: the browser talks only to :5173, and Vite
    // forwards API/SSE calls to the Go backend on :8080. Keeps everything
    // same-origin (no CORS) and lets the React code use relative URLs that
    // also work in prod, where Go serves both the static build and the API.
    proxy: {
      '/titles': 'http://localhost:8080',
      // '/paths' will go here too when we wire up the SSE crawl feed.
    },
  },
})
