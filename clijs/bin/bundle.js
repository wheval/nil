#!/usr/bin/env node
import fs from 'fs'
import { fileURLToPath } from 'url'
import { join, dirname } from 'path'
import { build } from 'esbuild'
import { execaCommandSync } from 'execa'
import { inject } from 'postject'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// esbuild
const result = await build({
  entryPoints: [join(__dirname, '../src/sea.ts')],
  outfile: join(__dirname, '../dist/cli.cjs'),
  bundle: true,
  format: 'cjs',
  platform: 'node',
  loader: {'.node': 'copy'},
  inject: [join(__dirname, './cjs-shims.js')],
  splitting: false,
  treeShaking: true,
  metafile: true,
  minify: true,
})
console.log('✅  esbuild done')

fs.writeFileSync(
  join(__dirname, '../dist/meta.json'),
  JSON.stringify(result.metafile)
);

// copy node executable
const nodePath = process.argv[0]
const binPath = join(__dirname, '../dist/clijs')
fs.copyFileSync(nodePath, binPath)
fs.chmodSync(binPath, 0o755)

// prepare sea-prep.blob
execaCommandSync(`${nodePath} --experimental-sea-config ./sea-config.json`, {
  stdio: 'inherit',
  cwd: join(__dirname, '..'),
})

// inject sea-prep.blob to executable
await inject(
  binPath,
  'NODE_SEA_BLOB',
  fs.readFileSync(join(__dirname, '../dist/sea-prep.blob')),
  {
    sentinelFuse: 'NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2',
  })

console.log('🎉  Done!')
