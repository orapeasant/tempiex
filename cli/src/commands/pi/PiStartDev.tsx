import React, { useEffect } from 'react';
import { Box, Text } from 'ink';

interface PiStartDevProps {
  onDone: () => void;
}

export function PiStartDev({ onDone }: PiStartDevProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  return (
    <Box flexDirection="column">
      <Text bold>Tempiex Pi Agent Harness</Text>
      <Text dimColor>
        Polls Tempiex activity task queues via the worker module and dispatches
        activities to AI agent tools. To start it, run from the project root:
      </Text>
      <Box marginTop={1} flexDirection="column">
        <Text>  cd /path/to/tempiex/pi</Text>
        <Text>  go run ./cmd/pi -config config/development.yaml</Text>
      </Box>
      <Box marginTop={1}>
        <Text dimColor>
          For a generic (non-AI) worker instead, use `tempiex worker start-dev`.
        </Text>
      </Box>
    </Box>
  );
}
