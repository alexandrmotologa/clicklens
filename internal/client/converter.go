package client

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/clicklens/internal/models"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress"
	"github.com/parquet-go/parquet-go/compress/snappy"
	"github.com/parquet-go/parquet-go/compress/zstd"
)

// FileConverter handles file format conversions between CSV, TSV, and Parquet.
type FileConverter struct {
	localRunner *LocalRunner
}

// NewFileConverter creates a new FileConverter.
func NewFileConverter(lr *LocalRunner) *FileConverter {
	return &FileConverter{localRunner: lr}
}

// ConvertFile converts a tabular file (CSV/TSV) into Parquet format.
func (fc *FileConverter) ConvertFile(ctx context.Context, req models.FileConvertRequest) (*models.FileConvertResponse, error) {
	srcInfo, err := os.Stat(req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("source file not found: %w", err)
	}

	targetPath := req.TargetPath
	if targetPath == "" {
		ext := filepath.Ext(req.SourcePath)
		targetPath = strings.TrimSuffix(req.SourcePath, ext) + ".parquet"
	}

	// Make sure target directory exists
	if dir := filepath.Dir(targetPath); dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// If clickhouse-local is available, use high-speed native converter
	if fc.localRunner != nil && fc.localRunner.IsAvailable() {
		srcFormat := "CSVWithNames"
		if strings.HasSuffix(strings.ToLower(req.SourcePath), ".tsv") {
			srcFormat = "TSVWithNames"
		}

		query := fmt.Sprintf("SELECT * FROM file('%s', '%s') INTO OUTFILE '%s' FORMAT Parquet",
			filepath.ToSlash(req.SourcePath), srcFormat, filepath.ToSlash(targetPath))

		_, err := fc.localRunner.ExecuteQuery(ctx, query)
		if err == nil {
			targetInfo, statErr := os.Stat(targetPath)
			if statErr == nil {
				ratio := float64(srcInfo.Size()) / float64(targetInfo.Size())
				return &models.FileConvertResponse{
					Success:          true,
					SourceSize:       srcInfo.Size(),
					TargetSize:       targetInfo.Size(),
					CompressionRatio: ratio,
					TargetPath:       targetPath,
					Message:          fmt.Sprintf("Successfully converted via clickhouse-local into Parquet (%.2fx compression)", ratio),
				}, nil
			}
		}
	}

	// Fallback: Pure Go CSV/TSV to Parquet converter
	srcFile, err := os.Open(req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	reader := csv.NewReader(srcFile)
	if strings.HasSuffix(strings.ToLower(req.SourcePath), ".tsv") {
		reader.Comma = '\t'
	}

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read table header: %w", err)
	}

	// Build dynamic Parquet schema with UTF8 String columns
	parquetGroup := parquet.Group{}
	for i, colName := range header {
		cleanCol := strings.TrimSpace(colName)
		if cleanCol == "" {
			cleanCol = fmt.Sprintf("col_%d", i)
		}
		parquetGroup[cleanCol] = parquet.String()
	}

	schema := parquet.NewSchema("dynamic_record", parquetGroup)

	outFile, err := os.Create(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer outFile.Close()

	var codec compress.Codec = &snappy.Codec{}
	if strings.ToUpper(req.Compression) == "ZSTD" {
		codec = &zstd.Codec{}
	}

	writer := parquet.NewGenericWriter[any](outFile, schema, parquet.Compression(codec))

	// Write rows as map[string]any records
	rowCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		rowMap := make(map[string]any)
		for i, val := range record {
			if i < len(header) {
				cleanCol := strings.TrimSpace(header[i])
				if cleanCol == "" {
					cleanCol = fmt.Sprintf("col_%d", i)
				}
				rowMap[cleanCol] = val
			}
		}

		if _, err := writer.Write([]any{rowMap}); err == nil {
			rowCount++
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat converted target: %w", err)
	}

	ratio := 1.0
	if targetInfo.Size() > 0 {
		ratio = float64(srcInfo.Size()) / float64(targetInfo.Size())
	}

	return &models.FileConvertResponse{
		Success:          true,
		SourceSize:       srcInfo.Size(),
		TargetSize:       targetInfo.Size(),
		CompressionRatio: ratio,
		TargetPath:       targetPath,
		Message:          fmt.Sprintf("Converted %d rows to Parquet using native Go engine (%.2fx compression)", rowCount, ratio),
	}, nil
}
