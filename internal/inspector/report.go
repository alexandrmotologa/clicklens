package inspector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// ReportGenerator compiles system telemetry into health audit reports.
type ReportGenerator struct {
	client      *client.Client
	partsInsp   *PartsInspector
	columnsInsp *ColumnsInspector
}

// NewReportGenerator creates a new ReportGenerator.
func NewReportGenerator(c *client.Client, pi *PartsInspector, ci *ColumnsInspector) *ReportGenerator {
	return &ReportGenerator{
		client:      c,
		partsInsp:   pi,
		columnsInsp: ci,
	}
}

// GenerateReport compiles a comprehensive diagnostic audit report in struct and markdown format.
func (rg *ReportGenerator) GenerateReport(ctx context.Context, database, table string) (*models.HealthAuditReport, error) {
	if database == "" {
		database = "telemetry"
	}
	if table == "" {
		table = "events"
	}

	status := rg.client.GetStatus(ctx)
	partsRes, err := rg.partsInsp.GetTableParts(ctx, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect parts: %w", err)
	}

	columnsRes, err := rg.columnsInsp.AuditColumns(ctx, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect columns: %w", err)
	}

	bottlenecks := make([]string, 0)
	recommendations := make([]string, 0)

	// Evaluate parts fragmentation
	if partsRes.MaxPartCount > 80 {
		bottlenecks = append(bottlenecks, fmt.Sprintf("Partition %s contains %d active parts, exceeding recommended threshold of 80.", partsRes.MaxPartsPartition, partsRes.MaxPartCount))
		recommendations = append(recommendations, fmt.Sprintf("Run 'OPTIMIZE TABLE %s.%s PARTITION '%s' FINAL' during low load.", database, table, partsRes.MaxPartsPartition))
	}

	// Evaluate micro-batching
	if partsRes.MicroBatchingWarning {
		bottlenecks = append(bottlenecks, partsRes.MicroBatchingDetails)
		recommendations = append(recommendations, "Implement client-side insert batching (min 10k rows or 5s windows) or use Buffer engine.")
	}

	// Evaluate column codecs
	if columnsRes.PotentialSavings > 0 {
		mbSavings := float64(columnsRes.PotentialSavings) / (1024 * 1024)
		bottlenecks = append(bottlenecks, fmt.Sprintf("Identified %.1f MB of reclaimable storage from sub-optimal column codecs.", mbSavings))
		recommendations = append(recommendations, "Execute the zero-downtime codec migration playbook generated in ClickLens Studio.")
	}

	// Format readable markdown report
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# ClickLens Cluster Diagnostic Report\n\n"))
	md.WriteString(fmt.Sprintf("- **Generated At**: %s\n", time.Now().UTC().Format(time.RFC3339)))
	md.WriteString(fmt.Sprintf("- **Target**: `%s.%s`\n", database, table))
	md.WriteString(fmt.Sprintf("- **Engine**: %s\n", status.Version))
	md.WriteString(fmt.Sprintf("- **Overall Health Rating**: **%s**\n\n", partsRes.HealthRating))

	md.WriteString("## 1. Storage & Partition Telemetry\n\n")
	md.WriteString("| Metric | Value |\n")
	md.WriteString("| :--- | :--- |\n")
	md.WriteString(fmt.Sprintf("| Total Active Parts | %d |\n", partsRes.TotalActiveParts))
	md.WriteString(fmt.Sprintf("| Total Inactive Parts | %d |\n", partsRes.TotalInactiveParts))
	md.WriteString(fmt.Sprintf("| Partition Count | %d |\n", len(partsRes.Partitions)))
	md.WriteString(fmt.Sprintf("| Peak Partition Parts | %d (`%s`) |\n", partsRes.MaxPartCount, partsRes.MaxPartsPartition))
	md.WriteString(fmt.Sprintf("| Compressed Disk Size | %.2f MB |\n", float64(partsRes.TotalBytesOnDisk)/(1024*1024)))
	md.WriteString(fmt.Sprintf("| Uncompressed Data Size | %.2f MB |\n", float64(partsRes.TotalUncompressed)/(1024*1024)))
	md.WriteString(fmt.Sprintf("| Overall Compression Ratio | %.2f:1 |\n\n", partsRes.OverallRatio))

	md.WriteString("## 2. Ingestion & Part Age Profile\n\n")
	md.WriteString("| Age Window | Active Parts | Total Rows | Total Size |\n")
	md.WriteString("| :--- | :--- | :--- | :--- |\n")
	for _, b := range partsRes.PartAgeBuckets {
		md.WriteString(fmt.Sprintf("| %s | %d | %d | %.2f MB |\n", b.Label, b.PartCount, b.TotalRows, float64(b.TotalBytes)/(1024*1024)))
	}
	md.WriteString("\n")

	md.WriteString("## 3. Storage Optimization Opportunities\n\n")
	md.WriteString("| Column | Type | Current Codec | Recommended Codec | Est. Savings |\n")
	md.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
	for _, col := range columnsRes.Columns {
		if col.SavingsEstimatePercent > 0 {
			md.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %.1f%% (%.1f MB) |\n",
				col.Column, col.Type, col.CompressionCodec, col.RecommendedCodec, col.SavingsEstimatePercent, float64(col.ProjectedSavingsBytes)/(1024*1024)))
		}
	}
	md.WriteString(fmt.Sprintf("\n**Total Projected Storage Savings**: %.2f MB\n\n", float64(columnsRes.PotentialSavings)/(1024*1024)))

	md.WriteString("## 4. Key Recommendations\n\n")
	for i, rec := range recommendations {
		md.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	return &models.HealthAuditReport{
		GeneratedAt:        time.Now().UTC(),
		Database:           database,
		Table:              table,
		ServerVersion:      status.Version,
		OverallHealth:      partsRes.HealthRating,
		ActiveParts:        partsRes.TotalActiveParts,
		PartitionCount:     len(partsRes.Partitions),
		MaxPartsPartition:  partsRes.MaxPartsPartition,
		MaxPartCount:       partsRes.MaxPartCount,
		DiskBytes:          partsRes.TotalBytesOnDisk,
		UncompressedBytes:  partsRes.TotalUncompressed,
		OverallRatio:       partsRes.OverallRatio,
		PotentialSavings:   columnsRes.PotentialSavings,
		MicroBatchingRisk:  partsRes.MicroBatchingWarning,
		Bottlenecks:        bottlenecks,
		TopRecommendations: recommendations,
		MarkdownContent:    md.String(),
	}, nil
}
