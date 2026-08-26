import React, { useEffect } from 'react';
import { Text } from 'ink';

interface StubProps {
  onDone: () => void;
}

export function WorkflowTerminate({ onDone }: StubProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);
  return <Text color="yellow">Not yet supported by server API</Text>;
}
