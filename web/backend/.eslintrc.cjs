// .eslintrc.js
module.exports = {
  root: true,
  env: {
    browser: true,
    node: true,
    es6: true,
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 2020,
    sourceType: 'module',
    jsxPragma: 'React',
    ecmaFeatures: {
      jsx: true,
    },
  },
  extends: [
    'plugin:vue/vue3-recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:prettier/recommended', // 保持最后，自动关闭冲突的格式规则
  ],
  rules: {
    // ESLint 核心规则（只保留“逻辑/规范”类，不保留格式化类）
    'no-var': 'error',
    'prefer-const': 'off',
    'no-use-before-define': 'off',

    // 注意：no-multiple-empty-lines 已移除，由 Prettier 负责

    // TypeScript 规则
    '@typescript-eslint/no-unused-vars': 'error',
    '@typescript-eslint/no-empty-function': 'error',
    '@typescript-eslint/prefer-ts-expect-error': 'error',
    '@typescript-eslint/ban-ts-comment': 'error',
    '@typescript-eslint/no-inferrable-types': 'off',
    '@typescript-eslint/no-namespace': 'off',
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/ban-types': 'off',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/no-non-null-assertion': 'off',

    // Vue 规则（只保留逻辑相关，移除格式相关如 attributes-order、html-closing-bracket-newline）
    'vue/script-setup-uses-vars': 'error',
    'vue/v-slot-style': 'error',
    'vue/no-mutating-props': 'error',
    // "vue/custom-event-name-casing": "error",  // 无冲突，但如果你想保留可以取消注释
    'vue/attribute-hyphenation': 'error',
    'vue/attributes-order': 'off',
    'vue/no-v-html': 'off',
    'vue/require-default-prop': 'off',
    'vue/multi-word-component-names': 'off',
    'vue/no-setup-props-destructure': 'off',
  },
}
