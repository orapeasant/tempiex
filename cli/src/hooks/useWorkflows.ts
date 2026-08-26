import { useEffect, useState } from 'react';
import { TempiexHttpClient, WorkflowsResponse } from '../client/http.js';

interface UseWorkflowsResult {
  data: WorkflowsResponse | null;
  loading: boolean;
  error: Error | null;
}

export function useWorkflows(
  client: TempiexHttpClient,
  namespace: string,
  query?: string,
  limit?: number,
): UseWorkflowsResult {
  const [data, setData] = useState<WorkflowsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client
      .listWorkflows(namespace, query, limit)
      .then((res) => {
        if (!cancelled) {
          setData(res);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [client, namespace, query, limit]);

  return { data, loading, error };
}
