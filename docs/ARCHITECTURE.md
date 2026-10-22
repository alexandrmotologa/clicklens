# ClickLens Architecture Specification

This document details the internal design, system data flows, and component models implemented in ClickLens.

## High-Level Architecture

ClickLens runs as a self-contained Go binary hosting an embedded React 19 web studio and a command-line interface.

```
+-------------------------------------------------------------------------------+
|                            React 19 Web Interface                             |
|  * Parts Heatmap Grid            * Query Pipeline DAG Visualizer              |
|  * Merges and Mutations Stream   * Column Codec Optimizer                     |
|  * Process and Query Watcher     * Local File Explorer (Parquet / CSV)        |
+---------------------------------------+---------------------------------------+
                                        | (REST APIs / WebSocket Stream)
+---------------------------------------v---------------------------------------+
|                             ClickLens Go Backend                              |
|                                                                               |
|  [HTTP Router / WS Broker] <---> [Diagnostics Engine]                         |
|                                        |                                      |
|            +---------------------------+---------------------------+          |
|            |                           |                           |          |
|   [Native Client]             [Local Runner]             [File Inspector]     |
|   (clickhouse-go/v2)          (clickhouse-local)         (Pure Go Parquet)    |
+------------+---------------------------+---------------------------+----------+
             |                           |                           |
             v                           v                           v
     ClickHouse Cluster          CLI Local Files             Local Parquet Files
    (Ports 9000 / 8123)       (.parquet, .csv, .ndjson)         (Schema Fallback)
```

## System Components

### 1. Client Subsystem (`internal/client`)

- `native.go`: Manages connection pools to remote or local ClickHouse servers via TCP port 9000 or HTTP port 8123. Parses connection options including timeouts, read buffers, user authentication, and secure TLS configurations.
- `local_runner.go`: Wraps the `clickhouse-local` executable to execute analytical queries directly against local files (`SELECT * FROM file(...)`). Detects binary presence in PATH, WSL environments, or custom directory locations.
- `file_inspector.go`: Provides pure Go schema, column type, and metadata extraction for Parquet and CSV files when `clickhouse-local` is not installed on the system.
- `simulator.go`: Realistic in-memory telemetry provider. Emulates ClickHouse system tables (`system.parts`, `system.merges`, `system.columns`, `system.processes`, `system.mutations`, `system.query_log`) with dynamic time-varying data. Allows developers to test and verify all visualizers without spinning up a live ClickHouse daemon.

### 2. Inspector Subsystem (`internal/inspector`)

- `parts.go`: Reads metadata from `system.parts`. Computes part counts, active versus inactive parts, uncompressed versus compressed bytes, marks count, and partition fragmentation scores. Identifies partitions exceeding 300 active parts, warning of pending write throttling.
- `merges.go`: Subscribes to `system.merges`. Computes elapsed durations, merge completion percentages, read and write throughput in bytes and rows per second, and detects large part mergers that could cause I/O stalls.
- `mutations.go`: Queries `system.mutations` to track asynchronous table mutations triggered by `ALTER TABLE ... UPDATE` or `DELETE`. Records mutation status, remaining parts, and failure reasons.
- `columns.go`: Scans `system.columns`. Calculates compression ratios ($C_r = \text{uncompressed} / \text{compressed}$) and evaluates column types against optimal codec rules (DoubleDelta, Gorilla, T64, LowCardinality, ZSTD).
- `processes.go`: Inspects `system.processes` to track running queries, allocated memory, CPU elapsed time, and issues `KILL QUERY` commands.

### 3. Profiler Subsystem (`internal/profiler`)

- `explain_parser.go`: Parses text trees from `EXPLAIN PIPELINE` and `EXPLAIN PLAN` into a structured Directed Acyclic Graph (DAG).
  - Graph nodes represent operators (`FilterTransform`, `AggregatingTransform`, `SortingTransform`, `ExpressionTransform`, `ReadFromMergeTree`).
  - Graph edges convey stream counts and data flow directions.
  - Concurrency analysis flags thread bottlenecks where multi-stream processing collapses into a single-threaded serialization step.
- `memory_tracker.go`: Correlates queries with `system.query_log` to detect external aggregation or sorting disk spills triggered when memory limits (`max_bytes_before_external_group_by`, `max_bytes_before_external_sort`) are reached.

### 4. Server Subsystem (`internal/server`)

- `api.go`: Standardized REST endpoints delivering JSON payloads for parts, merges, columns, mutations, processes, and query execution.
- `routes.go`: Registers HTTP endpoints, CORS headers, and WebSocket broadcast channels.
- `static.go`: Embeds the precompiled React 19 production bundle into the binary using Go `embed.FS`, enabling single-file deployment.

## Telemetry Data Contract

The WebSocket endpoint `/api/ws` pushes events matching the following schema:

```json
{
  "type": "merges_update",
  "timestamp": "2026-10-22T12:00:00Z",
  "data": {
    "active_merges": [
      {
        "database": "telemetry",
        "table": "metrics",
        "elapsed_seconds": 3.42,
        "progress": 68.5,
        "num_parts": 6,
        "result_part_name": "all_1_6_1",
        "bytes_read": 14502010,
        "bytes_written": 9820100
      }
    ],
    "merges_rate_bytes_sec": 4238012
  }
}
```

## Security Considerations

1. Read-only permissions: ClickLens queries ClickHouse system tables (`system.*`). A read-only ClickHouse user account (`readonly = 1`) suffices for all monitoring operations.
2. Query execution limits: Ad-hoc queries in the Query Console observe ClickHouse server-side quotas, timeouts (`max_execution_time`), and memory limits (`max_memory_usage`).
3. Local file sandboxing: `clickhouse-local` queries only access paths permitted by host file permissions.
