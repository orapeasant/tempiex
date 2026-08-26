import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { loadConfig, saveConfig, deleteProfileKey } from '../../config/loader.js';

interface ConfigDeleteProps {
  profile: string;
  propName: string | undefined;
  onDone: () => void;
}

export function ConfigDelete({ profile, propName, onDone }: ConfigDeleteProps): React.ReactElement {
  useEffect(() => {
    const config = loadConfig();
    if (!propName) {
      // Delete entire profile
      delete config.profiles[profile];
      saveConfig(config);
    } else {
      const updated = deleteProfileKey(profile, propName!, config);
      saveConfig(updated);
    }
    onDone();
  }, [profile, propName, onDone]);

  return (
    <Box>
      <Text color="green">✓ Deleted {propName ? `${profile}.${propName}` : `profile ${profile}`}</Text>
    </Box>
  );
}
