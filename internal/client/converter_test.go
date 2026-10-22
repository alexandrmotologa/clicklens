package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/clicklens/internal/models"
)

func TestFileConverter_CSVToParquet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clicklens_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	csvPath := filepath.Join(tmpDir, "test.csv")
	csvContent := "user_id,status,score,created_at\n1,active,98.5,2026-10-20\n2,pending,42.0,2026-10-21\n3,active,88.2,2026-10-22\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test csv: %v", err)
	}

	lr := NewLocalRunner()
	fc := NewFileConverter(lr)

	targetParquet := filepath.Join(tmpDir, "test.parquet")
	resp, err := fc.ConvertFile(context.Background(), models.FileConvertRequest{
		SourcePath:   csvPath,
		TargetPath:   targetParquet,
		TargetFormat: "Parquet",
		Compression:  "SNAPPY",
	})

	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	info, err := os.Stat(targetParquet)
	if err != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty output parquet file, err: %v", err)
	}
}
