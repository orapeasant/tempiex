import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { loadConfig, saveConfig, setProfileKey } from '../../config/loader.js';

interface ConfigSetProps {
  profile: string;
  propName: string;
  value: string;
  onDone: () => void;
}

export function ConfigSet({ profile, propName, value, onDone }: ConfigSetProps): React.ReactElement {
  useEffect(() => {
    const config = loadConfig();
    const updated = setProfileKey(profile, propName, value, config);
    saveConfig(updated);
    onDone();
  }, [profile, propName, value, onDone]);

  return (
    <Box>
      <Text color="green">✓ Set {profile}.{propName} = {value}</Text>
    </Box>
  );
}
