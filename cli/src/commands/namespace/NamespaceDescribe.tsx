import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient, Namespace } from '../../client/http.js';
import { useNamespaces } from '../../hooks/useNamespaces.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface NamespaceDescribeProps {
  client: TempiexHttpClient;
  namespace: string;
  output: OutputFormat;
  onDone: () => void;
}

export function NamespaceDescribe({
  client,
  namespace,
  output,
  onDone,
}: NamespaceDescribeProps): React.ReactElement {
  const { namespaces, loading, error } = useNamespaces(client);

  const ns: Namespace | undefined = namespaces.find(
    (n) => n.namespaceInfo.name === namespace || n.namespaceInfo.id === namespace,
  );

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(ns ?? null, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, ns, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Fetching namespace…" />;
  if (error) return <ErrorPanel error={error} />;
  if (!ns) return <Text color="red">Namespace not found: {namespace}</Text>;

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}><Text bold>Namespace Details</Text></Box>
      <Box><Text bold>Name:       </Text><Text>{ns.namespaceInfo.name}</Text></Box>
      <Box><Text bold>ID:         </Text><Text>{ns.namespaceInfo.id}</Text></Box>
      <Box><Text bold>Description:</Text><Text> {ns.namespaceInfo.description || '—'}</Text></Box>
      <Box><Text bold>Retention:  </Text><Text>{String(ns.config.workflowExecutionRetentionTtl?.days ?? '—')} days</Text></Box>
    </Box>
  );
}
