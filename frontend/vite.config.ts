import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// In development, proxy the API and signaling WebSocket to the Go backend so the
// app can use same-origin relative URLs (/api, /ws) in every environment.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
      },
    },
  },
});
