import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { loadConfig } from '../../config/loader.js';
import { Profile } from '../../config/types.js';

interface ConfigGetProps {
  profile: string;
  propName: string | undefined;
  onDone: () => void;
}

export function ConfigGet({ profile, propName, onDone }: ConfigGetProps): React.ReactElement {
  useEffect(() => {
    onDone();
  }, [onDone]);

  const config = loadConfig();
  const prof = config.profiles[profile] as Profile | undefined;

  if (!prof) {
    return <Text color="red">Profile not found: {profile}</Text>;
  }

  if (!propName) {
    // Show all values for the profile
    return (
      <Box flexDirection="column">
        <Text bold>Config Profile: {profile}</Text>
        <Text> </Text>
        {Object.entries(prof).map(([k, v]) => (
          <Box key={k}>
            <Box width={24}><Text>{k}</Text></Box>
            <Text>{String(v ?? '')}</Text>
          </Box>
        ))}
      </Box>
    );
  }

  const value = (prof as Record<string, unknown>)[propName];
  if (value === undefined) {
    return <Text color="red">Key not found: {propName}</Text>;
  }

  return (
    <Box>
      <Text>{String(value)}</Text>
    </Box>
  );
}
