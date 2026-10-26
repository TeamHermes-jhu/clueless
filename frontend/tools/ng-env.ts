// TODO: Replace with Eric's ENV additions.

import { spawnSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { createRequire } from 'node:module';

// Keys from .env that are exposed to the app as global constants of the same name.
const DEFINED_KEYS = ['PRIMENG_LICENSE'];

if (existsSync('.env')) {
  process.loadEnvFile('.env');
} else {
  console.warn('ng-env: no .env file found; copy .env.example to .env and fill it in.');
}

const defines = DEFINED_KEYS.flatMap((key) => {
  const value = process.env[key];
  if (value === undefined) {
    console.warn(`ng-env: ${key} is not set.`);
    return [];
  }
  return ['--define', `${key}=${JSON.stringify(value)}`];
});

const ng = createRequire(import.meta.url).resolve('@angular/cli/bin/ng.js');
const result = spawnSync(process.execPath, [ng, ...process.argv.slice(2), ...defines], {
  stdio: 'inherit',
});
process.exit(result.status ?? 1);
