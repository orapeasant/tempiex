import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { PiStartDev } from './PiStartDev.js';

export function registerPiCommands(program: Command): void {
  const pi = program.command('pi').description('AI agent harness management');

  pi.command('start-dev')
    .description('Start the development pi agent harness (instructions)')
    .action(() => {
      const { unmount } = render(
        React.createElement(PiStartDev, {
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
