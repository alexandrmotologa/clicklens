package inspector

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// MergesInspector monitors background merges from system.merges.
type MergesInspector struct {
	client *client.Client
}

// NewMergesInspector creates a MergesInspector.
func NewMergesInspector(c *client.Client) *MergesInspector {
	return &MergesInspector{client: c}
}

// GetActiveMerges retrieves currently running background merges.
func (mi *MergesInspector) GetActiveMerges(ctx context.Context) ([]models.MergeInfo, error) {
	if mi.client.IsSimulated() {
		return mi.client.Simulator().TickMerges(), nil
	}

	conn := mi.client.Conn()
	if conn == nil {
		return mi.client.Simulator().TickMerges(), nil
	}

	query := `
		SELECT
			database,
			table,
			elapsed,
			progress,
			num_parts,
			result_part_name,
			source_part_names,
			bytes_read_uncompressed,
			bytes_written_uncompressed,
			rows_read,
			rows_written,
			memory_usage
		FROM system.merges
		ORDER BY elapsed DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.merges: %w", err)
	}
	defer rows.Close()

	merges := []models.MergeInfo{}
	for rows.Next() {
		var (
			database        string
			table           string
			elapsed         float64
			progress        float64
			numParts        uint64
			resultPartName  string
			sourceParts     []string
			bytesRead       uint64
			bytesWritten    uint64
			rowsRead        uint64
			rowsWritten     uint64
			memoryUsage     uint64
		)

		if err := rows.Scan(
			&database,
			&table,
			&elapsed,
			&progress,
			&numParts,
			&resultPartName,
			&sourceParts,
			&bytesRead,
			&bytesWritten,
			&rowsRead,
			&rowsWritten,
			&memoryUsage,
		); err != nil {
			continue
		}

		readRate := 0.0
		writeRate := 0.0
		if elapsed > 0 {
			readRate = float64(bytesRead) / elapsed
			writeRate = float64(bytesWritten) / elapsed
		}

		merges = append(merges, models.MergeInfo{
			Database:        database,
			Table:           table,
			ElapsedSeconds:  elapsed,
			Progress:        progress * 100.0, // Convert 0-1 to percentage
			NumParts:        numParts,
			ResultPartName:  resultPartName,
			SourceParts:     sourceParts,
			BytesRead:       bytesRead,
			BytesWritten:    bytesWritten,
			RowsRead:        rowsRead,
			RowsWritten:     rowsWritten,
			MemoryUsage:     memoryUsage,
			ReadThroughput:  readRate,
			WriteThroughput: writeRate,
		})
	}

	return merges, nil
}
