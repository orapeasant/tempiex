# 04 - Temporal Web UI Specification

## Overview

**Component**: Temporal Web UI  
**Repository**: `ui/`  
**Language**: TypeScript  
**Framework**: SvelteKit + Svelte 5  
**Type**: Single Page Application (SPA)  
**Implementation Priority**: 4 (Depends on UI Server)

## Purpose

Temporal Web UI is a modern, responsive web application that provides a graphical interface for managing and monitoring Temporal workflows, namespaces, schedules, and clusters. It enables users to:

- **View and search workflows** with advanced filtering
- **Inspect workflow execution** with event history and stack traces
- **Manage namespaces** and their configurations
- **Monitor cluster health** and metrics
- **Interact with workflows** via signals, queries, and operations
- **Manage schedules** for periodic workflow execution
- **Visualize workflow graphs** and execution paths

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Web Browser                            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           SvelteKit Application                      │  │
│  │                                                      │  │
│  │  ┌────────────────────────────────────────────┐     │  │
│  │  │           Router (SvelteKit)               │     │  │
│  │  │  • /namespaces/:namespace/workflows       │     │  │
│  │  │  • /namespaces/:namespace/schedules       │     │  │
│  │  │  • /namespaces/:namespace/settings        │     │  │
│  │  └───────────────┬────────────────────────────┘     │  │
│  │                  │                                   │  │
│  │  ┌───────────────▼────────────────────────────┐     │  │
│  │  │         Svelte 5 Components                │     │  │
│  │  │  • Workflow List                           │     │  │
│  │  │  • Workflow Detail                         │     │  │
│  │  │  • Event History                           │     │  │
│  │  │  • Namespace Manager                       │     │  │
│  │  │  • Schedule Editor                         │     │  │
│  │  └───────────────┬────────────────────────────┘     │  │
│  │                  │                                   │  │
│  │  ┌───────────────▼────────────────────────────┐     │  │
│  │  │        State Management                    │     │  │
│  │  │  • Stores (Svelte runes)                   │     │  │
│  │  │  • Context providers                       │     │  │
│  │  │  • Local storage                           │     │  │
│  │  └───────────────┬────────────────────────────┘     │  │
│  │                  │                                   │  │
│  │  ┌───────────────▼────────────────────────────┐     │  │
│  │  │         API Client Layer                   │     │  │
│  │  │  • Fetch wrapper                           │     │  │
│  │  │  • Request/response transforms             │     │  │
│  │  │  • Error handling                          │     │  │
│  │  │  • Polling & SSE                           │     │  │
│  │  └───────────────┬────────────────────────────┘     │  │
│  │                  │ HTTP/REST                         │  │
│  └──────────────────┼──────────────────────────────────┘  │
└─────────────────────┼─────────────────────────────────────┘
                      │
                      ▼
            ┌─────────────────┐
            │   UI Server     │
            │   (Port 8080)   │
            └─────────────────┘
```

### Component Hierarchy

```
src/
├── routes/                    # SvelteKit routes (pages)
│   ├── (app)/                # Main application routes
│   │   ├── namespaces/
│   │   │   └── [namespace]/
│   │   │       ├── workflows/
│   │   │       ├── schedules/
│   │   │       ├── settings/
│   │   │       └── import-export/
│   │   └── cluster/
│   └── (login)/              # Login/auth routes
│
├── lib/
│   ├── components/           # Reusable Svelte components
│   │   ├── workflow/        # Workflow-specific components
│   │   ├── namespace/       # Namespace components
│   │   ├── schedule/        # Schedule components
│   │   └── common/          # Shared UI components
│   ├── holocene/            # Holocene design system
│   ├── stores/              # Global state stores
│   ├── services/            # API service layer
│   ├── models/              # TypeScript types/interfaces
│   ├── utilities/           # Helper functions
│   ├── i18n/                # Internationalization
│   └── theme/               # Theme configuration
│
└── app.html                 # HTML template
```

## Technical Stack

### Core Technologies

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | TypeScript | 6.0.3 |
| Framework | SvelteKit | 2.57.1 |
| UI Library | Svelte | 5.55.7 |
| Build Tool | Vite | 6.4.2 |
| Package Manager | pnpm | 10.10.0 |
| Node.js | Node.js | >=22.14.0 |

### UI Framework & Styling

| Component | Technology | Version |
|-----------|-----------|---------|
| CSS Framework | TailwindCSS | 3.4.1 |
| CSS Utility | class-variance-authority | 0.7.0 |
| CSS Merge | tailwind-merge | 1.14.0 |
| Design System | Holocene (custom) | - |
| Icons | Custom SVG components | - |

### Code Editor (Monaco-based)

| Component | Technology | Version |
|-----------|-----------|---------|
| Code Editor | CodeMirror 6 | 6.x |
| Autocomplete | @codemirror/autocomplete | 6.18.7 |
| Language Support | @codemirror/lang-* | 6.x |
| Syntax Highlighting | @lezer/highlight | 1.1.3 |

### Forms & Validation

| Component | Technology | Version |
|-----------|-----------|---------|
| Form Library | sveltekit-superforms | 2.30.2 |
| Validation | Zod | 4.1.12 |

### Date & Time

| Component | Technology | Version |
|-----------|-----------|---------|
| Date Utilities | date-fns | 2.29.x |
| Timezone Support | date-fns-tz | 1.3.x |
| Cron Expression | cronstrue | 3.14.0 |

### Data Handling

| Component | Technology | Version |
|-----------|-----------|---------|
| BigInt Support | json-bigint | 1.0.0 |
| Utilities | es-toolkit | 1.49.0 |

### Markdown & Content

| Component | Technology | Version |
|-----------|-----------|---------|
| Markdown Parser | mdast-util-from-markdown | 2.0.1 |
| Table Support | mdast-util-gfm-table | 2.0.0 |
| HTML Rendering | hast-util-to-html | 9.0.1 |
| Sanitization | hast-util-sanitize | 5.0.1 |

### Development Tools

| Component | Technology | Version |
|-----------|-----------|---------|
| Testing Framework | Vitest | 3.2.6 |
| E2E Testing | Playwright | 1.55.1 |
| Linter | ESLint | 9.39.2 |
| TypeScript Linter | typescript-eslint | 8.54.0 |
| Svelte Linter | eslint-plugin-svelte | 3.14.0 |
| Formatter | Prettier | 3.8.1 |
| Type Checker | svelte-check | 4.1.5 |

### Key Dependencies

```json
{
  // UI Framework
  "@sveltejs/kit": "2.57.1",
  "svelte": "5.55.7",
  "tailwindcss": "^3.4.1",
  
  // Forms & Validation
  "sveltekit-superforms": "2.30.2",
  "zod": "^4.1.12",
  
  // Code Editor
  "@codemirror/autocomplete": "^6.18.7",
  "@codemirror/lang-json": "^6.0.2",
  "@codemirror/lang-go": "^6.0.1",
  
  // Date/Time
  "date-fns": "2.29.x",
  "date-fns-tz": "1.3.x",
  
  // Utilities
  "es-toolkit": "^1.49.0",
  "json-bigint": "^1.0.0",
  
  // Internationalization
  "i18next": "^22.4.15",
  
  // Testing
  "vitest": "^3.2.6",
  "@playwright/test": "^1.55.1"
}
```

## Directory Structure (Detailed)

```
ui/
├── src/
│   ├── routes/                     # SvelteKit routes
│   │   ├── (app)/                 # Protected app routes
│   │   │   ├── namespaces/
│   │   │   │   └── [namespace]/
│   │   │   │       ├── workflows/
│   │   │   │       │   ├── +page.svelte              # Workflow list
│   │   │   │       │   └── [workflowId]/
│   │   │   │       │       ├── [runId]/
│   │   │   │       │       │   ├── +page.svelte      # Workflow detail
│   │   │   │       │       │   └── history/
│   │   │   │       │       │       └── +page.svelte  # Event history
│   │   │   │       ├── schedules/
│   │   │   │       │   ├── +page.svelte              # Schedule list
│   │   │   │       │   └── [scheduleId]/
│   │   │   │       │       └── +page.svelte          # Schedule detail
│   │   │   │       ├── archival/
│   │   │   │       ├── import-export/
│   │   │   │       └── settings/
│   │   │   └── cluster/
│   │   └── (login)/               # Auth routes
│   │       └── +page.svelte
│   │
│   ├── lib/
│   │   ├── components/            # Reusable components
│   │   │   ├── workflow/
│   │   │   │   ├── workflow-list.svelte
│   │   │   │   ├── workflow-details.svelte
│   │   │   │   ├── event-history.svelte
│   │   │   │   ├── workflow-actions.svelte
│   │   │   │   └── workflow-status.svelte
│   │   │   ├── namespace/
│   │   │   │   ├── namespace-select.svelte
│   │   │   │   └── namespace-settings.svelte
│   │   │   ├── schedule/
│   │   │   │   ├── schedule-form.svelte
│   │   │   │   └── schedule-calendar.svelte
│   │   │   └── common/
│   │   │       ├── button.svelte
│   │   │       ├── input.svelte
│   │   │       ├── table.svelte
│   │   │       ├── modal.svelte
│   │   │       └── code-editor.svelte
│   │   │
│   │   ├── holocene/              # Holocene design system
│   │   │   ├── badge.svelte
│   │   │   ├── button.svelte
│   │   │   ├── card.svelte
│   │   │   ├── checkbox.svelte
│   │   │   ├── dropdown.svelte
│   │   │   ├── input.svelte
│   │   │   ├── select.svelte
│   │   │   ├── table.svelte
│   │   │   └── tabs.svelte
│   │   │
│   │   ├── stores/                # Global state
│   │   │   ├── auth.ts
│   │   │   ├── namespace.ts
│   │   │   ├── workflow.ts
│   │   │   ├── settings.ts
│   │   │   └── theme.ts
│   │   │
│   │   ├── services/              # API services
│   │   │   ├── workflow-service.ts
│   │   │   ├── namespace-service.ts
│   │   │   ├── schedule-service.ts
│   │   │   └── cluster-service.ts
│   │   │
│   │   ├── models/                # TypeScript types
│   │   │   ├── workflow.ts
│   │   │   ├── namespace.ts
│   │   │   ├── schedule.ts
│   │   │   └── common.ts
│   │   │
│   │   ├── utilities/             # Helpers
│   │   │   ├── format-date.ts
│   │   │   ├── format-duration.ts
│   │   │   ├── parse-time.ts
│   │   │   ├── query-builder.ts
│   │   │   └── event-history.ts
│   │   │
│   │   ├── i18n/                  # Internationalization
│   │   │   ├── en.json
│   │   │   └── locales.ts
│   │   │
│   │   └── theme/                 # Theming
│   │       ├── colors.ts
│   │       └── theme.ts
│   │
│   ├── app.html                   # HTML template
│   ├── app.css                    # Global styles
│   ├── hooks.server.ts            # Server hooks
│   └── hooks.client.ts            # Client hooks
│
├── static/                        # Static assets
│   ├── favicon.png
│   └── icons/
│
├── tests/
│   ├── e2e/                       # E2E tests
│   └── integration/               # Integration tests
│
├── scripts/                       # Build scripts
│   ├── download-temporal.ts
│   ├── generate-locales.ts
│   └── start-temporal-server.ts
│
├── package.json
├── vite.config.ts                 # Vite configuration
├── svelte.config.js               # Svelte configuration
├── tailwind.config.ts             # Tailwind configuration
├── tsconfig.json                  # TypeScript configuration
└── playwright.config.ts           # Playwright configuration
```

## Key Features

### 1. Workflow Management

#### Workflow List
- **Advanced search** with query builder
- **Filters**: Status, time range, workflow type, workflow ID
- **Sorting**: Start time, close time, execution time
- **Pagination**: Efficient loading of large result sets
- **Batch operations**: Bulk terminate, cancel

#### Workflow Details
- **Execution summary**: Status, start time, close time, duration
- **Input/output payloads**: JSON viewer with syntax highlighting
- **Workflow actions**:
  - Terminate
  - Cancel
  - Signal
  - Query
  - Reset
  - Retry

#### Event History
- **Timeline view**: Visual event timeline
- **Event details**: Full event attributes
- **Event filtering**: Filter by event type
- **Event search**: Search within events
- **Compact/expanded modes**
- **Stack trace viewer**: For workflow failures

### 2. Namespace Management

- **List namespaces** with search and filtering
- **Create namespace** with configuration
- **Update namespace settings**:
  - Retention period
  - Archival configuration
  - Search attributes
  - Cluster configuration
- **Namespace metrics**: Workflow counts, task queue stats

### 3. Schedule Management

- **List schedules** with status
- **Create schedule** with:
  - Cron expression builder
  - Interval configuration
  - Calendar specification
- **Edit schedule** configuration
- **Pause/unpause** schedules
- **View schedule execution history**
- **Trigger schedule** manually

### 4. Advanced Search

**Query Builder**:
```
WorkflowType = "MyWorkflow" AND
ExecutionStatus = "Running" AND
StartTime > "2024-01-01"
```

**Search Attributes**:
- Custom indexed fields
- Type-safe queries
- Advanced operators (>, <, =, !=, IN)

### 5. Code Editor Integration

**CodeMirror 6** for:
- **JSON payload editing**: With validation
- **Query builder**: Syntax highlighting
- **Workflow input**: Autocomplete for search attributes
- **Schedule cron expressions**: Inline validation

**Features**:
- Syntax highlighting
- Auto-completion
- Lint errors inline
- Multiple language support (JSON, Go, Java, Python, PHP)

### 6. Data Visualization

- **Workflow execution graphs**: Visual workflow structure
- **Event timeline**: Interactive timeline
- **Metrics dashboards**: Workflow success/failure rates
- **Table views**: Sortable, filterable tables

### 7. Internationalization

**Supported Languages**:
- English (default)
- Extensible for other languages

**i18next Integration**:
```typescript
import { t } from '$lib/i18n';

// Usage
const title = $t('workflow.list.title');
```

### 8. Theming

**Dark/Light Mode**:
- Auto-detection from system preferences
- Manual toggle
- Persisted in local storage

**Holocene Design System**:
- Consistent component library
- Accessible components (ARIA-compliant)
- Responsive design

### 9. Real-Time Updates

**Polling**:
- Workflow status updates every 5 seconds
- Event history updates
- Schedule execution updates

**Server-Sent Events (SSE)** (future):
- Real-time event streaming

## Configuration

### Build Targets

```bash
# Local development (points to UI server)
pnpm dev

# Local Temporal Server (standalone)
pnpm dev:local-temporal

# Temporal CLI
pnpm dev:temporal-cli

# Docker environment
pnpm dev:docker
```

### Environment Variables

**Build-time**:
```bash
# API endpoint
VITE_API=http://localhost:8080

# Build target
VITE_TEMPORAL_UI_BUILD_TARGET=local

# Feature flags
VITE_FEATURE_NEXUS=true
VITE_FEATURE_BATCH_OPERATIONS=true
```

**Runtime** (via UI server):
- Authentication settings
- API base URL
- Feature flags

## Build & Deployment

### Development

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Type checking
pnpm check

# Linting
pnpm lint

# Testing
pnpm test
```

### Production Build

```bash
# Build for local preview
pnpm build:local

# Build for UI server
pnpm build:server

# Build for Docker
pnpm build:docker
```

### Build Outputs

| Target | Output Path | Purpose |
|--------|-------------|---------|
| Local | `./build` | Preview/testing |
| Server | `../ui-server/ui` | Embedded in UI server |
| Docker | `./build` | Docker image |

### Docker Deployment

The UI is typically deployed as static assets served by UI Server. See UI Server specification for deployment.

## API Integration

### API Client

**Base Service**:
```typescript
// src/lib/services/base-service.ts
export class BaseService {
  constructor(private baseUrl: string) {}
  
  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`);
    if (!response.ok) throw new Error(response.statusText);
    return response.json();
  }
}
```

**Workflow Service**:
```typescript
// src/lib/services/workflow-service.ts
export class WorkflowService extends BaseService {
  async listWorkflows(namespace: string, query: WorkflowQuery): Promise<WorkflowList> {
    const params = new URLSearchParams({
      pageSize: String(query.pageSize),
      query: query.query,
    });
    return this.get(`/api/v1/namespaces/${namespace}/workflows?${params}`);
  }
  
  async getWorkflow(namespace: string, workflowId: string, runId: string): Promise<Workflow> {
    return this.get(`/api/v1/namespaces/${namespace}/workflows/${workflowId}/runs/${runId}`);
  }
}
```

### State Management

**Svelte 5 Runes**:
```typescript
// src/lib/stores/workflow.ts
import { writable } from 'svelte/store';

export const workflowStore = writable<Workflow | null>(null);

// Usage in component
let workflow = $state<Workflow | null>(null);
let loading = $state(false);

$effect(() => {
  loadWorkflow();
});

async function loadWorkflow() {
  loading = true;
  workflow = await workflowService.getWorkflow(namespace, workflowId, runId);
  loading = false;
}
```

## Testing

### Unit Tests (Vitest)

```bash
# Run tests
pnpm test

# Watch mode
pnpm test:ui

# Coverage
pnpm test:coverage
```

### E2E Tests (Playwright)

```bash
# Run E2E tests
pnpm test:e2e

# Interactive mode
pnpm test:e2e:ui
```

### Integration Tests

```bash
# Run integration tests
pnpm test:integration
```

## Performance Optimizations

### Code Splitting
- Route-based code splitting
- Dynamic imports for heavy components
- Lazy loading for code editor

### Caching
- Workflow list pagination cache
- Namespace list cache
- Local storage for user preferences

### Rendering
- Virtual scrolling for large event histories
- Debounced search input
- Optimistic UI updates

## Accessibility

### ARIA Compliance
- Semantic HTML
- ARIA labels and roles
- Keyboard navigation
- Screen reader support

### Standards
- WCAG 2.1 Level AA compliance
- Focus management
- Color contrast ratios
- Alternative text for images

## Browser Support

| Browser | Version |
|---------|---------|
| Chrome | Last 2 versions |
| Firefox | Last 2 versions |
| Safari | Last 2 versions |
| Edge | Last 2 versions |

## Dependencies & Integration Points

### Upstream Dependencies
- **UI Server** (required) - HTTP API backend
- **Temporal Server** (via UI Server) - Data source

### Downstream Dependents
- None (end-user application)

## Security

### Content Security Policy
- Strict CSP headers
- No inline scripts (production build)
- Trusted sources only

### XSS Protection
- Sanitized markdown rendering
- Escaped user input
- Secure JSON parsing

### Authentication
- OIDC integration (via UI Server)
- Session cookie validation
- CSRF token handling

## Implementation Notes

### Svelte 5 Patterns

**Props**:
```typescript
let { class: className = '', adapter }: Props = $props();
```

**State**:
```typescript
let count = $state(0);
```

**Computed**:
```typescript
const doubled = $derived(count * 2);
```

**Effects**:
```typescript
$effect(() => {
  console.log('Count:', count);
  return () => cleanup();
});
```

### Import Order
1. Node.js built-ins
2. External libraries (svelte/** first)
3. SvelteKit imports ($app/**)
4. Internal imports ($lib/**)
5. Component imports (**.svelte)
6. Relative imports (./, ../)

### Code Style
- No code comments unless necessary
- TypeScript for type safety
- Prefer Holocene components over custom
- Ensure accessibility (ARIA attributes)

## References

- [Temporal UI Repository](https://github.com/temporalio/ui)
- [SvelteKit Documentation](https://kit.svelte.dev/)
- [Svelte 5 Documentation](https://svelte-5-preview.vercel.app/)
- [Holocene Design System](https://holocene.temporalio.com/)
