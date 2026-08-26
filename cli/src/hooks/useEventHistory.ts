import { useEffect, useState } from 'react';
import { TempiexHttpClient, HistoryEvent } from '../client/http.js';

interface UseEventHistoryResult {
  events: HistoryEvent[];
  loading: boolean;
  error: Error | null;
}

export function useEventHistory(
  client: TempiexHttpClient,
  namespace: string,
  workflowId: string,
  runId: string | undefined,
): UseEventHistoryResult {
  const [events, setEvents] = useState<HistoryEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const fetchData = async (): Promise<void> => {
      let resolvedRunId = runId;
      if (!resolvedRunId) {
        const list = await client.listWorkflows(namespace, undefined, 100);
        const match = list.executions.find((e) => e.execution.workflowId === workflowId);
        if (!match) throw new Error(`Workflow not found: ${workflowId}`);
        resolvedRunId = match.execution.runId;
      }
      const res = await client.getHistory(namespace, workflowId, resolvedRunId);
      if (!cancelled) {
        setEvents(res.history?.events ?? res.events ?? []);
        setLoading(false);
      }
    };

    fetchData().catch((err: unknown) => {
      if (!cancelled) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [client, namespace, workflowId, runId]);

  return { events, loading, error };
}
