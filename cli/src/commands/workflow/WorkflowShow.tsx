import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { useEventHistory } from '../../hooks/useEventHistory.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { formatDateTime } from '../../utils/datetime.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface WorkflowShowProps {
  client: TempiexHttpClient;
  namespace: string;
  workflowId: string;
  runId: string | undefined;
  reverse?: boolean;
  output: OutputFormat;
  onDone: () => void;
}

export function WorkflowShow({
  client,
  namespace,
  workflowId,
  runId,
  reverse,
  output,
  onDone,
}: WorkflowShowProps): React.ReactElement {
  const { events, loading, error } = useEventHistory(client, namespace, workflowId, runId);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(events, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, events, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Fetching history…" />;
  if (error) return <ErrorPanel error={error} />;

  const displayed = reverse ? [...events].reverse() : events;

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}>
        <Text bold>Event History </Text>
        <Text dimColor>({workflowId})</Text>
      </Box>
      {displayed.map((ev, i) => (
        <Box key={i} flexDirection="column" marginBottom={0}>
          <Box>
            <Box width={6}><Text dimColor>{String(ev.eventId)}</Text></Box>
            <Box width={20}><Text dimColor>{formatDateTime(ev.eventTime as string | undefined)}</Text></Box>
            <Box><Text color="cyan">{String(ev.eventType ?? '—')}</Text></Box>
          </Box>
        </Box>
      ))}
      {events.length === 0 && <Text dimColor>No events found.</Text>}
      <Box marginTop={1}>
        <Text dimColor>{events.length} event(s)</Text>
      </Box>
    </Box>
  );
}
