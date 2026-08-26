import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';

interface ConfirmProps {
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export function Confirm({ message, onConfirm, onCancel }: ConfirmProps): React.ReactElement {
  const [answer, setAnswer] = useState<string | null>(null);

  useInput((input) => {
    if (answer !== null) return;
    const lower = input.toLowerCase();
    if (lower === 'y') {
      setAnswer('y');
      onConfirm();
    } else if (lower === 'n') {
      setAnswer('n');
      onCancel();
    }
  });

  return (
    <Box>
      <Text>{message} </Text>
      <Text color="yellow">[y/N] </Text>
      {answer !== null && <Text>{answer}</Text>}
    </Box>
  );
}
