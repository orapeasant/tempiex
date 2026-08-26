import { defineConfig } from 'tsup';

export default defineConfig({
  entry: { tempiex: 'src/main.tsx' },
  outDir: 'dist',
  format: ['esm'],
  target: 'node22',
  outExtension: () => ({ js: '.js' }),
  banner: {
    js: '#!/usr/bin/env node',
  },
  bundle: true,
  splitting: false,
  clean: true,
  treeshake: true,
  external: [],
  esbuildOptions(options) {
    options.conditions = ['import', 'default'];
  },
});
