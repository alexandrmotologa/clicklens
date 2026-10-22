# ClickLens: Master Implementation Plan

> Local ClickHouse Observability Studio, MergeTree Parts Visualizer and Query Memory Profiler
>
> Developer studio and diagnostics workbench for ClickHouse. Inspects active MergeTree data parts and background merges, profiles query execution pipelines with memory spillover warnings, analyzes column compression efficiency, and integrates with clickhouse-local for serverless profiling.

---

## 1. Overview and Architecture

ClickHouse stores analytical workloads across partitions and parts using the MergeTree engine family. Operating ClickHouse at scale requires direct visibility into storage layout and pipeline execution:
1. Part lifecycle: MergeTree tables continually merge smaller parts into larger blocks. Ingestion spikes or poorly chosen partition keys cause part explosion, leading to write rejection when active parts exceed threshold limits.
2. Memory consumption: Analytical pipelines with hash joins, large aggregations, or external sorting allocate memory across multiple execution streams. Diagnosing spikes requires inspecting the pipeline DAG and query log spill markers.
3. Column codec efficiency: Default LZ4 compression frequently wastes disk capacity on monotonic timestamps, sequential IDs, and floating point telemetry where Gorilla, DoubleDelta, or ZSTD codecs provide superior ratios.
4. Serverless data inspection: Ad-hoc analysis of Parquet and CSV files with clickhouse-local requires complex shell arguments without visual diagnostics.

ClickLens provides a unified Go backend and React 19 interface to inspect, profile, and optimize ClickHouse storage and queries.

```
+-----------------------------------------------------------------------------------+
|                                     CLICKLENS                                     |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|    MergeTree Storage Engine     |                 |    Query Pipeline Profiler    |
|  * Active and Inactive Parts    |                 |  * EXPLAIN PIPELINE Parser    |
|  * Background Merge Tracker     |                 |  * Per-Stream Concurrency DAG |
|  * Partition Health and Bloat   |                 |  * Memory Spillover Detection |
+-------+-------------------------+                 +-------+-----------------------+
        |                                                   |
        +---------------------------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |    Codec and Compression Auditor  |
                        |  * Compressed vs Uncompressed     |
                        |  * Codec Recommendations (ZSTD,   |
                        |    Gorilla, DoubleDelta, T64)     |
                        |  * Mutations and TTL Progress     |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Interactive Web Studio and CLI  |
                        |  * Partition Heatmap Grid         |
                        |  * Execution Stream Graph         |
                        |  * Serverless Local File Explorer |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
clicklens/
|-- cmd/
|   `-- clicklens/
|       `-- main.go                 # CLI entrypoint with subcommands
|-- internal/
|   |-- client/
|   |   |-- native.go               # ClickHouse native TCP/HTTP driver
|   |   |-- local_runner.go         # clickhouse-local execution wrapper
|   |   |-- file_inspector.go       # Pure Go Parquet and CSV schema reader
|   |   `-- simulator.go            # Realistic in-memory telemetry provider
|   |-- inspector/
|   |   |-- parts.go                # system.parts scanner and partition health
|   |   |-- merges.go               # system.merges real-time tracker
|   |   |-- mutations.go            # system.mutations observer
|   |   |-- columns.go              # system.columns compression auditor
|   |   `-- processes.go            # system.processes active queries
|   |-- profiler/
|   |   |-- explain_parser.go       # EXPLAIN PIPELINE and PLAN parser
|   |   |-- memory_tracker.go       # Memory allocation and spill detector
|   |   `-- query_log.go            # system.query_log correlation
|   `-- server/
|       |-- api.go                  # REST and WebSocket HTTP handler
|       |-- routes.go               # Router registration
|       `-- static.go               # Embedded React 19 single-binary server
|-- web/
|   |-- src/
|   |   |-- components/
|   |   |   |-- Header.tsx          # Connection switcher and status
|   |   |   |-- PartsHeatmap.tsx    # Partition and data parts visual grid
|   |   |   |-- MergesMonitor.tsx   # Real-time merge progress and throughput
|   |   |   |-- PipelineGraph.tsx   # Thread execution stream DAG viewer
|   |   |   |-- CodecOptimizer.tsx  # Column compression savings calculator
|   |   |   |-- MutationsViewer.tsx # ALTER TABLE progress monitor
|   |   |   |-- LocalWorkbench.tsx  # Serverless Parquet and CSV explorer
|   |   |   `-- QueryConsole.tsx    # Interactive SQL editor with EXPLAIN
|   |   |-- hooks/
|   |   |   `-- useClickHouseStream.ts
|   |   |-- types.ts
|   |   |-- App.tsx
|   |   `-- main.tsx
|   |-- index.html
|   |-- package.json
|   |-- tailwind.config.js
|   `-- vite.config.ts
|-- docs/
|   |-- images/
|   |   `-- clicklens_demo.gif
|   `-- ARCHITECTURE.md
|-- go.mod
|-- go.sum
|-- Makefile
|-- LICENSE
`-- README.md
```

---

## 3. Detailed Work Breakdown

### Task 1: Repository Foundation, Documentation and Go Module
- Initialize Go module `github.com/alexandrmotologa/clicklens` targeting Go 1.23.
- Write production README.md adhering to the humanizer discipline.
- Document architectural specifications in `docs/ARCHITECTURE.md`.
- Generate project logo SVG for README header.

### Task 2: ClickHouse Client and Connection Subsystem
- Implement `internal/client/native.go`:
  - Native TCP connection via `clickhouse-go/v2` (port 9000).
  - HTTP protocol support (port 8123).
  - Dynamic DSN parser with credentials and TLS options.
  - Connection ping and version discovery (`SELECT version()`).
- Implement `internal/client/simulator.go`:
  - Built-in realistic dataset generator representing multi-partition MergeTree tables, active merges, mutations, and query logs.
  - Activated via `--demo` CLI flag or when no ClickHouse server is reachable, enabling complete end-to-end functionality verification.
- Implement `internal/client/local_runner.go` and `file_inspector.go`:
  - Locate `clickhouse-local` in system PATH, WSL, or custom binaries.
  - Execute queries against local Parquet, CSV, and JSON files without a server.
  - Include native Go Parquet metadata parser for schema and row count extraction when clickhouse-local is not present.

### Task 3: MergeTree Storage and Background Operations Inspector
- Implement `internal/inspector/parts.go`:
  - Query `system.parts` for partition, name, active, bytes_on_disk, data_uncompressed_bytes, marks, rows, modification_time.
  - Calculate partition health scores: part counts, average part size, inactive parts backlog.
  - Generate partition optimization SQL statements (`OPTIMIZE TABLE ... FINAL`).
- Implement `internal/inspector/merges.go`:
  - Query `system.merges` for database, table, elapsed, progress, num_parts, source parts, result part name.
  - Compute merge read/write throughput in bytes and rows per second.
- Implement `internal/inspector/mutations.go`:
  - Query `system.mutations` for mutation_id, command, create_time, parts_to_do, is_done, latest_fail_reason.
- Implement `internal/inspector/processes.go`:
  - Query `system.processes` for active running queries, memory usage, elapsed time, and query killing.

### Task 4: Column Codec and Compression Auditor
- Implement `internal/inspector/columns.go`:
  - Read column statistics from `system.columns`: data_compressed_bytes, data_uncompressed_bytes, compression_codec, type.
  - Compute compression ratio and savings potential.
  - Rule-based codec recommendation:
    - Monotonic timestamps and IDs: `DoubleDelta, ZSTD(1)` or `Gorilla, ZSTD(1)`.
    - Floating point metrics: `Gorilla, ZSTD(1)`.
    - Low-cardinality strings: `LowCardinality(String)` with `ZSTD(3)`.
    - High-entropy payloads: `ZSTD(6)`.
  - Generate actionable `ALTER TABLE ... MODIFY COLUMN ... CODEC(...)` SQL commands.

### Task 5: Query Pipeline Profiler (`EXPLAIN PIPELINE` DAG)
- Implement `internal/profiler/explain_parser.go`:
  - Execute and parse output from `EXPLAIN PIPELINE` and `EXPLAIN PLAN`.
  - Convert text stream representation into a typed Directed Acyclic Graph (DAG) with nodes (steps) and edges (stream transitions).
  - Extract parallel concurrency levels (stream counts) per stage.
  - Detect pipeline bottlenecks: fan-in transitions where high stream counts bottleneck into a single thread.
- Implement `internal/profiler/memory_tracker.go`:
  - Correlate query execution with `system.query_log`.
  - Flag memory spills (`max_bytes_before_external_group_by`, `max_bytes_before_external_sort`).

### Task 6: REST and WebSocket API Server
- Implement `internal/server/api.go` and `routes.go`:
  - Endpoints:
    - `GET /api/health`: Server and ClickHouse connectivity status.
    - `POST /api/connect`: Change active ClickHouse DSN.
    - `GET /api/databases`: List accessible databases and tables.
    - `GET /api/parts`: Table parts list and partition health metrics.
    - `GET /api/merges`: Active background merges.
    - `GET /api/mutations`: Active table mutations.
    - `GET /api/columns`: Column compression audit and recommendations.
    - `POST /api/explain`: Run query and return parsed pipeline DAG.
    - `GET /api/processes`: Active running queries.
    - `POST /api/kill`: Terminate running query.
    - `POST /api/local/query`: Run query via clickhouse-local on local files.
    - `POST /api/local/inspect`: Inspect schema of local Parquet or CSV file.
    - `GET /api/ws`: WebSocket endpoint pushing real-time merges and telemetry updates.
  - Embed compiled React 19 frontend assets via `embed.FS` in `internal/server/static.go`.

### Task 7: Modern React 19 Web Studio
- Setup Vite + React 19 + Tailwind CSS + Lucide icons + Recharts in `web/`.
- Build application views:
  - Header with connection status, database and table selectors, and demo mode toggle.
  - Parts Heatmap: visual grid representing partitions, active parts, inactive parts, disk size, and health warnings.
  - Merges Monitor: live progress bars, elapsed time, and source/result part details.
  - Pipeline Graph: interactive visual DAG with stream counts, step types, and bottleneck badges.
  - Codec Optimizer: interactive column audit with compression ratios, projected savings, and copyable SQL commands.
  - Serverless Local Workbench: file picker, schema preview, and instant SQL query runner.
  - Query Console: SQL editor with EXPLAIN PIPELINE, EXPLAIN PLAN, and execution metrics.

### Task 8: CLI Commands and Tooling
- Implement CLI subcommands in `cmd/clicklens/main.go`:
  - `clicklens serve [--addr :8080] [--dsn ...]`
  - `clicklens parts <database>.<table> [--dsn ...]`
  - `clicklens merges [--dsn ...]`
  - `clicklens audit <database>.<table> [--dsn ...]`
  - `clicklens explain "<query>" [--dsn ...]`
  - `clicklens profile <filepath>`
- Provide single-binary packaging with embedded frontend.

### Task 9: Verification, Tests, Visual Assets and Portfolio Sync Phase 2
- Run Go unit tests for parser, codec calculator, and simulator.
- Run frontend build with zero TypeScript and bundle errors.
- Generate UI screenshots and animated demo GIF in `docs/images/`.
- Execute Phase 2 portfolio sync: copy assets to `mtlg-site` and `mtlglabs-space`, set status to LIVE, re-compile and push.
