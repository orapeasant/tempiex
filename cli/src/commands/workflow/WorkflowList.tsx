import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { useWorkflows } from '../../hooks/useWorkflows.js';
import { StatusBadge } from '../../components/StatusBadge.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { formatRelative } from '../../utils/datetime.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface WorkflowListProps {
  client: TempiexHttpClient;
  namespace: string;
  query?: string;
  limit?: number;
  output: OutputFormat;
  onDone: () => void;
}

export function WorkflowList({
  client,
  namespace,
  query,
  limit,
  output,
  onDone,
}: WorkflowListProps): React.ReactElement {
  const { data, loading, error } = useWorkflows(client, namespace, query, limit);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(data?.executions ?? [], output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, data, output, onDone]);

  if (output !== 'text') return <></>;

  if (loading) return <Spinner label="Fetching workflows…" />;
  if (error) return <ErrorPanel error={error} />;

  const executions = data?.executions ?? [];

  const rows = executions.map((e) => ({
    workflowId: e.execution?.workflowId ?? '—',
    runId: e.execution?.runId?.slice(0, 8) ?? '—',
    type: e.type?.name ?? '—',
    status: e.status ?? '—',
    taskQueue: e.taskQueue ?? '—',
    started: formatRelative(e.startTime),
  }));

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}>
        <Text bold>Workflows </Text>
        <Text dimColor>(namespace: {namespace})</Text>
      </Box>
      <Box flexDirection="column">
        <Box>
          <Box width={32}><Text bold>WORKFLOW ID</Text></Box>
          <Box width={12}><Text bold>RUN ID</Text></Box>
          <Box width={24}><Text bold>TYPE</Text></Box>
          <Box width={14}><Text bold>STATUS</Text></Box>
          <Box width={20}><Text bold>TASK QUEUE</Text></Box>
          <Box width={20}><Text bold>STARTED</Text></Box>
        </Box>
        <Box>
          <Text dimColor>{'─'.repeat(120)}</Text>
        </Box>
        {rows.length === 0 && <Text dimColor>No workflows found.</Text>}
        {rows.map((row, i) => (
          <Box key={i}>
            <Box width={32}><Text>{row.workflowId.slice(0, 30)}</Text></Box>
            <Box width={12}><Text>{row.runId}</Text></Box>
            <Box width={24}><Text>{row.type.slice(0, 22)}</Text></Box>
            <Box width={14}><StatusBadge status={row.status} /></Box>
            <Box width={20}><Text>{row.taskQueue.slice(0, 18)}</Text></Box>
            <Box width={20}><Text dimColor>{row.started}</Text></Box>
          </Box>
        ))}
      </Box>
      <Box marginTop={1}>
        <Text dimColor>{executions.length} workflow(s)</Text>
      </Box>
    </Box>
  );
}
