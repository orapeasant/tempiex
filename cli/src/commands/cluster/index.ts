import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { TempiexHttpClient } from '../../client/http.js';
import { resolveConfig } from '../../config/loader.js';
import { ClusterDescribe } from './ClusterDescribe.js';
import { ClusterHealth } from './ClusterHealth.js';

export function registerClusterCommands(program: Command): void {
  const cluster = program.command('cluster').description('Cluster operations');

  cluster.command('describe')
    .description('Describe the cluster')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(ClusterDescribe, {
          client,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  cluster.command('health')
    .description('Check cluster health')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(ClusterHealth, {
          client,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
