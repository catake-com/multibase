import { fileURLToPath, URL } from "url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { quasar, transformAssetUrls } from "@quasar/vite-plugin";
import terminal from "vite-plugin-terminal";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    terminal({
      console: "terminal",
      output: ["terminal", "console"],
    }),

    vue({
      template: { transformAssetUrls },
    }),

    quasar({
      sassVariables: "src/quasar-variables.sass",
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `assets/[name].js`,
        chunkFileNames: `assets/[name].js`,
        assetFileNames: `assets/[name].[ext]`,
      },
      external: ["/@id/__x00__virtual:terminal/console"],
    },
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
