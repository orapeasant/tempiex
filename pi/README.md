# pi

Tempiex Pi — Agent Harness. A clean-room rewrite in Go of the
[`earendil-works/pi`](https://github.com/earendil-works/pi) agent harness: an
LLM agent runtime with a model-call loop, tool calling, durable sessions,
hooks, context compaction, and a remote-session protocol.

**Status: not yet implemented.** The Tempiex activity worker that previously
lived in this directory has moved to [`../worker`](../worker) — see
"What moved, and why" below. The harness itself is specified in
[`docs/spec/06-pi.md`](../docs/spec/06-pi.md) and is being written from
scratch.

## Scope

`pi` is an agent harness. It is **not** a Tempiex worker and has no dependency
on Tempiex:

- It does not poll task queues, and knows nothing about namespaces, workflow
  IDs, run IDs, activity task tokens, or `RespondActivityTask*` RPCs.
- `go list -deps ./...` from this directory must never contain
  `github.com/tempiex/tempiex` or `github.com/tempiex/worker`.

Planned packages (see the spec for detail): `ai` (multi-provider LLM API),
`agent` (turn loop, tools, queues), `harness` (durable sessions, lanes, hooks,
compaction, skills), `protocol`/`server`/`client` (remote sessions over framed
CBOR), `telemetry`, and the `pi` CLI.

## What moved, and why

The original `pi` mixed two unrelated things in one binary: the generic
poll/dispatch/shutdown lifecycle any Tempiex worker needs, and agent-harness
concerns. They are now separate modules.

| Was | Is now |
|---|---|
| `pi/internal/tool` | `worker/activity` (`Activity[I,O]`, `Registry`) |
| `pi/internal/activity` | `worker/activity` (`Executor`) |
| `pi/internal/session` | `worker/run` (`Run`, `Manager`, `Store`) |
| `pi/internal/event` | `worker/run` (`Bus`, `Event`) |
| `pi/internal/api` | `worker/api` (`/api/runs`, `/api/activities`) |
| `pi/internal/metrics` | `worker/metrics` (`worker.*` instruments) |
| `pi/internal/agent` | `worker/cmd/worker` (process wiring) |
| `pi/tools/{shell,file,httptool}` | `worker/activities/{shell,file,httptool}` |
| `pi/cmd/pi` | `worker/cmd/worker` |

Running an agent as a Tempiex activity will be a `worker/agentbridge` handler
that imports this module — the import points one way, `worker` → `pi`.

## Specs

- [`docs/spec/06-pi.md`](../docs/spec/06-pi.md) — this module
- [`docs/spec/08-worker.md`](../docs/spec/08-worker.md) — the worker it split from
