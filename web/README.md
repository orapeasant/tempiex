# web

Tempiex Web UI. A React 18 + TypeScript single-page application for monitoring and managing Tempiex workflows. Built with shadcn/ui, Tailwind CSS, and React Query.

## Quick Start

### Prerequisites

- Node.js 22+
- pnpm
- A running `server` instance on `http://localhost:8080`

### 1. Install dependencies

```bash
cd web/
pnpm install
```

### 2. Start the dev server

```bash
pnpm dev
```

Open [http://localhost:5173](http://localhost:5173).

The dev server proxies API requests to `http://localhost:8080`. To change the backend URL:

```bash
VITE_API_URL=http://localhost:8080 pnpm dev
```

## Installation (production build)

```bash
cd web/
pnpm install
pnpm build
# Output in dist/ — serve with any static file server
```

## Development

```bash
cd web/

# Start Vite dev server with hot reload
pnpm dev

# Type check (no emit)
pnpm typecheck

# Lint
pnpm lint

# Production build
pnpm build

# Preview production build locally
pnpm preview
```

## Project Structure

```
web/
├── src/
│   ├── pages/          # Route-level page components
│   ├── components/     # Shared UI components (shadcn/ui based)
│   ├── hooks/          # React Query hooks for data fetching
│   ├── lib/            # API client, utilities
│   ├── App.tsx         # Router setup
│   └── main.tsx        # Entry point
└── public/             # Static assets
```

## Tech Stack

| Library | Purpose |
|---|---|
| React 18 | UI rendering |
| TypeScript (strict) | Type safety |
| Vite | Build tool and dev server |
| shadcn/ui | Component library (Radix UI + Tailwind) |
| Tailwind CSS | Utility-first styling |
| React Query | Server state and data fetching |
| React Router | Client-side routing |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend server URL |

Set in a `.env.local` file at the `web/` root:

```bash
VITE_API_URL=http://localhost:8080
```
