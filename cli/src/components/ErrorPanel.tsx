import React from 'react';
import { Box, Text } from 'ink';

interface ErrorPanelProps {
  error: Error | string;
  title?: string;
}

export function ErrorPanel({ error, title }: ErrorPanelProps): React.ReactElement {
  const message = typeof error === 'string' ? error : error.message;
  return (
    <Box flexDirection="column" borderStyle="round" borderColor="red" padding={1}>
      <Text color="red" bold>
        {title ?? '✗ Error'}
      </Text>
      <Text>{message}</Text>
    </Box>
  );
}
