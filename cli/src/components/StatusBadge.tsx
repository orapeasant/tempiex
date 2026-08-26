import React from 'react';
import { Text } from 'ink';

type StatusColor = 'cyan' | 'green' | 'red' | 'yellow' | 'magenta' | 'white';

const STATUS_COLORS: Record<string, StatusColor> = {
  Running: 'cyan',
  RUNNING: 'cyan',
  Completed: 'green',
  COMPLETED: 'green',
  Failed: 'red',
  FAILED: 'red',
  Cancelled: 'yellow',
  CANCELLED: 'yellow',
  TimedOut: 'magenta',
  TIMED_OUT: 'magenta',
  Terminated: 'red',
  TERMINATED: 'red',
};

interface StatusBadgeProps {
  status: string | undefined;
}

export function StatusBadge({ status }: StatusBadgeProps): React.ReactElement {
  const label = status ?? 'Unknown';
  const color: StatusColor = STATUS_COLORS[label] ?? 'white';
  return <Text color={color}>{label}</Text>;
}
