import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import fsdPlugin from "@conarti/eslint-plugin-feature-sliced/dist/index.js";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    plugins: {
      "feature-sliced": fsdPlugin,
    },
    rules: {
      "feature-sliced/absolute-public-access": "error",
      "feature-sliced/layers-slices": "error",
      "feature-sliced/public-api": "error",
    },
  },
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;