import tailwindcss from "@tailwindcss/vite";
import mdx from "fumadocs-mdx/vite";
import { defineConfig } from "waku/config";

export default defineConfig({
  basePath: "/docs/",
  vite: {
    resolve: {
      tsconfigPaths: true,
      external: ["@takumi-rs/image-response"],
      dedupe: ["waku"],
    },

    plugins: [tailwindcss(), mdx()],
  },
});
