package client

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/models"
)

// Simulator generates dynamic telemetry matching ClickHouse system tables.
type Simulator struct {
	mu           sync.Mutex
	startTime    time.Time
	merges       []models.MergeInfo
	mutations    []models.MutationInfo
	processes    []models.ProcessInfo
	lastTickTime time.Time
}

// NewSimulator creates an initialized simulator with realistic state.
func NewSimulator() *Simulator {
	now := time.Now()
	s := &Simulator{
		startTime:    now.Add(-48 * time.Hour),
		lastTickTime: now,
	}
	s.initSampleMerges()
	s.initSampleMutations()
	s.initSampleProcesses()
	return s
}

func (s *Simulator) initSampleMerges() {
	s.merges = []models.MergeInfo{
		{
			Database:        "telemetry",
			Table:           "events",
			ElapsedSeconds:  4.2,
			Progress:        64.5,
			NumParts:        7,
			ResultPartName:  "202610_104_110_1",
			SourceParts:     []string{"202610_104_104_0", "202610_105_105_0", "202610_106_106_0", "202610_107_107_0", "202610_108_108_0", "202610_109_109_0", "202610_110_110_0"},
			BytesRead:       148201900,
			BytesWritten:    96140800,
			RowsRead:        1820400,
			RowsWritten:     1210900,
			MemoryUsage:     67108864, // 64 MB
			ReadThroughput:  35286166,
			WriteThroughput: 22890666,
		},
		{
			Database:        "analytics",
			Table:           "orders",
			ElapsedSeconds:  12.8,
			Progress:        89.2,
			NumParts:        4,
			ResultPartName:  "202610_88_91_2",
			SourceParts:     []string{"202610_88_88_0", "202610_89_89_0", "202610_90_90_1", "202610_91_91_0"},
			BytesRead:       481902000,
			BytesWritten:    398102000,
			RowsRead:        4910200,
			RowsWritten:     4380100,
			MemoryUsage:     134217728, // 128 MB
			ReadThroughput:  37648593,
			WriteThroughput: 31101718,
		},
	}
}

func (s *Simulator) initSampleMutations() {
	now := time.Now()
	s.mutations = []models.MutationInfo{
		{
			MutationID:       "mutation_1042.txt",
			Database:         "telemetry",
			Table:            "events",
			Command:          "ALTER TABLE telemetry.events DELETE WHERE timestamp < now() - INTERVAL 90 DAY",
			CreateTime:       now.Add(-45 * time.Minute),
			PartsToDo:        3,
			IsDone:           false,
			LatestFailedPart: "",
			LatestFailReason: "",
		},
		{
			MutationID:       "mutation_1041.txt",
			Database:         "analytics",
			Table:            "orders",
			Command:          "ALTER TABLE analytics.orders UPDATE status = 'archived' WHERE order_date < '2026-01-01'",
			CreateTime:       now.Add(-2 * time.Hour),
			PartsToDo:        0,
			IsDone:           true,
			LatestFailedPart: "",
			LatestFailReason: "",
		},
	}
}

func (s *Simulator) initSampleProcesses() {
	s.processes = []models.ProcessInfo{
		{
			QueryID:        "q-b472e819",
			User:           "default",
			Address:        "127.0.0.1:49210",
			ElapsedSeconds: 1.84,
			RowsRead:       14209000,
			BytesRead:      184029100,
			MemoryUsage:    245892010, // ~234 MB
			PeakMemory:     312019400,
			Query:          "SELECT event_type, host, count(), avg(duration_ms) FROM telemetry.events WHERE event_time >= now() - INTERVAL 1 HOUR GROUP BY event_type, host ORDER BY count() DESC",
		},
		{
			QueryID:        "q-a912fc43",
			User:           "grafana",
			Address:        "10.0.1.45:38192",
			ElapsedSeconds: 0.32,
			RowsRead:       420100,
			BytesRead:      5120400,
			MemoryUsage:    18201900,
			PeakMemory:     22104000,
			Query:          "SELECT toStartOfMinute(event_time) AS t, count() FROM telemetry.events WHERE event_time >= now() - INTERVAL 15 MINUTE GROUP BY t ORDER BY t",
		},
	}
}

// GetStatus returns simulated cluster health.
func (s *Simulator) GetStatus() models.ServerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	return models.ServerStatus{
		Connected:      true,
		Version:        "24.8.3.16 (ClickLens Simulated Engine)",
		CurrentDSN:     "demo",
		Protocol:       "native",
		IsDemo:         true,
		UptimeSeconds:  uint64(time.Since(s.startTime).Seconds()),
		ActiveQueries:  len(s.processes),
		ActiveMerges:   len(s.merges),
		TotalParts:     84,
		TotalBytesDisk: 2849102400, // ~2.84 GB
		LastChecked:    time.Now(),
	}
}

// GetDatabases returns simulated databases and tables.
func (s *Simulator) GetDatabases() []models.DatabaseInfo {
	return []models.DatabaseInfo{
		{
			Name:       "telemetry",
			Engine:     "Atomic",
			TableCount: 2,
			Tables:     []string{"events", "metrics"},
		},
		{
			Name:       "analytics",
			Engine:     "Atomic",
			TableCount: 2,
			Tables:     []string{"orders", "customers"},
		},
		{
			Name:       "logs",
			Engine:     "Atomic",
			TableCount: 1,
			Tables:     []string{"access_log"},
		},
		{
			Name:       "system",
			Engine:     "System",
			TableCount: 5,
			Tables:     []string{"parts", "merges", "columns", "processes", "query_log"},
		},
	}
}

// GetTableParts generates realistic parts and partition health for the given table.
func (s *Simulator) GetTableParts(database, table string) models.TablePartsResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	if database == "" {
		database = "telemetry"
	}
	if table == "" {
		table = "events"
	}

	partitions := []models.PartitionSummary{}
	partitionKeys := []string{"202610", "202609", "202608"}

	totalActive := 0
	totalInactive := 0
	var totalDiskBytes uint64 = 0
	var totalUncompressedBytes uint64 = 0

	for i, pKey := range partitionKeys {
		var activeCount int
		var inactiveCount int
		var healthStatus string
		var fragScore float64

		if i == 0 {
			// Current partition: high ingestion rate, multiple parts pending background merge
			activeCount = 28
			inactiveCount = 6
			healthStatus = "WARNING"
			fragScore = 48.5
		} else if i == 1 {
			// Previous month: well merged
			activeCount = 6
			inactiveCount = 0
			healthStatus = "HEALTHY"
			fragScore = 12.0
		} else {
			// Older month: consolidated
			activeCount = 3
			inactiveCount = 0
			healthStatus = "HEALTHY"
			fragScore = 5.2
		}

		parts := []models.PartInfo{}
		var pDisk uint64 = 0
		var pUncompressed uint64 = 0
		var pRows uint64 = 0

		// Generate active parts
		for p := 1; p <= activeCount; p++ {
			level := 0
			if p%4 == 0 {
				level = 1
			}
			rows := uint64(50000 + (p * 18200))
			uncompressed := rows * 128
			diskBytes := uint64(float64(uncompressed) / (3.2 + float64(p%3)*0.5))

			pRows += rows
			pDisk += diskBytes
			pUncompressed += uncompressed

			parts = append(parts, models.PartInfo{
				Partition:             pKey,
				Name:                  fmt.Sprintf("%s_%d_%d_%d", pKey, p*3, p*3+2, level),
				Active:                true,
				BytesOnDisk:           diskBytes,
				DataUncompressedBytes: uncompressed,
				Marks:                 (rows / 8192) + 1,
				Rows:                  rows,
				ModificationTime:      time.Now().Add(-time.Duration(p*40) * time.Minute),
				MinBlockNumber:        int64(p * 3),
				MaxBlockNumber:        int64(p*3 + 2),
				Level:                 level,
				CompressionRatio:      float64(uncompressed) / float64(diskBytes),
			})
		}

		// Generate inactive parts (waiting for garbage collection)
		for p := 1; p <= inactiveCount; p++ {
			rows := uint64(15000 + p*5000)
			uncompressed := rows * 128
			diskBytes := uint64(float64(uncompressed) / 3.0)

			parts = append(parts, models.PartInfo{
				Partition:             pKey,
				Name:                  fmt.Sprintf("%s_%d_%d_0", pKey, p+100, p+100),
				Active:                false,
				BytesOnDisk:           diskBytes,
				DataUncompressedBytes: uncompressed,
				Marks:                 (rows / 8192) + 1,
				Rows:                  rows,
				ModificationTime:      time.Now().Add(-time.Duration(p*5) * time.Minute),
				MinBlockNumber:        int64(p + 100),
				MaxBlockNumber:        int64(p + 100),
				Level:                 0,
				CompressionRatio:      float64(uncompressed) / float64(diskBytes),
			})
		}

		ratio := 1.0
		if pDisk > 0 {
			ratio = float64(pUncompressed) / float64(pDisk)
		}

		optSQL := fmt.Sprintf("OPTIMIZE TABLE %s.%s PARTITION '%s' FINAL;", database, table, pKey)

		partitions = append(partitions, models.PartitionSummary{
			Partition:          pKey,
			ActiveParts:        activeCount,
			InactiveParts:      inactiveCount,
			TotalRows:          pRows,
			TotalBytesOnDisk:   pDisk,
			TotalUncompressed:  pUncompressed,
			CompressionRatio:   ratio,
			HealthStatus:       healthStatus,
			FragmentationScore: fragScore,
			OptimizationSQL:    optSQL,
			Parts:              parts,
		})

		totalActive += activeCount
		totalInactive += inactiveCount
		totalDiskBytes += pDisk
		totalUncompressedBytes += pUncompressed
	}

	overallRatio := 1.0
	if totalDiskBytes > 0 {
		overallRatio = float64(totalUncompressedBytes) / float64(totalDiskBytes)
	}

	healthRating := "HEALTHY"
	if totalActive > 50 {
		healthRating = "DEGRADED"
	}

	// Calculate realistic simulated part age buckets
	ageBuckets := []models.AgeBucket{
		{Label: "< 5 min", PartCount: 14, TotalRows: 840000, TotalBytes: 68400000},
		{Label: "5-30 min", PartCount: 10, TotalRows: 720000, TotalBytes: 82000000},
		{Label: "30m-2h", PartCount: 4, TotalRows: 480000, TotalBytes: 64000000},
		{Label: "2h-24h", PartCount: 6, TotalRows: 1200000, TotalBytes: 145000000},
		{Label: "> 24h", PartCount: 3, TotalRows: 2400000, TotalBytes: 310000000},
	}

	microBatchingWarning := true
	microBatchingDetails := "High ingestion frequency detected: 14 active parts created in the last 5 minutes. Consider using buffer tables or streaming batches to avoid write throttling."

	return models.TablePartsResponse{
		Database:             database,
		Table:                table,
		Engine:               "MergeTree",
		TotalActiveParts:     totalActive,
		TotalInactiveParts:   totalInactive,
		TotalBytesOnDisk:     totalDiskBytes,
		TotalUncompressed:    totalUncompressedBytes,
		OverallRatio:         overallRatio,
		MaxPartsPartition:    "202610",
		MaxPartCount:         28,
		HealthRating:         healthRating,
		PartAgeBuckets:       ageBuckets,
		MicroBatchingWarning: microBatchingWarning,
		MicroBatchingDetails: microBatchingDetails,
		Partitions:           partitions,
	}
}

// TickMerges updates progress of background merges to create real-time dynamics.
func (s *Simulator) TickMerges() []models.MergeInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	delta := now.Sub(s.lastTickTime).Seconds()
	s.lastTickTime = now

	for i := range s.merges {
		s.merges[i].ElapsedSeconds += delta
		s.merges[i].Progress += delta * (4.0 + rand.Float64()*3.0)

		if s.merges[i].Progress >= 100.0 {
			// Finish merge and start a new simulated merge
			s.merges[i].Progress = 5.0
			s.merges[i].ElapsedSeconds = 0.5
			s.merges[i].NumParts = uint64(3 + rand.Intn(6))
			s.merges[i].ResultPartName = fmt.Sprintf("202610_%d_%d_2", 120+rand.Intn(40), 160+rand.Intn(40))
		}
	}

	res := make([]models.MergeInfo, len(s.merges))
	copy(res, s.merges)
	return res
}

// GetMutations returns simulated mutations.
func (s *Simulator) GetMutations() []models.MutationInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := make([]models.MutationInfo, len(s.mutations))
	copy(res, s.mutations)
	return res
}

// GetProcesses returns simulated running queries.
func (s *Simulator) GetProcesses() []models.ProcessInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := make([]models.ProcessInfo, len(s.processes))
	copy(res, s.processes)
	return res
}

// KillProcess emulates killing an active query.
func (s *Simulator) KillProcess(queryID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.processes {
		if p.QueryID == queryID {
			s.processes = append(s.processes[:i], s.processes[i+1:]...)
			return true
		}
	}
	return false
}

// GetColumnsAudit evaluates column compression efficiency for the given table.
func (s *Simulator) GetColumnsAudit(database, table string) models.TableColumnsResponse {
	if database == "" {
		database = "telemetry"
	}
	if table == "" {
		table = "events"
	}

	columns := []models.ColumnAudit{
		{
			Database:               database,
			Table:                  table,
			Column:                 "event_time",
			Type:                   "DateTime64(3)",
			DefaultExpression:      "",
			CompressionCodec:       "Default(LZ4)",
			DataCompressedBytes:    142019400,
			DataUncompressedBytes:  220419200,
			CompressionRatio:       1.55,
			RecommendedCodec:       "CODEC(DoubleDelta, ZSTD(1))",
			SavingsEstimatePercent: 62.4,
			ProjectedSavingsBytes:  88619000,
			AlterSQL:               fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN event_time DateTime64(3) CODEC(DoubleDelta, ZSTD(1));", database, table),
			Rationale:              "Monotonic timestamps exhibit uniform delta spacing. DoubleDelta reduces bit width by 80% before ZSTD entropy compression.",
		},
		{
			Database:               database,
			Table:                  table,
			Column:                 "seq_id",
			Type:                   "UInt64",
			DefaultExpression:      "",
			CompressionCodec:       "Default(LZ4)",
			DataCompressedBytes:    98104000,
			DataUncompressedBytes:  220419200,
			CompressionRatio:       2.24,
			RecommendedCodec:       "CODEC(T64, ZSTD(1))",
			SavingsEstimatePercent: 54.0,
			ProjectedSavingsBytes:  52976160,
			AlterSQL:               fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN seq_id UInt64 CODEC(T64, ZSTD(1));", database, table),
			Rationale:              "Sequential integer IDs benefit from T64 transposition encoding, grouping identical high-order bits into compact runs.",
		},
		{
			Database:               database,
			Table:                  table,
			Column:                 "cpu_usage",
			Type:                   "Float64",
			DefaultExpression:      "",
			CompressionCodec:       "Default(LZ4)",
			DataCompressedBytes:    164920000,
			DataUncompressedBytes:  220419200,
			CompressionRatio:       1.33,
			RecommendedCodec:       "CODEC(Gorilla, ZSTD(1))",
			SavingsEstimatePercent: 58.2,
			ProjectedSavingsBytes:  95983440,
			AlterSQL:               fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN cpu_usage Float64 CODEC(Gorilla, ZSTD(1));", database, table),
			Rationale:              "Floating-point sensor metrics have repeated XOR exponents. Gorilla compression exploits continuous value continuity.",
		},
		{
			Database:               database,
			Table:                  table,
			Column:                 "host",
			Type:                   "String",
			DefaultExpression:      "",
			CompressionCodec:       "Default(LZ4)",
			DataCompressedBytes:    72019000,
			DataUncompressedBytes:  310492000,
			CompressionRatio:       4.31,
			RecommendedCodec:       "LowCardinality(String) CODEC(ZSTD(3))",
			SavingsEstimatePercent: 44.5,
			ProjectedSavingsBytes:  32048455,
			AlterSQL:               fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN host LowCardinality(String) CODEC(ZSTD(3));", database, table),
			Rationale:              "Host names have low distinct cardinality. Replacing String with LowCardinality dictionary encoding cuts index size by 70%.",
		},
		{
			Database:               database,
			Table:                  table,
			Column:                 "payload_json",
			Type:                   "String",
			DefaultExpression:      "",
			CompressionCodec:       "CODEC(ZSTD(3))",
			DataCompressedBytes:    340192000,
			DataUncompressedBytes:  1840291000,
			CompressionRatio:       5.41,
			RecommendedCodec:       "CODEC(ZSTD(6))",
			SavingsEstimatePercent: 18.0,
			ProjectedSavingsBytes:  61234560,
			AlterSQL:               fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN payload_json String CODEC(ZSTD(6));", database, table),
			Rationale:              "High-entropy JSON text achieves an additional 18% compression ratio jump when elevating ZSTD level from 3 to 6.",
		},
	}

	var totalComp uint64 = 0
	var totalUncomp uint64 = 0
	var potSavings uint64 = 0

	for _, col := range columns {
		totalComp += col.DataCompressedBytes
		totalUncomp += col.DataUncompressedBytes
		potSavings += col.ProjectedSavingsBytes
	}

	migrationScript := fmt.Sprintf(`-- ==============================================================================
-- Zero-Downtime Codec Migration Playbook for %s.%s
-- Generated automatically by ClickLens Studio
-- ==============================================================================

-- Step 1: Create a staging shadow table with the existing schema
CREATE TABLE IF NOT EXISTS %s.%s_new AS %s.%s;

-- Step 2: Apply optimized compression codecs to new table definition
ALTER TABLE %s.%s_new MODIFY COLUMN event_time DateTime64(3) CODEC(DoubleDelta, ZSTD(1));
ALTER TABLE %s.%s_new MODIFY COLUMN seq_id UInt64 CODEC(T64, ZSTD(1));
ALTER TABLE %s.%s_new MODIFY COLUMN cpu_usage Float64 CODEC(Gorilla, ZSTD(1));
ALTER TABLE %s.%s_new MODIFY COLUMN host LowCardinality(String) CODEC(ZSTD(3));
ALTER TABLE %s.%s_new MODIFY COLUMN payload_json String CODEC(ZSTD(6));

-- Step 3: Populate new table with compressed data partition by partition
-- (Run during low-traffic windows or batch by month)
INSERT INTO %s.%s_new SELECT * FROM %s.%s;

-- Step 4: Atomic zero-downtime swap of tables (seamless to active queries)
EXCHANGE TABLES %s.%s AND %s.%s_new;

-- Step 5: Verify row counts and drop the old legacy uncompressed table
-- SELECT count() FROM %s.%s;
-- DROP TABLE %s.%s_new;
`, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table, database, table)

	return models.TableColumnsResponse{
		Database:                    database,
		Table:                       table,
		TotalColumns:                len(columns),
		TotalCompressed:             totalComp,
		TotalUncompressed:           totalUncomp,
		PotentialSavings:            potSavings,
		ZeroDowntimeMigrationScript: migrationScript,
		Columns:                     columns,
	}
}

// GenerateSamplePipelineAnalysis returns a structured DAG simulating EXPLAIN PIPELINE output.
func (s *Simulator) GenerateSamplePipelineAnalysis(query string) models.PipelineAnalysis {
	nodes := []models.DAGNode{
		{
			ID:          "step-1",
			Label:       "ReadFromMergeTree",
			StepType:    "ReadFromMergeTree",
			StreamCount: 16,
			Details:     "telemetry.events (14 parts, 1,420,180 rows, 142 MB)",
			Outputs:     []string{"step-2"},
		},
		{
			ID:          "step-2",
			Label:       "FilterTransform",
			StepType:    "FilterTransform",
			StreamCount: 16,
			Details:     "event_time >= now() - INTERVAL 1 HOUR",
			Inputs:      []string{"step-1"},
			Outputs:     []string{"step-3"},
		},
		{
			ID:          "step-3",
			Label:       "AggregatingTransform",
			StepType:    "AggregatingTransform",
			StreamCount: 16,
			Details:     "GROUP BY event_type, host (hash aggregation)",
			Inputs:      []string{"step-2"},
			Outputs:     []string{"step-4"},
		},
		{
			ID:           "step-4",
			Label:        "MergeAggregatedTransform",
			StepType:     "MergeAggregatedTransform",
			StreamCount:  1,
			Details:      "Fan-in from 16 parallel thread streams to 1 single thread",
			Inputs:       []string{"step-3"},
			Outputs:      []string{"step-5"},
			IsBottleneck: true,
		},
		{
			ID:          "step-5",
			Label:       "SortingTransform",
			StepType:    "SortingTransform",
			StreamCount: 1,
			Details:     "ORDER BY count() DESC LIMIT 100",
			Inputs:      []string{"step-4"},
			Outputs:     []string{"step-6"},
		},
		{
			ID:          "step-6",
			Label:       "LimitsCheckingTransform",
			StepType:    "LimitsCheckingTransform",
			StreamCount: 1,
			Details:     "LIMIT 100",
			Inputs:      []string{"step-5"},
		},
	}

	edges := []models.DAGEdge{
		{From: "step-1", To: "step-2", StreamCount: 16, Label: "16 streams"},
		{From: "step-2", To: "step-3", StreamCount: 16, Label: "16 streams"},
		{From: "step-3", To: "step-4", StreamCount: 1, Label: "Fan-in (16 -> 1)"},
		{From: "step-4", To: "step-5", StreamCount: 1, Label: "1 stream"},
		{From: "step-5", To: "step-6", StreamCount: 1, Label: "1 stream"},
	}

	bottlenecks := []models.PipelineBottleneck{
		{
			NodeID:          "step-4",
			StepType:        "MergeAggregatedTransform",
			PreviousStreams: 16,
			CurrentStreams:  1,
			Severity:        "WARNING",
			Description:     "Parallel stream fan-in bottleneck: 16 worker threads collapsed into 1 single coordinator thread.",
			Recommendation: "Consider setting 'max_threads = 16' and checking whether 'totals_having_mode' can be calculated in parallel.",
		},
	}

	rawText := `(Expression)
ExpressionTransform × 1
  (Sorting)
  MergingSortedTransform 16 → 1
    LimitsCheckingTransform × 16
      PartialSortingTransform × 16
        (Aggregating)
        Resize 16 → 16
          AggregatingTransform × 16
            (Filter)
            FilterTransform × 16
              (ReadFromMergeTree)
              MergeTreeSelect(pool: 16) × 16`

	return models.PipelineAnalysis{
		Query:          query,
		Mode:           "PIPELINE",
		RawText:        rawText,
		TotalStreams:   16,
		MaxConcurrency: 16,
		Nodes:          nodes,
		Edges:          edges,
		Bottlenecks:    bottlenecks,
		SpillAlerts:    []models.SpillAlert{},
	}
}

// GetQueryLog returns realistic simulated query log history.
func (s *Simulator) GetQueryLog(filter string, limit int) *models.QueryLogResponse {
	now := time.Now()

	allEntries := []models.QueryLogEntry{
		{
			QueryID:       "q-9817bf44-32",
			EventTime:     now.Add(-2 * time.Minute),
			DurationMs:    18452.3,
			MemoryUsage:   4820194816, // 4.8 GB
			ReadRows:      48201940,
			ReadBytes:     3841029400,
			ResultRows:    100,
			ResultBytes:   14200,
			QueryText:     "SELECT host, event_type, count(), uniqExact(user_id), sum(duration_ms) FROM telemetry.events WHERE event_time >= now() - INTERVAL 7 DAY GROUP BY host, event_type ORDER BY count() DESC LIMIT 100",
			ExceptionCode: 0,
			HasSpill:      true,
			Type:          "QueryFinish",
		},
		{
			QueryID:       "q-7740ca91-11",
			EventTime:     now.Add(-6 * time.Minute),
			DurationMs:    12840.1,
			MemoryUsage:   2940192800, // ~2.9 GB
			ReadRows:      28491024,
			ReadBytes:     1940291000,
			ResultRows:    50000,
			ResultBytes:   4920190,
			QueryText:     "SELECT * FROM telemetry.events WHERE payload_json LIKE '%payment_error%' ORDER BY event_time DESC LIMIT 50000",
			ExceptionCode: 0,
			HasSpill:      false,
			Type:          "QueryFinish",
		},
		{
			QueryID:       "q-3312fe88-99",
			EventTime:     now.Add(-14 * time.Minute),
			DurationMs:    3410.5,
			MemoryUsage:   10737418240, // 10 GB (hit limit)
			ReadRows:      88102940,
			ReadBytes:     7102948100,
			ResultRows:    0,
			ResultBytes:   0,
			QueryText:     "SELECT user_id, groupArray(payload_json) FROM telemetry.events GROUP BY user_id",
			ExceptionCode: 241,
			Exception:     "Code: 241. DB::Exception: Memory limit (total) exceeded: would use 10.00 GiB (limit: 8.00 GiB). (MEMORY_LIMIT_EXCEEDED)",
			HasSpill:      false,
			Type:          "ExceptionWhileProcessing",
		},
		{
			QueryID:       "q-5510ab12-05",
			EventTime:     now.Add(-22 * time.Minute),
			DurationMs:    8940.0,
			MemoryUsage:   1840192000,
			ReadRows:      18491020,
			ReadBytes:     1240192000,
			ResultRows:    1200,
			ResultBytes:   84019,
			QueryText:     "SELECT toStartOfHour(event_time) AS h, status, count() FROM analytics.orders GROUP BY h, status ORDER BY h",
			ExceptionCode: 0,
			HasSpill:      true,
			Type:          "QueryFinish",
		},
		{
			QueryID:       "q-1194dd77-43",
			EventTime:     now.Add(-35 * time.Minute),
			DurationMs:    142.1,
			MemoryUsage:   64102940,
			ReadRows:      0,
			ReadBytes:     0,
			ResultRows:    0,
			ResultBytes:   0,
			QueryText:     "SELECT * FROM telemetry.non_existent_table WHERE id = 12",
			ExceptionCode: 60,
			Exception:     "Code: 60. DB::Exception: Table telemetry.non_existent_table doesn't exist. (UNKNOWN_TABLE)",
			HasSpill:      false,
			Type:          "ExceptionWhileProcessing",
		},
		{
			QueryID:       "q-8491bb33-67",
			EventTime:     now.Add(-48 * time.Minute),
			DurationMs:    6240.8,
			MemoryUsage:   3410291000,
			ReadRows:      14201940,
			ReadBytes:     940291000,
			ResultRows:    25,
			ResultBytes:   3200,
			QueryText:     "SELECT host, quantile(0.99)(cpu_usage) FROM telemetry.metrics WHERE event_time >= now() - INTERVAL 1 DAY GROUP BY host",
			ExceptionCode: 0,
			HasSpill:      false,
			Type:          "QueryFinish",
		},
	}

	filtered := make([]models.QueryLogEntry, 0)
	for _, entry := range allEntries {
		switch filter {
		case "memory":
			if entry.MemoryUsage > 1024*1024*1024 { // > 1 GB
				filtered = append(filtered, entry)
			}
		case "spill":
			if entry.HasSpill {
				filtered = append(filtered, entry)
			}
		case "errors":
			if entry.ExceptionCode != 0 || entry.Type == "ExceptionWhileProcessing" {
				filtered = append(filtered, entry)
			}
		default: // "slow" or all
			filtered = append(filtered, entry)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return &models.QueryLogResponse{
		Filter:  filter,
		Total:   len(filtered),
		Entries: filtered,
	}
}

// ExecuteQuery parses and evaluates SQL queries for the simulated telemetry engine.
func (s *Simulator) ExecuteQuery(query string) models.QueryResult {
	clean := strings.TrimSpace(query)
	upper := strings.ToUpper(clean)
	t0 := time.Now()

	switch {
	case strings.HasPrefix(upper, "SHOW TABLES"):
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"name"},
			Rows:            [][]any{{"events"}, {"metrics"}, {"orders"}, {"customers"}, {"access_log"}},
			RowsCount:       5,
			ExecutionTimeMs: 1.2,
		}

	case strings.HasPrefix(upper, "SHOW DATABASES"):
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"name"},
			Rows:            [][]any{{"telemetry"}, {"analytics"}, {"logs"}, {"system"}},
			RowsCount:       4,
			ExecutionTimeMs: 0.8,
		}

	case strings.Contains(upper, "VERSION()"):
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"version()"},
			Rows:            [][]any{{"24.8.3.16 (ClickLens Simulated Engine)"}},
			RowsCount:       1,
			ExecutionTimeMs: 0.5,
		}

	case strings.Contains(upper, "COUNT()") && strings.Contains(upper, "EVENTS"):
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"count()"},
			Rows:            [][]any{{1420180}},
			RowsCount:       1,
			ExecutionTimeMs: 2.4,
			RowsRead:        1420180,
			BytesRead:       14201800,
		}

	case strings.Contains(upper, "ORDERS"):
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"order_id", "customer_id", "status", "amount_usd", "created_at"},
			Rows: [][]any{
				{1001, 401, "completed", 189.50, "2026-10-22 14:20:10"},
				{1002, 402, "processing", 42.00, "2026-10-22 14:21:05"},
				{1003, 405, "completed", 1250.00, "2026-10-22 14:23:44"},
				{1004, 408, "refunded", 99.99, "2026-10-22 14:25:12"},
				{1005, 412, "completed", 310.25, "2026-10-22 14:28:30"},
			},
			RowsCount:       5,
			ExecutionTimeMs: 4.1,
			RowsRead:        50000,
			BytesRead:       1840000,
		}

	default: // Default telemetry events
		now := time.Now().Format("2006-01-02 15:04:05")
		return models.QueryResult{
			Query:           clean,
			Columns:         []string{"event_time", "event_type", "host", "duration_ms", "user_id"},
			Rows: [][]any{
				{now, "api_request", "prod-worker-01", 14.2, 8912},
				{now, "db_query", "prod-worker-02", 8.4, 9041},
				{now, "checkout", "prod-worker-01", 48.9, 1204},
				{now, "login", "auth-gateway", 22.1, 7731},
				{now, "cache_hit", "cache-node-03", 1.8, 5509},
			},
			RowsCount:       5,
			ExecutionTimeMs: float64(time.Since(t0).Microseconds()) / 1000.0,
			RowsRead:        142090,
			BytesRead:       4810240,
		}
	}
}
