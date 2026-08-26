import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { useNamespaces } from '../../hooks/useNamespaces.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface NamespaceListProps {
  client: TempiexHttpClient;
  output: OutputFormat;
  onDone: () => void;
}

export function NamespaceList({
  client,
  output,
  onDone,
}: NamespaceListProps): React.ReactElement {
  const { namespaces, loading, error } = useNamespaces(client);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(namespaces, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, namespaces, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Fetching namespaces…" />;
  if (error) return <ErrorPanel error={error} />;

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}>
        <Text bold>Namespaces</Text>
      </Box>
      <Box>
        <Box width={36}><Text bold>NAME</Text></Box>
        <Box width={40}><Text bold>ID</Text></Box>
        <Box width={12}><Text bold>RETENTION</Text></Box>
      </Box>
      <Box><Text dimColor>{'─'.repeat(88)}</Text></Box>
      {namespaces.map((ns, i) => (
        <Box key={i}>
          <Box width={36}><Text>{ns.namespaceInfo.name}</Text></Box>
          <Box width={40}><Text dimColor>{ns.namespaceInfo.id}</Text></Box>
          <Box width={12}><Text>{String(ns.config.workflowExecutionRetentionTtl?.days ?? '—')}d</Text></Box>
        </Box>
      ))}
      <Box marginTop={1}><Text dimColor>{namespaces.length} namespace(s)</Text></Box>
    </Box>
  );
}
