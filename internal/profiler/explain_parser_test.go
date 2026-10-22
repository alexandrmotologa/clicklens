package profiler

import (
	"context"
	"testing"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

func TestExplainParser_Simulated(t *testing.T) {
	cli, err := client.NewClient(models.ConnectionConfig{IsDemo: true})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	parser := NewExplainParser(cli)
	analysis, err := parser.ParseQueryPipeline(context.Background(), "SELECT count() FROM telemetry.events", "PIPELINE")
	if err != nil {
		t.Fatalf("ParseQueryPipeline failed: %v", err)
	}

	if len(analysis.Nodes) == 0 {
		t.Errorf("expected non-empty DAG nodes, got 0")
	}

	if analysis.MaxConcurrency < 1 {
		t.Errorf("expected max concurrency >= 1, got %d", analysis.MaxConcurrency)
	}

	// Verify fan-in bottleneck detected
	foundBottleneck := false
	for _, b := range analysis.Bottlenecks {
		if b.CurrentStreams < b.PreviousStreams {
			foundBottleneck = true
			break
		}
	}

	if !foundBottleneck {
		t.Errorf("expected stream fan-in bottleneck to be detected")
	}
}

func TestExplainParser_BuildDAGFromExplainLines(t *testing.T) {
	cli, _ := client.NewClient(models.ConnectionConfig{IsDemo: true})
	parser := NewExplainParser(cli)

	sampleLines := []string{
		"(Expression)",
		"ExpressionTransform × 1",
		"  (Sorting)",
		"  MergingSortedTransform 16 → 1",
		"    (Aggregating)",
		"    AggregatingTransform × 16",
		"      (ReadFromMergeTree)",
		"      MergeTreeSelect(pool: 16) × 16",
	}

	analysis := parser.buildDAGFromExplainLines("SELECT 1", "PIPELINE", "raw", sampleLines)

	if len(analysis.Nodes) != 4 {
		t.Fatalf("expected 4 DAG nodes, got %d", len(analysis.Nodes))
	}

	if analysis.MaxConcurrency != 16 {
		t.Errorf("expected max concurrency 16, got %d", analysis.MaxConcurrency)
	}

	if len(analysis.Bottlenecks) == 0 {
		t.Errorf("expected fan-in bottleneck on MergingSortedTransform 16 -> 1")
	}
}
