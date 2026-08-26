import React, { useEffect } from 'react';
import { Box, Text } from 'ink';

interface TaskQueueDescribeProps {
  taskQueue: string;
  namespace: string;
  onDone: () => void;
}

export function TaskQueueDescribe({
  taskQueue,
  namespace,
  onDone,
}: TaskQueueDescribeProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  return (
    <Box flexDirection="column">
      <Text color="yellow">Task queue describe not yet supported by server API</Text>
      <Text dimColor>Task Queue: {taskQueue} | Namespace: {namespace}</Text>
    </Box>
  );
}
