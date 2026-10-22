package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// QueryLogInspector queries system.query_log to identify slow queries, memory spikes, and disk spills.
type QueryLogInspector struct {
	client *client.Client
}

// NewQueryLogInspector creates a QueryLogInspector.
func NewQueryLogInspector(c *client.Client) *QueryLogInspector {
	return &QueryLogInspector{client: c}
}

// GetQueryLog retrieves historical queries according to the requested filter.
func (qi *QueryLogInspector) GetQueryLog(ctx context.Context, filter string, limit int) (*models.QueryLogResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if filter == "" {
		filter = "slow"
	}

	if qi.client.IsSimulated() {
		return qi.client.Simulator().GetQueryLog(filter, limit), nil
	}

	conn := qi.client.Conn()
	if conn == nil {
		return qi.client.Simulator().GetQueryLog(filter, limit), nil
	}

	whereClause := "type = 'QueryFinish'"
	orderBy := "query_duration_ms DESC"

	switch strings.ToLower(filter) {
	case "memory":
		whereClause = "type = 'QueryFinish'"
		orderBy = "memory_usage DESC"
	case "spill":
		whereClause = "type = 'QueryFinish' AND (ProfileEvents['ExternalProcessingRows'] > 0 OR ProfileEvents['DiskReadElapsedMicroseconds'] > 500000)"
		orderBy = "query_duration_ms DESC"
	case "errors":
		whereClause = "type = 'ExceptionWhileProcessing'"
		orderBy = "event_time DESC"
	default: // "slow"
		whereClause = "type = 'QueryFinish'"
		orderBy = "query_duration_ms DESC"
	}

	query := fmt.Sprintf(`
		SELECT
			query_id,
			event_time,
			query_duration_ms,
			memory_usage,
			read_rows,
			read_bytes,
			result_rows,
			result_bytes,
			query,
			exception_code,
			exception,
			if(ProfileEvents['ExternalProcessingRows'] > 0, 1, 0) AS has_spill,
			type
		FROM system.query_log
		WHERE %s
		ORDER BY %s
		LIMIT %d
	`, whereClause, orderBy, limit)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.query_log: %w", err)
	}
	defer rows.Close()

	entries := []models.QueryLogEntry{}
	for rows.Next() {
		var (
			queryID        string
			eventTime      time.Time
			durationMs     float64
			memoryUsage    uint64
			readRows       uint64
			readBytes      uint64
			resultRows     uint64
			resultBytes    uint64
			queryText      string
			exceptionCode  int
			exception      string
			hasSpill       uint8
			entryType      string
		)

		if err := rows.Scan(
			&queryID,
			&eventTime,
			&durationMs,
			&memoryUsage,
			&readRows,
			&readBytes,
			&resultRows,
			&resultBytes,
			&queryText,
			&exceptionCode,
			&exception,
			&hasSpill,
			&entryType,
		); err != nil {
			continue
		}

		entries = append(entries, models.QueryLogEntry{
			QueryID:       queryID,
			EventTime:     eventTime,
			DurationMs:    durationMs,
			MemoryUsage:   memoryUsage,
			ReadRows:      readRows,
			ReadBytes:     readBytes,
			ResultRows:    resultRows,
			ResultBytes:   resultBytes,
			QueryText:     queryText,
			ExceptionCode: exceptionCode,
			Exception:     exception,
			HasSpill:      hasSpill == 1,
			Type:          entryType,
		})
	}

	return &models.QueryLogResponse{
		Filter:  filter,
		Total:   len(entries),
		Entries: entries,
	}, nil
}
