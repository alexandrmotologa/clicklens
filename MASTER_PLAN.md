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
|   |   |-- converter.go            # CSV/TSV to compressed Parquet converter
|   |   `-- simulator.go            # Realistic in-memory telemetry provider
|   |-- inspector/
|   |   |-- parts.go                # system.parts scanner, age buckets & micro-batching
|   |   |-- merges.go               # system.merges real-time tracker
|   |   |-- mutations.go            # system.mutations observer
|   |   |-- columns.go              # system.columns compression auditor & playbook
|   |   |-- benchmark.go            # Live codec benchmark simulator
|   |   |-- query_log.go            # system.query_log multi-filter inspector
|   |   |-- report.go               # Cluster health diagnostic report generator
|   |   `-- processes.go            # system.processes active queries
|   |-- profiler/
|   |   |-- explain_parser.go       # EXPLAIN PIPELINE and PLAN parser
|   |   `-- memory_tracker.go       # Memory allocation and spill detector
|   `-- server/
|       |-- api.go                  # REST and WebSocket HTTP handler
|       |-- routes.go               # Router registration
|       |-- static.go               # Embedded React 19 single-binary server
|       `-- web_dist/               # Embedded production React frontend
|-- web/
|   |-- src/
|   |   |-- components/
|   |   |   |-- Header.tsx          # Connection switcher, status & diagnostic report
|   |   |   |-- PartsHeatmap.tsx    # Partition grid, age histogram & anomaly banner
|   |   |   |-- MergesMonitor.tsx   # Real-time merge progress and throughput
|   |   |   |-- PipelineGraph.tsx   # Thread execution stream DAG viewer
|   |   |   |-- CodecOptimizer.tsx  # Codecs, benchmark simulator & migration playbook
|   |   |   |-- QueryLogViewer.tsx  # Historical query log analytics
|   |   |   |-- LocalWorkbench.tsx  # Serverless file inspector & converter
|   |   |   |-- CommandPalette.tsx  # Ctrl+K modal quick jump
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
|   |   |-- clicklens_demo.gif
|   |   `-- clicklens_*.png
|   `-- ARCHITECTURE.md
|-- go.mod
|-- go.sum
|-- Makefile
|-- LICENSE
`-- README.md
```

---

## 3. Detailed Work Breakdown & Status

- [x] **Task 1: Repository Foundation, Documentation and Go Module**
  - Go module `github.com/alexandrmotologa/clicklens` initialized.
  - Public GitHub repository created with 13 domain topics.
  - Factual README.md adhering to humanizer guidelines.
  - Complete architecture specification in `docs/ARCHITECTURE.md`.
  - Scalable vector mascot logo in `assets/logo.svg`.

- [x] **Task 2: ClickHouse Client and Connection Subsystem**
  - Native TCP connection via `clickhouse-go/v2`.
  - Realistic in-memory simulator (`internal/client/simulator.go`) for zero-dependency operation.
  - `clickhouse-local` process runner and pure Go `parquet-go` fallback.

- [x] **Task 3: MergeTree Storage and Background Operations Inspector**
  - Part and partition metrics computed with fragmentation scores.
  - Real-time WebSocket streaming of `system.merges` every 500ms.
  - Asynchronous table mutations monitoring (`system.mutations`).

- [x] **Task 4: Column Codec and Compression Auditor**
  - Ratio calculator comparing raw and compressed bytes.
  - Type-aware heuristic recommendation engine.

- [x] **Task 5: Query Pipeline Profiler (`EXPLAIN PIPELINE` DAG)**
  - Regex-free parser converting raw pipeline output into typed execution DAG.
  - Detection of thread fan-in bottlenecks and disk spill risks.

- [x] **Task 6: REST and WebSocket API Server**
  - Full suite of REST endpoints registered in `internal/server/routes.go`.
  - WebSocket broadcaster for live telemetry.
  - Embedded React 19 assets via Go `embed.FS`.

- [x] **Task 7: Modern React 19 Web Studio**
  - Dark-themed studio built with React 19, Tailwind CSS, Lucide, and Recharts.
  - 100% functional components with zero mockups.

- [x] **Task 8: Production CLI Commands**
  - Commands `serve`, `parts`, `merges`, `audit`, `explain`, `profile` implemented in `cmd/clicklens/main.go`.

- [x] **Task 9: Advanced Architectural Enhancements (100% Completed)**
  - 1. Historical Query Log Analytics (`system.query_log`) with slow, memory, spill, and error filters.
  - 2. Part Age Distribution Histogram (5 age windows) and Ingestion Micro-Batching Anomaly Detector.
  - 3. Zero-Downtime Codec Migration Playbook Generator (`table_new`, `INSERT INTO ... SELECT`, `EXCHANGE TABLES`).
  - 4. Live Codec Compression Benchmark Simulator with throughput and ratio metrics.
  - 5. Serverless Local File Converter (CSV/TSV to Parquet with SNAPPY/ZSTD).
  - 6. Cluster Health Diagnostic Report Export in Markdown and JSON.
  - 7. Global Command Palette (`Ctrl+K` / `Cmd+K`) with keyboard navigation.

- [x] **Task 10: Verification, Assets and Dual-Portfolio Synchronization**
  - All unit and integration tests passing (`go test ./... -v`).
  - Frontend production build verified (`npm run build`).
  - Headless Chrome and FFmpeg animated GIF and screenshot suite generated in `docs/images/`.
  - Dual-portfolio sync executed across `mtlg-site` and `mtlglabs-space`.
