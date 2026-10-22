package profiler

import (
	"context"
	"strings"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// MemoryTracker evaluates query memory allocations and flags spillover risks.
type MemoryTracker struct {
	client *client.Client
}

// NewMemoryTracker creates a MemoryTracker.
func NewMemoryTracker(c *client.Client) *MemoryTracker {
	return &MemoryTracker{client: c}
}

// DetectSpillRisks inspects query syntax and query_log for memory overflow potential.
func (mt *MemoryTracker) DetectSpillRisks(ctx context.Context, query string) []models.SpillAlert {
	alerts := []models.SpillAlert{}
	upper := strings.ToUpper(query)

	if strings.Contains(upper, "GROUP BY") && !strings.Contains(upper, "MAX_BYTES_BEFORE_EXTERNAL_GROUP_BY") {
		alerts = append(alerts, models.SpillAlert{
			SpillType:          "EXTERNAL_GROUP_BY",
			TriggerMemoryBytes: 10 * 1024 * 1024 * 1024, // 10 GB default ceiling
			Advice:             "High-cardinality GROUP BY may exceed server memory limit. Set 'max_bytes_before_external_group_by = 20000000000' to enable two-level hash table disk spillover.",
		})
	}

	if strings.Contains(upper, "ORDER BY") && !strings.Contains(upper, "MAX_BYTES_BEFORE_EXTERNAL_SORT") {
		alerts = append(alerts, models.SpillAlert{
			SpillType:          "EXTERNAL_SORT",
			TriggerMemoryBytes: 5 * 1024 * 1024 * 1024, // 5 GB default ceiling
			Advice:             "Sorting broad result sets in memory risks node OOM. Set 'max_bytes_before_external_sort = 10000000000' to enable external merge sorting on disk.",
		})
	}

	return alerts
}
