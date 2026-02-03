import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
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
      },
    },
  },
  build: {
    // Code splitting for better caching
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'antd-vendor': ['antd', '@ant-design/icons'],
          'query-vendor': ['@tanstack/react-query'],

          // Feature chunks
          'k8s': [
            './src/pages/ClusterListPage.tsx',
            './src/pages/ClusterDetailPage.tsx',
            './src/pages/ClusterMonitoringPage.tsx',
            './src/pages/WorkloadListPage.tsx',
            './src/pages/PodDetailPage.tsx',
          ],
          'host': [
            './src/pages/HostListPage.tsx',
            './src/pages/HostDetailPage.tsx',
          ],
          'observability': [
            './src/pages/OtelCollectorPage.tsx',
            './src/pages/PrometheusDataSourcePage.tsx',
            './src/pages/PrometheusAlertRulesPage.tsx',
            './src/pages/GrafanaInstancesPage.tsx',
          ],
          'ai': [
            './src/pages/AnomalyDetectionPage.tsx',
          ],
          'security': [
            './src/pages/UserManagementPage.tsx',
            './src/pages/RolesPage.tsx',
          ],
        },
      },
    },
    // Chunk size warning limit
    chunkSizeWarningLimit: 1000,
    // Target modern browsers
    target: 'es2015',
    // Minification
    minify: 'esbuild',
    // Source maps
    sourcemap: false,
    // CSS code splitting
    cssCodeSplit: true,
  },
  // Optimize dependency pre-bundling
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'antd',
      '@ant-design/icons',
      '@tanstack/react-query',
      'axios',
      'dayjs',
    ],
  },
  // CSS optimization
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
        modifyVars: {
          // Theme customization
        },
      },
    },
  },
  // Define global constants
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
    __API_URL__: JSON.stringify(process.env.VITE_API_BASE_URL || 'http://localhost:8080'),
  },
})
