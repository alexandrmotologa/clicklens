<div align="center">
  <img src="assets/logo.svg" width="96" height="96" alt="ClickLens Logo" />
  <h1>ClickLens</h1>
  <p>Local Observability Studio, MergeTree Parts Visualizer and Query Pipeline Profiler for ClickHouse</p>

  <p>
    <a href="#quick-start">Quick Start</a> &bull;
    <a href="#core-capabilities">Capabilities</a> &bull;
    <a href="#cli-usage">CLI Usage</a> &bull;
    <a href="#architecture">Architecture</a> &bull;
    <a href="#license">License</a>
  </p>
</div>

---

ClickLens provides low-level operational visibility into ClickHouse storage mechanics and query execution. It monitors MergeTree parts lifecycles, tracks background merges and mutations in real time, parses `EXPLAIN PIPELINE` output into execution stream DAGs with memory spill warnings, calculates column-level compression ratios with codec recommendations, and runs ad-hoc queries against local Parquet and CSV files through `clickhouse-local`.

![ClickLens Studio Overview](docs/images/clicklens_demo.gif)

## Core Capabilities

### 1. MergeTree Parts and Partition Diagnostics
- Maps active and inactive data parts across table partitions.
- Computes partition fragmentation scores, identifying partitions at risk of write stalls due to part count limits.
- Displays marks count, row counts, disk bytes, uncompressed bytes, and modification timestamps per part.
- Generates targeted `OPTIMIZE TABLE ... PARTITION ... FINAL` queries to consolidate fragmented partitions.

### 2. Real-Time Background Merge and Mutation Monitor
- Polls `system.merges` over WebSocket connections every 500ms to monitor background merges.
- Tracks source parts, target part name, elapsed time, progress percentage, and memory usage.
- Measures merge read and write throughput in bytes and rows per second.
- Monitors `system.mutations` for long-running `ALTER TABLE ... UPDATE/DELETE` statements with failure diagnostics.

### 3. Query Pipeline Profiler and Memory Spill Detection
- Parses `EXPLAIN PIPELINE` output into an interactive Directed Acyclic Graph (DAG).
- Inspects thread concurrency and stream splits/merges across pipeline operators (`FilterTransform`, `AggregatingTransform`, `SortingTransform`).
- Flags thread fan-in bottlenecks where parallel streams collapse into single-threaded execution.
- Detects external disk spill events when queries exceed `max_bytes_before_external_group_by` or `max_bytes_before_external_sort`.

### 4. Column Codec and Compression Auditor
- Inspects storage consumption across table columns using `system.columns`.
- Calculates exact compression ratios comparing uncompressed bytes to compressed bytes on disk.
- Generates algorithmic recommendations:
  - Monotonic timestamps and numeric IDs: `DoubleDelta, ZSTD(1)` or `Gorilla, ZSTD(1)`.
  - Floating point metrics: `Gorilla, ZSTD(1)`.
  - Low-cardinality string columns: `LowCardinality(String)` with `ZSTD(3)`.
  - Text payloads: `ZSTD(6)`.
- Calculates projected disk savings and generates copyable `ALTER TABLE MODIFY COLUMN` statements.

### 5. Serverless clickhouse-local Workbench
- Queries local files directly using `clickhouse-local` without a running ClickHouse daemon.
- Supports Parquet, CSV, TSV, JSONEachRow, and Arrow formats.
- Extracts schema, column datatypes, and row counts instantly.
- Includes pure Go Parquet metadata fallback inspection when `clickhouse-local` is not installed on the host.

---

## Quick Start

### Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/alexandrmotologa/clicklens.git
cd clicklens

# Build the web studio bundle
cd web
npm install
npm run build
cd ..

# Build the single Go binary
go build -o bin/clicklens.exe ./cmd/clicklens
```

### Launch the Observability Studio

Run the studio connected to a local ClickHouse server:

```bash
./bin/clicklens.exe serve --addr :8080 --dsn "clickhouse://default:password@localhost:9000/default"
```

If you do not have a live ClickHouse server running, launch with `--demo` to explore the full studio with realistic telemetry:

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
