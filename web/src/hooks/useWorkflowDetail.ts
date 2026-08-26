import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useWorkflowDetail(namespace: string, wfId: string, runId: string) {
  return useQuery({
    queryKey: ['workflow-detail', namespace, wfId, runId],
    queryFn: () => api.getWorkflowDetail(namespace, wfId, runId),
    refetchInterval: (query) => {
      const data = query.state.data;
      return data?.workflowExecutionInfo?.status === 'Running' ? 5000 : false;
    },
  });
}
