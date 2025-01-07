/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { mergeConfig } from 'vite';
import { defineConfig as defineVitestConfig } from 'vitest/config';

const viteConfig = defineConfig({
  plugins: [react()],
  publicDir: 'public',
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      'weather-icons-npm': path.resolve(__dirname, 'node_modules/weather-icons-npm'),
    },
  },
  optimizeDeps: {
    include: ['weather-icons-npm'],
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern-compiler' // or "modern"
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo: { names?: string[] }) => {
          if (assetInfo.names) {
            if (assetInfo.names.some(name => name.match(/\.(woff|woff2|eot|ttf|svg)$/))) {
              return 'assets/fonts/[name]-[hash][extname]';
            }
          }
          return 'assets/[name]-[hash][extname]';
        },
      },
    },
  },
  server: {
    port: 3000,
    host: '0.0.0.0',
    watch: {
      usePolling: true,
    },
  },
});

const vitestConfig = defineVitestConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: true,
  },
});

export default mergeConfig(viteConfig, vitestConfig);