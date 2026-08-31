import React, { useEffect } from 'react';
import { Box, Text } from 'ink';

interface WorkerStartDevProps {
  onDone: () => void;
}

export function WorkerStartDev({ onDone }: WorkerStartDevProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  return (
    <Box flexDirection="column">
      <Text bold>Tempiex Generic Worker</Text>
      <Text dimColor>
        Polls Tempiex activity task queues and dispatches to an ActivityHandler.
        To start it, run from the project root:
      </Text>
      <Box marginTop={1} flexDirection="column">
        <Text>  cd /path/to/tempiex/worker</Text>
        <Text>  go run ./cmd/worker -config config/development.yaml</Text>
      </Box>
      <Box marginTop={1}>
        <Text dimColor>
          For an AI-agent-harness worker instead, use `tempiex pi start-dev`.
        </Text>
      </Box>
    </Box>
  );
}
