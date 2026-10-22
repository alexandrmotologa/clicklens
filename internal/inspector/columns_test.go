package inspector

import (
	"context"
	"strings"
	"testing"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

func TestColumnsInspector_Recommendations(t *testing.T) {
	cli, _ := client.NewClient(models.ConnectionConfig{IsDemo: true})
	insp := NewColumnsInspector(cli)

	res, err := insp.AuditColumns(context.Background(), "telemetry", "events")
	if err != nil {
		t.Fatalf("AuditColumns failed: %v", err)
	}

	if len(res.Columns) == 0 {
		t.Fatalf("expected columns in audit, got 0")
	}

	for _, col := range res.Columns {
		if strings.Contains(strings.ToUpper(col.Type), "DATETIME") {
			if !strings.Contains(col.RecommendedCodec, "DoubleDelta") {
				t.Errorf("expected DoubleDelta recommendation for timestamp, got: %s", col.RecommendedCodec)
			}
		}
		if strings.Contains(strings.ToUpper(col.Type), "FLOAT") {
			if !strings.Contains(col.RecommendedCodec, "Gorilla") {
				t.Errorf("expected Gorilla recommendation for float, got: %s", col.RecommendedCodec)
			}
		}
	}
}

func TestRecommendCodecHeuristics(t *testing.T) {
	// Test datetime
	rec, savings, _ := recommendCodec("created_at", "DateTime64(3)", "Default(LZ4)")
	if !strings.Contains(rec, "DoubleDelta") || savings < 50 {
		t.Errorf("unexpected recommendation for DateTime: %s (savings: %.1f)", rec, savings)
	}

	// Test float
	rec, savings, _ = recommendCodec("temperature", "Float64", "Default(LZ4)")
	if !strings.Contains(rec, "Gorilla") || savings < 50 {
		t.Errorf("unexpected recommendation for Float64: %s (savings: %.1f)", rec, savings)
	}

	// Test categorical string
	rec, savings, _ = recommendCodec("event_type", "String", "Default(LZ4)")
	if !strings.Contains(rec, "LowCardinality") {
		t.Errorf("unexpected recommendation for categorical string: %s", rec)
	}
}
