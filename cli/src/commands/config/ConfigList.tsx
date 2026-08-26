import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { loadConfig } from '../../config/loader.js';

interface ConfigListProps {
  profile: string;
  onDone: () => void;
}

export function ConfigList({ profile, onDone }: ConfigListProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  const config = loadConfig();
  const prof = config.profiles[profile];

  if (!prof) {
    return <Text color="red">Profile not found: {profile}</Text>;
  }

  return (
    <Box flexDirection="column">
      <Box marginBottom={1}><Text bold>Config Profile: {profile}</Text></Box>
      {Object.entries(prof).map(([k, v]) => (
        <Box key={k}>
          <Box width={24}><Text dimColor>{k}</Text></Box>
          <Text>{String(v)}</Text>
        </Box>
      ))}
    </Box>
  );
}
