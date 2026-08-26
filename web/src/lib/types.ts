export interface WorkflowExecution {
  execution: { workflowId: string; runId: string };
  type: { name: string };
  startTime: string;
  closeTime: string | null;
  status: string;
  taskQueue: string;
  historyLength: number;
}

export interface WorkflowListResponse {
  executions: WorkflowExecution[];
  nextPageToken: string;
}

export interface NamespaceInfo {
  namespaceInfo: {
    name: string;
    id: string;
    description: string;
  };
  config: {
    workflowExecutionRetentionTtl: { days: number };
  };
}

export interface NamespaceListResponse {
  namespaces: NamespaceInfo[];
}

export interface WorkflowDetailResponse {
  workflowExecutionInfo: {
    execution: { workflowId: string; runId: string };
    type: { name: string };
    startTime: string;
    closeTime: string | null;
    status: string;
    taskQueue: string;
  };
}

export interface HistoryEvent {
  eventId: string;
  eventTime: string;
  eventType: string;
  attributes: Record<string, unknown>;
}

export interface HistoryResponse {
  history: { events: HistoryEvent[] };
}

export interface SettingsResponse {
  auth: { enabled: boolean };
  codec: { endpoint: string };
  version: string;
}
