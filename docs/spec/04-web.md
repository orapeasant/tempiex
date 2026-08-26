# 04 — Web: React UI Specification

## Overview

**Component**: Tempiex Web UI  
**Folder**: `web/`  
**Language**: TypeScript (strict)  
**Framework**: React 18 + Vite  
**UI Library**: shadcn/ui + Radix UI + Tailwind CSS  
**Package Manager**: pnpm  
**Implementation Priority**: 4 — Depends on server (HTTP API)

## Purpose

`web` is a single-page application (SPA) that provides a modern, responsive interface for managing Tempiex workflows, namespaces, schedules, and clusters. It communicates exclusively with the `server` component via HTTP/REST.

Users can:
- View, search, and filter workflow executions
- Inspect event history, stack traces, and payloads
- Manage namespaces and their configurations
- Create, view, and control schedules
- Monitor cluster health
- Interact with workflows via signals, queries, and updates
- Visualize workflow execution graphs

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    Web Browser                       │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │       React Router v6 (SPA routing)          │   │
│  │  /namespaces/:ns/workflows                   │   │
│  │  /namespaces/:ns/schedules                   │   │
│  │  /namespaces/:ns/settings                    │   │
│  │  /cluster                                    │   │
│  └───────────────────┬──────────────────────────┘   │
│                      │                               │
│  ┌───────────────────▼──────────────────────────┐   │
│  │       React Components (shadcn/ui)           │   │
│  │  Workflow List │ Workflow Detail │ History    │   │
│  │  Namespace Mgr │ Schedule Editor│ Cluster    │   │
│  └───────────────────┬──────────────────────────┘   │
│                      │                               │
│  ┌───────────────────▼──────────────────────────┐   │
│  │     State Layer                              │   │
│  │  React Query — server state (cache+refetch)  │   │
│  │  Zustand — client-only UI state              │   │
│  └───────────────────┬──────────────────────────┘   │
│                      │                               │
│  ┌───────────────────▼──────────────────────────┐   │
│  │     API Client (fetch wrapper)               │   │
│  │  Request/response transforms, error handling │   │
│  │  Polling for in-progress workflows           │   │
│  └───────────────────┬──────────────────────────┘   │
│                      │ HTTP/REST                      │
└──────────────────────┼───────────────────────────────┘
                       ▼
              ┌────────────────┐
              │  server :8080  │
              └────────────────┘
```

## Technical Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | TypeScript | 5.x strict |
| UI Framework | React | 18+ |
| Build Tool | Vite | 5+ |
| Routing | React Router | v6 |
| Component Library | shadcn/ui | latest |
| Primitive Components | Radix UI | latest |
| CSS | Tailwind CSS | 3.x |
| Server State | @tanstack/react-query | v5 |
| Client State | Zustand | v4 |
| Forms | react-hook-form + Zod | |
| Code Editor | CodeMirror 6 | 6.x |
| Date/Time | date-fns + date-fns-tz | 3.x |
| Icons | lucide-react | latest |
| Testing | Vitest + React Testing Library | |
| E2E Testing | Playwright | 1.x |
| Linter | ESLint + @typescript-eslint | 9.x |
| Formatter | Prettier | 3.x |

## Folder Structure

```
web/
├── src/
│   ├── main.tsx                  # App entry point
│   ├── App.tsx                   # Root component, Router setup
│   ├── routes/                   # Page-level route components
│   │   ├── namespaces/
│   │   │   └── [namespace]/
│   │   │       ├── workflows/
│   │   │       │   ├── WorkflowList.tsx
│   │   │       │   └── [workflowId]/
│   │   │       │       ├── WorkflowDetail.tsx
│   │   │       │       └── history/
│   │   │       │           └── EventHistory.tsx
│   │   │       ├── schedules/
│   │   │       │   ├── ScheduleList.tsx
│   │   │       │   └── [scheduleId]/
│   │   │       │       └── ScheduleDetail.tsx
│   │   │       └── settings/
│   │   │           └── NamespaceSettings.tsx
│   │   ├── cluster/
│   │   │   └── ClusterDashboard.tsx
│   │   └── Login.tsx
│   │
│   ├── components/               # Reusable UI components
│   │   ├── workflow/
│   │   │   ├── WorkflowStatus.tsx
│   │   │   ├── WorkflowActions.tsx
│   │   │   ├── EventHistoryTable.tsx
│   │   │   ├── PayloadViewer.tsx
│   │   │   └── WorkflowGraph.tsx
│   │   ├── schedule/
│   │   │   ├── CronEditor.tsx
│   │   │   └── ScheduleCalendar.tsx
│   │   ├── namespace/
│   │   │   └── NamespaceSelector.tsx
│   │   └── common/
│   │       ├── DataTable.tsx
│   │       ├── SearchBar.tsx
│   │       ├── Pagination.tsx
│   │       ├── CodeEditor.tsx
│   │       ├── DateTimeDisplay.tsx
│   │       └── ErrorBoundary.tsx
│   │
│   ├── lib/
│   │   ├── api/                  # API client layer
│   │   │   ├── client.ts         # fetch wrapper (base URL, CSRF header, error handling)
│   │   │   ├── workflows.ts      # workflow API calls
│   │   │   ├── namespaces.ts     # namespace API calls
│   │   │   ├── schedules.ts      # schedule API calls
│   │   │   └── cluster.ts        # cluster API calls
│   │   ├── hooks/                # React Query hooks per domain
│   │   │   ├── useWorkflows.ts
│   │   │   ├── useWorkflowDetail.ts
│   │   │   ├── useEventHistory.ts
│   │   │   ├── useNamespaces.ts
│   │   │   └── useCluster.ts
│   │   ├── store/                # Zustand stores
│   │   │   ├── namespaceStore.ts
│   │   │   └── uiStore.ts        # sidebar, theme, etc.
│   │   ├── types/                # TypeScript interfaces
│   │   │   ├── workflow.ts
│   │   │   ├── namespace.ts
│   │   │   ├── schedule.ts
│   │   │   └── cluster.ts
│   │   └── utils/
│   │       ├── datetime.ts
│   │       ├── duration.ts
│   │       └── payload.ts        # JSON / base64 payload formatting
│   │
│   └── assets/                   # Static assets
│
├── test/
│   ├── unit/                     # Vitest unit tests (co-located or here)
│   └── e2e/                      # Playwright E2E tests
│       ├── workflows.spec.ts
│       ├── namespaces.spec.ts
│       └── fixtures/
│
├── public/
│   └── favicon.ico
│
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── eslint.config.mjs
├── playwright.config.ts
├── package.json
└── pnpm-lock.yaml
```

## Key Pages & Components

### Workflow List (`/namespaces/:ns/workflows`)
- Searchable, filterable table with React Query + infinite scroll / pagination
- Filter by: status, task queue, workflow type, search attribute, date range
- Columns: workflow ID, type, status, start time, end time, run ID
- Actions: terminate, signal (bulk)
- Bulk operations: select multiple rows → bulk terminate or cancel via confirmation dialog.

### Workflow Detail (`/namespaces/:ns/workflows/:wfId/:runId`)
- Header: status badge, timing, task queue, workflow type
- Tabs: Event History | Pending Activities | Stack Trace | Queries | Input/Result
- Actions: terminate, signal, query, update, reset

### Event History (`history/`)
- Timeline view with visual event progression
- Expandable rows showing full event attributes
- Filter events by type (e.g. show only ActivityTask events)
- Compact / expanded view toggle
- Stack trace viewer for workflow failures
- Search within event payloads
- CodeMirror JSON viewer for payload inspection

### Namespace Settings
- Retention, archival config, search attributes, RBAC

### Schedule Editor
- Cron expression editor with human-readable preview
- Calendar visualization of upcoming runs
- Supports both cron expressions and interval/calendar specifications.

### Cluster Dashboard
- Service health, shard counts, replication status

## State Management

### React Query (server state)
- All API data fetched and cached via React Query hooks.
- Stale-while-revalidate with configurable refetch intervals for in-progress workflows.
- Optimistic updates for workflow actions (terminate, signal).

### Zustand (client state)
- Selected namespace (persisted to localStorage).
- Sidebar collapsed state, theme preference.
- No server data in Zustand.

## API Client

```typescript
// lib/api/client.ts
const API_BASE = import.meta.env.VITE_API_BASE ?? '';

export async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}/api/v1${path}`, {
    credentials: 'include',
    headers: { 'X-CSRF-Token': getCsrfToken() },
  });
  if (!res.ok) throw new ApiError(res.status, await res.json());
  return res.json();
}
```

CSRF token is read from a cookie set by `server` on every response.

## Vite Configuration

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  build: {
    outDir: 'dist',
  },
});
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE` | `''` (same origin) | API base URL |
| `VITE_TEMPIEX_VERSION` | `''` | Display version |

## shadcn/ui Usage

Components are added via the shadcn CLI into `src/components/ui/`. Do not edit generated files directly; extend via composition in `src/components/`.

```bash
pnpm dlx shadcn@latest add button table dialog sheet select badge
```

## Build & Test Commands

```bash
cd web/

# Install
pnpm install

# Dev server (requires server running at :8080)
pnpm dev

# Type check
pnpm typecheck        # tsc --noEmit

# Lint
pnpm lint             # eslint src/

# Format check
pnpm format:check     # prettier --check

# Unit tests
pnpm test             # vitest run

# Unit tests (watch)
pnpm test:watch       # vitest

# E2E tests (requires server + tempiex running)
pnpm test:e2e         # playwright test

# Production build
pnpm build            # outputs to dist/

# Preview production build
pnpm preview
```

## Linting & Formatting Config

```json
// eslint.config.mjs (flat config)
// rules: @typescript-eslint/recommended + react-hooks + import order
```

```json
// .prettierrc
{
  "semi": true,
  "singleQuote": true,
  "trailingComma": "all",
  "printWidth": 100
}
```

## Testing Strategy

- **Unit**: Vitest + React Testing Library. Test hooks, utils, and complex components.
- **E2E**: Playwright. Cover critical user flows (list workflows, inspect detail, send signal, terminate).
- No snapshot tests — prefer behavior assertions.

## Performance

### Code Splitting
- Route-based splitting via React Router lazy loading.
- CodeMirror 6 editor loaded only when a payload viewer is mounted.

### Virtual Scrolling
- Event history tables use virtual scrolling (react-virtual or similar) for histories with thousands of events.

### Caching
- React Query caches workflow list, namespace list, and cluster info with configurable stale times.
- User preferences (selected namespace, theme) persisted to localStorage.

### Polling Intervals
- In-progress workflow detail: 5s refetch interval.
- Workflow list: 15s background refetch.
- Completed/terminated workflows: no polling.

## Browser Support

| Browser | Supported versions |
|---------|--------------------|
| Chrome | Last 2 |
| Firefox | Last 2 |
| Safari | Last 2 |
| Edge | Last 2 |

## Security

### Content Security Policy
The server sets `Content-Security-Policy: default-src 'self'` in production. No inline scripts in the production build.

### XSS Protection
- User-supplied payload content rendered via CodeMirror (sandboxed editor), never via `dangerouslySetInnerHTML`.
- Markdown memos sanitized with `hast-util-sanitize` before rendering.

### Authentication
- OIDC handled by the `server` component; web only stores the session cookie.
- CSRF token fetched on load; sent in `X-CSRF-Token` header for all mutations.

## Internationalization

All user-facing strings use `i18next`. English is the default and only shipped locale; the setup is extensible. String keys follow the pattern `<domain>.<page>.<key>` (e.g. `workflow.list.title`).

```typescript
import { useTranslation } from 'react-i18next';
const { t } = useTranslation();
<h1>{t('workflow.list.title')}</h1>
```

## Theming

Dark/light mode follows the system `prefers-color-scheme` media query by default. The user can override with a toggle persisted to `localStorage`. Tailwind's `darkMode: 'class'` strategy is used; the root `<html>` element gets `class="dark"` or `class="light"`.

## Accessibility

- All interactive shadcn/Radix components are accessible by default (ARIA, keyboard nav).
- Color contrast meets WCAG AA.
- Route changes announce to screen readers via `aria-live`.

## Dependencies (key)

```json
{
  "react": "^18",
  "react-dom": "^18",
  "react-router-dom": "^6",
  "@tanstack/react-query": "^5",
  "zustand": "^4",
  "react-hook-form": "^7",
  "zod": "^3",
  "@codemirror/state": "^6",
  "@codemirror/view": "^6",
  "@codemirror/lang-json": "^6",
  "date-fns": "^3",
  "date-fns-tz": "^3",
  "lucide-react": "latest",
  "tailwindcss": "^3",
  "class-variance-authority": "^0.7",
  "clsx": "^2",
  "tailwind-merge": "^2"
}
```

```json
{
  "devDependencies": {
    "vite": "^5",
    "@vitejs/plugin-react": "^4",
    "typescript": "^5",
    "vitest": "^1",
    "@testing-library/react": "^14",
    "@playwright/test": "^1",
    "eslint": "^9",
    "@typescript-eslint/eslint-plugin": "^7",
    "prettier": "^3"
  }
}
```
