// @ts-check
// Baseline ESLint flat config for Angular + TypeScript.
// Install: ng add angular-eslint   (or: npm i -D eslint @eslint/js typescript-eslint angular-eslint)
const eslint = require("@eslint/js");
const { defineConfig } = require("eslint/config");
const tseslint = require("typescript-eslint");
const angular = require("angular-eslint");

module.exports = defineConfig([
  {
    ignores: ["dist/", ".angular/", "coverage/", "node_modules/"],
  },

  // ---------- TypeScript files ----------
  {
    files: ["**/*.ts"],
    extends: [
      eslint.configs.recommended,
      tseslint.configs.recommendedTypeChecked,
      tseslint.configs.stylisticTypeChecked,
      angular.configs.tsRecommended,
    ],
    languageOptions: {
      parserOptions: {
        projectService: true, // type-aware linting via your tsconfig
        tsconfigRootDir: __dirname,
      },
    },
    processor: angular.processInlineTemplates, // lint inline templates too
    rules: {
      // Angular conventions (change "app" to your project prefix)
      "@angular-eslint/component-selector": [
        "error",
        { type: "element", prefix: "app", style: "kebab-case" },
      ],
      "@angular-eslint/directive-selector": [
        "error",
        { type: "attribute", prefix: "app", style: "camelCase" },
      ],
      "@angular-eslint/prefer-on-push-component-change-detection": "warn",
      "@angular-eslint/no-output-on-prefix": "error",
      "@angular-eslint/use-lifecycle-interface": "error",

      // TypeScript hygiene
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/consistent-type-imports": "error",
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/explicit-member-accessibility": [
        "warn",
        { accessibility: "no-public" },
      ],

      // General
      eqeqeq: ["error", "smart"],
      "no-console": ["warn", { allow: ["warn", "error"] }],
      "prefer-const": "error",
    },
  },

  // ---------- Angular HTML templates ----------
  {
    files: ["**/*.html"],
    extends: [
      angular.configs.templateRecommended,
      angular.configs.templateAccessibility,
    ],
    rules: {
      "@angular-eslint/template/prefer-control-flow": "error", // @if/@for over *ngIf/*ngFor
      "@angular-eslint/template/prefer-self-closing-tags": "warn",
      "@angular-eslint/template/no-any": "warn",
      "@angular-eslint/template/eqeqeq": "error",
    },
  },
]);