import js from "@eslint/js";
import globals from "globals";
import { defineConfig } from "eslint/config";
import jsdoc from "eslint-plugin-jsdoc";
import pluginImport from "eslint-plugin-import";

export default defineConfig([
  {
    files: ["**/*.{js,mjs,cjs}"],
    plugins: {
      jsdoc,
      import: pluginImport,
    },
    languageOptions: {
      globals: globals.browser,
      ecmaVersion: 2020,
      sourceType: "module",
    },
    rules: {
      ...js.configs.recommended.rules,
      ...jsdoc.configs["flat/recommended"].rules,

      // JSDoc rules
      "jsdoc/check-types": "error",
      "jsdoc/require-param-type": "error",
      "jsdoc/require-returns-type": "error",
      "jsdoc/check-param-names": "error",
      "jsdoc/require-param": "error",
      "jsdoc/require-returns": "error",
      "jsdoc/valid-types": "error",

      // Additional rules
      "default-param-last": "error",
      "prefer-spread": "error",
      "import/no-mutable-exports": "error",
      "import/prefer-default-export": "error",
      "new-cap": "error",
      "no-duplicate-imports": "warn",
      "no-iterator": "error",
      "no-loop-func": "error",
      "no-new-func": "error",
      "no-param-reassign": "error",
      "no-undef": "error",
      "no-unused-vars": "error",
      "prefer-const": "error",
      "prefer-rest-params": "error",
    },
  },
]);
