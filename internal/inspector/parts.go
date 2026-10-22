package inspector

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// PartsInspector audits system.parts and evaluates partition fragmentation.
type PartsInspector struct {
	client *client.Client
}

// NewPartsInspector creates a new PartsInspector.
func NewPartsInspector(c *client.Client) *PartsInspector {
	return &PartsInspector{client: c}
}

// GetTableParts retrieves parts and computes partition diagnostics.
func (pi *PartsInspector) GetTableParts(ctx context.Context, database, table string) (*models.TablePartsResponse, error) {
	if pi.client.IsSimulated() {
		res := pi.client.Simulator().GetTableParts(database, table)
		return &res, nil
	}

	conn := pi.client.Conn()
	if conn == nil {
		res := pi.client.Simulator().GetTableParts(database, table)
		return &res, nil
	}

	query := `
		SELECT
			partition,
			name,
			active,
			bytes_on_disk,
			data_uncompressed_bytes,
			marks,
			rows,
			modification_time,
			min_block_number,
			max_block_number,
			level
		FROM system.parts
		WHERE database = ? AND table = ?
		ORDER BY partition DESC, name ASC
	`

	rows, err := conn.Query(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.parts for %s.%s: %w", database, table, err)
	}
	defer rows.Close()

	partitionMap := make(map[string]*models.PartitionSummary)
	partitionOrder := make([]string, 0)

	var totalActive int
	var totalInactive int
	var totalDiskBytes uint64
	var totalUncompressed uint64

	for rows.Next() {
		var (
			partition        string
			name             string
			active           uint8
			bytesOnDisk      uint64
			uncompressed     uint64
			marks            uint64
			partRows         uint64
			modificationTime time.Time
			minBlock         int64
			maxBlock         int64
			level            int
		)

		if err := rows.Scan(
			&partition,
			&name,
			&active,
			&bytesOnDisk,
			&uncompressed,
			&marks,
			&partRows,
			&modificationTime,
			&minBlock,
			&maxBlock,
			&level,
		); err != nil {
			continue
		}

		isActive := active == 1
		ratio := 1.0
		if bytesOnDisk > 0 {
			ratio = float64(uncompressed) / float64(bytesOnDisk)
		}

		part := models.PartInfo{
			Partition:             partition,
			Name:                  name,
			Active:                isActive,
			BytesOnDisk:           bytesOnDisk,
			DataUncompressedBytes: uncompressed,
			Marks:                 marks,
			Rows:                  partRows,
			ModificationTime:      modificationTime,
			MinBlockNumber:        minBlock,
			MaxBlockNumber:        maxBlock,
			Level:                 level,
			CompressionRatio:      ratio,
		}

		summary, exists := partitionMap[partition]
		if !exists {
			summary = &models.PartitionSummary{
				Partition:       partition,
				OptimizationSQL: fmt.Sprintf("OPTIMIZE TABLE %s.%s PARTITION '%s' FINAL;", database, table, partition),
				Parts:           []models.PartInfo{},
			}
			partitionMap[partition] = summary
			partitionOrder = append(partitionOrder, partition)
		}

		summary.Parts = append(summary.Parts, part)
		summary.TotalRows += partRows
		summary.TotalBytesOnDisk += bytesOnDisk
		summary.TotalUncompressed += uncompressed

		if isActive {
			summary.ActiveParts++
			totalActive++
		} else {
			summary.InactiveParts++
			totalInactive++
		}
		totalDiskBytes += bytesOnDisk
		totalUncompressed += uncompressed
	}

	partitions := make([]models.PartitionSummary, 0, len(partitionOrder))
	maxPartCount := 0
	maxPartition := ""

	for _, pKey := range partitionOrder {
		summary := partitionMap[pKey]

		if summary.TotalBytesOnDisk > 0 {
			summary.CompressionRatio = float64(summary.TotalUncompressed) / float64(summary.TotalBytesOnDisk)
		} else {
			summary.CompressionRatio = 1.0
		}

		// Calculate health and fragmentation
		// If active parts > 300, ClickHouse default throws "Too many parts in all data parts in table"
		if summary.ActiveParts > 300 {
			summary.HealthStatus = "CRITICAL"
		} else if summary.ActiveParts > 80 {
			summary.HealthStatus = "WARNING"
		} else {
			summary.HealthStatus = "HEALTHY"
		}

		// Fragmentation score: proportion of small parts (< 10MB) combined with active part count
		smallParts := 0
		for _, pt := range summary.Parts {
			if pt.Active && pt.BytesOnDisk < 10*1024*1024 {
				smallParts++
			}
		}
		var frag float64 = 0
		if summary.ActiveParts > 0 {
			frag = (float64(smallParts) / float64(summary.ActiveParts)) * 60.0
			if summary.ActiveParts > 50 {
				frag += 40.0
			} else {
				frag += float64(summary.ActiveParts) * 0.8
			}
		}
		if frag > 100.0 {
			frag = 100.0
		}
		summary.FragmentationScore = frag

		if summary.ActiveParts > maxPartCount {
			maxPartCount = summary.ActiveParts
			maxPartition = pKey
		}

		partitions = append(partitions, *summary)
	}

	overallRatio := 1.0
	if totalDiskBytes > 0 {
		overallRatio = float64(totalUncompressed) / float64(totalDiskBytes)
	}

	healthRating := "HEALTHY"
	if maxPartCount > 300 {
		healthRating = "CRITICAL"
	} else if maxPartCount > 80 {
		healthRating = "DEGRADED"
	}

	// Calculate Part Age Buckets
	now := time.Now()
	ageBuckets := []models.AgeBucket{
		{Label: "< 5 min"},
		{Label: "5-30 min"},
		{Label: "30m-2h"},
		{Label: "2h-24h"},
		{Label: "> 24h"},
	}

	recentPartCount := 0
	for _, summary := range partitionMap {
		for _, pt := range summary.Parts {
			if !pt.Active {
				continue
			}
			age := now.Sub(pt.ModificationTime)
			idx := 4
			switch {
			case age < 5*time.Minute:
				idx = 0
				recentPartCount++
			case age < 30*time.Minute:
				idx = 1
			case age < 2*time.Hour:
				idx = 2
			case age < 24*time.Hour:
				idx = 3
			}
			ageBuckets[idx].PartCount++
			ageBuckets[idx].TotalRows += pt.Rows
			ageBuckets[idx].TotalBytes += pt.BytesOnDisk
		}
	}

	microBatchingWarning := false
	microBatchingDetails := ""
	if recentPartCount >= 10 {
		microBatchingWarning = true
		microBatchingDetails = fmt.Sprintf("High ingestion frequency detected: %d parts created within the last 5 minutes. Consider using buffer tables or batching inserts to avoid write throttling.", recentPartCount)
	}

	return &models.TablePartsResponse{
		Database:             database,
		Table:                table,
		Engine:               "MergeTree",
		TotalActiveParts:     totalActive,
		TotalInactiveParts:   totalInactive,
		TotalBytesOnDisk:     totalDiskBytes,
		TotalUncompressed:    totalUncompressed,
		OverallRatio:         overallRatio,
		MaxPartsPartition:    maxPartition,
		MaxPartCount:         maxPartCount,
		HealthRating:         healthRating,
		PartAgeBuckets:       ageBuckets,
		MicroBatchingWarning: microBatchingWarning,
		MicroBatchingDetails: microBatchingDetails,
		Partitions:           partitions,
	}, nil
}
