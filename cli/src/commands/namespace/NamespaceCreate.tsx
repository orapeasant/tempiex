import React, { useEffect, useState } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { Spinner } from '../../components/Spinner.js';
import { ErrorPanel } from '../../components/ErrorPanel.js';
import { OutputFormat } from '../../utils/output.js';

interface NamespaceCreateProps {
  client: TempiexHttpClient;
  namespace: string;
  retention?: string;
  description?: string;
  output: OutputFormat;
  onDone: () => void;
}

export function NamespaceCreate({
  client,
  namespace,
  retention,
  description,
  output: _output,
  onDone,
}: NamespaceCreateProps): React.ReactElement {
  const [done, setDone] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    client
      .createNamespace({
        name: namespace,
        description: description ?? '',
        workflowExecutionRetentionPeriod: retention ?? '72h',
      })
      .then(() => {
        setDone(true);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err : new Error(String(err)));
        setDone(true);
      });
  }, [client, namespace, retention, description]);

  useEffect(() => {
    if (done) onDone();
  }, [done, onDone]);

  if (!done) return <Spinner label="Creating namespace…" />;
  if (error) return <ErrorPanel error={error} />;

  return (
    <Box>
      <Text color="green">✓ Namespace created: {namespace}</Text>
    </Box>
  );
}
