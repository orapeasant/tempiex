# Temporal CLI Specification

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Global Flags](#global-flags)
   - [Common Flags](#common-flags)
   - [Client / Connection Flags](#client--connection-flags)
4. [Command Reference](#command-reference)
   - [temporal activity](#temporal-activity)
   - [temporal batch](#temporal-batch)
   - [temporal config](#temporal-config)
   - [temporal env](#temporal-env)
   - [temporal nexus operation](#temporal-nexus-operation)
   - [temporal operator cluster](#temporal-operator-cluster)
   - [temporal operator namespace](#temporal-operator-namespace)
   - [temporal operator nexus endpoint](#temporal-operator-nexus-endpoint)
   - [temporal operator search-attribute](#temporal-operator-search-attribute)
   - [temporal schedule](#temporal-schedule)
   - [temporal server start-dev](#temporal-server-start-dev)
   - [temporal task-queue](#temporal-task-queue)
   - [temporal worker](#temporal-worker)
   - [temporal workflow](#temporal-workflow)
5. [Option Sets](#option-sets)
6. [Dev Server](#dev-server)
7. [Output Formats](#output-formats)
8. [Environment & Configuration](#environment--configuration)
9. [Codec Server Integration](#codec-server-integration)

---

## Overview

The **Temporal CLI** (`temporal`) is the official command-line tool for managing, monitoring, and debugging Temporal applications. It allows you to:

- Run a local Temporal development service
- Start, list, cancel, terminate, and inspect Workflow Executions
- Send Signals and Updates to running Workflows
- Manage Schedules, Namespaces, Search Attributes, Nexus Endpoints
- Operate on Activities (including Standalone Activities)
- Manage Worker Deployments and versioning

**Go module path:** `github.com/temporalio/cli`

### Install via Homebrew

```bash
brew install temporal
```

### Install via Binary Download

Download the archive for your platform from:

| Platform | URL |
|---|---|
| Linux amd64 | https://temporal.download/cli/archive/latest?platform=linux&arch=amd64 |
| Linux arm64 | https://temporal.download/cli/archive/latest?platform=linux&arch=arm64 |
| macOS amd64 | https://temporal.download/cli/archive/latest?platform=darwin&arch=amd64 |
| macOS arm64 | https://temporal.download/cli/archive/latest?platform=darwin&arch=arm64 |
| Windows amd64 | https://temporal.download/cli/archive/latest?platform=windows&arch=amd64 |

Extract the archive and add the `temporal` binary to your `PATH`.

### Run via Docker

```bash
docker run --rm temporalio/temporal --help

# Dev server accessible from host:
docker run --rm -p 7233:7233 -p 8233:8233 temporalio/temporal:latest server start-dev --ip 0.0.0.0
```

### Build from Source

```bash
# Install Go, clone the repository, then:
go build ./cmd/temporal
```

---

## Architecture

The CLI is organized into focused Go packages:

| Package | Role |
|---|---|
| `cmd/temporal` | Entry point; wires up the CLI and calls `temporalcli.Execute()` |
| `internal/temporalcli` | Core command implementations; auto-generated structs from `commands.yaml` |
| `internal/devserver` | Embedded Temporal development server (SQLite-backed, includes UI) |
| `internal/tracer` | Workflow execution tracer used by `temporal workflow trace` |
| `internal/printer` | Output formatting (text tables, JSON, JSONL, colorized output) |
| `cliext` | Flag extensions and shared option sets (common, client); OAuth support |
| `internal/commandsgen` | Code-generation tooling that reads `commands.yaml` to produce Go structs and docs |

---

## Global Flags

These flags apply to every `temporal` command. They control output, logging, configuration profiles, and server connectivity.

### Common Flags

Available on all commands (from the `common` option set in `cliext/option-sets.yaml`):

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--env` | string | `default` | No | Active environment name. Env var: `TEMPORAL_ENV` |
| `--env-file` | string | — | No | Path to environment settings file. Env var: `TEMPORAL_ENV_FILE` |
| `--config-file` | string | — | No | TOML config file path. Env var: `TEMPORAL_CONFIG_FILE` |
| `--profile` | string | `default` | No | Configuration profile to use. Env var: `TEMPORAL_PROFILE` |
| `--disable-config-file` | bool | false | No | Disable loading config from file |
| `--disable-config-env` | bool | false | No | Disable loading config from environment variables |
| `--log-level` | string-enum | `never` | No | Log level: `debug`, `info`, `warn`, `error`, `never` |
| `--log-format` | string-enum | `text` | No | Log format: `text`, `json` |
| `--output` / `-o` | string-enum | `text` | No | Non-logging data output format: `text`, `json`, `jsonl`, `none` |
| `--time-format` | string-enum | `relative` | No | Time format: `relative`, `iso`, `raw` |
| `--color` | string-enum | `auto` | No | Output coloring: `always`, `never`, `auto` |
| `--no-json-shorthand-payloads` | bool | false | No | Raw payload output, even if the JSON option was used |
| `--command-timeout` | duration | — | No | Command execution timeout |
| `--client-connect-timeout` | duration | — | No | Client connection timeout |

### Client / Connection Flags

Available on commands that connect to a Temporal Service (from the `client` option set). Values resolve in order: CLI flag > environment variable > config file.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--address` | string | `localhost:7233` | No | Temporal Service gRPC endpoint. Env var: `TEMPORAL_ADDRESS` |
| `--client-authority` | string | — | No | Temporal gRPC client `:authority` pseudoheader |
| `--namespace` / `-n` | string | `default` | No | Temporal Service Namespace. Env var: `TEMPORAL_NAMESPACE` |
| `--api-key` | string | — | No | API key for request. Env var: `TEMPORAL_API_KEY` |
| `--grpc-meta` | string[] | — | No | HTTP headers for requests (`KEY=VALUE`, repeatable) |
| `--tls` | bool | false | No | Enable base TLS encryption. Auto-enabled when api-key or TLS options set. Env var: `TEMPORAL_TLS` |
| `--tls-cert-path` | string | — | No | Path to x509 certificate. Env var: `TEMPORAL_TLS_CLIENT_CERT_PATH` |
| `--tls-cert-data` | string | — | No | Inline x509 certificate data. Env var: `TEMPORAL_TLS_CLIENT_CERT_DATA` |
| `--tls-key-path` | string | — | No | Path to x509 private key. Env var: `TEMPORAL_TLS_CLIENT_KEY_PATH` |
| `--tls-key-data` | string | — | No | Inline x509 private key data. Env var: `TEMPORAL_TLS_CLIENT_KEY_DATA` |
| `--tls-ca-path` | string | — | No | Path to server CA certificate. Env var: `TEMPORAL_TLS_SERVER_CA_CERT_PATH` |
| `--tls-ca-data` | string | — | No | Inline server CA certificate data. Env var: `TEMPORAL_TLS_SERVER_CA_CERT_DATA` |
| `--tls-disable-host-verification` | bool | false | No | Disable TLS host-name verification. Env var: `TEMPORAL_TLS_DISABLE_HOST_VERIFICATION` |
| `--tls-server-name` | string | — | No | Override target TLS server name. Env var: `TEMPORAL_TLS_SERVER_NAME` |
| `--codec-endpoint` | string | — | No | Remote Codec Server endpoint. Env var: `TEMPORAL_CODEC_ENDPOINT` |
| `--codec-auth` | string | — | No | Authorization header for Codec Server requests. Env var: `TEMPORAL_CODEC_AUTH` |
| `--codec-header` | string[] | — | No | HTTP headers for codec server (`KEY=VALUE`, repeatable) |
| `--identity` | string | — | No | Identity of the client submitting requests |

---

## Command Reference

### temporal activity

**Summary:** Operate on Activity Executions.

```bash
temporal activity start \
    --activity-id YourActivityId \
    --type YourActivity \
    --task-queue YourTaskQueue \
    --start-to-close-timeout 5m
```

Uses the `client` option set (see [Client / Connection Flags](#client--connection-flags)).

---

#### temporal activity cancel

🧪 Experimental

**Summary:** Request cancellation of a Standalone Activity.

Transitions the Activity's run state to `CancelRequested`. If the Activity is heartbeating, a cancellation error is raised on next heartbeat response.

```bash
temporal activity cancel \
    --activity-id YourActivityId

# Bulk cancellation:
temporal activity cancel \
    --query YourQuery \
    --reason YourReason
```

Uses the `activity-reference-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for cancellation. Also used as reason for batch operation with `--query` |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm. Only allowed when `--query` is present |

---

#### temporal activity complete

**Summary:** Mark an Activity as completed successfully with a result.

```bash
temporal activity complete \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId \
    --result '{"YourResultKey": "YourResultVal"}'
```

Omit `--workflow-id` to target a Standalone Activity.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | **Yes** | Activity ID (workflow-invoked or Standalone) |
| `--workflow-id` / `-w` | string | — | No | Workflow ID. Required for workflow Activities |
| `--run-id` / `-r` | string | — | No | Run ID (Workflow Run ID for workflow Activities; Activity Run ID for Standalone) |
| `--result` | string | — | **Yes** | Result JSON to return |

---

#### temporal activity count

🧪 Experimental

**Summary:** Count Standalone Activities matching a query.

```bash
temporal activity count \
    --query 'ActivityType="YourActivity"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | Query to filter Activity Executions to count |

---

#### temporal activity delete

🧪 Experimental

**Summary:** Delete a Standalone Activity Execution and its Event History.

```bash
temporal activity delete \
    --activity-id YourActivityId

# Bulk delete:
temporal activity delete \
    --query YourQuery
```

Uses the `activity-reference-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for batch operation. Only use with `--query` |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm |

---

#### temporal activity describe

🧪 Experimental

**Summary:** Show detailed info for a Standalone Activity.

```bash
temporal activity describe \
    --activity-id YourActivityId
```

Uses the `activity-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--raw` | bool | false | No | Print properties without changing their format |

---

#### temporal activity execute

🧪 Experimental

**Summary:** Start a new Standalone Activity and wait for its result.

```bash
temporal activity execute \
    --activity-id YourActivityId \
    --type YourActivity \
    --task-queue YourTaskQueue \
    --start-to-close-timeout 30s \
    --input '{"some-key": "some-value"}'
```

Uses `activity-start` and `payload-input` option sets.

---

#### temporal activity fail

**Summary:** Mark an Activity as completed unsuccessfully with an error.

```bash
temporal activity fail \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | **Yes** | Activity ID |
| `--workflow-id` / `-w` | string | — | No | Workflow ID. Required for workflow Activities |
| `--run-id` / `-r` | string | — | No | Run ID |
| `--detail` | string | — | No | Failure detail (JSON) |
| `--reason` | string | — | No | Failure reason (message) |

---

#### temporal activity list

🧪 Experimental

**Summary:** List Standalone Activities matching a query.

```bash
temporal activity list \
    --query 'ActivityType="YourActivity"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | Query to filter Activity Executions to list |
| `--limit` | int | 0 | No | Maximum number of Activity Executions to display |
| `--page-size` | int | 0 | No | Maximum number of Activity Executions to fetch at a time |

---

#### temporal activity pause

**Summary:** Pause an Activity.

If the Activity is currently running, it will continue until it fails, completes, or times out, then the pause kicks in. Does not stop or extend the Activity's Schedule-To-Close Timeout.

```bash
# Pause a workflow Activity:
temporal activity pause \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId

# Pause a standalone Activity:
temporal activity pause \
    --activity-id YourActivityId \
    --run-id YourRunId
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | No | The Activity ID to pause |
| `--workflow-id` / `-w` | string | — | No | Workflow ID. Set to target a workflow Activity |
| `--run-id` / `-r` | string | — | No | Run ID |
| `--identity` | string | — | No | Identity of the user submitting this request |
| `--reason` | string | — | No | Reason for pausing the Activity |

---

#### temporal activity reset

**Summary:** Reset an Activity.

Restarts the activity from scratch: attempt count returns to one, per-attempt timeouts re-armed, heartbeat details cleared.

```bash
temporal activity reset \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId \
    --keep-paused

# Bulk reset:
temporal activity reset \
    --query 'WorkflowType="YourWorkflow"'
```

Uses `single-activity-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | No | The Activity ID to reset. Mutually exclusive with `--query` |
| `--keep-paused` | bool | false | No | If the activity was paused, it will stay paused |
| `--jitter` | duration | — | No | Random time within this duration before reset. Only with `--query` |
| `--restore-original-options` | bool | false | No | Restore the original options of the activity |

---

#### temporal activity result

🧪 Experimental

**Summary:** Wait for and output the result of a Standalone Activity.

```bash
temporal activity result \
    --activity-id YourActivityId
```

Uses the `activity-reference` option set.

---

#### temporal activity start

🧪 Experimental

**Summary:** Start a new Standalone Activity.

```bash
temporal activity start \
    --activity-id YourActivityId \
    --type YourActivity \
    --task-queue YourTaskQueue \
    --start-to-close-timeout 5m \
    --input '{"some-key": "some-value"}'
```

Uses `activity-start` and `payload-input` option sets.

---

#### temporal activity terminate

🧪 Experimental

**Summary:** Forcefully end a Standalone Activity. Activity code cannot see or respond to terminations.

```bash
temporal activity terminate \
    --activity-id YourActivityId \
    --reason YourReason

# Bulk termination:
temporal activity terminate \
    --query YourQuery \
    --reason YourReason
```

Uses the `activity-reference-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for termination. Defaults to current user's name |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm. Only allowed when `--query` is present |

---

#### temporal activity unpause

**Summary:** Unpause an Activity.

Re-schedules a previously-paused Activity for execution.

```bash
temporal activity unpause \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId

# Bulk unpause:
temporal activity unpause \
    --query 'TemporalPauseInfo IS NOT NULL'
```

Uses `single-activity-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | No | The Activity ID to unpause. Mutually exclusive with `--query` |
| `--jitter` | duration | — | No | Random time within this duration before start. Only with `--query` |

---

#### temporal activity update-options

**Summary:** Change the values of options affecting an Activity.

```bash
temporal activity update-options \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId \
    --task-queue NewTaskQueueName \
    --schedule-to-close-timeout DURATION \
    --start-to-close-timeout DURATION

# Bulk update:
temporal activity update-options \
    --query 'WorkflowType="YourWorkflow"' \
    --task-queue NewTaskQueueName
```

Uses `single-activity-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--activity-id` / `-a` | string | — | No | Activity ID. Mutually exclusive with `--query` |
| `--task-queue` | string | — | No | Name of the task queue for the Activity |
| `--schedule-to-close-timeout` | duration | — | No | Max time caller waits for completion |
| `--schedule-to-start-timeout` | duration | — | No | Max time a task can stay in queue before pickup |
| `--start-to-close-timeout` | duration | — | No | Max execution time for a single attempt |
| `--heartbeat-timeout` | duration | — | No | Max time between successful heartbeats |
| `--start-delay` | duration | — | No | Change when first Activity Task becomes available (Standalone only, before first dispatch) |
| `--retry-initial-interval` | duration | — | No | Interval of the first retry |
| `--retry-maximum-interval` | duration | — | No | Maximum interval between retries |
| `--retry-backoff-coefficient` | float | — | No | Coefficient for next retry interval (must be ≥ 1) |
| `--retry-maximum-attempts` | int | — | No | Maximum retry attempts (0 = unlimited, 1 = no retries) |
| `--restore-original-options` | bool | false | No | Restore the original options of the activity |

---

### temporal batch

**Summary:** Manage running batch jobs.

A batch job executes a command on multiple Workflow Executions at once. Create batch jobs by passing `--query` to commands that support it.

```bash
temporal workflow cancel \
    --query 'ExecutionStatus = "Running" AND WorkflowType="YourWorkflow"' \
    --reason "Testing"
```

Uses the `client` option set.

---

#### temporal batch describe

**Summary:** Show batch job progress.

```bash
temporal batch describe \
    --job-id YourJobId
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--job-id` | string | — | **Yes** | Batch job ID |

---

#### temporal batch list

**Summary:** List all batch jobs.

```bash
temporal batch list \
    --namespace YourNamespace
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--limit` | int | 0 | No | Maximum number of batch jobs to display |

---

#### temporal batch terminate

**Summary:** Forcefully end a batch job.

```bash
temporal batch terminate \
    --job-id YourJobId \
    --reason YourTerminationReason
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--job-id` | string | — | **Yes** | Job ID to terminate |
| `--reason` | string | — | **Yes** | Reason for terminating the batch job |

---

### temporal config

🧪 Experimental

**Summary:** Manage TOML config files with profiles.

```bash
temporal config set \
    --profile YourProfile \
    --prop address \
    --value us-west-2.aws.api.temporal.io:7233
```

The default config file path is `$CONFIG_PATH/temporalio/temporal.toml` where `$CONFIG_PATH` is `$HOME/.config` (Unix), `$HOME/Library/Application Support` (macOS), or `%AppData%` (Windows). Override with `TEMPORAL_CONFIG_FILE` or `--config-file`.

---

#### temporal config delete

🧪 Experimental

**Summary:** Delete a config file property.

```bash
temporal config delete \
    --profile YourProfile \
    --prop tls.client_cert_path
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--prop` / `-p` | string | — | **Yes** | Property to delete. If unset, deletes entire profile |

---

#### temporal config delete-profile

🧪 Experimental

**Summary:** Delete an entire config profile.

```bash
temporal config delete-profile \
    --profile YourProfile
```

---

#### temporal config get

🧪 Experimental

**Summary:** Show config file properties.

```bash
temporal config get \
    --profile YourProfile \
    --prop address

# Show all:
temporal config get
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--prop` / `-p` | string | — | No | Specific property to get |

---

#### temporal config list

🧪 Experimental

**Summary:** Show config file profiles (list profile names).

```bash
temporal config list
```

---

#### temporal config set

🧪 Experimental

**Summary:** Set config file properties.

```bash
temporal config set \
    --profile YourProfile \
    --prop address \
    --value us-west-2.aws.api.temporal.io:7233
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--prop` / `-p` | string | — | **Yes** | Property name |
| `--value` / `-v` | string | — | **Yes** | Property value |

---

### temporal env

**Summary:** Manage environments (key-value presets for CLI options).

> Note: `temporal env` is slated for deprecation in favor of `temporal config`.

```bash
temporal env set \
    --env prod \
    --key address \
    --value production.f45a2.tmprl.cloud:7233
```

Environments are stored in `$HOME/.config/temporalio/temporal.yaml`.

---

#### temporal env delete

**Summary:** Delete an environment or environment property.

```bash
temporal env delete \
    --env YourEnvironment

temporal env delete \
    --env prod \
    --key tls-key-path
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--key` / `-k` | string | — | No | Property name to delete |

---

#### temporal env get

**Summary:** Show environment properties.

```bash
temporal env get \
    --env YourEnvironment \
    --key YourPropertyKey
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--key` / `-k` | string | — | No | Property name |

---

#### temporal env list

**Summary:** Show environment names.

```bash
temporal env list
```

---

#### temporal env set

**Summary:** Set environment properties.

```bash
temporal env set \
    --env environment \
    --key property \
    --value value
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--key` / `-k` | string | — | No | Property name |
| `--value` / `-v` | string | — | No | Property value |

---

### temporal nexus operation

🧪 Experimental

**Summary:** Commands for managing Nexus Operations.

```bash
temporal nexus operation [command] [options]
```

Uses the `client` option set.

---

#### temporal nexus operation cancel

🧪 Experimental

**Summary:** Request cancellation of a Nexus Operation.

```bash
temporal nexus operation cancel \
    --operation-id YourOperationId
```

Uses `nexus-operation-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for cancellation |

---

#### temporal nexus operation count

🧪 Experimental

**Summary:** Count Nexus Operations matching a query.

```bash
temporal nexus operation count \
    --query 'Endpoint="YourEndpoint"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | Query to filter Nexus Operation Executions to count |

---

#### temporal nexus operation describe

🧪 Experimental

**Summary:** Show detailed info for a Nexus Operation.

```bash
temporal nexus operation describe \
    --operation-id YourOperationId
```

Uses `nexus-operation-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--raw` | bool | false | No | Print properties without changing their format |

---

#### temporal nexus operation execute

🧪 Experimental

**Summary:** Start a new Nexus Operation and wait for its result.

```bash
temporal nexus operation execute \
    --endpoint YourEndpoint \
    --service YourService \
    --operation YourOperation \
    --operation-id YourOperationId \
    --input '{"some-key": "some-value"}'
```

Uses `nexus-operation-start` and `payload-input` option sets.

---

#### temporal nexus operation list

🧪 Experimental

**Summary:** List Nexus Operations matching a query.

```bash
temporal nexus operation list \
    --query 'Endpoint="YourEndpoint"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | Query to filter Nexus Operation Executions |
| `--limit` | int | 0 | No | Maximum number of Nexus Operation Executions to display |
| `--page-size` | int | 0 | No | Maximum number to fetch at a time from the server |

---

#### temporal nexus operation result

🧪 Experimental

**Summary:** Wait for and output the result of a Nexus Operation.

```bash
temporal nexus operation result \
    --operation-id YourOperationId
```

Uses `nexus-operation-reference` option set.

---

#### temporal nexus operation start

🧪 Experimental

**Summary:** Start a new Nexus Operation.

```bash
temporal nexus operation start \
    --endpoint YourEndpoint \
    --service YourService \
    --operation YourOperation \
    --operation-id YourOperationId \
    --input '{"some-key": "some-value"}'
```

Uses `nexus-operation-start` and `payload-input` option sets.

---

#### temporal nexus operation terminate

🧪 Experimental

**Summary:** Forcefully end a Nexus Operation. Operation handlers cannot see or respond to terminations.

```bash
temporal nexus operation terminate \
    --operation-id YourOperationId \
    --reason YourReason
```

Uses `nexus-operation-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for termination. Defaults to current user's name |

---

### temporal operator cluster

**Summary:** Manage a Temporal Cluster (Service).

```bash
temporal operator cluster [subcommand] [options]
```

Uses the `client` option set.

---

#### temporal operator cluster describe

**Summary:** Show Temporal Cluster information (name, persistence store, visibility store).

```bash
temporal operator cluster describe [--detail]
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--detail` | bool | false | No | Show history shard count and Cluster/Service version information |

---

#### temporal operator cluster health

**Summary:** Check Temporal Service health.

```bash
temporal operator cluster health
```

---

#### temporal operator cluster list

**Summary:** Show Temporal Clusters registered to the local Service.

```bash
temporal operator cluster list [--limit max-count]
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--limit` | int | 0 | No | Maximum number of Clusters to display |

---

#### temporal operator cluster remove

**Summary:** Remove a registered remote Temporal Cluster.

```bash
temporal operator cluster remove \
    --name YourClusterName
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` | string | — | **Yes** | Cluster/Service name |

---

#### temporal operator cluster system

**Summary:** Show Temporal Server information (version, scheduling support, etc.).

```bash
temporal operator cluster system \
    --frontend-address "YourRemoteEndpoint:YourRemotePort"
```

---

#### temporal operator cluster upsert

**Summary:** Add or update a registered Temporal Cluster.

```bash
temporal operator cluster upsert \
    --frontend-address "YourRemoteEndpoint:YourRemotePort" \
    --enable-connection false
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--frontend-address` | string | — | **Yes** | Remote endpoint |
| `--enable-connection` | bool | false | No | Set the connection to "enabled" |
| `--enable-replication` | bool | false | No | Set the replication to "enabled" |

---

### temporal operator namespace

**Summary:** Namespace operations (create, delete, describe, list, update).

```bash
temporal operator namespace [command] [command options]
```

---

#### temporal operator namespace create

**Summary:** Register a new Namespace.

```bash
temporal operator namespace create \
    --namespace YourNewNamespaceName

# Multi-region:
temporal operator namespace create \
    --global \
    --namespace YourNewNamespaceName

# With archival:
temporal operator namespace create \
    --retention 5d \
    --visibility-archival-state enabled \
    --visibility-uri YourURI \
    --namespace YourNewNamespaceName
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--active-cluster` | string | — | No | Active Cluster (Service) name |
| `--cluster` | string[] | — | No | Cluster names. Can be passed multiple times |
| `--data` | string[] | — | No | Namespace data as `KEY=VALUE` pairs (JSON values). Repeatable |
| `--description` | string | — | No | Namespace description |
| `--email` | string | — | No | Owner email |
| `--global` | bool | false | No | Enable multi-region data replication |
| `--history-archival-state` | string-enum | `disabled` | No | History archival state: `disabled`, `enabled` |
| `--history-uri` | string | — | No | Archive history to this URI. Once enabled, can't be changed |
| `--retention` | duration | `72h` | No | Time to preserve closed Workflows before deletion |
| `--visibility-archival-state` | string-enum | `disabled` | No | Visibility archival state: `disabled`, `enabled` |
| `--visibility-uri` | string | — | No | Archive visibility data to this URI. Once enabled, can't be changed |

---

#### temporal operator namespace delete

**Summary:** Delete a Namespace.

```bash
temporal operator namespace delete \
    --namespace YourNamespaceName
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm deletion |

---

#### temporal operator namespace describe

**Summary:** Describe a Namespace by ID or name.

```bash
temporal operator namespace describe \
    --namespace-id YourNamespaceId

temporal operator namespace describe \
    --namespace YourNamespaceName
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--namespace-id` | string | — | No | Namespace ID |

---

#### temporal operator namespace list

**Summary:** Display a detailed listing for all Namespaces.

```bash
temporal operator namespace list
```

---

#### temporal operator namespace update

**Summary:** Update a Namespace.

```bash
temporal operator namespace update \
    --namespace YourNamespaceName \
    --active-cluster NewActiveCluster

# Promote for multi-region:
temporal operator namespace update \
    --namespace YourNamespaceName \
    --promote-global
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--active-cluster` | string | — | No | Active Cluster (Service) name |
| `--cluster` | string[] | — | No | Cluster (Service) names |
| `--data` | string[] | — | No | Namespace data as `KEY=VALUE` pairs. Repeatable |
| `--description` | string | — | No | Namespace description |
| `--email` | string | — | No | Owner email |
| `--promote-global` | bool | false | No | Enable multi-region data replication |
| `--history-archival-state` | string-enum | — | No | History archival state: `disabled`, `enabled` |
| `--history-uri` | string | — | No | Archive history URI. Once enabled, can't be changed |
| `--replication-state` | string-enum | — | No | Replication state: `normal`, `handover` |
| `--retention` | duration | — | No | Length of time a closed Workflow is preserved |
| `--visibility-archival-state` | string-enum | — | No | Visibility archival state: `disabled`, `enabled` |
| `--visibility-uri` | string | — | No | Archive visibility URI. Once enabled, can't be changed |

---

### temporal operator nexus endpoint

**Summary:** Commands for managing Nexus Endpoints.

```bash
temporal operator nexus endpoint [command] [options]
```

---

#### temporal operator nexus endpoint create

**Summary:** Create a Nexus Endpoint on the Server.

```bash
temporal operator nexus endpoint create \
    --name your-endpoint \
    --target-namespace your-namespace \
    --target-task-queue your-task-queue \
    --description-file DESCRIPTION.md
```

Uses `nexus-endpoint-identity` and `nexus-endpoint-config` option sets.

---

#### temporal operator nexus endpoint delete

**Summary:** Delete a Nexus Endpoint.

```bash
temporal operator nexus endpoint delete --name your-endpoint
```

Uses `nexus-endpoint-identity` option set.

---

#### temporal operator nexus endpoint get

🧪 Experimental

**Summary:** Get a Nexus Endpoint by name.

```bash
temporal operator nexus endpoint get --name your-endpoint
```

Uses `nexus-endpoint-identity` option set.

---

#### temporal operator nexus endpoint list

**Summary:** List all Nexus Endpoints.

```bash
temporal operator nexus endpoint list
```

---

#### temporal operator nexus endpoint update

**Summary:** Update an existing Nexus Endpoint. The Endpoint is patched; existing fields not provided are left unchanged.

```bash
temporal operator nexus endpoint update \
    --name your-endpoint \
    --target-task-queue your-other-queue

temporal operator nexus endpoint update \
    --name your-endpoint \
    --description-file DESCRIPTION.md
```

Uses `nexus-endpoint-identity` and `nexus-endpoint-config` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--unset-description` | bool | false | No | Unset the description |

---

### temporal operator search-attribute

**Summary:** Create, list, or remove Search Attribute fields.

Supported types: `Text`, `Keyword`, `Int`, `Double`, `Bool`, `Datetime`, `KeywordList`.

---

#### temporal operator search-attribute create

**Summary:** Add custom Search Attributes.

```bash
temporal operator search-attribute create \
    --name YourAttributeName \
    --type Keyword
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` | string[] | — | **Yes** | Search Attribute name. Can be passed multiple times |
| `--type` | string-enum[] | — | **Yes** | Search Attribute type: `Text`, `Keyword`, `Int`, `Double`, `Bool`, `Datetime`, `KeywordList` |

---

#### temporal operator search-attribute list

**Summary:** List active Search Attributes.

```bash
temporal operator search-attribute list
```

---

#### temporal operator search-attribute remove

**Summary:** Remove custom Search Attributes.

```bash
temporal operator search-attribute remove \
    --name YourAttributeName \
    --yes
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` | string[] | — | **Yes** | Search Attribute name. Can be passed multiple times |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm removal |

---

### temporal schedule

**Summary:** Create, use, and update Schedules for Workflow Executions.

```bash
temporal schedule describe \
    --schedule-id "YourScheduleId"
```

Uses the `client` option set.

---

#### temporal schedule backfill

**Summary:** Batch-execute actions for a specified time interval.

```bash
temporal schedule backfill \
    --schedule-id "YourScheduleId" \
    --start-time "2022-05-01T00:00:00Z" \
    --end-time "2022-05-31T23:59:59Z" \
    --overlap-policy BufferAll
```

Uses `overlap-policy` and `schedule-id` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--end-time` | timestamp | — | **Yes** | Backfill end time |
| `--start-time` | timestamp | — | **Yes** | Backfill start time |

---

#### temporal schedule create

**Summary:** Create a new Schedule.

```bash
temporal schedule create \
    --schedule-id "YourScheduleId" \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"30"}' \
    --workflow-id YourBaseWorkflowIdName \
    --task-queue YourTaskQueue \
    --type YourWorkflowType
```

Uses `schedule-configuration`, `schedule-create-only`, `schedule-id`, `overlap-policy`, `shared-workflow-start`, and `payload-input` option sets.

---

#### temporal schedule delete

**Summary:** Remove a Schedule (does not affect running Workflow Executions).

```bash
temporal schedule delete \
    --schedule-id YourScheduleId
```

Uses `schedule-id` option set.

---

#### temporal schedule describe

**Summary:** Display Schedule state including past, current, and future runs.

```bash
temporal schedule describe \
    --schedule-id YourScheduleId
```

Uses `schedule-id` option set.

---

#### temporal schedule list

**Summary:** List Schedules hosted by a Namespace.

```bash
temporal schedule list \
    --namespace YourNamespace
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--long` / `-l` | bool | false | No | Show detailed information |
| `--really-long` | bool | false | No | Show extensive information in non-table form |
| `--query` / `-q` | string | — | No | Filter results using given List Filter |

---

#### temporal schedule list-matching-times

🧪 Experimental

**Summary:** List times a Schedule's spec would match within a given range.

```bash
temporal schedule list-matching-times \
    --schedule-id "YourScheduleId" \
    --start-time "2024-01-01T00:00:00Z" \
    --end-time "2024-01-31T23:59:59Z"
```

Uses `schedule-id` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--start-time` | timestamp | — | **Yes** | Start of time range |
| `--end-time` | timestamp | — | **Yes** | End of time range |

---

#### temporal schedule toggle

**Summary:** Pause or unpause a Schedule.

```bash
temporal schedule toggle \
    --schedule-id "YourScheduleId" \
    --pause \
    --reason "YourReason"

temporal schedule toggle \
    --schedule-id "YourScheduleId" \
    --unpause \
    --reason "YourReason"
```

Uses `schedule-id` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--pause` | bool | false | No | Pause the Schedule |
| `--unpause` | bool | false | No | Unpause the Schedule |
| `--reason` | string | `(no reason provided)` | No | Reason for pausing or unpausing |

---

#### temporal schedule trigger

**Summary:** Trigger a Schedule to run immediately.

```bash
temporal schedule trigger \
    --schedule-id "YourScheduleId"
```

Uses `schedule-id` and `overlap-policy` option sets.

---

#### temporal schedule update

**Summary:** Update Schedule details (full replacement).

```bash
temporal schedule update \
    --schedule-id "YourScheduleId" \
    --workflow-id YourBaseWorkflowIdName \
    --task-queue YourTaskQueue \
    --type YourWorkflowType
```

> **Note:** This performs a full replacement. Re-specify all desired options. Memo and search attributes cannot be updated here; they are set only at creation.

Uses `schedule-configuration`, `schedule-id`, `overlap-policy`, `shared-workflow-start`, and `payload-input` option sets.

---

### temporal server start-dev

**Summary:** Start Temporal development server.

> ⚠️ **WARNING:** The development server is not intended for production use. It skips certain HTTP security checks.

```bash
temporal server start-dev

# With persistent storage:
temporal server start-dev \
    --db-filename path-to-your-local-persistent-store

# Custom ports:
temporal server start-dev \
    --port 7234 \
    --ui-port 8234 \
    --metrics-port 57271
```

The Web UI is available at `http://localhost:8233` by default.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--db-filename` / `-f` | string | — | No | Path for persistent Temporal state store (SQLite). Without this, state is lost on restart |
| `--namespace` / `-n` | string[] | — | No | Namespaces to create at launch. "default" is always created |
| `--port` / `-p` | int | `7233` | No | Port for the front-end gRPC Service |
| `--http-port` | int | `0` (random) | No | Port for the HTTP API service |
| `--metrics-port` | int | — (random) | No | Port for the `/metrics` HTTP endpoint |
| `--ui-port` | int | `port + 1000` | No | Port for the Web UI |
| `--headless` | bool | false | No | Disable the Web UI |
| `--ip` | string | `localhost` | No | IP address bound to the front-end Service |
| `--ui-ip` | string | same as `--ip` | No | IP address bound to the Web UI |
| `--ui-public-path` | string | `/` | No | The public base path for the Web UI |
| `--ui-asset-path` | string | — | No | UI custom assets path |
| `--ui-codec-endpoint` | string | — | No | UI remote codec HTTP endpoint |
| `--ui-disable-news-fetch` | bool | false | No | Disable the Web UI newsfeed |
| `--sqlite-pragma` | string[] | — | No | SQLite pragma statements in `PRAGMA=VALUE` format |
| `--dynamic-config-value` | string[] | — | No | Dynamic config values as `KEY=VALUE` pairs (JSON values). Repeatable |
| `--log-config` | bool | false | No | Print the server config to stderr |
| `--search-attribute` | string[] | — | No | Register Search Attributes as `KEY=VALUE` pairs (value = type name). Repeatable |

---

### temporal task-queue

**Summary:** Inspect and update Task Queues.

```bash
temporal task-queue [command] [command options] \
    --task-queue YourTaskQueue
```

Uses the `client` option set.

---

#### temporal task-queue describe

**Summary:** Show active Workers and statistics for a Task Queue.

```bash
temporal task-queue describe \
    --task-queue YourTaskQueue

# Specific type:
temporal task-queue describe \
    --task-queue YourTaskQueue \
    --task-queue-type "activity"

# With reachability:
temporal task-queue describe \
    --task-queue YourTaskQueue \
    --select-build-id "YourBuildId" \
    --report-reachability
```

Task reachability states (deprecated in favor of Worker Deployment Drainage Status):
- `Reachable`: Build ID may be used by new or open executions
- `ClosedWorkflowsOnly`: No open executions; may have closed within retention
- `Unreachable`: Not used for any executions within retention period

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |
| `--task-queue-type` | string-enum[] | — | No | Task Queue type(s): `workflow`, `activity`, `nexus`. All types if omitted |
| `--select-build-id` | string[] | — | No | Filter by Build ID. Repeatable |
| `--select-unversioned` | bool | false | No | Include the unversioned queue |
| `--select-all-active` | bool | false | No | Include all active versions |
| `--report-reachability` | bool | false | No | Display task reachability information |
| `--legacy-mode` | bool | false | No | Enable legacy mode for servers without rules-based versioning |
| `--task-queue-type-legacy` | string-enum | `workflow` | No | Task Queue type (legacy mode only): `workflow`, `activity` |
| `--partitions-legacy` | int | `1` | No | Query partitions 1 through N (legacy mode only) |
| `--disable-stats` | bool | false | No | Disable task queue statistics |
| `--report-config` | bool | false | No | Include task queue configuration (rate limits) in the response |

---

#### temporal task-queue get-build-id-reachability

⚠️ Deprecated

**Summary:** Show Build ID availability.

```bash
temporal task-queue get-build-id-reachability \
    --task-queue YourTaskQueue \
    --build-id "YourBuildId"
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--build-id` | string[] | — | No | One or more Build ID strings. Repeatable |
| `--reachability-type` | string-enum | `existing` | No | Reachability filter: `open`, `closed`, `existing` |
| `--task-queue` / `-t` | string[] | — | No | Limit to specific task queue(s). Repeatable |

---

#### temporal task-queue get-build-ids

⚠️ Deprecated

**Summary:** Fetch Build ID versions for a Task Queue.

```bash
temporal task-queue get-build-ids \
    --task-queue YourTaskQueue
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |
| `--max-sets` | int | `0` | No | Max return count (0 = all sets, 1 = default major version) |

---

#### temporal task-queue list-partition

**Summary:** List Task Queue partitions with assigned matching nodes.

```bash
temporal task-queue list-partition \
    --task-queue YourTaskQueue
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |

---

#### temporal task-queue update-build-ids

⚠️ Deprecated

**Summary:** Manage Build IDs for a Task Queue (Worker versioning).

Subcommands:

##### temporal task-queue update-build-ids add-new-compatible

Add a compatible Build ID to an existing version set:

```bash
temporal task-queue update-build-ids add-new-compatible \
    --task-queue YourTaskQueue \
    --existing-compatible-build-id "YourExistingBuildId" \
    --build-id "YourNewBuildId"
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--build-id` | string | — | **Yes** | Build ID to be added |
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |
| `--existing-compatible-build-id` | string | — | **Yes** | Pre-existing Build ID in this Task Queue |
| `--set-as-default` | bool | false | No | Set the expanded Build ID set as the Task Queue default |

##### temporal task-queue update-build-ids add-new-default

⚠️ Deprecated

Create a new Task Queue Build ID set and make it the overall default:

```bash
temporal task-queue update-build-ids add-new-default \
    --task-queue YourTaskQueue \
    --build-id "YourNewBuildId"
```

---

#### temporal task-queue versioning

⚠️ Deprecated — This API has been deprecated by Worker Deployment.

**Summary:** Manage Task Queue Build ID assignment and redirect rules.

Subcommands (all deprecated):

| Subcommand | Summary |
|---|---|
| `add-redirect-rule` | Add a redirect rule for a source Build ID to a target Build ID |
| `commit-build-id` | Complete a Build ID rollout and clean up rules |
| `delete-assignment-rule` | Delete assignment rule by index |
| `delete-redirect-rule` | Delete redirect rule for a source Build ID |
| `get-rules` | Fetch all assignment and redirect rules |
| `insert-assignment-rule` | Insert a new assignment rule at an index |
| `replace-assignment-rule` | Replace assignment rule at an index |
| `replace-redirect-rule` | Replace target Build ID for a redirect rule |

---

#### temporal task-queue config get

**Summary:** Retrieve the current configuration for a Task Queue.

```bash
temporal task-queue config get \
    --task-queue YourTaskQueue \
    --task-queue-type activity
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |
| `--task-queue-type` | string-enum | — | **Yes** | Task Queue type: `workflow`, `activity`, `nexus` |

---

#### temporal task-queue config set

**Summary:** Update configuration settings for a Task Queue.

```bash
temporal task-queue config set \
    --task-queue YourTaskQueue \
    --task-queue-type activity \
    --queue-rps-limit 100.0 \
    --fairness-key-weight HighPriority=2.0 \
    --fairness-key-weight LowPriority=0.5
```

Pass `default` to unset a rate limit (e.g., `--queue-rps-limit default`).

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--task-queue` / `-t` | string | — | **Yes** | Task Queue name |
| `--task-queue-type` | string-enum | — | **Yes** | Task Queue type: `workflow`, `activity`, `nexus` |
| `--queue-rps-limit` | float\|default | — | No | Queue rate limit in requests per second. Pass `default` to unset |
| `--queue-rps-limit-reason` | string | — | No | Reason for queue rate limit update |
| `--fairness-key-rps-limit-default` | float\|default | — | No | Fairness key rate limit default in RPS. Pass `default` to unset |
| `--fairness-key-rps-limit-reason` | string | — | No | Reason for fairness key rate limit update |
| `--fairness-key-weight` | string[] | — | No | Fairness key weight overrides as `key=weight` or `key=default`. Repeatable |
| `--fairness-key-weight-clear-all` | bool | false | No | Unset all fairness key weight overrides. Cannot be used with `--fairness-key-weight` |

---

### temporal worker

**Summary:** Read or update Worker state.

```bash
temporal worker deployment
```

Uses the `client` option set.

---

#### temporal worker deployment

**Summary:** Describe, list, and operate on Worker Deployments and Versions.

```bash
temporal worker deployment [command] [options]
```

---

##### temporal worker deployment create

🧪 Experimental

**Summary:** Create a new Worker Deployment.

Uses `deployment-name` option set.

---

##### temporal worker deployment describe

**Summary:** Show properties of a Worker Deployment (versions, routing info, creation time).

```bash
temporal worker deployment describe \
    --name YourDeploymentName
```

Uses `deployment-name` option set.

---

##### temporal worker deployment delete

**Summary:** Delete a Worker Deployment (must have no Versions).

```bash
temporal worker deployment delete \
    --name YourDeploymentName
```

Uses `deployment-name` option set.

---

##### temporal worker deployment list

**Summary:** List existing Worker Deployments.

```bash
temporal worker deployment list \
    --namespace YourDeploymentNamespace
```

---

##### temporal worker deployment create-version

🧪 Experimental

**Summary:** Create a new Worker Deployment Version with compute configuration (AWS Lambda or GCP Cloud Run).

```bash
temporal worker deployment create-version \
    --deployment-name YourDeploymentName \
    --build-id YourBuildID \
    --aws-lambda-function-arn LambdaFunctionARN \
    --aws-lambda-assume-role-arn LambdaAssumeRoleARN \
    --aws-lambda-assume-role-external-id LambdaAssumeRoleExternalID
```

Uses `deployment-version` option set. See the full list of AWS Lambda and GCP Cloud Run flags in the YAML source.

---

##### temporal worker deployment describe-version

**Summary:** Show properties of a Worker Deployment Version.

```bash
temporal worker deployment describe-version \
    --deployment-name YourDeploymentName --build-id YourBuildID
```

Uses `deployment-version` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--report-task-queue-stats` | bool | false | No | Report stats for task queues in this version |

---

##### temporal worker deployment delete-version

**Summary:** Delete a Worker Deployment Version (must not be Current/Ramping and have no active pollers).

```bash
temporal worker deployment delete-version \
    --deployment-name YourDeploymentName --build-id YourBuildID \
    --skip-drainage
```

Uses `deployment-version` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--skip-drainage` | bool | false | No | Ignore the deletion requirement of not draining |

---

##### temporal worker deployment set-current-version

**Summary:** Make a Worker Deployment Version Current for a Deployment.

```bash
temporal worker deployment set-current-version \
    --deployment-name YourDeploymentName --build-id YourBuildID

# Set to unversioned:
temporal worker deployment set-current-version \
    --deployment-name YourDeploymentName --unversioned
```

Uses `deployment-version-or-unversioned` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--ignore-missing-task-queues` | bool | false | No | Override protection to accidentally remove task queues |
| `--allow-no-pollers` | bool | false | No | Set version as current even if it has no pollers |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm |

---

##### temporal worker deployment set-ramping-version

**Summary:** Change Version Ramping settings for a Worker Deployment.

```bash
temporal worker deployment set-ramping-version \
    --deployment-name YourDeploymentName --build-id YourBuildID \
    --percentage 10.0

# Remove ramping:
temporal worker deployment set-ramping-version \
    --deployment-name YourDeploymentName --build-id YourBuildID \
    --delete
```

Uses `deployment-version-or-unversioned` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--percentage` | float | — | No | Percentage of tasks redirected to the Ramping Version (0–100) |
| `--delete` | bool | false | No | Delete the Ramping Version |
| `--ignore-missing-task-queues` | bool | false | No | Override protection to accidentally remove task queues |
| `--allow-no-pollers` | bool | false | No | Set version as ramping even if it has no pollers |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm |

---

##### temporal worker deployment update-version-metadata

**Summary:** Update user-provided metadata for a Deployment Version.

```bash
temporal worker deployment update-version-metadata \
    --deployment-name YourDeploymentName --build-id YourBuildID \
    --metadata bar=1 \
    --metadata foo=true
```

Uses `deployment-version` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--metadata` | string[] | — | No | Set deployment metadata as `KEY="VALUE"` pairs. Repeatable |
| `--remove-entries` | string[] | — | No | Keys to delete from metadata. Repeatable |

---

##### temporal worker deployment manager-identity set / unset

**Summary:** Set or unset the Manager Identity of a Worker Deployment.

```bash
temporal worker deployment manager-identity set \
    --deployment-name DeploymentName \
    --self

temporal worker deployment manager-identity unset \
    --deployment-name YourDeploymentName
```

---

#### temporal worker list

🧪 Experimental

**Summary:** List worker status information in a namespace.

```bash
temporal worker list --namespace YourNamespace --query 'TaskQueue="YourTaskQueue"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter |
| `--limit` | int | 0 | No | Maximum number of workers to display |
| `--include-system-workers` | bool | false | No | Include system workers created by the server |

---

#### temporal worker count

🧪 Experimental

**Summary:** Count workers in a namespace.

```bash
temporal worker count --namespace YourNamespace --query 'TaskQueue="YourTaskQueue"'
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter |
| `--include-system-workers` | bool | false | No | Include system workers created by the server |

---

#### temporal worker describe

🧪 Experimental

**Summary:** Returns information about a specific worker.

```bash
temporal worker describe --namespace YourNamespace --worker-instance-key YourKey
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--worker-instance-key` | string | — | **Yes** | Worker instance key to describe |

---

### temporal workflow

**Summary:** Start, list, and operate on Workflows.

```bash
temporal workflow [command] [options]
```

Uses the `client` option set.

---

#### temporal workflow cancel

**Summary:** Send cancellation to Workflow Execution.

Records a `WorkflowExecutionCancelRequested` event; allows cleanup work.

```bash
temporal workflow cancel \
    --workflow-id YourWorkflowId

# Bulk:
temporal workflow cancel \
    --query YourQuery
```

Uses `single-workflow-or-batch` option set.

---

#### temporal workflow count

**Summary:** Count Workflow Executions.

```bash
temporal workflow count \
    --query YourQuery
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter |

---

#### temporal workflow delete

**Summary:** Remove Workflow Execution and its Event History.

```bash
temporal workflow delete \
    --workflow-id YourWorkflowId
```

> **Note:** Deleting in a global Namespace removes from all replicas. Use `--grpc-meta xdc-redirection=false` to target passive clusters directly.

Uses `single-workflow-or-batch` option set.

---

#### temporal workflow describe

**Summary:** Show Workflow Execution info.

```bash
temporal workflow describe \
    --workflow-id YourWorkflowId

temporal workflow describe \
    --workflow-id YourWorkflowId \
    --reset-points true
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reset-points` | bool | false | No | Show auto-reset points only |
| `--raw` | bool | false | No | Print properties without changing their format |

---

#### temporal workflow execute

**Summary:** Start new Workflow Execution and wait for it to complete (blocks).

```bash
temporal workflow execute \
    --workflow-id YourWorkflowId \
    --type YourWorkflow \
    --task-queue YourTaskQueue \
    --input '{"some-key": "some-value"}'
```

Uses `shared-workflow-start`, `workflow-start`, and `payload-input` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--detailed` | bool | false | No | Display events as sections instead of table (not for JSON output) |

---

#### temporal workflow fix-history-json

**Summary:** Reserialize an Event History JSON file.

```bash
temporal workflow fix-history-json \
    --source /path/to/original.json \
    --target /path/to/reserialized.json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--source` / `-s` | string | — | **Yes** | Path to the original file |
| `--target` / `-t` | string | — | No | Path to results file. If omitted, output goes to stdout |

---

#### temporal workflow list

**Summary:** Show Workflow Executions.

```bash
temporal workflow list \
    --query YourQuery

# Archived:
temporal workflow list \
    --archived
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter |
| `--archived` | bool | false | No | 🧪 Limit output to archived Workflow Executions |
| `--limit` | int | 0 | No | Maximum number of Workflow Executions to display |
| `--page-size` | int | 0 | No | Maximum number to fetch at a time from the server |

---

#### temporal workflow metadata

**Summary:** Query the Workflow for user-specified metadata (summary and details).

```bash
temporal workflow metadata \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` and `query-modifiers` option sets.

---

#### temporal workflow pause

🧪 Experimental

**Summary:** Pause a Workflow Execution.

```bash
temporal workflow pause \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for pausing. Defaults to current user's name |

---

#### temporal workflow query

**Summary:** Retrieve Workflow Execution state via a Query.

```bash
temporal workflow query \
    --workflow-id YourWorkflowId \
    --name YourQueryType \
    --input '{"YourInputKey": "YourInputValue"}'
```

Uses `payload-input`, `workflow-reference`, and `query-modifiers` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` / `--type` | string | — | **Yes** | Query Type/Name |

---

#### temporal workflow reset

**Summary:** Move Workflow Execution history point (reset without losing progress).

```bash
temporal workflow reset \
    --workflow-id YourWorkflowId \
    --event-id YourLastEvent

temporal workflow reset \
    --workflow-id YourWorkflowId \
    --type LastContinuedAsNew
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--workflow-id` / `-w` | string | — | No | Workflow ID (required for non-batch reset) |
| `--run-id` / `-r` | string | — | No | Run ID |
| `--event-id` / `-e` | int | — | No | Event ID to reset to |
| `--reason` | string | — | **Yes** | Reason for reset |
| `--reapply-type` | string-enum | `All` | No | ⚠️ Deprecated. Use `--reapply-exclude` instead. Types: `All`, `Signal`, `None` |
| `--reapply-exclude` | string-enum[] | — | No | Exclude event types from re-application: `All`, `Signal`, `Update` |
| `--type` / `-t` | string-enum | — | No | Event type: `FirstWorkflowTask`, `LastWorkflowTask`, `LastContinuedAsNew`, `BuildId` |
| `--build-id` | string | — | No | Build ID. Use only with `BuildId` `--type` |
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter (for batch reset) |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm. Only when `--query` is present |

Subcommand:

##### temporal workflow reset with-workflow-update-options

Run Workflow Update Options atomically after the Workflow is reset. Uses `workflow-update-options` option set.

---

#### temporal workflow result

**Summary:** Wait for and show the result of a Workflow Execution.

```bash
temporal workflow result \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` option set.

---

#### temporal workflow show

**Summary:** Display Event History.

```bash
temporal workflow show \
    --workflow-id YourWorkflowId \
    --output json
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--follow` / `-f` | bool | false | No | Follow Workflow Execution progress in real time |
| `--detailed` | bool | false | No | Display events as detailed sections instead of table |
| `--reverse` | bool | false | No | Fetch Event History newest-event-first. Cannot combine with `--follow` |

---

#### temporal workflow signal

**Summary:** Send an asynchronous Signal to a running Workflow Execution.

```bash
temporal workflow signal \
    --workflow-id YourWorkflowId \
    --name YourSignal \
    --input '{"YourInputKey": "YourInputValue"}'
```

Uses `single-workflow-or-batch` and `payload-input` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` / `--type` | string | — | **Yes** | Signal name |

---

#### temporal workflow signal-with-start

**Summary:** Send a Signal to a Workflow, starting it if not running.

```bash
temporal workflow signal-with-start \
    --signal-name YourSignal \
    --signal-input '{"some-key": "some-value"}' \
    --workflow-id YourWorkflowId \
    --type YourWorkflowType \
    --task-queue YourTaskQueue \
    --input '{"some-key": "some-value"}'
```

Uses `shared-workflow-start`, `workflow-start`, and `payload-input` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--signal-name` / `--signal-type` | string | — | **Yes** | Signal name |
| `--signal-input` | string[] | — | No | Signal input value. Repeatable for multiple arguments |
| `--signal-input-file` | string[] | — | No | Path(s) to signal input file(s). Can't combine with `--signal-input` |
| `--signal-input-meta` | string[] | — | No | Signal payload metadata as `KEY=VALUE` pairs. Repeatable |
| `--signal-input-base64` | bool | false | No | Assume signal inputs are base64-encoded |

---

#### temporal workflow stack

**Summary:** Trace a Workflow Execution's current call stack.

```bash
temporal workflow stack \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reject-condition` | string-enum | — | No | Reject Queries based on Workflow state: `not_open`, `not_completed_cleanly` |

---

#### temporal workflow start

**Summary:** Initiate a Workflow Execution. Returns Workflow- and Run-IDs.

```bash
temporal workflow start \
    --workflow-id YourWorkflowId \
    --type YourWorkflow \
    --task-queue YourTaskQueue \
    --input '{"some-key": "some-value"}'
```

Uses `shared-workflow-start`, `workflow-start`, and `payload-input` option sets.

---

#### temporal workflow terminate

**Summary:** Forcefully end a Workflow Execution.

```bash
temporal workflow terminate \
    --reason YourReasonForTermination \
    --workflow-id YourWorkflowId

# Bulk termination:
temporal workflow terminate \
    --query YourQuery \
    --reason YourReasonForTermination
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--workflow-id` / `-w` | string | — | No | Workflow ID. Either this or `--query` required |
| `--query` / `-q` | string | — | No | SQL-like QUERY List Filter. Either this or `--workflow-id` required |
| `--run-id` / `-r` | string | — | No | Run ID. Only with `--workflow-id` |
| `--reason` | string | — | No | Reason for termination. Defaults to current user's name |
| `--yes` / `-y` | bool | false | No | Don't prompt to confirm. Only with `--query` |
| `--rps` | float | — | No | Limit batch requests per second. Only if query is present |

---

#### temporal workflow trace

**Summary:** Display live progress of a Workflow Execution and its child workflows.

```bash
temporal workflow trace \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--fold` | string[] | — | No | Fold away Child Workflows with specified statuses. Repeatable |
| `--no-fold` | bool | false | No | Disable folding; fetch and display all Child Workflows |
| `--depth` | int | `-1` | No | Depth for Child Workflow fetches. -1 for any depth |
| `--concurrency` | int | `10` | No | Number of Workflow Histories to fetch at a time |

---

#### temporal workflow unpause

🧪 Experimental

**Summary:** Unpause a previously paused Workflow Execution.

```bash
temporal workflow unpause \
    --workflow-id YourWorkflowId
```

Uses `workflow-reference` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--reason` | string | — | No | Reason for unpausing. Defaults to current user's name |

---

#### temporal workflow update

**Summary:** Send and interact with Workflow Updates.

An Update is a synchronous call that can change Workflow state, control its flow, and return a result.

Subcommands:

##### temporal workflow update describe

Obtain status info about a specific Update.

```bash
temporal workflow update describe \
    --workflow-id YourWorkflowId \
    --update-id YourUpdateId
```

Uses `update-targeting` option set.

##### temporal workflow update execute

Send an Update and wait for it to complete.

```bash
temporal workflow update execute \
    --workflow-id YourWorkflowId \
    --name YourUpdate \
    --input '{"some-key": "some-value"}'
```

Uses `update-starting` and `payload-input` option sets.

##### temporal workflow update result

Wait for a specific Update to complete.

```bash
temporal workflow update result \
    --workflow-id YourWorkflowId \
    --update-id YourUpdateId
```

Uses `update-targeting` option set.

##### temporal workflow update start

Send an Update and wait for it to be accepted or rejected.

```bash
temporal workflow update start \
    --workflow-id YourWorkflowId \
    --name YourUpdate \
    --input '{"some-key": "some-value"}' \
    --wait-for-stage accepted
```

Uses `update-starting` and `payload-input` option sets.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--wait-for-stage` | string-enum | — | **Yes** | Update stage to wait for. Only valid value: `accepted` |

---

#### temporal workflow update-options

**Summary:** Change Workflow Execution Options (versioning overrides).

```bash
# Auto-upgrade:
temporal workflow update-options \
    --workflow-id YourWorkflowId \
    --versioning-override-behavior auto_upgrade

# Pin to a deployment version:
temporal workflow update-options \
    --workflow-id YourWorkflowId \
    --versioning-override-behavior pinned \
    --versioning-override-deployment-name YourDeploymentName \
    --versioning-override-build-id YourDeploymentBuildId

# Remove overrides:
temporal workflow update-options \
    --workflow-id YourWorkflowId \
    --versioning-override-behavior unspecified
```

Uses `single-workflow-or-batch` option set.

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--versioning-override-behavior` | string-enum | — | **Yes** | Override behavior: `unspecified`, `pinned`, `one_time`, `auto_upgrade` |
| `--versioning-override-deployment-name` | string | — | No | Deployment Name for `pinned` or `one_time` behaviors |
| `--versioning-override-build-id` | string | — | No | Build ID for `pinned` or `one_time` behaviors |

---

#### temporal workflow start-update-with-start

🧪 Experimental

**Summary:** Send an Update-With-Start and wait for it to be accepted or rejected.

```bash
temporal workflow start-update-with-start \
    --update-name YourUpdate \
    --update-input '{"update-key": "update-value"}' \
    --update-wait-for-stage accepted \
    --workflow-id YourWorkflowId \
    --type YourWorkflowType \
    --task-queue YourTaskQueue \
    --id-conflict-policy Fail \
    --input '{"wf-key": "wf-value"}'
```

Uses `shared-workflow-start`, `workflow-start`, and `payload-input` option sets.

---

#### temporal workflow execute-update-with-start

🧪 Experimental

**Summary:** Send an Update-With-Start and wait for it to complete.

```bash
temporal workflow execute-update-with-start \
    --update-name YourUpdate \
    --update-input '{"update-key": "update-value"}' \
    --workflow-id YourWorkflowId \
    --type YourWorkflowType \
    --task-queue YourTaskQueue \
    --id-conflict-policy Fail \
    --input '{"wf-key": "wf-value"}'
```

Uses `shared-workflow-start`, `workflow-start`, and `payload-input` option sets.

---

## Option Sets

Reusable option sets that are composed into commands.

### `common`

Applied to every command. Controls output formatting, logging, and configuration. See [Common Flags](#common-flags).

### `client`

Applied to commands that connect to a Temporal Service. See [Client / Connection Flags](#client--connection-flags).

### `payload-input`

Input payload options used by commands that accept data:

| Flag | Type | Description |
|---|---|---|
| `--input` / `-i` | string[] | Input value (JSON). Repeatable for multiple arguments. Can't combine with `--input-file` |
| `--input-file` | string[] | Path(s) to input file(s). Can't combine with `--input`. Repeatable |
| `--input-meta` | string[] | Payload metadata as `KEY=VALUE`. Repeatable. Repeated keys applied in order to corresponding inputs |
| `--input-base64` | bool | Assume inputs are base64-encoded |

### `workflow-reference`

Used by commands targeting a specific Workflow Execution:

| Flag | Type | Required | Description |
|---|---|---|---|
| `--workflow-id` / `-w` | string | **Yes** | Workflow ID |
| `--run-id` / `-r` | string | No | Run ID |

### `single-workflow-or-batch`

Used by commands that operate on one Workflow or a batch via query:

| Flag | Type | Required | Description |
|---|---|---|---|
| `--workflow-id` / `-w` | string | No | Workflow ID. Either this or `--query` required |
| `--query` / `-q` | string | No | SQL-like QUERY List Filter. Either this or `--workflow-id` required |
| `--run-id` / `-r` | string | No | Run ID. Only with `--workflow-id` |
| `--reason` | string | No | Reason for batch operation |
| `--yes` / `-y` | bool | No | Don't prompt to confirm. Only when `--query` is present |
| `--rps` | float | No | Limit batch requests per second |
| `--headers` | string[] | No | Temporal workflow headers as `KEY=VALUE` pairs. Repeatable |

### `shared-workflow-start`

Common options for starting a Workflow Execution:

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--workflow-id` / `-w` | string | — | No | Workflow ID (auto-generated if not provided) |
| `--type` / `--name` | string | — | **Yes** | Workflow Type name |
| `--task-queue` / `-t` | string | — | **Yes** | Workflow Task queue |
| `--run-timeout` | duration | — | No | Fail a Workflow Run after this duration |
| `--execution-timeout` | duration | — | No | Fail a WorkflowExecution after this duration (includes retries and ContinueAsNew) |
| `--task-timeout` | duration | `10s` | No | Fail a Workflow Task after this duration |
| `--search-attribute` | string[] | — | No | Search Attributes as `KEY=VALUE` pairs. Repeatable |
| `--headers` | string[] | — | No | Workflow headers as `KEY=VALUE` pairs. Repeatable |
| `--memo` | string[] | — | No | Memo as `KEY="VALUE"` pairs |
| `--static-summary` | string | — | No | 🧪 Static Workflow summary (single line, Temporal Markdown) |
| `--static-details` | string | — | No | 🧪 Static Workflow details (multi-line, Temporal Markdown) |
| `--priority-key` | int | `3` | No | Priority key (1–5, lower = higher priority) |
| `--fairness-key` | string | — | No | Fairness key (max 64 bytes) for proportional task dispatch |
| `--fairness-weight` | float | — | No | Weight [0.001–1000] for this fairness key |

### `workflow-start`

Additional options for workflow start/execute:

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--cron` | string | — | No | ⚠️ Deprecated. Use Schedules instead |
| `--fail-existing` | bool | false | No | Fail if the Workflow already exists |
| `--start-delay` | duration | — | No | Delay before starting. Can't combine with cron |
| `--id-reuse-policy` | string-enum | — | No | Workflow ID reuse policy: `AllowDuplicate`, `AllowDuplicateFailedOnly`, `RejectDuplicate`, `TerminateIfRunning` |
| `--id-conflict-policy` | string-enum | — | No | Workflow ID conflict policy for open executions: `Fail`, `UseExisting`, `TerminateExisting` |

### `schedule-configuration`

Options for defining a Schedule's time spec:

| Flag | Type | Description |
|---|---|---|
| `--calendar` | string[] | Calendar specification in JSON. Repeatable |
| `--catchup-window` | duration | Maximum catch-up time when Service is unavailable |
| `--cron` | string[] | Calendar spec in cron string format. Repeatable |
| `--end-time` | timestamp | Schedule end time |
| `--interval` | string[] | Interval duration (e.g., `90m` or `6h/5h`). Repeatable |
| `--jitter` | duration | Random variation from spec start time |
| `--notes` | string | Initial notes field value |
| `--paused` | bool | Pause the Schedule immediately on creation |
| `--pause-on-failure` | bool | Pause schedule after Workflow failures |
| `--remaining-actions` | int | Total allowed actions (0 = unlimited) |
| `--start-time` | timestamp | Schedule start time |
| `--time-zone` | string | Time zone for calendar specs |

### `overlap-policy`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--overlap-policy` | string-enum | `Skip` | Policy for overlapping Workflow Executions: `Skip`, `BufferOne`, `BufferAll`, `CancelOther`, `TerminateOther`, `AllowAll` |

### `activity-reference`

| Flag | Type | Required | Description |
|---|---|---|---|
| `--activity-id` / `-a` | string | **Yes** | Activity ID |
| `--run-id` / `-r` | string | No | Activity Run ID. If not set, targets the latest run |

### `activity-start`

Options for starting an Activity (used by `temporal activity start` and `temporal activity execute`):

| Flag | Type | Required | Description |
|---|---|---|---|
| `--activity-id` / `-a` | string | **Yes** | Activity ID |
| `--type` | string | **Yes** | Activity Type name |
| `--task-queue` / `-t` | string | **Yes** | Activity task queue |
| `--schedule-to-close-timeout` | duration | No | Max time for Execution including all retries |
| `--schedule-to-start-timeout` | duration | No | Max time task can wait in queue before pickup |
| `--start-to-close-timeout` | duration | No | Max time for a single attempt |
| `--heartbeat-timeout` | duration | No | Max time between heartbeats |
| `--start-delay` | duration | No | Delay before dispatching the first Activity task |
| `--retry-initial-interval` | duration | No | Interval of the first retry |
| `--retry-maximum-interval` | duration | No | Maximum retry interval |
| `--retry-backoff-coefficient` | float | No | Retry backoff coefficient (≥ 1) |
| `--retry-maximum-attempts` | int | No | Max retry attempts (0 = unlimited, 1 = disabled) |
| `--id-reuse-policy` | string-enum | No | `AllowDuplicate`, `AllowDuplicateFailedOnly`, `RejectDuplicate` |
| `--id-conflict-policy` | string-enum | No | `Fail`, `UseExisting` |
| `--search-attribute` | string[] | No | Search Attributes as `KEY=VALUE` pairs. Repeatable |
| `--headers` | string[] | No | Activity headers as `KEY=VALUE` pairs. Repeatable |
| `--static-summary` | string | No | 🧪 Static Activity summary |
| `--static-details` | string | No | 🧪 Static Activity details |
| `--priority-key` | int | No | Priority key (1–5, lower = higher priority, default 3) |
| `--fairness-key` | string | No | Fairness key (max 64 bytes) |
| `--fairness-weight` | float | No | Weight [0.001–1000] for this fairness key |

### `nexus-endpoint-config`

| Flag | Type | Description |
|---|---|---|
| `--description` | string | Nexus Endpoint description (Markdown supported) |
| `--description-file` | string | Path to description file (Markdown supported) |
| `--target-namespace` | string | Namespace where handler Worker polls for Nexus tasks |
| `--target-task-queue` | string | Task Queue that handler Worker polls |
| `--target-url` | string | 🧪 External Nexus Endpoint URL (alternative to namespace+task-queue) |

### `nexus-operation-start`

| Flag | Type | Required | Description |
|---|---|---|---|
| `--endpoint` | string | **Yes** | Nexus Endpoint name |
| `--service` | string | **Yes** | Nexus Service name |
| `--operation` | string | **Yes** | Nexus Operation name |
| `--operation-id` | string | **Yes** | Nexus Operation ID |
| `--schedule-to-close-timeout` | duration | No | Total time operation is allowed to run |
| `--schedule-to-start-timeout` | duration | No | Max time to wait for operation to be started |
| `--start-to-close-timeout` | duration | No | Max time for async operation after start |
| `--id-conflict-policy` | string-enum | No | `Fail`, `UseExisting`, `TerminateExisting` |
| `--id-reuse-policy` | string-enum | No | `AllowDuplicate`, `RejectDuplicate` |
| `--search-attribute` | string[] | No | Search Attributes as `KEY=VALUE`. Repeatable |
| `--static-summary` | string | No | 🧪 Static summary (single line, Temporal Markdown) |

### `update-starting`

| Flag | Type | Required | Description |
|---|---|---|---|
| `--name` / `--type` | string | **Yes** | Handler method name |
| `--workflow-id` / `-w` | string | **Yes** | Workflow ID |
| `--update-id` | string | No | Update ID. Defaults to UUID |
| `--run-id` / `-r` | string | No | Run ID |
| `--first-execution-run-id` | string | No | Parent Run ID for chained Workflow Executions |
| `--headers` | string[] | No | Workflow headers as `KEY=VALUE`. Repeatable |

### `deployment-version`

| Flag | Type | Required | Description |
|---|---|---|---|
| `--deployment-name` | string | **Yes** | Name of the Worker Deployment |
| `--build-id` | string | **Yes** | Build ID of the Worker Deployment Version |

### `deployment-version-or-unversioned`

| Flag | Type | Required | Description |
|---|---|---|---|
| `--deployment-name` | string | **Yes** | Name of the Worker Deployment |
| `--build-id` | string | No | Build ID. Required unless `--unversioned` is specified |
| `--unversioned` | bool | No | Set unversioned workers as the target version |

---

## Dev Server

`temporal server start-dev` runs an embedded, single-binary Temporal development server suitable for local development.

**Key characteristics:**

- **Embedded SQLite:** Uses an in-memory SQLite database by default. Pass `--db-filename` to persist state across server restarts.
- **Embedded UI:** The Temporal Web UI is served at `http://localhost:<ui-port>` (default: gRPC port + 1000 = `8233`). Disable with `--headless`.
- **Dynamic Config:** Inject server dynamic config values at startup with `--dynamic-config-value KEY=VALUE` (values must be JSON). Can be passed multiple times.
- **Namespaces:** The `default` Namespace is always created. Add more at startup with `--namespace`. Separate multiple namespace flags.
- **Search Attributes:** Pre-register Search Attributes at startup with `--search-attribute KEY=TYPE`.
- **IP Binding:** By default, the server binds to `localhost`. For Docker or remote access, bind with `--ip 0.0.0.0`.
- **SQLite pragmas:** Tune SQLite performance with `--sqlite-pragma PRAGMA=VALUE`.

```bash
# Full-featured dev server example:
temporal server start-dev \
    --db-filename ./temporal.db \
    --namespace my-namespace \
    --search-attribute CustomAttr=Keyword \
    --dynamic-config-value frontend.enableUpdateWorkflowExecution=true \
    --ui-codec-endpoint http://localhost:8081 \
    --port 7233 \
    --ui-port 8233
```

> ⚠️ Not intended for production. See https://docs.temporal.io/production-deployment for production deployments.

---

## Output Formats

The `--output` / `-o` flag controls the format of non-logging data output:

| Value | Description |
|---|---|
| `text` (default) | Human-readable tabular output with ANSI color when `--color auto` or `always` |
| `json` | Single JSON object or array per command. Use with SDK replay for event history |
| `jsonl` | JSON Lines (one JSON object per line). Useful for streaming or `jq` processing |
| `none` | Suppress all non-logging output |

**Examples:**

```bash
# JSON output for workflow list:
temporal workflow list -o json

# JSONL for streaming processing:
temporal workflow list -o jsonl | jq '.workflowId'

# JSON event history for replay:
temporal workflow show --workflow-id YourWorkflowId --output json
```

The `--no-json-shorthand-payloads` flag forces raw payload output even when JSON output mode is selected.

The `--time-format` flag controls how timestamps are displayed:
- `relative` (default): Human-friendly relative times (e.g., "2 minutes ago")
- `iso`: ISO 8601 format
- `raw`: Raw numeric value

---

## Environment & Configuration

The Temporal CLI supports two complementary configuration systems:

### Environment Files (`temporal env`)

Legacy key-value stores saved in `$HOME/.config/temporalio/temporal.yaml`. Each environment (e.g., `dev`, `prod`) is isolated.

```bash
# Set up a production environment:
temporal env set \
    --env prod \
    --key address \
    --value production.f45a2.tmprl.cloud:7233

# Use the environment:
temporal workflow list --env prod
```

The active environment is resolved in this order:
1. `--env` flag
2. `TEMPORAL_ENV` environment variable
3. Falls back to `default`

### Config Profiles (`temporal config`)

🧪 Experimental — TOML-based profiles stored at:
- Unix: `$HOME/.config/temporalio/temporal.toml`
- macOS: `$HOME/Library/Application Support/temporalio/temporal.toml`
- Windows: `%AppData%\temporalio\temporal.toml`

Override the config file path with `TEMPORAL_CONFIG_FILE` or `--config-file`.

```bash
# Set a value in a profile:
temporal config set \
    --profile prod \
    --prop address \
    --value us-east-1.aws.api.temporal.io:7233

# Read from a profile:
temporal workflow list --profile prod
```

The active profile is resolved in order:
1. `--profile` flag
2. `TEMPORAL_PROFILE` environment variable
3. Falls back to `default`

### Implied Environment Variables

Client flags can be set via environment variables (resolving order: flag > env var > config file):

| Flag | Environment Variable |
|---|---|
| `--address` | `TEMPORAL_ADDRESS` |
| `--namespace` | `TEMPORAL_NAMESPACE` |
| `--api-key` | `TEMPORAL_API_KEY` |
| `--tls` | `TEMPORAL_TLS` |
| `--tls-cert-path` | `TEMPORAL_TLS_CLIENT_CERT_PATH` |
| `--tls-cert-data` | `TEMPORAL_TLS_CLIENT_CERT_DATA` |
| `--tls-key-path` | `TEMPORAL_TLS_CLIENT_KEY_PATH` |
| `--tls-key-data` | `TEMPORAL_TLS_CLIENT_KEY_DATA` |
| `--tls-ca-path` | `TEMPORAL_TLS_SERVER_CA_CERT_PATH` |
| `--tls-ca-data` | `TEMPORAL_TLS_SERVER_CA_CERT_DATA` |
| `--tls-disable-host-verification` | `TEMPORAL_TLS_DISABLE_HOST_VERIFICATION` |
| `--tls-server-name` | `TEMPORAL_TLS_SERVER_NAME` |
| `--codec-endpoint` | `TEMPORAL_CODEC_ENDPOINT` |
| `--codec-auth` | `TEMPORAL_CODEC_AUTH` |
| `--env` | `TEMPORAL_ENV` |
| `--env-file` | `TEMPORAL_ENV_FILE` |
| `--config-file` | `TEMPORAL_CONFIG_FILE` |
| `--profile` | `TEMPORAL_PROFILE` |

---

## Codec Server Integration

A Codec Server decodes/encodes Temporal payloads (e.g., encrypted or compressed data) for display by the CLI and Web UI.

### Flags

| Flag | Env Var | Description |
|---|---|---|
| `--codec-endpoint` | `TEMPORAL_CODEC_ENDPOINT` | Remote Codec Server HTTP endpoint |
| `--codec-auth` | `TEMPORAL_CODEC_AUTH` | Authorization header value sent with Codec Server requests |
| `--codec-header` | — | Additional HTTP headers for codec server (`KEY=VALUE`, repeatable) |

### Usage

```bash
# Decode payloads using a local codec server:
temporal workflow show \
    --workflow-id YourWorkflowId \
    --codec-endpoint http://localhost:8081 \
    --codec-auth "Bearer YourToken"
```

For the Web UI, use `--ui-codec-endpoint` with `temporal server start-dev`:

```bash
temporal server start-dev \
    --ui-codec-endpoint http://localhost:8081
```

The CLI sends payload data to the codec endpoint for decoding before display. The `--codec-auth` flag sets the `Authorization` HTTP header on these requests. Use `--codec-header` for any additional headers required by your codec server.
