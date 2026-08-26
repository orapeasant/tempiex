import React, { useEffect, useState } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient, HealthResponse } from '../../client/http.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface ClusterHealthProps {
  client: TempiexHttpClient;
  output: OutputFormat;
  onDone: () => void;
}

export function ClusterHealth({
  client,
  output,
  onDone,
}: ClusterHealthProps): React.ReactElement {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    client
      .health()
      .then((res) => {
        if (!cancelled) { setHealth(res); setLoading(false); }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
          setLoading(false);
        }
      });
    return () => { cancelled = true; };
  }, [client]);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(health, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, health, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Checking cluster health…" />;
  if (error) return <ErrorPanel error={error} />;

  const ok = health?.status === 'ok';
  return (
    <Box flexDirection="column">
      <Box>
        <Text bold>Status:  </Text>
        <Text color={ok ? 'green' : 'red'}>{health?.status ?? 'unknown'}</Text>
      </Box>
      <Box>
        <Text bold>Version: </Text>
        <Text>{health?.version ?? '—'}</Text>
      </Box>
    </Box>
  );
}
