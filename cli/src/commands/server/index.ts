import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { ServerStartDev } from './ServerStartDev.js';

export function registerServerCommands(program: Command): void {
  const server = program.command('server').description('Server management');

  server.command('start-dev')
    .description('Start the development server (instructions)')
    .action(() => {
      const { unmount } = render(
        React.createElement(ServerStartDev, {
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
