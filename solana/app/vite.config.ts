import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  // @solana/web3.js (v1, via wallet-adapter) references a `global`.
  define: { global: "globalThis" },
  resolve: {
    alias: {
      // Import the generated kit client straight from source (sibling folder,
      // mounted at /clients in the app container — see docker-compose.yml).
      "@counter-client": fileURLToPath(
        new URL("../clients/ts/src/generated", import.meta.url),
      ),
    },
  },
  // The generated client lives in a sibling folder (outside the app root), so
  // let Vite's dev server read beyond its root.
  server: { port: 5173, fs: { strict: false } },
});
