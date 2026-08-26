import React, { useEffect, useState } from 'react';
import { Box, Text } from 'ink';
import { TempiexHttpClient } from '../../client/http.js';
import { Spinner } from '../../components/Spinner.js';
import { printOutput, OutputFormat } from '../../utils/output.js';

interface WorkflowStartProps {
  client: TempiexHttpClient;
  namespace: string;
  workflowId?: string;
  workflowType: string;
  taskQueue: string;
  input?: string;
  output: OutputFormat;
  onDone: () => void;
}

interface StartResult {
  workflowId: string;
  runId: string;
}

export function WorkflowStart({
  client: _client,
  namespace: _namespace,
  workflowId,
  workflowType,
  taskQueue,
  input: _input,
  output,
  onDone,
}: WorkflowStartProps): React.ReactElement {
  const [result, setResult] = useState<StartResult | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Stub: workflow start requires gRPC
    const stub: StartResult = {
      workflowId: workflowId ?? `wf-${Date.now()}`,
      runId: 'not-supported',
    };
    setTimeout(() => {
      console.error(
        `Starting workflow type="${workflowType}" on task-queue="${taskQueue}" is not yet supported via server API.`,
      );
      setResult(stub);
      setLoading(false);
    }, 0);
  }, [workflowId, workflowType, taskQueue]);

  useEffect(() => {
    if (!loading) {
      if (output !== 'text') {
        printOutput(result, output);
      }
      onDone();
    }
  }, [loading, result, output, onDone]);

  if (output !== 'text') return <></>;
  if (loading) return <Spinner label="Starting workflow…" />;

  return (
    <Box flexDirection="column">
      <Text color="yellow">Not yet supported by server API</Text>
    </Box>
  );
}
