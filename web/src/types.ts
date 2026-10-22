export interface ServerStatus {
  connected: boolean;
  version: string;
  current_dsn: string;
  protocol: string;
  is_demo: boolean;
  uptime_seconds: number;
  active_queries: number;
  active_merges: number;
  total_parts: number;
  total_bytes_disk: number;
  last_checked: string;
}

export interface DatabaseInfo {
  name: string;
  engine: string;
  table_count: number;
  tables?: string[];
}

export interface PartInfo {
  partition: string;
  name: string;
  active: boolean;
  bytes_on_disk: number;
  data_uncompressed_bytes: number;
  marks: number;
  rows: number;
  modification_time: string;
  min_block_number: number;
  max_block_number: number;
  level: number;
  compression_ratio: number;
}

export interface AgeBucket {
  label: string;
  part_count: number;
  total_rows: number;
  total_bytes: number;
}

export interface PartitionSummary {
  partition: string;
  active_parts: number;
  inactive_parts: number;
  total_rows: number;
  total_bytes_on_disk: number;
  total_uncompressed: number;
  compression_ratio: number;
  health_status: 'HEALTHY' | 'WARNING' | 'CRITICAL';
  fragmentation_score: number;
  optimization_sql: string;
  parts?: PartInfo[];
}

export interface TablePartsResponse {
  database: string;
  table: string;
  engine: string;
  total_active_parts: number;
  total_inactive_parts: number;
  total_bytes_on_disk: number;
  total_uncompressed: number;
  overall_ratio: number;
  max_parts_partition: string;
  max_part_count: number;
  health_rating: 'HEALTHY' | 'DEGRADED' | 'CRITICAL';
  part_age_buckets?: AgeBucket[];
  micro_batching_warning?: boolean;
  micro_batching_details?: string;
  partitions: PartitionSummary[];
}

export interface MergeInfo {
  database: string;
  table: string;
  elapsed_seconds: number;
  progress: number;
  num_parts: number;
  result_part_name: string;
  source_parts: string[];
  bytes_read: number;
  bytes_written: number;
  rows_read: number;
  rows_written: number;
  memory_usage: number;
  read_throughput_bytes_sec: number;
  write_throughput_bytes_sec: number;
}

export interface MutationInfo {
  mutation_id: string;
  database: string;
  table: string;
  command: string;
  create_time: string;
  parts_to_do: number;
  is_done: boolean;
  latest_failed_part: string;
  latest_fail_reason: string;
}

export interface ColumnAudit {
  database: string;
  table: string;
  column: string;
  type: string;
  default_expression: string;
  compression_codec: string;
  data_compressed_bytes: number;
  data_uncompressed_bytes: number;
  compression_ratio: number;
  recommended_codec: string;
  savings_estimate_percent: number;
  projected_savings_bytes: number;
  alter_sql: string;
  rationale: string;
}

export interface TableColumnsResponse {
  database: string;
  table: string;
  total_columns: number;
  total_compressed_bytes: number;
  total_uncompressed_bytes: number;
  potential_savings_bytes: number;
  zero_downtime_migration_script?: string;
  columns: ColumnAudit[];
}

export interface CodecBenchmarkItem {
  codec: string;
  compressed_bytes: number;
  ratio: number;
  compress_time_ms: number;
  decompress_time_ms: number;
  throughput_mb_s: number;
}

export interface CodecBenchmarkResponse {
  database: string;
  table: string;
  column: string;
  original_bytes: number;
  sample_rows: number;
  results: CodecBenchmarkItem[];
}

export interface QueryLogEntry {
  query_id: string;
  event_time: string;
  duration_ms: number;
  memory_usage: number;
  read_rows: number;
  read_bytes: number;
  result_rows: number;
  result_bytes: number;
  query_text: string;
  exception_code: number;
  exception?: string;
  has_spill: boolean;
  type: string;
}

export interface QueryLogResponse {
  filter: string;
  total: number;
  entries: QueryLogEntry[];
}

export interface FileConvertRequest {
  source_path: string;
  target_path?: string;
  target_format?: string;
  compression?: string;
}

export interface FileConvertResponse {
  success: boolean;
  source_size: number;
  target_size: number;
  compression_ratio: number;
  target_path: string;
  message: string;
}

export interface HealthAuditReport {
  generated_at: string;
  database: string;
  table: string;
  server_version: string;
  overall_health: string;
  active_parts: number;
  partition_count: number;
  max_parts_partition: string;
  max_part_count: number;
  disk_bytes: number;
  uncompressed_bytes: number;
  overall_ratio: number;
  potential_savings_bytes: number;
  micro_batching_risk: boolean;
  bottlenecks: string[];
  top_recommendations: string[];
  markdown_content: string;
}

export interface DAGNode {
  id: string;
  label: string;
  step_type: string;
  stream_count: number;
  details?: string;
  inputs?: string[];
  outputs?: string[];
  is_bottleneck: boolean;
}

export interface DAGEdge {
  from: string;
  to: string;
  stream_count: number;
  label?: string;
}

export interface PipelineBottleneck {
  node_id: string;
  step_type: string;
  previous_streams: number;
  current_streams: number;
  severity: 'WARNING' | 'CRITICAL';
  description: string;
  recommendation: string;
}

export interface SpillAlert {
  query_id?: string;
  spill_type: string;
  trigger_memory_bytes: number;
  advice: string;
}

export interface PipelineAnalysis {
  query: string;
  mode: 'PIPELINE' | 'PLAN';
  raw_text: string;
  total_streams: number;
  max_concurrency: number;
  nodes: DAGNode[];
  edges: DAGEdge[];
  bottlenecks: PipelineBottleneck[];
  spill_alerts: SpillAlert[];
}

export interface ProcessInfo {
  query_id: string;
  user: string;
  address: string;
  elapsed_seconds: number;
  rows_read: number;
  bytes_read: number;
  memory_usage: number;
  peak_memory_usage: number;
  query: string;
}

export interface LocalColumnMetadata {
  name: string;
  type: string;
  nullable: boolean;
}

export interface LocalFileProfile {
  file_path: string;
  file_name: string;
  format: string;
  file_size_bytes: number;
  row_count: number;
  columns: LocalColumnMetadata[];
  sample_rows: Record<string, any>[];
  runner_used: string;
}

export interface QueryResult {
  query: string;
  columns: string[];
  rows: any[][];
  rows_count: number;
  execution_time_ms: number;
  bytes_read?: number;
  rows_read?: number;
  error?: string;
}
