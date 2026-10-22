<div align="center">
  <img src="assets/logo.svg" width="96" height="96" alt="ClickLens Logo" />
  <h1>ClickLens</h1>
  <p>Local Observability Studio, MergeTree Parts Visualizer and Query Pipeline Profiler for ClickHouse</p>

  <p>
    <a href="#quick-start">Quick Start</a> &bull;
    <a href="#core-capabilities">Capabilities</a> &bull;
    <a href="#cli-usage">CLI Usage</a> &bull;
    <a href="#screenshots">Screenshots</a> &bull;
    <a href="#architecture">Architecture</a> &bull;
    <a href="#license">License</a>
  </p>
</div>

---

ClickLens provides low-level operational visibility into ClickHouse storage mechanics and query execution. It monitors MergeTree parts lifecycles, tracks background merges and mutations in real time, parses `EXPLAIN PIPELINE` output into execution stream DAGs with memory spill warnings, calculates column-level compression ratios with codec recommendations, and runs ad-hoc queries against local Parquet and CSV files through `clickhouse-local`.

![ClickLens Studio Overview](docs/images/clicklens_demo.gif)

## Core Capabilities

### 1. MergeTree Parts and Ingestion Anomaly Detection
- Visualizes active and inactive data parts across partitions with a color-coded fragmentation meter.
- Detects ingestion micro-batching anomalies when small parts accumulate faster than ClickHouse can merge them.
- Provides a Part Age Distribution Histogram grouping parts into 5 time windows (< 5 min, 5-30 min, 30m-2h, 2h-24h, > 24h).
- Generates targeted `OPTIMIZE TABLE ... PARTITION ... FINAL` queries to consolidate fragmented partitions.

### 2. Live Background Merge and Mutation Monitor
- Streams `system.merges` telemetry over WebSocket connections at 500ms intervals.
- Tracks source parts, target part name, elapsed time, progress percentage, memory usage, and throughput in bytes/rows per second.
- Monitors `system.mutations` for running `ALTER TABLE ... UPDATE/DELETE` statements with failure diagnostics.

### 3. Query Pipeline Profiler and Memory Spill Detection
- Parses ClickHouse `EXPLAIN PIPELINE` output into an interactive Directed Acyclic Graph (DAG).
- Inspects thread concurrency and stream transformations across `FilterTransform`, `AggregatingTransform`, and `SortingTransform`.
- Detects thread fan-in choke points where parallel streams collapse into single-threaded bottlenecks.
- Flags external memory spill risks when queries exceed `max_bytes_before_external_group_by` or `max_bytes_before_external_sort`.

### 4. Historical Query Log Analytics
- Surfaces telemetry from `system.query_log` across four operational dimensions:
  - Slow Queries: Queries with excessive duration.
  - Memory Spenders: High RAM allocation (> 1 GB) and peak memory tracking.
  - Disk Spills: Queries that spilled intermediate aggregate or sort buffers to disk.
  - Errors and OOMs: ClickHouse exceptions including `MEMORY_LIMIT_EXCEEDED` and syntax errors.
- One-click action to transfer any historical query directly into the Query Pipeline Profiler.

### 5. Column Codec Optimizer and Migration Playbook
- Audits column storage consumption using `system.columns` and compares raw versus compressed disk bytes.
- Recommends specialized codecs based on data types:
  - Monotonic timestamps and numeric IDs: `DoubleDelta, ZSTD(1)` or `T64, ZSTD(1)`.
  - Floating point metrics: `Gorilla, ZSTD(1)`.
  - Repetitive text: `LowCardinality(String)` with `ZSTD(3)`.
  - Unstructured text and JSON: `ZSTD(6)`.
- Live Codec Compression Benchmark Simulator: Measures compression ratio and MB/s throughput across candidate codecs.
- Zero-Downtime Migration Playbook Generator: Generates an atomic 5-step SQL playbook using shadow tables (`table_new`, `INSERT INTO ... SELECT`, `EXCHANGE TABLES`) so data is converted with zero downtime.

### 6. Serverless Local File Workbench and Converter
- Inspects and queries local analytical files without a running ClickHouse server.
- Supports Parquet, CSV, TSV, and JSON formats.
- In-memory tabular file converter: Converts CSV and TSV files into compressed Parquet (`SNAPPY` or `ZSTD`) using native ClickHouse-local execution with a pure Go fallback engine.

### 7. Health Diagnostic Report and Command Palette
- Compiles an end-to-end ClickHouse cluster health audit into downloadable Markdown (`.md`) and JSON (`.json`) reports.
- Includes a Command Palette (`Ctrl+K` or `Cmd+K`) for rapid navigation across tables, views, and studio diagnostics.

---

## Screenshots

<div align="center">
  <p><strong>Historical Query Log Analytics</strong></p>
  <img src="docs/images/clicklens_query_log.png" width="900" alt="Query Log Analytics" />
  <br/><br/>
  <p><strong>Live Codec Benchmark Simulator</strong></p>
  <img src="docs/images/clicklens_codec_benchmark.png" width="900" alt="Codec Benchmark Simulator" />
  <br/><br/>
  <p><strong>Cluster Health Diagnostic Report Export</strong></p>
  <img src="docs/images/clicklens_diagnostic_report.png" width="900" alt="Diagnostic Audit Report" />
  <br/><br/>
  <p><strong>Command Palette (Ctrl+K)</strong></p>
  <img src="docs/images/clicklens_command_palette.png" width="900" alt="Command Palette" />
</div>

---

## Quick Start

### Installation

Clone the repository and compile the single binary:

```bash
git clone https://github.com/alexandrmotologa/clicklens.git
cd clicklens

# Build the web studio bundle
cd web
npm install
npm run build
cd ..

# Build the standalone Go binary (with embedded frontend)
go build -o bin/clicklens.exe ./cmd/clicklens
```

### Launch the Observability Studio

Run the studio connected to a live ClickHouse server:

```bash
./bin/clicklens.exe serve --addr :8080 --dsn "clickhouse://default:password@localhost:9000/default"
```

If you do not have a live ClickHouse instance available, launch with `--demo` to explore the complete studio with simulated production telemetry:

```bash
./bin/clicklens.exe serve --addr :8080 --demo
```

Open `http://localhost:8080` in your web browser.

---

## CLI Usage

ClickLens operates as a web studio or as a standalone terminal inspection utility.

### Inspect Table Parts

```bash
clicklens parts system.query_log --dsn "clickhouse://localhost:9000"
```

Output:
```
Partition: 202610 | Parts: 14 active (2 inactive) | Rows: 1,420,180 | Disk: 142.6 MB (Ratio: 4.82x)
Partition: 202609 | Parts: 8 active (0 inactive)  | Rows: 890,200   | Disk: 94.1 MB (Ratio: 5.10x)
```

### Monitor Background Merges

```bash
clicklens merges --dsn "clickhouse://localhost:9000"
```

### Audit Column Compression

```bash
clicklens audit analytics.events --dsn "clickhouse://localhost:9000"
```

Output:
```
Column: event_time | Type: DateTime | Codec: Default(LZ4) | Ratio: 1.4x
  -> Recommended: CODEC(DoubleDelta, ZSTD(1)) | Projected Savings: 62%
  -> SQL: ALTER TABLE analytics.events MODIFY COLUMN event_time CODEC(DoubleDelta, ZSTD(1));
```

### Profile Query Execution Pipeline

```bash
clicklens explain "SELECT event_type, count() FROM analytics.events GROUP BY event_type"
```

### Profile Local Parquet or CSV File

```bash
clicklens profile ./testdata/telemetry.parquet
```

---

## Architecture

ClickLens consists of three integrated layers:
1. Native Protocol Ingestion: Uses `clickhouse-go/v2` to query system tables (`system.parts`, `system.merges`, `system.columns`, `system.processes`, `system.query_log`) over native TCP (port 9000) or HTTP (port 8123).
2. Diagnostics Engine: Evaluates partition fragmentation limits, parses pipeline syntax into DAG structures, evaluates compression models, and wraps `clickhouse-local` for serverless inspection.
3. React 19 Studio: A dark-themed cockpit built with Tailwind CSS and Recharts, served directly from the single Go binary via `embed.FS`.

For full design specifications, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## License

MIT License. See [LICENSE](LICENSE) for details.
