import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { resolveConfig } from '../../config/loader.js';
import { TaskQueueDescribe } from './TaskQueueDescribe.js';

export function registerTaskQueueCommands(program: Command): void {
  const tq = program.command('taskqueue').description('Manage Task Queues');

  tq.command('describe')
    .description('Describe a Task Queue')
    .requiredOption('-t, --task-queue <name>', 'Task queue name')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts) => {
      const cfg = resolveConfig(opts);
      const { unmount } = render(
        React.createElement(TaskQueueDescribe, {
          taskQueue: opts.taskQueue as string,
          namespace: cfg.namespace,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
