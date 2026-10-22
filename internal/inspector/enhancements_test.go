package inspector

import (
	"context"
	"testing"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

func TestQueryLogInspector(t *testing.T) {
	c, err := client.NewClient(models.ConnectionConfig{IsDemo: true})
	if err != nil {
		t.Fatalf("failed to init client: %v", err)
	}
	qi := NewQueryLogInspector(c)

	ctx := context.Background()

	// Test slow queries filter
	res, err := qi.GetQueryLog(ctx, "slow", 10)
	if err != nil {
		t.Fatalf("unexpected error fetching query log: %v", err)
	}
	if len(res.Entries) == 0 {
		t.Fatalf("expected non-empty query log entries")
	}

	// Test memory filter
	memRes, err := qi.GetQueryLog(ctx, "memory", 10)
	if err != nil {
		t.Fatalf("unexpected error fetching memory query log: %v", err)
	}
	for _, entry := range memRes.Entries {
		if entry.MemoryUsage <= 1024*1024*1024 {
			t.Errorf("expected memory usage > 1GB, got %d", entry.MemoryUsage)
		}
	}

	// Test spill filter
	spillRes, err := qi.GetQueryLog(ctx, "spill", 10)
	if err != nil {
		t.Fatalf("unexpected error fetching spill query log: %v", err)
	}
	for _, entry := range spillRes.Entries {
		if !entry.HasSpill {
			t.Errorf("expected entry to have spill flag true")
		}
	}

	// Test errors filter
	errRes, err := qi.GetQueryLog(ctx, "errors", 10)
	if err != nil {
		t.Fatalf("unexpected error fetching errors query log: %v", err)
	}
	for _, entry := range errRes.Entries {
		if entry.ExceptionCode == 0 && entry.Type != "ExceptionWhileProcessing" {
			t.Errorf("expected exception in error log entry")
		}
	}
}

func TestCodecBenchmarkInspector(t *testing.T) {
	c, err := client.NewClient(models.ConnectionConfig{IsDemo: true})
	if err != nil {
		t.Fatalf("failed to init client: %v", err)
	}
	bi := NewCodecBenchmarkInspector(c)

	ctx := context.Background()
	res, err := bi.BenchmarkColumn(ctx, "telemetry", "events", "event_time")
	if err != nil {
		t.Fatalf("benchmark failed: %v", err)
	}

	if len(res.Results) == 0 {
		t.Fatalf("expected benchmark results")
	}

	foundDoubleDelta := false
	for _, item := range res.Results {
		if item.Codec == "DoubleDelta, ZSTD(1)" {
			foundDoubleDelta = true
			if item.Ratio <= 1.0 {
				t.Errorf("expected ratio > 1.0 for DoubleDelta, got %.2f", item.Ratio)
			}
		}
	}
	if !foundDoubleDelta {
		t.Errorf("expected DoubleDelta in benchmark results")
	}
}

func TestReportGenerator(t *testing.T) {
	c, err := client.NewClient(models.ConnectionConfig{IsDemo: true})
	if err != nil {
		t.Fatalf("failed to init client: %v", err)
	}
	pi := NewPartsInspector(c)
	ci := NewColumnsInspector(c)
	rg := NewReportGenerator(c, pi, ci)

	ctx := context.Background()
	report, err := rg.GenerateReport(ctx, "telemetry", "events")
	if err != nil {
		t.Fatalf("report generation failed: %v", err)
	}

	if report.OverallHealth == "" {
		t.Errorf("expected health rating to be populated")
	}
	if len(report.MarkdownContent) == 0 {
		t.Errorf("expected markdown content in diagnostic report")
	}
}
