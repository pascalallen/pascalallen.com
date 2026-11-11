import js from '@eslint/js';
import prettier from 'eslint-plugin-prettier/recommended';
import tseslint from 'typescript-eslint';

const domRule = 'Build DOM with createElement and textContent.';

export default tseslint.config(
  { ignores: ['node_modules/', '.test-build/'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  prettier,
  {
    files: ['**/*.ts'],
    rules: {
      '@typescript-eslint/no-unused-vars': 'error',
      'no-restricted-properties': [
        'error',
        { property: 'innerHTML', message: domRule },
        { property: 'outerHTML', message: domRule },
        { property: 'insertAdjacentHTML', message: domRule },
        { object: 'document', property: 'write', message: domRule },
        { object: 'document', property: 'writeln', message: domRule }
      ]
    }
  },
  {
    files: ['**/*.js', '**/*.mjs'],
    languageOptions: {
      globals: {
        require: 'readonly',
        module: 'writable',
        __dirname: 'readonly',
        process: 'readonly',
        console: 'readonly',
        URL: 'readonly'
      }
    },
    rules: { '@typescript-eslint/no-require-imports': 'off' }
  }
);
