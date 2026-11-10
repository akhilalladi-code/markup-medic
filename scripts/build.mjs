import { cp, mkdir, rm, access } from 'node:fs/promises';
import { constants } from 'node:fs';
import { spawnSync, execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';

const root = fileURLToPath(new URL('../', import.meta.url));
const dist = new URL('../dist/', import.meta.url);
await rm(dist, { recursive: true, force: true });
await mkdir(dist, { recursive: true });

for (const path of ['index.html', 'styles.css', 'src', 'assets']) {
  await cp(new URL(`../${path}`, import.meta.url), new URL(`../dist/${path}`, import.meta.url), { recursive: true });
}

const wasmOutput = fileURLToPath(new URL('../dist/audit.wasm', import.meta.url));
const wasm = spawnSync('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', wasmOutput, './cmd/wasm'], {
  cwd: root,
  env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
  stdio: 'inherit'
});
if (wasm.status !== 0) process.exit(wasm.status ?? 1);

const goroot = execFileSync('go', ['env', 'GOROOT'], { encoding: 'utf8' }).trim();
await cp(join(goroot, 'misc', 'wasm', 'wasm_exec.js'), new URL('../dist/wasm_exec.js', import.meta.url));

for (const path of ['index.html', 'styles.css', 'src/app.js', 'assets/favicon.svg', 'audit.wasm', 'wasm_exec.js']) {
  await access(new URL(`../dist/${path}`, import.meta.url), constants.R_OK);
}
console.log('Go WebAssembly production build created in dist/.');
