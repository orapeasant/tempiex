import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useWorkflows(namespace: string, status?: string) {
  return useQuery({
    queryKey: ['workflows', namespace, status],
    queryFn: () => api.getWorkflows(namespace, { status: status || undefined, pageSize: 50 }),
    refetchInterval: 15000,
  });
}
