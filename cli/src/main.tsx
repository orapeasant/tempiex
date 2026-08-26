import { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { registerWorkflowCommands } from './commands/workflow/index.js';
import { registerNamespaceCommands } from './commands/namespace/index.js';
import { registerClusterCommands } from './commands/cluster/index.js';
import { registerTaskQueueCommands } from './commands/taskqueue/index.js';
import { registerConfigCommands } from './commands/config/index.js';
import { registerServerCommands } from './commands/server/index.js';
import { App } from './app/App.js';
import { resolveConfig } from './config/loader.js';
import { TempiexHttpClient } from './client/http.js';

const program = new Command();

program
  .name('tempiex')
  .description('Tempiex CLI — manage workflow executions and namespaces')
  .version('0.1.0')
  .enablePositionalOptions()
  .option('--address <addr>', 'Server address (HTTP)')
  .option('-n, --namespace <ns>', 'Namespace')
  .option('--api-key <key>', 'API key')
  .option('-o, --output <format>', 'Output format: text|json|yaml')
  .option('--profile <profile>', 'Config profile')
  .option('--log-level <level>', 'Log level');

registerWorkflowCommands(program);
registerNamespaceCommands(program);
registerClusterCommands(program);
registerTaskQueueCommands(program);
registerConfigCommands(program);
registerServerCommands(program);

// If no sub-command is given, launch interactive TUI
program.action((opts) => {
  const cfg = resolveConfig(opts as Record<string, string>);
  const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
  const { unmount } = render(
    React.createElement(App, {
      client,
      onDone: () => { unmount(); process.exit(0); },
    }),
  );
});

program.parse(process.argv);
