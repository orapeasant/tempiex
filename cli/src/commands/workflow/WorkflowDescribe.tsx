import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { useWorkflowDetail } from '../../hooks/useWorkflowDetail.js';
import { StatusBadge } from '../../components/StatusBadge.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { formatDateTime, formatRelative } from '../../utils/datetime.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface WorkflowDescribeProps {
  client: TempiexHttpClient;
  namespace: string;
  workflowId: string;
  runId: string | undefined;
  output: OutputFormat;
  onDone: () => void;
}

export function WorkflowDescribe({
  client,
  namespace,
  workflowId,
  runId,
  output,
  onDone,
}: WorkflowDescribeProps): React.ReactElement {
  const { data, loading, error } = useWorkflowDetail(client, namespace, workflowId, runId);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(data, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, data, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Fetching workflow…" />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return <Text>Workflow not found.</Text>;

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}>
        <Text bold>Workflow Details</Text>
      </Box>
      <Box flexDirection="column">
        <Box><Text bold>Workflow ID:  </Text><Text>{data.execution?.workflowId}</Text></Box>
        <Box><Text bold>Run ID:       </Text><Text>{data.execution?.runId}</Text></Box>
        <Box><Text bold>Type:         </Text><Text>{data.type?.name ?? '—'}</Text></Box>
        <Box><Text bold>Status:       </Text><StatusBadge status={data.status} /></Box>
        <Box><Text bold>Task Queue:   </Text><Text>{data.taskQueue ?? '—'}</Text></Box>
        <Box><Text bold>Start Time:   </Text><Text>{formatDateTime(data.startTime)} ({formatRelative(data.startTime)})</Text></Box>
        <Box><Text bold>Close Time:   </Text><Text>{data.closeTime ? formatDateTime(data.closeTime) : '—'}</Text></Box>
        <Box><Text bold>History Len:  </Text><Text>{String(data.historyLength ?? '—')}</Text></Box>
      </Box>
    </Box>
  );
}
