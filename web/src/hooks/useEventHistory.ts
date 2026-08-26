import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useEventHistory(namespace: string, wfId: string, runId: string) {
  return useQuery({
    queryKey: ['event-history', namespace, wfId, runId],
    queryFn: () => api.getWorkflowHistory(namespace, wfId, runId),
  });
}
