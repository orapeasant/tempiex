import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'fs';
import { join, dirname } from 'path';
import { homedir } from 'os';
import TOML from '@iarna/toml';
import type { Command } from 'commander';
import { Config, ConfigSchema, GlobalFlags, Profile } from './types.js';

function getConfigPath(): string {
  return (
    process.env['TEMPIEX_CONFIG_FILE'] ??
    join(homedir(), '.config', 'tempiex', 'config.toml')
  );
}

const defaultConfig: Config = {
  version: 1,
  activeProfile: 'default',
  profiles: {
    default: {
      address: 'http://localhost:8080',
      namespace: 'default',
      tls: false,
      output: 'text',
    },
  },
};

export function loadConfig(): Config {
  const configPath = getConfigPath();
  if (!existsSync(configPath)) {
    saveConfig(defaultConfig);
    return defaultConfig;
  }
  try {
    const raw = readFileSync(configPath, 'utf-8');
    const parsed = TOML.parse(raw);
    return ConfigSchema.parse(parsed);
  } catch {
    return defaultConfig;
  }
}

export function saveConfig(config: Config): void {
  const configPath = getConfigPath();
  mkdirSync(dirname(configPath), { recursive: true });
  writeFileSync(configPath, TOML.stringify(config as TOML.JsonMap), 'utf-8');
}

/** Walk parent chain and merge opts: child opts override parent opts. */
function collectParentOpts(cmd: Command): Record<string, unknown> {
  const chain: Record<string, unknown>[] = [];
  let cur: Command | null = cmd.parent;
  while (cur) {
    chain.unshift(cur.opts<Record<string, unknown>>());
    cur = cur.parent;
  }
  return Object.assign({}, ...chain);
}

export function resolveConfig(
  opts: Partial<GlobalFlags> & { _cmd?: Command },
  config?: Config,
): Profile & { address: string; namespace: string; output: 'text' | 'json' | 'yaml' } {
  const cfg = config ?? loadConfig();

  // Merge parent-chain opts (lower priority) with current opts (higher priority)
  const parentOpts = opts._cmd ? collectParentOpts(opts._cmd) : {};
  const merged = { ...parentOpts, ...opts } as Partial<GlobalFlags>;

  const profileName =
    merged.profile ?? process.env['TEMPIEX_PROFILE'] ?? cfg.activeProfile ?? 'default';
  const profile = cfg.profiles[profileName] ?? defaultConfig.profiles['default']!;

  return {
    address:
      merged.address ??
      process.env['TEMPIEX_ADDRESS'] ??
      profile.address ??
      'http://localhost:8080',
    namespace:
      merged.namespace ??
      process.env['TEMPIEX_NAMESPACE'] ??
      profile.namespace ??
      'default',
    apiKey: merged.apiKey ?? process.env['TEMPIEX_API_KEY'] ?? profile.apiKey,
    tls: merged.tls ?? profile.tls ?? false,
    tlsCertPath:
      merged.tlsCertPath ??
      process.env['TEMPIEX_TLS_CLIENT_CERT_PATH'] ??
      profile.tlsCertPath,
    tlsKeyPath:
      merged.tlsKeyPath ??
      process.env['TEMPIEX_TLS_CLIENT_KEY_PATH'] ??
      profile.tlsKeyPath,
    tlsCaPath:
      merged.tlsCaPath ??
      process.env['TEMPIEX_TLS_SERVER_CA_CERT_PATH'] ??
      profile.tlsCaPath,
    codecEndpoint:
      merged.codecEndpoint ?? process.env['TEMPIEX_CODEC_ENDPOINT'] ?? profile.codecEndpoint,
    output: merged.output ?? profile.output ?? 'text',
  };
}

export function getProfile(name: string, config?: Config): Profile | undefined {
  const cfg = config ?? loadConfig();
  return cfg.profiles[name];
}

export function setProfileKey(
  profileName: string,
  key: string,
  value: string,
  config?: Config,
): Config {
  const cfg = config ?? loadConfig();
  if (!cfg.profiles[profileName]) {
    cfg.profiles[profileName] = {
      address: 'http://localhost:8080',
      namespace: 'default',
      tls: false,
      output: 'text',
    };
  }
  (cfg.profiles[profileName] as Record<string, unknown>)[key] = value;
  return cfg;
}

export function deleteProfileKey(profileName: string, key: string, config?: Config): Config {
  const cfg = config ?? loadConfig();
  if (cfg.profiles[profileName]) {
    delete (cfg.profiles[profileName] as Record<string, unknown>)[key];
  }
  return cfg;
}
