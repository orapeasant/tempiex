import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { writeFileSync, mkdirSync, rmSync, existsSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import TOML from '@iarna/toml';
import { loadConfig, saveConfig, resolveConfig } from '../../src/config/loader.js';

const testConfigDir = join(homedir(), '.config', 'tempiex-test-' + Date.now());
const testConfigFile = join(testConfigDir, 'config.toml');

describe('config loader', () => {
  beforeEach(() => {
    process.env['TEMPIEX_CONFIG_FILE'] = testConfigFile;
  });

  afterEach(() => {
    delete process.env['TEMPIEX_CONFIG_FILE'];
    delete process.env['TEMPIEX_NAMESPACE'];
    if (existsSync(testConfigDir)) {
      rmSync(testConfigDir, { recursive: true });
    }
  });

  it('creates default config on first load', () => {
    const config = loadConfig();
    expect(config.version).toBe(1);
    expect(config.activeProfile).toBe('default');
    expect(config.profiles['default']).toBeDefined();
    expect(existsSync(testConfigFile)).toBe(true);
  });

  it('roundtrip: save and load config', () => {
    mkdirSync(testConfigDir, { recursive: true });
    const toml = TOML.stringify({
      version: 1,
      activeProfile: 'myprofile',
      profiles: {
        myprofile: {
          address: 'http://custom:9090',
          namespace: 'testns',
          tls: false,
          output: 'json',
        },
      },
    });
    writeFileSync(testConfigFile, toml, 'utf-8');

    const config = loadConfig();
    expect(config.activeProfile).toBe('myprofile');
    expect(config.profiles['myprofile']?.address).toBe('http://custom:9090');
    expect(config.profiles['myprofile']?.namespace).toBe('testns');
    expect(config.profiles['myprofile']?.output).toBe('json');
  });

  it('saveConfig writes valid TOML that can be reloaded', () => {
    mkdirSync(testConfigDir, { recursive: true });
    const original = loadConfig();
    saveConfig(original);
    const reloaded = loadConfig();
    expect(reloaded.version).toBe(original.version);
    expect(reloaded.activeProfile).toBe(original.activeProfile);
  });

  it('resolveConfig prefers CLI flag over env var over profile', () => {
    mkdirSync(testConfigDir, { recursive: true });
    const toml = TOML.stringify({
      version: 1,
      activeProfile: 'default',
      profiles: {
        default: {
          address: 'http://profile:1234',
          namespace: 'profile-ns',
          tls: false,
          output: 'text',
        },
      },
    });
    writeFileSync(testConfigFile, toml, 'utf-8');

    const config = loadConfig();

    // CLI flag wins
    const r1 = resolveConfig({ namespace: 'cli-ns' }, config);
    expect(r1.namespace).toBe('cli-ns');

    // env var wins over profile
    process.env['TEMPIEX_NAMESPACE'] = 'env-ns';
    const r2 = resolveConfig({}, config);
    expect(r2.namespace).toBe('env-ns');
    delete process.env['TEMPIEX_NAMESPACE'];

    // profile value used when no flag/env
    const r3 = resolveConfig({}, config);
    expect(r3.namespace).toBe('profile-ns');

    // address from CLI flag
    const r4 = resolveConfig({ address: 'http://override:5678' }, config);
    expect(r4.address).toBe('http://override:5678');
  });
});
