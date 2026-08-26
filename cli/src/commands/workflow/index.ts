import type { Command } from 'commander';
import { render } from 'ink';
import React from 'react';
import { TempiexHttpClient } from '../../client/http.js';
import { resolveConfig } from '../../config/loader.js';
import type { GlobalFlags } from '../../config/types.js';
import { WorkflowList } from './WorkflowList.js';
import { WorkflowDescribe } from './WorkflowDescribe.js';
import { WorkflowStart } from './WorkflowStart.js';
import { WorkflowShow } from './WorkflowShow.js';
import { WorkflowCount } from './WorkflowCount.js';

export function registerWorkflowCommands(program: Command): void {
  const wf = program.command('workflow').description('Manage Workflow Executions');

  wf.command('list')
    .description('List Workflow Executions')
    .option('-q, --query <query>', 'SQL-like filter')
    .option('--limit <n>', 'Max executions', '50')
    .option('--archived', 'Show archived executions')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts, cmd) => {
      const cfg = resolveConfig({ ...opts as Partial<GlobalFlags>, _cmd: cmd });
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(WorkflowList, {
          client,
          namespace: cfg.namespace,
          query: opts.query as string | undefined,
          limit: opts.limit ? Number(opts.limit) : 50,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  wf.command('describe')
    .description('Show Workflow Execution details')
    .requiredOption('-w, --workflow-id <id>', 'Workflow ID')
    .option('-r, --run-id <runId>', 'Run ID')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts, cmd) => {
      const cfg = resolveConfig({ ...opts as Partial<GlobalFlags>, _cmd: cmd });
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(WorkflowDescribe, {
          client,
          namespace: cfg.namespace,
          workflowId: opts.workflowId as string,
          runId: opts.runId as string | undefined,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  wf.command('start')
    .description('Start a Workflow Execution')
    .option('-w, --workflow-id <id>', 'Workflow ID')
    .requiredOption('--type <type>', 'Workflow type name')
    .requiredOption('-t, --task-queue <queue>', 'Task queue name')
    .option('-i, --input <json>', 'JSON input')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts, cmd) => {
      const cfg = resolveConfig({ ...opts as Partial<GlobalFlags>, _cmd: cmd });
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(WorkflowStart, {
          client,
          namespace: cfg.namespace,
          workflowId: opts.workflowId as string | undefined,
          workflowType: opts.type as string,
          taskQueue: opts.taskQueue as string,
          input: opts.input as string | undefined,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  wf.command('show')
    .description('Show Event History')
    .requiredOption('-w, --workflow-id <id>', 'Workflow ID')
    .option('-r, --run-id <runId>', 'Run ID')
    .option('-f, --follow', 'Follow execution progress')
    .option('--reverse', 'Newest events first')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts, cmd) => {
      const cfg = resolveConfig({ ...opts as Partial<GlobalFlags>, _cmd: cmd });
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(WorkflowShow, {
          client,
          namespace: cfg.namespace,
          workflowId: opts.workflowId as string,
          runId: opts.runId as string | undefined,
          reverse: opts.reverse as boolean | undefined,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });

  wf.command('terminate')
    .description('Terminate a Workflow Execution (stub)')
    .option('-w, --workflow-id <id>', 'Workflow ID')
    .option('--reason <reason>', 'Reason')
    .action(() => {
      console.log('Not yet supported by server API');
      process.exit(0);
    });

  wf.command('cancel')
    .description('Cancel a Workflow Execution (stub)')
    .option('-w, --workflow-id <id>', 'Workflow ID')
    .action(() => {
      console.log('Not yet supported by server API');
      process.exit(0);
    });

  wf.command('signal')
    .description('Signal a Workflow Execution (stub)')
    .option('-w, --workflow-id <id>', 'Workflow ID')
    .option('--name <name>', 'Signal name')
    .action(() => {
      console.log('Not yet supported by server API');
      process.exit(0);
    });

  wf.command('count')
    .description('Count Workflow Executions')
    .option('-q, --query <query>', 'SQL-like filter')
    .option('-o, --output <format>', 'Output format: text|json|yaml')
    .option('-n, --namespace <ns>', 'Namespace')
    .option('--address <addr>', 'Server address')
    .action((opts, cmd) => {
      const cfg = resolveConfig({ ...opts as Partial<GlobalFlags>, _cmd: cmd });
      const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
      const { unmount } = render(
        React.createElement(WorkflowCount, {
          client,
          namespace: cfg.namespace,
          query: opts.query as string | undefined,
          output: cfg.output,
          onDone: () => { unmount(); process.exit(0); },
        }),
      );
    });
}
