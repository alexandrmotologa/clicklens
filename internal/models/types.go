package models

import "time"

// ConnectionConfig holds connection parameters.
type ConnectionConfig struct {
	DSN      string `json:"dsn"`
	Protocol string `json:"protocol"` // "native" (tcp:9000) or "http" (8123)
	IsDemo   bool   `json:"is_demo"`
}

// ServerStatus holds database connection health and version.
type ServerStatus struct {
	Connected      bool      `json:"connected"`
	Version        string    `json:"version"`
	CurrentDSN     string    `json:"current_dsn"`
	Protocol       string    `json:"protocol"`
	IsDemo         bool      `json:"is_demo"`
	UptimeSeconds  uint64    `json:"uptime_seconds"`
	ActiveQueries  int       `json:"active_queries"`
	ActiveMerges   int       `json:"active_merges"`
	TotalParts     int       `json:"total_parts"`
	TotalBytesDisk uint64    `json:"total_bytes_disk"`
	LastChecked    time.Time `json:"last_checked"`
}

// DatabaseInfo represents a ClickHouse database.
type DatabaseInfo struct {
	Name       string   `json:"name"`
	Engine     string   `json:"engine"`
	TableCount int      `json:"table_count"`
	Tables     []string `json:"tables,omitempty"`
}

// TableInfo represents a ClickHouse table.
type TableInfo struct {
	Database    string `json:"database"`
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	TotalRows   uint64 `json:"total_rows"`
	TotalBytes  uint64 `json:"total_bytes"`
	ActiveParts int    `json:"active_parts"`
}

// PartInfo represents a single MergeTree data part from system.parts.
type PartInfo struct {
	Partition              string    `json:"partition"`
	Name                   string    `json:"name"`
	Active                 bool      `json:"active"`
	BytesOnDisk            uint64    `json:"bytes_on_disk"`
	DataUncompressedBytes  uint64    `json:"data_uncompressed_bytes"`
	Marks                  uint64    `json:"marks"`
	Rows                   uint64    `json:"rows"`
	ModificationTime       time.Time `json:"modification_time"`
	MinBlockNumber         int64     `json:"min_block_number"`
	MaxBlockNumber         int64     `json:"max_block_number"`
	Level                  int       `json:"level"`
	CompressionRatio       float64   `json:"compression_ratio"`
}

// AgeBucket groups parts by how recently they were modified or ingested.
type AgeBucket struct {
	Label      string `json:"label"` // e.g. "< 5 min", "5-30 min", "30m-2h", "2h-24h", "> 24h"
	PartCount  int    `json:"part_count"`
	TotalRows  uint64 `json:"total_rows"`
	TotalBytes uint64 `json:"total_bytes"`
}

// PartitionSummary aggregates part metrics for a partition.
type PartitionSummary struct {
	Partition           string     `json:"partition"`
	ActiveParts         int        `json:"active_parts"`
	InactiveParts       int        `json:"inactive_parts"`
	TotalRows           uint64     `json:"total_rows"`
	TotalBytesOnDisk    uint64     `json:"total_bytes_on_disk"`
	TotalUncompressed   uint64     `json:"total_uncompressed"`
	CompressionRatio    float64    `json:"compression_ratio"`
	HealthStatus        string     `json:"health_status"` // "HEALTHY", "WARNING", "CRITICAL"
	FragmentationScore  float64    `json:"fragmentation_score"` // 0.0 - 100.0 (high means fragmented)
	OptimizationSQL     string     `json:"optimization_sql"`
	Parts               []PartInfo `json:"parts,omitempty"`
}

// TablePartsResponse aggregates table partition analysis.
type TablePartsResponse struct {
	Database               string             `json:"database"`
	Table                  string             `json:"table"`
	Engine                 string             `json:"engine"`
	TotalActiveParts       int                `json:"total_active_parts"`
	TotalInactiveParts     int                `json:"total_inactive_parts"`
	TotalBytesOnDisk       uint64             `json:"total_bytes_on_disk"`
	TotalUncompressed      uint64             `json:"total_uncompressed"`
	OverallRatio           float64            `json:"overall_ratio"`
	MaxPartsPartition      string             `json:"max_parts_partition"`
	MaxPartCount           int                `json:"max_part_count"`
	HealthRating           string             `json:"health_rating"` // "HEALTHY", "DEGRADED", "CRITICAL"
	PartAgeBuckets         []AgeBucket        `json:"part_age_buckets"`
	MicroBatchingWarning   bool               `json:"micro_batching_warning"`
	MicroBatchingDetails   string             `json:"micro_batching_details"`
	Partitions             []PartitionSummary `json:"partitions"`
}

// MergeInfo represents active background merges from system.merges.
type MergeInfo struct {
	Database        string   `json:"database"`
	Table           string   `json:"table"`
	ElapsedSeconds  float64  `json:"elapsed_seconds"`
	Progress        float64  `json:"progress"` // 0.0 - 100.0
	NumParts        uint64   `json:"num_parts"`
	ResultPartName  string   `json:"result_part_name"`
	SourceParts     []string `json:"source_parts"`
	BytesRead       uint64   `json:"bytes_read"`
	BytesWritten    uint64   `json:"bytes_written"`
	RowsRead        uint64   `json:"rows_read"`
	RowsWritten     uint64   `json:"rows_written"`
	MemoryUsage     uint64   `json:"memory_usage"`
	ReadThroughput  float64  `json:"read_throughput_bytes_sec"`
	WriteThroughput float64  `json:"write_throughput_bytes_sec"`
}

// MutationInfo represents asynchronous table mutations from system.mutations.
type MutationInfo struct {
	MutationID       string    `json:"mutation_id"`
	Database         string    `json:"database"`
	Table            string    `json:"table"`
	Command          string    `json:"command"`
	CreateTime       time.Time `json:"create_time"`
	PartsToDo        int64     `json:"parts_to_do"`
	IsDone           bool      `json:"is_done"`
	LatestFailedPart string    `json:"latest_failed_part"`
	LatestFailReason string    `json:"latest_fail_reason"`
}

// ColumnAudit audits a column's compression efficiency and suggests optimal codecs.
type ColumnAudit struct {
	Database               string  `json:"database"`
	Table                  string  `json:"table"`
	Column                 string  `json:"column"`
	Type                   string  `json:"type"`
	DefaultExpression      string  `json:"default_expression"`
	CompressionCodec       string  `json:"compression_codec"`
	DataCompressedBytes    uint64  `json:"data_compressed_bytes"`
	DataUncompressedBytes  uint64  `json:"data_uncompressed_bytes"`
	CompressionRatio       float64 `json:"compression_ratio"`
	RecommendedCodec       string  `json:"recommended_codec"`
	SavingsEstimatePercent float64 `json:"savings_estimate_percent"`
	ProjectedSavingsBytes  uint64  `json:"projected_savings_bytes"`
	AlterSQL               string  `json:"alter_sql"`
	Rationale              string  `json:"rationale"`
}

// TableColumnsResponse holds column compression audit results.
type TableColumnsResponse struct {
	Database                      string        `json:"database"`
	Table                         string        `json:"table"`
	TotalColumns                  int           `json:"total_columns"`
	TotalCompressed               uint64        `json:"total_compressed_bytes"`
	TotalUncompressed             uint64        `json:"total_uncompressed_bytes"`
	PotentialSavings              uint64        `json:"potential_savings_bytes"`
	ZeroDowntimeMigrationScript   string        `json:"zero_downtime_migration_script"`
	Columns                       []ColumnAudit `json:"columns"`
}

// CodecBenchmarkItem holds the live benchmark metrics of a specific codec.
type CodecBenchmarkItem struct {
	Codec             string  `json:"codec"`
	CompressedBytes   uint64  `json:"compressed_bytes"`
	Ratio             float64 `json:"ratio"`
	CompressTimeMs    float64 `json:"compress_time_ms"`
	DecompressTimeMs  float64 `json:"decompress_time_ms"`
	ThroughputMBs     float64 `json:"throughput_mb_s"`
}

// CodecBenchmarkResponse holds the results of benchmarking multiple codecs on sample column data.
type CodecBenchmarkResponse struct {
	Database       string               `json:"database"`
	Table          string               `json:"table"`
	Column         string               `json:"column"`
	OriginalBytes  uint64               `json:"original_bytes"`
	SampleRows     int                  `json:"sample_rows"`
	Results        []CodecBenchmarkItem `json:"results"`
}

// ProcessInfo represents an active query from system.processes.
type ProcessInfo struct {
	QueryID        string  `json:"query_id"`
	User           string  `json:"user"`
	Address        string  `json:"address"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
	RowsRead       uint64  `json:"rows_read"`
	BytesRead      uint64  `json:"bytes_read"`
	MemoryUsage    uint64  `json:"memory_usage"`
	PeakMemory     uint64  `json:"peak_memory_usage"`
	Query          string  `json:"query"`
}

// QueryLogEntry represents historical query telemetry from system.query_log.
type QueryLogEntry struct {
	QueryID        string    `json:"query_id"`
	EventTime      time.Time `json:"event_time"`
	DurationMs     float64   `json:"duration_ms"`
	MemoryUsage    uint64    `json:"memory_usage"`
	ReadRows       uint64    `json:"read_rows"`
	ReadBytes      uint64    `json:"read_bytes"`
	ResultRows     uint64    `json:"result_rows"`
	ResultBytes    uint64    `json:"result_bytes"`
	QueryText      string    `json:"query_text"`
	ExceptionCode  int       `json:"exception_code"`
	Exception      string    `json:"exception,omitempty"`
	HasSpill       bool      `json:"has_spill"`
	Type           string    `json:"type"` // "QueryFinish", "ExceptionWhileProcessing"
}

// QueryLogResponse returns filtered query history.
type QueryLogResponse struct {
	Filter  string          `json:"filter"`
	Total   int             `json:"total"`
	Entries []QueryLogEntry `json:"entries"`
}

// DAGNode represents a step in EXPLAIN PIPELINE or EXPLAIN PLAN.
type DAGNode struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	StepType    string   `json:"step_type"`
	StreamCount int      `json:"stream_count"`
	Details     string   `json:"details,omitempty"`
	Inputs      []string `json:"inputs,omitempty"`
	Outputs     []string `json:"outputs,omitempty"`
	IsBottleneck bool    `json:"is_bottleneck"`
}

// DAGEdge represents data flow between execution steps.
type DAGEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	StreamCount int    `json:"stream_count"`
	Label       string `json:"label,omitempty"`
}

// PipelineBottleneck identifies a concurrency choke point.
type PipelineBottleneck struct {
	NodeID         string `json:"node_id"`
	StepType       string `json:"step_type"`
	PreviousStreams int    `json:"previous_streams"`
	CurrentStreams  int    `json:"current_streams"`
	Severity       string `json:"severity"` // "WARNING", "CRITICAL"
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// SpillAlert represents an external memory spill event.
type SpillAlert struct {
	QueryID           string `json:"query_id,omitempty"`
	SpillType         string `json:"spill_type"` // "EXTERNAL_GROUP_BY" or "EXTERNAL_SORT"
	TriggerMemoryBytes uint64 `json:"trigger_memory_bytes"`
	Advice            string `json:"advice"`
}

// PipelineAnalysis is the parsed EXPLAIN result.
type PipelineAnalysis struct {
	Query       string               `json:"query"`
	Mode        string               `json:"mode"` // "PIPELINE" or "PLAN"
	RawText     string               `json:"raw_text"`
	TotalStreams int                 `json:"total_streams"`
	MaxConcurrency int               `json:"max_concurrency"`
	Nodes       []DAGNode            `json:"nodes"`
	Edges       []DAGEdge            `json:"edges"`
	Bottlenecks []PipelineBottleneck `json:"bottlenecks"`
	SpillAlerts []SpillAlert         `json:"spill_alerts"`
}

// LocalColumnMetadata defines schema column info for local files.
type LocalColumnMetadata struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// LocalFileProfile holds inspection results for a local file.
type LocalFileProfile struct {
	FilePath      string                `json:"file_path"`
	FileName      string                `json:"file_name"`
	Format        string                `json:"format"` // Parquet, CSV, TSV, JSONEachRow
	FileSizeBytes int64                 `json:"file_size_bytes"`
	RowCount      int64                 `json:"row_count"`
	Columns       []LocalColumnMetadata `json:"columns"`
	SampleRows    []map[string]any      `json:"sample_rows"`
	RunnerUsed    string                `json:"runner_used"` // "clickhouse-local" or "pure-go-engine"
}

// QueryResult represents the execution output of a query.
type QueryResult struct {
	Query           string           `json:"query"`
	Columns         []string         `json:"columns"`
	Rows            [][]any          `json:"rows"`
	RowsCount       int              `json:"rows_count"`
	ExecutionTimeMs float64          `json:"execution_time_ms"`
	BytesRead       uint64           `json:"bytes_read"`
	RowsRead        uint64           `json:"rows_read"`
	Error           string           `json:"error,omitempty"`
}

// FileConvertRequest defines arguments for file format conversion.
type FileConvertRequest struct {
	SourcePath   string `json:"source_path"`
	TargetPath   string `json:"target_path"`
	TargetFormat string `json:"target_format"` // "Parquet", "CSV", "TSV"
	Compression  string `json:"compression"`   // "ZSTD", "SNAPPY", "NONE"
}

// FileConvertResponse holds the outcome of a file conversion.
type FileConvertResponse struct {
	Success          bool    `json:"success"`
	SourceSize       int64   `json:"source_size"`
	TargetSize       int64   `json:"target_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	TargetPath       string  `json:"target_path"`
	Message          string  `json:"message"`
}

// HealthAuditReport contains a diagnostic audit ready for export.
type HealthAuditReport struct {
	GeneratedAt        time.Time `json:"generated_at"`
	Database           string    `json:"database"`
	Table              string    `json:"table"`
	ServerVersion      string    `json:"server_version"`
	OverallHealth      string    `json:"overall_health"`
	ActiveParts        int       `json:"active_parts"`
	PartitionCount     int       `json:"partition_count"`
	MaxPartsPartition  string    `json:"max_parts_partition"`
	MaxPartCount       int       `json:"max_part_count"`
	DiskBytes          uint64    `json:"disk_bytes"`
	UncompressedBytes  uint64    `json:"uncompressed_bytes"`
	OverallRatio       float64   `json:"overall_ratio"`
	PotentialSavings   uint64    `json:"potential_savings_bytes"`
	MicroBatchingRisk  bool      `json:"micro_batching_risk"`
	Bottlenecks        []string  `json:"bottlenecks"`
	TopRecommendations []string  `json:"top_recommendations"`
	MarkdownContent    string    `json:"markdown_content"`
}

// TelemetryWebSocketMessage defines the live stream payload over WebSocket.
type TelemetryWebSocketMessage struct {
	Type      string      `json:"type"` // "merges_update", "status_update", "processes_update"
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}
