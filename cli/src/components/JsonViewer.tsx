import React from 'react';
import { Box, Text } from 'ink';

interface JsonViewerProps {
  data: unknown;
  indent?: number;
}

export function JsonViewer({ data, indent = 2 }: JsonViewerProps): React.ReactElement {
  const json = JSON.stringify(data, null, indent);
  return (
    <Box flexDirection="column">
      {json.split('\n').map((line, i) => (
        <Text key={i}>{line}</Text>
      ))}
    </Box>
  );
}
