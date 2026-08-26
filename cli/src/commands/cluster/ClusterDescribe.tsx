import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { useCluster } from '../../hooks/useCluster.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface ClusterDescribeProps {
  client: TempiexHttpClient;
  output: OutputFormat;
  onDone: () => void;
}

export function ClusterDescribe({
  client,
  output,
  onDone,
}: ClusterDescribeProps): React.ReactElement {
  const { cluster, loading, error } = useCluster(client);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(cluster, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, cluster, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Fetching cluster info…" />;
  if (error) return <ErrorPanel error={error} />;

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}><Text bold>Cluster Information</Text></Box>
      <Box><Text bold>ID:           </Text><Text>{cluster?.clusterId ?? '—'}</Text></Box>
      <Box><Text bold>Name:         </Text><Text>{cluster?.clusterName ?? '—'}</Text></Box>
      <Box><Text bold>Version:      </Text><Text>{cluster?.serverVersion ?? '—'}</Text></Box>
      <Box><Text bold>History Shards:</Text><Text> {String(cluster?.historyShardCount ?? '—')}</Text></Box>
    </Box>
  );
}
