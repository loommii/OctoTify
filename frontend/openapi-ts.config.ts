import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  client: false,
  input: 'http://localhost:34123/openapi.json',
  output: 'apps/web-ele/src/api/generated',
  plugins: [
    '@hey-api/typescript',
    '@hey-api/schemas',
  ],
});
