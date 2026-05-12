import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: 'http://localhost:34123/openapi.json',
  output: 'src/api/generated',
  plugins: [
    '@hey-api/client-axios',
    '@hey-api/typescript',
    {
      name: '@hey-api/sdk',
      operations: {
        strategy: 'single',
      },
    },
  ],
})
