import { useEffect, useState } from 'react';
import { TempiexHttpClient, WorkflowExecutionInfo } from '../client/http.js';

interface UseWorkflowDetailResult {
  data: WorkflowExecutionInfo | null;
  loading: boolean;
  error: Error | null;
}

export function useWorkflowDetail(
  client: TempiexHttpClient,
  namespace: string,
  workflowId: string,
  runId: string | undefined,
): UseWorkflowDetailResult {
  const [data, setData] = useState<WorkflowExecutionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const fetchData = async (): Promise<void> => {
      let resolvedRunId = runId;
      if (!resolvedRunId) {
        // Get the latest run by listing workflows and finding this wfId
        const list = await client.listWorkflows(namespace, undefined, 100);
        const match = list.executions.find((e) => e.execution.workflowId === workflowId);
        if (!match) throw new Error(`Workflow not found: ${workflowId}`);
        resolvedRunId = match.execution.runId;
      }
      return client.getWorkflow(namespace, workflowId, resolvedRunId).then((res) => {
        if (!cancelled) {
          setData(res);
          setLoading(false);
        }
      });
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

  return { data, loading, error };
}
