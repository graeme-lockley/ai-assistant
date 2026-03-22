import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const root = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  test: {
    include: [
      "packages/core/src/**/*.test.ts",
      "packages/consolidate/src/**/*.test.ts",
    ],
    environment: "node",
  },
  resolve: {
    alias: {
      "@ai-assistant/core": path.join(root, "packages/core/src/index.ts"),
    },
  },
});
