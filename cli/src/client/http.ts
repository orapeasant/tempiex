export class HttpError extends Error {
  constructor(
    public readonly status: number,
    public readonly body: string,
    message: string,
  ) {
    super(message);
    this.name = 'HttpError';
  }
}

export interface NamespaceInfo {
  id: string;
  name: string;
  description?: string;
  data?: Record<string, string>;
}

export interface NamespaceConfig {
  workflowExecutionRetentionTtl?: { days?: number };
}

export interface Namespace {
  namespaceInfo: NamespaceInfo;
  config: NamespaceConfig;
}

export interface NamespacesResponse {
  namespaces: Namespace[];
}

export interface WorkflowExecution {
  workflowId: string;
  runId: string;
}

export interface WorkflowExecutionInfo {
  execution: WorkflowExecution;
  type?: { name: string };
  startTime?: string;
  closeTime?: string;
  status?: string;
  taskQueue?: string;
  historyLength?: number;
  memo?: Record<string, unknown>;
  searchAttributes?: Record<string, unknown>;
}

export interface WorkflowsResponse {
  executions: WorkflowExecutionInfo[];
  nextPageToken?: string;
}

export interface HistoryEvent {
  eventId: number;
  eventTime?: string;
  eventType?: string;
  [key: string]: unknown;
}

export interface HistoryResponse {
  history?: { events: HistoryEvent[] };
  events?: HistoryEvent[];
}

export interface ClusterInfo {
  clusterId?: string;
  clusterName?: string;
  historyShardCount?: number;
  serverVersion?: string;
}

export interface ClusterResponse {
  clusterInfo: ClusterInfo;
}

export interface HealthResponse {
  status: string;
  version?: string;
}

export interface Settings {
  [key: string]: unknown;
}

export class TempiexHttpClient {
  private baseUrl: string;
  private apiKey?: string;

  constructor(address: string, apiKey?: string) {
    const base = address.startsWith('http') ? address : `http://${address}`;
    this.baseUrl = base.replace(/\/$/, '');
    this.apiKey = apiKey;
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.apiKey) h['Authorization'] = `Bearer ${this.apiKey}`;
    return h;
  }

  private async request<T>(path: string, init?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const res = await fetch(url, { ...init, headers: { ...this.headers(), ...(init?.headers as Record<string, string> ?? {}) } });
    const text = await res.text();
    if (!res.ok) {
      throw new HttpError(res.status, text, `HTTP ${res.status}: ${text}`);
    }
    return JSON.parse(text) as T;
  }

  async listNamespaces(): Promise<NamespacesResponse> {
    return this.request<NamespacesResponse>('/api/v1/namespaces');
  }

  async listWorkflows(ns: string, query?: string, limit?: number): Promise<WorkflowsResponse> {
    const params = new URLSearchParams();
    if (query) params.set('query', query);
    if (limit) params.set('limit', String(limit));
    const qs = params.toString();
    return this.request<WorkflowsResponse>(`/api/v1/namespaces/${encodeURIComponent(ns)}/workflows${qs ? '?' + qs : ''}`);
  }

  async getWorkflow(ns: string, wfId: string, runId: string): Promise<WorkflowExecutionInfo> {
    const res = await this.request<{ workflowExecutionInfo?: WorkflowExecutionInfo } & WorkflowExecutionInfo>(
      `/api/v1/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(wfId)}/${encodeURIComponent(runId)}`,
    );
    // API wraps in workflowExecutionInfo
    return (res as { workflowExecutionInfo?: WorkflowExecutionInfo }).workflowExecutionInfo ?? res;
  }

  async getHistory(ns: string, wfId: string, runId: string): Promise<HistoryResponse> {
    return this.request<HistoryResponse>(
      `/api/v1/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(wfId)}/${encodeURIComponent(runId)}/history`,
    );
  }

  async createNamespace(body: Record<string, unknown>): Promise<unknown> {
    return this.request<unknown>('/api/v1/namespaces', {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  async getCluster(): Promise<ClusterResponse> {
    return this.request<ClusterResponse>('/api/v1/cluster');
  }

  async getSettings(): Promise<Settings> {
    return this.request<Settings>('/api/v1/settings');
  }

  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>('/health');
  }

  async ready(): Promise<HealthResponse> {
    return this.request<HealthResponse>('/ready');
  }
}
