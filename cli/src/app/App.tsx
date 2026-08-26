import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';
import { TempiexHttpClient, Namespace } from '../client/http.js';
import { useNamespaces } from '../hooks/useNamespaces.js';
import { useWorkflows } from '../hooks/useWorkflows.js';
import { Spinner } from '../components/Spinner.js';
import { StatusBadge } from '../components/StatusBadge.js';
import { ErrorPanel } from '../components/ErrorPanel.js';

interface AppProps {
  client: TempiexHttpClient;
  onDone: () => void;
}

export function App({ client, onDone }: AppProps): React.ReactElement {
  const { namespaces, loading: nsLoading, error: nsError } = useNamespaces(client);
  const [selectedNsIdx, setSelectedNsIdx] = useState(0);
  const [selectedWfIdx, setSelectedWfIdx] = useState(0);
  const [focus, setFocus] = useState<'ns' | 'wf'>('ns');

  const selectedNs: Namespace | undefined = namespaces[selectedNsIdx];
  const nsName = selectedNs?.namespaceInfo.name ?? 'default';

  const { data: wfData, loading: wfLoading } = useWorkflows(client, nsName, undefined, 50);
  const workflows = wfData?.executions ?? [];

  useInput((input, key) => {
    if (input === 'q') {
      onDone();
      return;
    }
    if (key.tab) {
      setFocus((f) => (f === 'ns' ? 'wf' : 'ns'));
      return;
    }
    if (focus === 'ns') {
      if (key.upArrow) setSelectedNsIdx((i) => Math.max(0, i - 1));
      if (key.downArrow) setSelectedNsIdx((i) => Math.min(namespaces.length - 1, i + 1));
      if (key.return) setFocus('wf');
    } else {
      if (key.upArrow) setSelectedWfIdx((i) => Math.max(0, i - 1));
      if (key.downArrow) setSelectedWfIdx((i) => Math.min(workflows.length - 1, i + 1));
      if (key.escape) setFocus('ns');
    }
  });

  if (nsLoading) return <Spinner label="Loading…" />;
  if (nsError) return <ErrorPanel error={nsError} />;

  return (
    <Box flexDirection="column" height={30}>
      <Box marginBottom={1}>
        <Text bold color="cyan">Tempiex TUI </Text>
        <Text dimColor>— [Tab] switch panel | [↑↓] navigate | [q] quit</Text>
      </Box>
      <Box flexDirection="row" flexGrow={1}>
        {/* Namespace sidebar */}
        <Box flexDirection="column" width={32} marginRight={2}>
          <Text bold underline>Namespaces {focus === 'ns' ? '●' : '○'}</Text>
          {namespaces.slice(0, 20).map((ns, i) => (
            <Box key={i}>
              <Text
                color={i === selectedNsIdx ? 'cyan' : undefined}
                bold={i === selectedNsIdx}
              >
                {i === selectedNsIdx ? '▶ ' : '  '}
                {ns.namespaceInfo.name.slice(0, 26)}
              </Text>
            </Box>
          ))}
        </Box>

        {/* Workflow list */}
        <Box flexDirection="column" flexGrow={1}>
          <Text bold underline>Workflows: {nsName} {focus === 'wf' ? '●' : '○'}</Text>
          {wfLoading && <Spinner label="Loading workflows…" />}
          {!wfLoading && workflows.slice(0, 20).map((wf, i) => {
            const isSelected = focus === 'wf' && i === selectedWfIdx;
            return (
              <Box key={i}>
                <Text color={isSelected ? 'cyan' : undefined} bold={isSelected}>
                  {isSelected ? '▶ ' : '  '}
                </Text>
                <Box width={30}>
                  <Text color={isSelected ? 'cyan' : undefined}>
                    {(wf.execution?.workflowId ?? '—').slice(0, 28)}
                  </Text>
                </Box>
                <Box width={16}>
                  <StatusBadge status={wf.status} />
                </Box>
                <Box>
                  <Text dimColor>{wf.type?.name?.slice(0, 20) ?? '—'}</Text>
                </Box>
              </Box>
            );
          })}
          {!wfLoading && workflows.length === 0 && (
            <Text dimColor>No workflows in this namespace.</Text>
          )}
        </Box>
      </Box>
    </Box>
  );
}
