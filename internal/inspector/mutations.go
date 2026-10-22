package inspector

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// MutationsInspector monitors asynchronous table mutations from system.mutations.
type MutationsInspector struct {
	client *client.Client
}

// NewMutationsInspector creates a MutationsInspector.
func NewMutationsInspector(c *client.Client) *MutationsInspector {
	return &MutationsInspector{client: c}
}

// GetMutations retrieves table mutations.
func (mi *MutationsInspector) GetMutations(ctx context.Context, database, table string) ([]models.MutationInfo, error) {
	if mi.client.IsSimulated() {
		return mi.client.Simulator().GetMutations(), nil
	}

	conn := mi.client.Conn()
	if conn == nil {
		return mi.client.Simulator().GetMutations(), nil
	}

	query := `
		SELECT
			mutation_id,
			database,
			table,
			command,
			create_time,
			parts_to_do,
			is_done,
			latest_failed_part,
			latest_fail_reason
		FROM system.mutations
		WHERE (database = ? OR ? = '') AND (table = ? OR ? = '')
		ORDER BY create_time DESC
		LIMIT 50
	`

	rows, err := conn.Query(ctx, query, database, database, table, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.mutations: %w", err)
	}
	defer rows.Close()

	mutations := []models.MutationInfo{}
	for rows.Next() {
		var (
			mutationID       string
			db               string
			tbl              string
			command          string
			createTime       time.Time
			partsToDo        int64
			isDone           uint8
			latestFailedPart string
			latestFailReason string
		)

		if err := rows.Scan(
			&mutationID,
			&db,
			&tbl,
			&command,
			&createTime,
			&partsToDo,
			&isDone,
			&latestFailedPart,
			&latestFailReason,
		); err != nil {
			continue
		}

		mutations = append(mutations, models.MutationInfo{
			MutationID:       mutationID,
			Database:         db,
			Table:            tbl,
			Command:          command,
			CreateTime:       createTime,
			PartsToDo:        partsToDo,
			IsDone:           isDone == 1,
			LatestFailedPart: latestFailedPart,
			LatestFailReason: latestFailReason,
		})
	}

	return mutations, nil
}
