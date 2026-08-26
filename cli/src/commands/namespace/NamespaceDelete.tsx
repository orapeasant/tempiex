import React, { useEffect } from 'react';
import { Text } from 'ink';

interface NamespaceDeleteProps {
  namespace: string;
  onDone: () => void;
}

export function NamespaceDelete({ namespace, onDone }: NamespaceDeleteProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);
  return <Text color="yellow">Namespace delete not yet supported by server API: {namespace}</Text>;
}
