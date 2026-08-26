import type { SettingsResponse, NamespaceListResponse, WorkflowListResponse, WorkflowDetailResponse, HistoryResponse } from './types';

const BASE = '';

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(`${BASE}${url}`);
  if (!res.ok) throw new Error(`API error: ${res.status} ${res.statusText}`);
  return res.json() as Promise<T>;
}

export const api = {
  getSettings: () => fetchJSON<SettingsResponse>('/api/v1/settings'),
  getNamespaces: () => fetchJSON<NamespaceListResponse>('/api/v1/namespaces'),
  getWorkflows: (ns: string, params?: { status?: string; pageSize?: number }) => {
    const search = new URLSearchParams();
    if (params?.pageSize) search.set('pageSize', String(params.pageSize));
    if (params?.status) search.set('status', params.status);
    const qs = search.toString();
    return fetchJSON<WorkflowListResponse>(`/api/v1/namespaces/${ns}/workflows${qs ? '?' + qs : ''}`);
  },
  getWorkflowDetail: (ns: string, wfId: string, runId: string) =>
    fetchJSON<WorkflowDetailResponse>(`/api/v1/namespaces/${ns}/workflows/${wfId}/${runId}`),
  getWorkflowHistory: (ns: string, wfId: string, runId: string) =>
    fetchJSON<HistoryResponse>(`/api/v1/namespaces/${ns}/workflows/${wfId}/${runId}/history`),
};
