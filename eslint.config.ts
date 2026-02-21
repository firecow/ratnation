import {defineConfig} from "eslint/config";
import eslint from "@eslint/js";
import stylistic from "@stylistic/eslint-plugin";
import tseslint from "typescript-eslint";

export default defineConfig(
    {
        ignores: [
            "node_modules",
            "**/*.js",
            "**/*.d.ts",
            "eslint.config.ts",
        ],
    },

    eslint.configs.recommended,
    stylistic.configs.all,

    {
        files: ["**/*.{ts,tsx}"],
        extends: [
            ...tseslint.configs.strictTypeChecked,
            ...tseslint.configs.stylisticTypeChecked,
        ],
        languageOptions: {
            parserOptions: {
                projectService: true,
            },
        },
        rules: {
            "@stylistic/padded-blocks": ["error", {blocks: "never", classes: "always", switches: "never"}],
            "@stylistic/comma-dangle": ["error", "always-multiline"],
            "@stylistic/quote-props": ["error", "consistent"],
            "@stylistic/array-element-newline": ["error", "consistent"],
            "@stylistic/object-property-newline": ["error", {allowAllPropertiesOnSameLine: true}],
            "@stylistic/multiline-ternary": ["error", "always-multiline"],
            "@stylistic/newline-per-chained-call": ["error", {ignoreChainWithDepth: 3}],
            "@typescript-eslint/restrict-template-expressions": ["error", {allow: ["Error"], allowNumber: true}],
            "@typescript-eslint/no-misused-promises": ["error", {checksVoidReturn: false}],
            "@typescript-eslint/promise-function-async": "error",

            "@stylistic/function-call-argument-newline": "off",
        },
    },
);
