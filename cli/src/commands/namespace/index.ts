import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { TempiexHttpClient } from '../../client/http.js';
import { resolveConfig } from '../../config/loader.js';
import { NamespaceList } from './NamespaceList.js';
import { NamespaceDescribe } from './NamespaceDescribe.js';
import { NamespaceCreate } from './NamespaceCreate.js';
import { NamespaceDelete } from './NamespaceDelete.js';

export function registerNamespaceCommands(program: Command): void {
  const ns = program.command('namespace').description('Manage Namespaces');

  ns.command('list')
    .description('List all Namespaces')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(NamespaceList, {
          client,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  ns.command('describe')
    .description('Describe a Namespace')
    .option('-n, --namespace <ns>', 'Namespace name or ID')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(NamespaceDescribe, {
          client,
          namespace: cfg.namespace,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  ns.command('create')
    .description('Create a Namespace')
    .requiredOption('-n, --namespace <ns>', 'Namespace name')
    .option('--retention <duration>', 'Retention period', '72h')
    .option('--description <desc>', 'Description')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(NamespaceCreate, {
          client,
          namespace: opts.namespace as string,
          retention: opts.retention as string | undefined,
          description: opts.description as string | undefined,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  ns.command('delete')
    .description('Delete a Namespace')
    .requiredOption('-n, --namespace <ns>', 'Namespace name')
    .option('-y, --yes', 'Skip confirmation')
    .action((opts) => {
      console.log(`Not yet supported by server API: ${opts.namespace as string}`);
      process.exit(0);
    });

  // avoid lint warnings
  void NamespaceDelete;
}
