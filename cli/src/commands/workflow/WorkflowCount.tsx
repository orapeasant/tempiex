import React, { useEffect, useState } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface WorkflowCountProps {
  client: TempiexHttpClient;
  namespace: string;
  query?: string;
  output: OutputFormat;
  onDone: () => void;
}

export function WorkflowCount({
  client,
  namespace,
  query,
  output,
  onDone,
}: WorkflowCountProps): React.ReactElement {
  const [count, setCount] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    client
      .listWorkflows(namespace, query, 1000)
      .then((res) => {
        if (!cancelled) {
          setCount(res.executions?.length ?? 0);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
          setLoading(false);
        }
      });
    return () => { cancelled = true; };
  }, [client, namespace, query]);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput({ count }, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, count, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Counting workflows…" />;
  if (error) return <ErrorPanel error={error} />;

  return (
    <Box>
      <Text bold>Count: </Text>
      <Text>{String(count)}</Text>
    </Box>
  );
}
