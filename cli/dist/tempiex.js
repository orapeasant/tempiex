#!/usr/bin/env node
import { Command } from 'commander';
import { render, useInput, Box, Text } from 'ink';
import React7, { useState, useEffect } from 'react';
import { existsSync, readFileSync, mkdirSync, writeFileSync } from 'fs';
import { dirname, join } from 'path';
import { homedir } from 'os';
import TOML from '@iarna/toml';
import { z } from 'zod';
import { jsx, jsxs, Fragment } from 'react/jsx-runtime';
import { formatDistanceToNow } from 'date-fns';
import { stringify } from 'yaml';

// src/client/http.ts
var HttpError = class extends Error {
  constructor(status, body, message) {
    super(message);
    this.status = status;
    this.body = body;
    this.name = "HttpError";
  }
  status;
  body;
};
var TempiexHttpClient = class {
  baseUrl;
  apiKey;
  constructor(address, apiKey) {
    const base = address.startsWith("http") ? address : `http://${address}`;
    this.baseUrl = base.replace(/\/$/, "");
    this.apiKey = apiKey;
  }
  headers() {
    const h = { "Content-Type": "application/json" };
    if (this.apiKey) h["Authorization"] = `Bearer ${this.apiKey}`;
    return h;
  }
  async request(path, init) {
    const url = `${this.baseUrl}${path}`;
    const res = await fetch(url, { ...init, headers: { ...this.headers(), ...init?.headers ?? {} } });
    const text = await res.text();
    if (!res.ok) {
      throw new HttpError(res.status, text, `HTTP ${res.status}: ${text}`);
    }
    return JSON.parse(text);
  }
  async listNamespaces() {
    return this.request("/api/v1/namespaces");
  }
  async listWorkflows(ns, query, limit) {
    const params = new URLSearchParams();
    if (query) params.set("query", query);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return this.request(`/api/v1/namespaces/${encodeURIComponent(ns)}/workflows${qs ? "?" + qs : ""}`);
  }
  async getWorkflow(ns, wfId, runId) {
    const res = await this.request(
      `/api/v1/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(wfId)}/${encodeURIComponent(runId)}`
    );
    return res.workflowExecutionInfo ?? res;
  }
  async getHistory(ns, wfId, runId) {
    return this.request(
      `/api/v1/namespaces/${encodeURIComponent(ns)}/workflows/${encodeURIComponent(wfId)}/${encodeURIComponent(runId)}/history`
    );
  }
  async createNamespace(body) {
    return this.request("/api/v1/namespaces", {
      method: "POST",
      body: JSON.stringify(body)
    });
  }
  async getCluster() {
    return this.request("/api/v1/cluster");
  }
  async getSettings() {
    return this.request("/api/v1/settings");
  }
  async health() {
    return this.request("/health");
  }
  async ready() {
    return this.request("/ready");
  }
};
var ProfileSchema = z.object({
  address: z.string().default("http://localhost:8080"),
  namespace: z.string().default("default"),
  apiKey: z.string().optional(),
  tls: z.boolean().default(false),
  tlsCertPath: z.string().optional(),
  tlsKeyPath: z.string().optional(),
  tlsCaPath: z.string().optional(),
  codecEndpoint: z.string().optional(),
  output: z.enum(["text", "json", "yaml"]).default("text")
});
var ConfigSchema = z.object({
  version: z.number().default(1),
  activeProfile: z.string().default("default"),
  profiles: z.record(ProfileSchema).default({})
});
z.object({
  address: z.string().optional(),
  namespace: z.string().optional(),
  apiKey: z.string().optional(),
  tls: z.boolean().optional(),
  tlsCertPath: z.string().optional(),
  tlsKeyPath: z.string().optional(),
  tlsCaPath: z.string().optional(),
  codecEndpoint: z.string().optional(),
  output: z.enum(["text", "json", "yaml"]).optional(),
  color: z.enum(["always", "never", "auto"]).optional(),
  profile: z.string().optional(),
  logLevel: z.enum(["debug", "info", "warn", "error", "never"]).optional()
});

// src/config/loader.ts
function getConfigPath() {
  return process.env["TEMPIEX_CONFIG_FILE"] ?? join(homedir(), ".config", "tempiex", "config.toml");
}
var defaultConfig = {
  version: 1,
  activeProfile: "default",
  profiles: {
    default: {
      address: "http://localhost:8080",
      namespace: "default",
      tls: false,
      output: "text"
    }
  }
};
function loadConfig() {
  const configPath = getConfigPath();
  if (!existsSync(configPath)) {
    saveConfig(defaultConfig);
    return defaultConfig;
  }
  try {
    const raw = readFileSync(configPath, "utf-8");
    const parsed = TOML.parse(raw);
    return ConfigSchema.parse(parsed);
  } catch {
    return defaultConfig;
  }
}
function saveConfig(config) {
  const configPath = getConfigPath();
  mkdirSync(dirname(configPath), { recursive: true });
  writeFileSync(configPath, TOML.stringify(config), "utf-8");
}
function collectParentOpts(cmd) {
  const chain = [];
  let cur = cmd.parent;
  while (cur) {
    chain.unshift(cur.opts());
    cur = cur.parent;
  }
  return Object.assign({}, ...chain);
}
function resolveConfig(opts, config) {
  const cfg = loadConfig();
  const parentOpts = opts._cmd ? collectParentOpts(opts._cmd) : {};
  const merged = { ...parentOpts, ...opts };
  const profileName = merged.profile ?? process.env["TEMPIEX_PROFILE"] ?? cfg.activeProfile ?? "default";
  const profile = cfg.profiles[profileName] ?? defaultConfig.profiles["default"];
  return {
    address: merged.address ?? process.env["TEMPIEX_ADDRESS"] ?? profile.address ?? "http://localhost:8080",
    namespace: merged.namespace ?? process.env["TEMPIEX_NAMESPACE"] ?? profile.namespace ?? "default",
    apiKey: merged.apiKey ?? process.env["TEMPIEX_API_KEY"] ?? profile.apiKey,
    tls: merged.tls ?? profile.tls ?? false,
    tlsCertPath: merged.tlsCertPath ?? process.env["TEMPIEX_TLS_CLIENT_CERT_PATH"] ?? profile.tlsCertPath,
    tlsKeyPath: merged.tlsKeyPath ?? process.env["TEMPIEX_TLS_CLIENT_KEY_PATH"] ?? profile.tlsKeyPath,
    tlsCaPath: merged.tlsCaPath ?? process.env["TEMPIEX_TLS_SERVER_CA_CERT_PATH"] ?? profile.tlsCaPath,
    codecEndpoint: merged.codecEndpoint ?? process.env["TEMPIEX_CODEC_ENDPOINT"] ?? profile.codecEndpoint,
    output: merged.output ?? profile.output ?? "text"
  };
}
function setProfileKey(profileName, key, value, config) {
  const cfg = config ?? loadConfig();
  if (!cfg.profiles[profileName]) {
    cfg.profiles[profileName] = {
      address: "http://localhost:8080",
      namespace: "default",
      tls: false,
      output: "text"
    };
  }
  cfg.profiles[profileName][key] = value;
  return cfg;
}
function deleteProfileKey(profileName, key, config) {
  const cfg = config ?? loadConfig();
  if (cfg.profiles[profileName]) {
    delete cfg.profiles[profileName][key];
  }
  return cfg;
}
function useWorkflows(client, namespace, query, limit) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client.listWorkflows(namespace, query, limit).then((res) => {
      if (!cancelled) {
        setData(res);
        setLoading(false);
      }
    }).catch((err) => {
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
var STATUS_COLORS = {
  Running: "cyan",
  RUNNING: "cyan",
  Completed: "green",
  COMPLETED: "green",
  Failed: "red",
  FAILED: "red",
  Cancelled: "yellow",
  CANCELLED: "yellow",
  TimedOut: "magenta",
  TIMED_OUT: "magenta",
  Terminated: "red",
  TERMINATED: "red"
};
function StatusBadge({ status }) {
  const label = status ?? "Unknown";
  const color = STATUS_COLORS[label] ?? "white";
  return /* @__PURE__ */ jsx(Text, { color, children: label });
}
var FRAMES = ["\u280B", "\u2819", "\u2839", "\u2838", "\u283C", "\u2834", "\u2826", "\u2827", "\u2807", "\u280F"];
function Spinner({ label }) {
  const [frame, setFrame] = useState(0);
  useEffect(() => {
    const timer = setInterval(() => {
      setFrame((f) => (f + 1) % FRAMES.length);
    }, 80);
    return () => clearInterval(timer);
  }, []);
  return /* @__PURE__ */ jsxs(Text, { children: [
    /* @__PURE__ */ jsxs(Text, { color: "cyan", children: [
      FRAMES[frame],
      " "
    ] }),
    label ?? "Loading\u2026"
  ] });
}
function ErrorPanel({ error, title }) {
  const message = typeof error === "string" ? error : error.message;
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", borderStyle: "round", borderColor: "red", padding: 1, children: [
    /* @__PURE__ */ jsx(Text, { color: "red", bold: true, children: title ?? "\u2717 Error" }),
    /* @__PURE__ */ jsx(Text, { children: message })
  ] });
}
function formatRelative(dateStr) {
  if (!dateStr) return "\u2014";
  try {
    return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
  } catch {
    return dateStr;
  }
}
function formatDateTime(dateStr) {
  if (!dateStr) return "\u2014";
  try {
    return new Date(dateStr).toISOString().replace("T", " ").replace("Z", " UTC");
  } catch {
    return dateStr;
  }
}
function renderJson(data) {
  return JSON.stringify(data, null, 2);
}
function renderYaml(data) {
  return stringify(data);
}
function printOutput(data, format) {
  if (format === "json") {
    console.log(renderJson(data));
  } else if (format === "yaml") {
    console.log(renderYaml(data));
  }
}
function WorkflowList({
  client,
  namespace,
  query,
  limit,
  output,
  onDone
}) {
  const { data, loading, error } = useWorkflows(client, namespace, query, limit);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(data?.executions ?? [], output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, data, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching workflows\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  const executions = data?.executions ?? [];
  const rows = executions.map((e) => ({
    workflowId: e.execution?.workflowId ?? "\u2014",
    runId: e.execution?.runId?.slice(0, 8) ?? "\u2014",
    type: e.type?.name ?? "\u2014",
    status: e.status ?? "\u2014",
    taskQueue: e.taskQueue ?? "\u2014",
    started: formatRelative(e.startTime)
  }));
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsxs(Box, { marginBottom: 1, children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Workflows " }),
      /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
        "(namespace: ",
        namespace,
        ")"
      ] })
    ] }),
    /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Box, { width: 32, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "WORKFLOW ID" }) }),
        /* @__PURE__ */ jsx(Box, { width: 12, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "RUN ID" }) }),
        /* @__PURE__ */ jsx(Box, { width: 24, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "TYPE" }) }),
        /* @__PURE__ */ jsx(Box, { width: 14, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "STATUS" }) }),
        /* @__PURE__ */ jsx(Box, { width: 20, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "TASK QUEUE" }) }),
        /* @__PURE__ */ jsx(Box, { width: 20, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "STARTED" }) })
      ] }),
      /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: "\u2500".repeat(120) }) }),
      rows.length === 0 && /* @__PURE__ */ jsx(Text, { dimColor: true, children: "No workflows found." }),
      rows.map((row, i) => /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Box, { width: 32, children: /* @__PURE__ */ jsx(Text, { children: row.workflowId.slice(0, 30) }) }),
        /* @__PURE__ */ jsx(Box, { width: 12, children: /* @__PURE__ */ jsx(Text, { children: row.runId }) }),
        /* @__PURE__ */ jsx(Box, { width: 24, children: /* @__PURE__ */ jsx(Text, { children: row.type.slice(0, 22) }) }),
        /* @__PURE__ */ jsx(Box, { width: 14, children: /* @__PURE__ */ jsx(StatusBadge, { status: row.status }) }),
        /* @__PURE__ */ jsx(Box, { width: 20, children: /* @__PURE__ */ jsx(Text, { children: row.taskQueue.slice(0, 18) }) }),
        /* @__PURE__ */ jsx(Box, { width: 20, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: row.started }) })
      ] }, i))
    ] }),
    /* @__PURE__ */ jsx(Box, { marginTop: 1, children: /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
      executions.length,
      " workflow(s)"
    ] }) })
  ] });
}
function useWorkflowDetail(client, namespace, workflowId, runId) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    const fetchData = async () => {
      let resolvedRunId = runId;
      if (!resolvedRunId) {
        const list = await client.listWorkflows(namespace, void 0, 100);
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
    fetchData().catch((err) => {
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
function WorkflowDescribe({
  client,
  namespace,
  workflowId,
  runId,
  output,
  onDone
}) {
  const { data, loading, error } = useWorkflowDetail(client, namespace, workflowId, runId);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(data, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, data, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching workflow\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  if (!data) return /* @__PURE__ */ jsx(Text, { children: "Workflow not found." });
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Box, { marginBottom: 1, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "Workflow Details" }) }),
    /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Workflow ID:  " }),
        /* @__PURE__ */ jsx(Text, { children: data.execution?.workflowId })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Run ID:       " }),
        /* @__PURE__ */ jsx(Text, { children: data.execution?.runId })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Type:         " }),
        /* @__PURE__ */ jsx(Text, { children: data.type?.name ?? "\u2014" })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Status:       " }),
        /* @__PURE__ */ jsx(StatusBadge, { status: data.status })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Task Queue:   " }),
        /* @__PURE__ */ jsx(Text, { children: data.taskQueue ?? "\u2014" })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Start Time:   " }),
        /* @__PURE__ */ jsxs(Text, { children: [
          formatDateTime(data.startTime),
          " (",
          formatRelative(data.startTime),
          ")"
        ] })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "Close Time:   " }),
        /* @__PURE__ */ jsx(Text, { children: data.closeTime ? formatDateTime(data.closeTime) : "\u2014" })
      ] }),
      /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Text, { bold: true, children: "History Len:  " }),
        /* @__PURE__ */ jsx(Text, { children: String(data.historyLength ?? "\u2014") })
      ] })
    ] })
  ] });
}
function WorkflowStart({
  client: _client,
  namespace: _namespace,
  workflowId,
  workflowType,
  taskQueue,
  input: _input,
  output,
  onDone
}) {
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const stub = {
      workflowId: workflowId ?? `wf-${Date.now()}`,
      runId: "not-supported"
    };
    setTimeout(() => {
      console.error(
        `Starting workflow type="${workflowType}" on task-queue="${taskQueue}" is not yet supported via server API.`
      );
      setResult(stub);
      setLoading(false);
    }, 0);
  }, [workflowId, workflowType, taskQueue]);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(result, output);
      }
      onDone();
    }
  }, [loading, result, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Starting workflow\u2026" });
  return /* @__PURE__ */ jsx(Box, { flexDirection: "column", children: /* @__PURE__ */ jsx(Text, { color: "yellow", children: "Not yet supported by server API" }) });
}
function useEventHistory(client, namespace, workflowId, runId) {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    const fetchData = async () => {
      let resolvedRunId = runId;
      if (!resolvedRunId) {
        const list = await client.listWorkflows(namespace, void 0, 100);
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
    fetchData().catch((err) => {
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
function WorkflowShow({
  client,
  namespace,
  workflowId,
  runId,
  reverse,
  output,
  onDone
}) {
  const { events, loading, error } = useEventHistory(client, namespace, workflowId, runId);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(events, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, events, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching history\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  const displayed = reverse ? [...events].reverse() : events;
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsxs(Box, { marginBottom: 1, children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Event History " }),
      /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
        "(",
        workflowId,
        ")"
      ] })
    ] }),
    displayed.map((ev, i) => /* @__PURE__ */ jsx(Box, { flexDirection: "column", marginBottom: 0, children: /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Box, { width: 6, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: String(ev.eventId) }) }),
      /* @__PURE__ */ jsx(Box, { width: 20, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: formatDateTime(ev.eventTime) }) }),
      /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsx(Text, { color: "cyan", children: String(ev.eventType ?? "\u2014") }) })
    ] }) }, i)),
    events.length === 0 && /* @__PURE__ */ jsx(Text, { dimColor: true, children: "No events found." }),
    /* @__PURE__ */ jsx(Box, { marginTop: 1, children: /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
      events.length,
      " event(s)"
    ] }) })
  ] });
}
function WorkflowCount({
  client,
  namespace,
  query,
  output,
  onDone
}) {
  const [count, setCount] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    client.listWorkflows(namespace, query, 1e3).then((res) => {
      if (!cancelled) {
        setCount(res.executions?.length ?? 0);
        setLoading(false);
      }
    }).catch((err) => {
      if (!cancelled) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [client, namespace, query]);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput({ count }, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, count, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Counting workflows\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  return /* @__PURE__ */ jsxs(Box, { children: [
    /* @__PURE__ */ jsx(Text, { bold: true, children: "Count: " }),
    /* @__PURE__ */ jsx(Text, { children: String(count) })
  ] });
}

// src/commands/workflow/index.ts
function registerWorkflowCommands(program2) {
  const wf = program2.command("workflow").description("Manage Workflow Executions");
  wf.command("list").description("List Workflow Executions").option("-q, --query <query>", "SQL-like filter").option("--limit <n>", "Max executions", "50").option("--archived", "Show archived executions").option("-o, --output <format>", "Output format: text|json|yaml").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts, cmd) => {
    const cfg = resolveConfig({ ...opts, _cmd: cmd });
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(WorkflowList, {
        client,
        namespace: cfg.namespace,
        query: opts.query,
        limit: opts.limit ? Number(opts.limit) : 50,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  wf.command("describe").description("Show Workflow Execution details").requiredOption("-w, --workflow-id <id>", "Workflow ID").option("-r, --run-id <runId>", "Run ID").option("-o, --output <format>", "Output format: text|json|yaml").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts, cmd) => {
    const cfg = resolveConfig({ ...opts, _cmd: cmd });
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(WorkflowDescribe, {
        client,
        namespace: cfg.namespace,
        workflowId: opts.workflowId,
        runId: opts.runId,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  wf.command("start").description("Start a Workflow Execution").option("-w, --workflow-id <id>", "Workflow ID").requiredOption("--type <type>", "Workflow type name").requiredOption("-t, --task-queue <queue>", "Task queue name").option("-i, --input <json>", "JSON input").option("-o, --output <format>", "Output format: text|json|yaml").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts, cmd) => {
    const cfg = resolveConfig({ ...opts, _cmd: cmd });
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(WorkflowStart, {
        client,
        namespace: cfg.namespace,
        workflowId: opts.workflowId,
        workflowType: opts.type,
        taskQueue: opts.taskQueue,
        input: opts.input,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  wf.command("show").description("Show Event History").requiredOption("-w, --workflow-id <id>", "Workflow ID").option("-r, --run-id <runId>", "Run ID").option("-f, --follow", "Follow execution progress").option("--reverse", "Newest events first").option("-o, --output <format>", "Output format: text|json|yaml").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts, cmd) => {
    const cfg = resolveConfig({ ...opts, _cmd: cmd });
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(WorkflowShow, {
        client,
        namespace: cfg.namespace,
        workflowId: opts.workflowId,
        runId: opts.runId,
        reverse: opts.reverse,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  wf.command("terminate").description("Terminate a Workflow Execution (stub)").option("-w, --workflow-id <id>", "Workflow ID").option("--reason <reason>", "Reason").action(() => {
    console.log("Not yet supported by server API");
    process.exit(0);
  });
  wf.command("cancel").description("Cancel a Workflow Execution (stub)").option("-w, --workflow-id <id>", "Workflow ID").action(() => {
    console.log("Not yet supported by server API");
    process.exit(0);
  });
  wf.command("signal").description("Signal a Workflow Execution (stub)").option("-w, --workflow-id <id>", "Workflow ID").option("--name <name>", "Signal name").action(() => {
    console.log("Not yet supported by server API");
    process.exit(0);
  });
  wf.command("count").description("Count Workflow Executions").option("-q, --query <query>", "SQL-like filter").option("-o, --output <format>", "Output format: text|json|yaml").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts, cmd) => {
    const cfg = resolveConfig({ ...opts, _cmd: cmd });
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(WorkflowCount, {
        client,
        namespace: cfg.namespace,
        query: opts.query,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
}
function useNamespaces(client) {
  const [namespaces, setNamespaces] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client.listNamespaces().then((res) => {
      if (!cancelled) {
        setNamespaces(res.namespaces ?? []);
        setLoading(false);
      }
    }).catch((err) => {
      if (!cancelled) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [client]);
  return { namespaces, loading, error };
}
function NamespaceList({
  client,
  output,
  onDone
}) {
  const { namespaces, loading, error } = useNamespaces(client);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(namespaces, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, namespaces, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching namespaces\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Box, { marginBottom: 1, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "Namespaces" }) }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Box, { width: 36, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "NAME" }) }),
      /* @__PURE__ */ jsx(Box, { width: 40, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "ID" }) }),
      /* @__PURE__ */ jsx(Box, { width: 12, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "RETENTION" }) })
    ] }),
    /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: "\u2500".repeat(88) }) }),
    namespaces.map((ns, i) => /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Box, { width: 36, children: /* @__PURE__ */ jsx(Text, { children: ns.namespaceInfo.name }) }),
      /* @__PURE__ */ jsx(Box, { width: 40, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: ns.namespaceInfo.id }) }),
      /* @__PURE__ */ jsx(Box, { width: 12, children: /* @__PURE__ */ jsxs(Text, { children: [
        String(ns.config.workflowExecutionRetentionTtl?.days ?? "\u2014"),
        "d"
      ] }) })
    ] }, i)),
    /* @__PURE__ */ jsx(Box, { marginTop: 1, children: /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
      namespaces.length,
      " namespace(s)"
    ] }) })
  ] });
}
function NamespaceDescribe({
  client,
  namespace,
  output,
  onDone
}) {
  const { namespaces, loading, error } = useNamespaces(client);
  const ns = namespaces.find(
    (n) => n.namespaceInfo.name === namespace || n.namespaceInfo.id === namespace
  );
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(ns ?? null, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, ns, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching namespace\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  if (!ns) return /* @__PURE__ */ jsxs(Text, { color: "red", children: [
    "Namespace not found: ",
    namespace
  ] });
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Box, { marginBottom: 1, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "Namespace Details" }) }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Name:       " }),
      /* @__PURE__ */ jsx(Text, { children: ns.namespaceInfo.name })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "ID:         " }),
      /* @__PURE__ */ jsx(Text, { children: ns.namespaceInfo.id })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Description:" }),
      /* @__PURE__ */ jsxs(Text, { children: [
        " ",
        ns.namespaceInfo.description || "\u2014"
      ] })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Retention:  " }),
      /* @__PURE__ */ jsxs(Text, { children: [
        String(ns.config.workflowExecutionRetentionTtl?.days ?? "\u2014"),
        " days"
      ] })
    ] })
  ] });
}
function NamespaceCreate({
  client,
  namespace,
  retention,
  description,
  output: _output,
  onDone
}) {
  const [done, setDone] = useState(false);
  const [error, setError] = useState(null);
  useEffect(() => {
    client.createNamespace({
      name: namespace,
      description: description ?? "",
      workflowExecutionRetentionPeriod: retention ?? "72h"
    }).then(() => {
      setDone(true);
    }).catch((err) => {
      setError(err instanceof Error ? err : new Error(String(err)));
      setDone(true);
    });
  }, [client, namespace, retention, description]);
  useEffect(() => {
    if (done) onDone();
  }, [done, onDone]);
  if (!done) return /* @__PURE__ */ jsx(Spinner, { label: "Creating namespace\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  return /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsxs(Text, { color: "green", children: [
    "\u2713 Namespace created: ",
    namespace
  ] }) });
}

// src/commands/namespace/index.ts
function registerNamespaceCommands(program2) {
  const ns = program2.command("namespace").description("Manage Namespaces");
  ns.command("list").description("List all Namespaces").option("-o, --output <format>", "Output format: text|json|yaml").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(NamespaceList, {
        client,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  ns.command("describe").description("Describe a Namespace").option("-n, --namespace <ns>", "Namespace name or ID").option("-o, --output <format>", "Output format: text|json|yaml").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(NamespaceDescribe, {
        client,
        namespace: cfg.namespace,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  ns.command("create").description("Create a Namespace").requiredOption("-n, --namespace <ns>", "Namespace name").option("--retention <duration>", "Retention period", "72h").option("--description <desc>", "Description").option("-o, --output <format>", "Output format: text|json|yaml").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(NamespaceCreate, {
        client,
        namespace: opts.namespace,
        retention: opts.retention,
        description: opts.description,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  ns.command("delete").description("Delete a Namespace").requiredOption("-n, --namespace <ns>", "Namespace name").option("-y, --yes", "Skip confirmation").action((opts) => {
    console.log(`Not yet supported by server API: ${opts.namespace}`);
    process.exit(0);
  });
}
function useCluster(client) {
  const [cluster, setCluster] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client.getCluster().then((res) => {
      if (!cancelled) {
        setCluster(res.clusterInfo);
        setLoading(false);
      }
    }).catch((err) => {
      if (!cancelled) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [client]);
  return { cluster, loading, error };
}
function ClusterDescribe({
  client,
  output,
  onDone
}) {
  const { cluster, loading, error } = useCluster(client);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(cluster, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, cluster, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Fetching cluster info\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Box, { marginBottom: 1, children: /* @__PURE__ */ jsx(Text, { bold: true, children: "Cluster Information" }) }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "ID:           " }),
      /* @__PURE__ */ jsx(Text, { children: cluster?.clusterId ?? "\u2014" })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Name:         " }),
      /* @__PURE__ */ jsx(Text, { children: cluster?.clusterName ?? "\u2014" })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Version:      " }),
      /* @__PURE__ */ jsx(Text, { children: cluster?.serverVersion ?? "\u2014" })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "History Shards:" }),
      /* @__PURE__ */ jsxs(Text, { children: [
        " ",
        String(cluster?.historyShardCount ?? "\u2014")
      ] })
    ] })
  ] });
}
function ClusterHealth({
  client,
  output,
  onDone
}) {
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  useEffect(() => {
    let cancelled = false;
    client.health().then((res) => {
      if (!cancelled) {
        setHealth(res);
        setLoading(false);
      }
    }).catch((err) => {
      if (!cancelled) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [client]);
  useEffect(() => {
    if (!loading) {
      if (output !== "text") {
        printOutput(health, output);
        onDone();
      } else {
        onDone();
      }
    }
  }, [loading, health, output, onDone]);
  if (output !== "text") return /* @__PURE__ */ jsx(Fragment, {});
  if (loading) return /* @__PURE__ */ jsx(Spinner, { label: "Checking cluster health\u2026" });
  if (error) return /* @__PURE__ */ jsx(ErrorPanel, { error });
  const ok = health?.status === "ok";
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Status:  " }),
      /* @__PURE__ */ jsx(Text, { color: ok ? "green" : "red", children: health?.status ?? "unknown" })
    ] }),
    /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Text, { bold: true, children: "Version: " }),
      /* @__PURE__ */ jsx(Text, { children: health?.version ?? "\u2014" })
    ] })
  ] });
}

// src/commands/cluster/index.ts
function registerClusterCommands(program2) {
  const cluster = program2.command("cluster").description("Cluster operations");
  cluster.command("describe").description("Describe the cluster").option("-o, --output <format>", "Output format: text|json|yaml").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(ClusterDescribe, {
        client,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  cluster.command("health").description("Check cluster health").option("-o, --output <format>", "Output format: text|json|yaml").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
    const { unmount } = render(
      React7.createElement(ClusterHealth, {
        client,
        output: cfg.output,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
}
function TaskQueueDescribe({
  taskQueue,
  namespace,
  onDone
}) {
  useEffect(() => {
    onDone();
  }, [onDone]);
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Text, { color: "yellow", children: "Task queue describe not yet supported by server API" }),
    /* @__PURE__ */ jsxs(Text, { dimColor: true, children: [
      "Task Queue: ",
      taskQueue,
      " | Namespace: ",
      namespace
    ] })
  ] });
}

// src/commands/taskqueue/index.ts
function registerTaskQueueCommands(program2) {
  const tq = program2.command("taskqueue").description("Manage Task Queues");
  tq.command("describe").description("Describe a Task Queue").requiredOption("-t, --task-queue <name>", "Task queue name").option("-n, --namespace <ns>", "Namespace").option("--address <addr>", "Server address").action((opts) => {
    const cfg = resolveConfig(opts);
    const { unmount } = render(
      React7.createElement(TaskQueueDescribe, {
        taskQueue: opts.taskQueue,
        namespace: cfg.namespace,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
}
function ConfigList({ profile, onDone }) {
  useEffect(() => {
    onDone();
  }, [onDone]);
  const config = loadConfig();
  const prof = config.profiles[profile];
  if (!prof) {
    return /* @__PURE__ */ jsxs(Text, { color: "red", children: [
      "Profile not found: ",
      profile
    ] });
  }
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Box, { marginBottom: 1, children: /* @__PURE__ */ jsxs(Text, { bold: true, children: [
      "Config Profile: ",
      profile
    ] }) }),
    Object.entries(prof).map(([k, v]) => /* @__PURE__ */ jsxs(Box, { children: [
      /* @__PURE__ */ jsx(Box, { width: 24, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: k }) }),
      /* @__PURE__ */ jsx(Text, { children: String(v) })
    ] }, k))
  ] });
}
function ConfigGet({ profile, propName, onDone }) {
  useEffect(() => {
    onDone();
  }, [onDone]);
  const config = loadConfig();
  const prof = config.profiles[profile];
  if (!prof) {
    return /* @__PURE__ */ jsxs(Text, { color: "red", children: [
      "Profile not found: ",
      profile
    ] });
  }
  if (!propName) {
    return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
      /* @__PURE__ */ jsxs(Text, { bold: true, children: [
        "Config Profile: ",
        profile
      ] }),
      /* @__PURE__ */ jsx(Text, { children: " " }),
      Object.entries(prof).map(([k, v]) => /* @__PURE__ */ jsxs(Box, { children: [
        /* @__PURE__ */ jsx(Box, { width: 24, children: /* @__PURE__ */ jsx(Text, { children: k }) }),
        /* @__PURE__ */ jsx(Text, { children: String(v ?? "") })
      ] }, k))
    ] });
  }
  const value = prof[propName];
  if (value === void 0) {
    return /* @__PURE__ */ jsxs(Text, { color: "red", children: [
      "Key not found: ",
      propName
    ] });
  }
  return /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsx(Text, { children: String(value) }) });
}
function ConfigSet({ profile, propName, value, onDone }) {
  useEffect(() => {
    const config = loadConfig();
    const updated = setProfileKey(profile, propName, value, config);
    saveConfig(updated);
    onDone();
  }, [profile, propName, value, onDone]);
  return /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsxs(Text, { color: "green", children: [
    "\u2713 Set ",
    profile,
    ".",
    propName,
    " = ",
    value
  ] }) });
}
function ConfigDelete({ profile, propName, onDone }) {
  useEffect(() => {
    const config = loadConfig();
    if (!propName) {
      delete config.profiles[profile];
      saveConfig(config);
    } else {
      const updated = deleteProfileKey(profile, propName, config);
      saveConfig(updated);
    }
    onDone();
  }, [profile, propName, onDone]);
  return /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsxs(Text, { color: "green", children: [
    "\u2713 Deleted ",
    propName ? `${profile}.${propName}` : `profile ${profile}`
  ] }) });
}

// src/commands/config/index.ts
function registerConfigCommands(program2) {
  const cfg = program2.command("config").description("Manage CLI configuration");
  cfg.command("list").description("List all config values for a profile").option("--profile <profile>", "Profile name", "default").action((opts) => {
    const { unmount } = render(
      React7.createElement(ConfigList, {
        profile: opts.profile,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  cfg.command("get").description("Get a config value").option("--profile <profile>", "Profile name", "default").option("-p, --prop <key>", "Property name (optional \u2014 shows all if omitted)").action((opts) => {
    const { unmount } = render(
      React7.createElement(ConfigGet, {
        profile: opts.profile,
        propName: opts.prop,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  cfg.command("set").description("Set a config value").option("--profile <profile>", "Profile name", "default").requiredOption("-p, --prop <key>", "Property name").requiredOption("-v, --value <val>", "Property value").action((opts) => {
    const { unmount } = render(
      React7.createElement(ConfigSet, {
        profile: opts.profile,
        propName: opts.prop,
        value: opts.value,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
  cfg.command("delete").description("Delete a config value or entire profile").option("--profile <profile>", "Profile name", "default").option("-p, --prop <key>", "Property to delete (omit to delete entire profile)").action((opts) => {
    const { unmount } = render(
      React7.createElement(ConfigDelete, {
        profile: opts.profile,
        propName: opts.prop,
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
}
function ServerStartDev({ onDone }) {
  useEffect(() => {
    onDone();
  }, [onDone]);
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", children: [
    /* @__PURE__ */ jsx(Text, { bold: true, children: "Tempiex Dev Server" }),
    /* @__PURE__ */ jsx(Text, { dimColor: true, children: "To start the server, run from the project root:" }),
    /* @__PURE__ */ jsxs(Box, { marginTop: 1, flexDirection: "column", children: [
      /* @__PURE__ */ jsx(Text, { children: "  cd /path/to/tempiex/server" }),
      /* @__PURE__ */ jsx(Text, { children: "  go run . serve" })
    ] }),
    /* @__PURE__ */ jsx(Box, { marginTop: 1, children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: "gRPC: :8133  |  HTTP: :8080" }) })
  ] });
}

// src/commands/server/index.ts
function registerServerCommands(program2) {
  const server = program2.command("server").description("Server management");
  server.command("start-dev").description("Start the development server (instructions)").action(() => {
    const { unmount } = render(
      React7.createElement(ServerStartDev, {
        onDone: () => {
          unmount();
          process.exit(0);
        }
      })
    );
  });
}
function App({ client, onDone }) {
  const { namespaces, loading: nsLoading, error: nsError } = useNamespaces(client);
  const [selectedNsIdx, setSelectedNsIdx] = useState(0);
  const [selectedWfIdx, setSelectedWfIdx] = useState(0);
  const [focus, setFocus] = useState("ns");
  const selectedNs = namespaces[selectedNsIdx];
  const nsName = selectedNs?.namespaceInfo.name ?? "default";
  const { data: wfData, loading: wfLoading } = useWorkflows(client, nsName, void 0, 50);
  const workflows = wfData?.executions ?? [];
  useInput((input, key) => {
    if (input === "q") {
      onDone();
      return;
    }
    if (key.tab) {
      setFocus((f) => f === "ns" ? "wf" : "ns");
      return;
    }
    if (focus === "ns") {
      if (key.upArrow) setSelectedNsIdx((i) => Math.max(0, i - 1));
      if (key.downArrow) setSelectedNsIdx((i) => Math.min(namespaces.length - 1, i + 1));
      if (key.return) setFocus("wf");
    } else {
      if (key.upArrow) setSelectedWfIdx((i) => Math.max(0, i - 1));
      if (key.downArrow) setSelectedWfIdx((i) => Math.min(workflows.length - 1, i + 1));
      if (key.escape) setFocus("ns");
    }
  });
  if (nsLoading) return /* @__PURE__ */ jsx(Spinner, { label: "Loading\u2026" });
  if (nsError) return /* @__PURE__ */ jsx(ErrorPanel, { error: nsError });
  return /* @__PURE__ */ jsxs(Box, { flexDirection: "column", height: 30, children: [
    /* @__PURE__ */ jsxs(Box, { marginBottom: 1, children: [
      /* @__PURE__ */ jsx(Text, { bold: true, color: "cyan", children: "Tempiex TUI " }),
      /* @__PURE__ */ jsx(Text, { dimColor: true, children: "\u2014 [Tab] switch panel | [\u2191\u2193] navigate | [q] quit" })
    ] }),
    /* @__PURE__ */ jsxs(Box, { flexDirection: "row", flexGrow: 1, children: [
      /* @__PURE__ */ jsxs(Box, { flexDirection: "column", width: 32, marginRight: 2, children: [
        /* @__PURE__ */ jsxs(Text, { bold: true, underline: true, children: [
          "Namespaces ",
          focus === "ns" ? "\u25CF" : "\u25CB"
        ] }),
        namespaces.slice(0, 20).map((ns, i) => /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsxs(
          Text,
          {
            color: i === selectedNsIdx ? "cyan" : void 0,
            bold: i === selectedNsIdx,
            children: [
              i === selectedNsIdx ? "\u25B6 " : "  ",
              ns.namespaceInfo.name.slice(0, 26)
            ]
          }
        ) }, i))
      ] }),
      /* @__PURE__ */ jsxs(Box, { flexDirection: "column", flexGrow: 1, children: [
        /* @__PURE__ */ jsxs(Text, { bold: true, underline: true, children: [
          "Workflows: ",
          nsName,
          " ",
          focus === "wf" ? "\u25CF" : "\u25CB"
        ] }),
        wfLoading && /* @__PURE__ */ jsx(Spinner, { label: "Loading workflows\u2026" }),
        !wfLoading && workflows.slice(0, 20).map((wf, i) => {
          const isSelected = focus === "wf" && i === selectedWfIdx;
          return /* @__PURE__ */ jsxs(Box, { children: [
            /* @__PURE__ */ jsx(Text, { color: isSelected ? "cyan" : void 0, bold: isSelected, children: isSelected ? "\u25B6 " : "  " }),
            /* @__PURE__ */ jsx(Box, { width: 30, children: /* @__PURE__ */ jsx(Text, { color: isSelected ? "cyan" : void 0, children: (wf.execution?.workflowId ?? "\u2014").slice(0, 28) }) }),
            /* @__PURE__ */ jsx(Box, { width: 16, children: /* @__PURE__ */ jsx(StatusBadge, { status: wf.status }) }),
            /* @__PURE__ */ jsx(Box, { children: /* @__PURE__ */ jsx(Text, { dimColor: true, children: wf.type?.name?.slice(0, 20) ?? "\u2014" }) })
          ] }, i);
        }),
        !wfLoading && workflows.length === 0 && /* @__PURE__ */ jsx(Text, { dimColor: true, children: "No workflows in this namespace." })
      ] })
    ] })
  ] });
}

// src/main.tsx
var program = new Command();
program.name("tempiex").description("Tempiex CLI \u2014 manage workflow executions and namespaces").version("0.1.0").enablePositionalOptions().option("--address <addr>", "Server address (HTTP)").option("-n, --namespace <ns>", "Namespace").option("--api-key <key>", "API key").option("-o, --output <format>", "Output format: text|json|yaml").option("--profile <profile>", "Config profile").option("--log-level <level>", "Log level");
registerWorkflowCommands(program);
registerNamespaceCommands(program);
registerClusterCommands(program);
registerTaskQueueCommands(program);
registerConfigCommands(program);
registerServerCommands(program);
program.action((opts) => {
  const cfg = resolveConfig(opts);
  const client = new TempiexHttpClient(cfg.address, cfg.apiKey);
  const { unmount } = render(
    React7.createElement(App, {
      client,
      onDone: () => {
        unmount();
        process.exit(0);
      }
    })
  );
});
program.parse(process.argv);
