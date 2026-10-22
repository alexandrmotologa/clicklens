package inspector

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// ProcessesInspector monitors running queries from system.processes and handles cancellation.
type ProcessesInspector struct {
	client *client.Client
}

// NewProcessesInspector creates a ProcessesInspector.
func NewProcessesInspector(c *client.Client) *ProcessesInspector {
	return &ProcessesInspector{client: c}
}

// GetActiveProcesses returns all running queries.
func (pi *ProcessesInspector) GetActiveProcesses(ctx context.Context) ([]models.ProcessInfo, error) {
	if pi.client.IsSimulated() {
		return pi.client.Simulator().GetProcesses(), nil
	}

	conn := pi.client.Conn()
	if conn == nil {
		return pi.client.Simulator().GetProcesses(), nil
	}

	query := `
		SELECT
			query_id,
			user,
			client_hostname,
			elapsed,
			read_rows,
			read_bytes,
			memory_usage,
			peak_memory_usage,
			query
		FROM system.processes
		WHERE is_cancelled = 0
		ORDER BY elapsed DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.processes: %w", err)
	}
	defer rows.Close()

	processes := []models.ProcessInfo{}
	for rows.Next() {
		var (
			queryID        string
			user           string
			address        string
			elapsed        float64
			readRows       uint64
			readBytes      uint64
			memoryUsage    uint64
			peakMemory     uint64
			qText          string
		)

		if err := rows.Scan(
			&queryID,
			&user,
			&address,
			&elapsed,
			&readRows,
			&readBytes,
			&memoryUsage,
			&peakMemory,
			&qText,
		); err != nil {
			continue
		}

		processes = append(processes, models.ProcessInfo{
			QueryID:        queryID,
			User:           user,
			Address:        address,
			ElapsedSeconds: elapsed,
			RowsRead:       readRows,
			BytesRead:      readBytes,
			MemoryUsage:    memoryUsage,
			PeakMemory:     peakMemory,
			Query:          qText,
		})
	}

	return processes, nil
}

// KillQuery issues a KILL QUERY statement to terminate a query.
func (pi *ProcessesInspector) KillQuery(ctx context.Context, queryID string) error {
	if pi.client.IsSimulated() {
		pi.client.Simulator().KillProcess(queryID)
		return nil
	}

	conn := pi.client.Conn()
	if conn == nil {
		return fmt.Errorf("clickhouse connection not available")
	}

	killSQL := fmt.Sprintf("KILL QUERY WHERE query_id = '%s' SYNC", queryID)
	return conn.Exec(ctx, killSQL)
}
