import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { ConfigList } from './ConfigList.js';
import { ConfigGet } from './ConfigGet.js';
import { ConfigSet } from './ConfigSet.js';
import { ConfigDelete } from './ConfigDelete.js';

export function registerConfigCommands(program: Command): void {
  const cfg = program.command('config').description('Manage CLI configuration');

  cfg.command('list')
    .description('List all config values for a profile')
    .option('--profile <profile>', 'Profile name', 'default')
    .action((opts) => {
      const { unmount } = render(
        React.createElement(ConfigList, {
          profile: opts.profile as string,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  cfg.command('get')
    .description('Get a config value')
    .option('--profile <profile>', 'Profile name', 'default')
    .option('-p, --prop <key>', 'Property name (optional — shows all if omitted)')
    .action((opts) => {
      const { unmount } = render(
        React.createElement(ConfigGet, {
          profile: opts.profile as string,
          propName: opts.prop as string | undefined,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  cfg.command('set')
    .description('Set a config value')
    .option('--profile <profile>', 'Profile name', 'default')
    .requiredOption('-p, --prop <key>', 'Property name')
    .requiredOption('-v, --value <val>', 'Property value')
    .action((opts) => {
      const { unmount } = render(
        React.createElement(ConfigSet, {
          profile: opts.profile as string,
          propName: opts.prop as string,
          value: opts.value as string,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  cfg.command('delete')
    .description('Delete a config value or entire profile')
    .option('--profile <profile>', 'Profile name', 'default')
    .option('-p, --prop <key>', 'Property to delete (omit to delete entire profile)')
    .action((opts) => {
      const { unmount } = render(
        React.createElement(ConfigDelete, {
          profile: opts.profile as string,
          propName: opts.prop as string | undefined,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
