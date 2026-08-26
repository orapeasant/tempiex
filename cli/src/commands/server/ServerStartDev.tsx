import React, { useEffect } from 'react';
import { Box, Text } from 'ink';

interface ServerStartDevProps {
  onDone: () => void;
}

export function ServerStartDev({ onDone }: ServerStartDevProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  return (
    <Box flexDirection="column">
      <Text bold>Tempiex Dev Server</Text>
      <Text dimColor>To start the server, run from the project root:</Text>
      <Box marginTop={1} flexDirection="column">
        <Text>  cd /path/to/tempiex/server</Text>
        <Text>  go run . serve</Text>
      </Box>
      <Box marginTop={1}>
        <Text dimColor>gRPC: :8133  |  HTTP: :8080</Text>
      </Box>
    </Box>
  );
}
