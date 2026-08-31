import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { WorkerStartDev } from './WorkerStartDev.js';

export function registerWorkerCommands(program: Command): void {
  const worker = program.command('worker').description('Generic Tempiex activity worker management');

  worker.command('start-dev')
    .description('Start the development worker (instructions)')
    .action(() => {
      const { unmount } = render(
        React.createElement(WorkerStartDev, {
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
