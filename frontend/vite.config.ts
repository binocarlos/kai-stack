import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [
    react({
      jsxImportSource: '@emotion/react',
      babel: {
        plugins: ['babel-plugin-react-compiler', '@emotion/babel-plugin']
      }
    })
  ],
  resolve: {
    extensions: ['.js', '.jsx', '.ts', '.tsx'],
    alias: {
      'react/compiler-runtime': 'react-compiler-runtime',
    },
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: undefined
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 8080,
    allowedHosts: [
      'localhost',
      '.ngrok-free.app',
      '.ngrok.io',
      '.ngrok.app'
    ]
  },
  publicDir: 'assets',
  optimizeDeps: {
    include: ['@emotion/react', '@emotion/styled', '@mui/material', '@mui/x-tree-view']
  }
})