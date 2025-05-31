// vite.config.ts
import { defineConfig } from "vite"
import electron from "vite-plugin-electron"
import react from "@vitejs/plugin-react"
import path from "path"

export default defineConfig({
  plugins: [
    react(),
    electron([
      {
        // main.ts
        entry: "electron/main.ts",
        vite: {
          build: {
            outDir: "dist-electron",
            sourcemap: true,
            minify: false,
            rollupOptions: {
              external: ["electron"]
            }
          }
        }
      },
      {
        // preload.ts
        entry: "electron/preload.ts",
        vite: {
          build: {
            outDir: "dist-electron",
            sourcemap: true,
            rollupOptions: {
              external: ["electron"]
            }
          }
        }
      },
      {
        // solutionPopupPreload.js
        entry: "electron/solutionPopupPreload.js",
        vite: {
          build: {
            outDir: "dist-electron",
            sourcemap: true, // Optional: sourcemap for preload
            minify: false, // Optional: minify preload
            rollupOptions: {
              external: ["electron"]
            }
          }
        }
      }
    ])
  ],
  base: process.env.NODE_ENV === "production" ? "./" : "/",
  server: {
    port: 54321,
    strictPort: true,
    watch: {
      usePolling: true
    }
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      input: {
        main: path.resolve(__dirname, "renderer/public/index.html"), // Assuming main index.html is here
        solutionPopup: path.resolve(__dirname, "renderer/public/solutionPopup.html")
      }
    }
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src")
    }
  }
})
