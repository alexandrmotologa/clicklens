package client

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/models"
	"github.com/parquet-go/parquet-go"
)

// FileInspector extracts schema and previews local analytical files.
type FileInspector struct {
	localRunner *LocalRunner
}

// NewFileInspector creates a new FileInspector.
func NewFileInspector(lr *LocalRunner) *FileInspector {
	return &FileInspector{
		localRunner: lr,
	}
}

// InspectFile inspects a local Parquet, CSV, TSV, or JSON file.
func (fi *FileInspector) InspectFile(ctx context.Context, path string) (*models.LocalFileProfile, error) {
	fiInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	fileName := filepath.Base(path)

	// If clickhouse-local is available, try describing via clickhouse-local first
	if fi.localRunner != nil && fi.localRunner.IsAvailable() {
		format := "Parquet"
		if ext == ".csv" {
			format = "CSVWithNames"
		} else if ext == ".tsv" {
			format = "TSVWithNames"
		} else if ext == ".json" || ext == ".ndjson" {
			format = "JSONEachRow"
		}

		descQuery := fmt.Sprintf("DESCRIBE TABLE file('%s', '%s')", filepath.ToSlash(path), format)
		res, err := fi.localRunner.ExecuteQuery(ctx, descQuery)
		if err == nil && len(res.Rows) > 0 {
			columns := make([]models.LocalColumnMetadata, 0, len(res.Rows))
			for _, r := range res.Rows {
				if len(r) >= 2 {
					name, _ := r[0].(string)
					tp, _ := r[1].(string)
					columns = append(columns, models.LocalColumnMetadata{
						Name:     name,
						Type:     tp,
						Nullable: strings.HasPrefix(tp, "Nullable"),
					})
				}
			}

			// Fetch count and sample
			countQuery := fmt.Sprintf("SELECT count() AS cnt FROM file('%s', '%s')", filepath.ToSlash(path), format)
			countRes, _ := fi.localRunner.ExecuteQuery(ctx, countQuery)
			var rowCount int64 = 0
			if countRes != nil && len(countRes.Rows) > 0 && len(countRes.Rows[0]) > 0 {
				if cntFloat, ok := countRes.Rows[0][0].(float64); ok {
					rowCount = int64(cntFloat)
				}
			}

			previewQuery := fmt.Sprintf("SELECT * FROM file('%s', '%s') LIMIT 15", filepath.ToSlash(path), format)
			previewRes, _ := fi.localRunner.ExecuteQuery(ctx, previewQuery)

			sampleRows := make([]map[string]any, 0)
			if previewRes != nil {
				for _, row := range previewRes.Rows {
					item := make(map[string]any)
					for i, col := range previewRes.Columns {
						if i < len(row) {
							item[col] = row[i]
						}
					}
					sampleRows = append(sampleRows, item)
				}
			}

			return &models.LocalFileProfile{
				FilePath:      path,
				FileName:      fileName,
				Format:        format,
				FileSizeBytes: fiInfo.Size(),
				RowCount:      rowCount,
				Columns:       columns,
				SampleRows:    sampleRows,
				RunnerUsed:    "clickhouse-local",
			}, nil
		}
	}

	// Fallback to pure Go engine for Parquet and CSV
	if ext == ".parquet" {
		return fi.inspectParquetPureGo(path, fiInfo.Size())
	} else if ext == ".csv" || ext == ".tsv" {
		return fi.inspectCSVPureGo(path, ext == ".tsv", fiInfo.Size())
	}

	return nil, fmt.Errorf("unsupported file format %s (supported: .parquet, .csv, .tsv)", ext)
}

func (fi *FileInspector) inspectParquetPureGo(path string, size int64) (*models.LocalFileProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open parquet file: %w", err)
	}
	defer file.Close()

	pfile, err := parquet.OpenFile(file, size)
	if err != nil {
		return nil, fmt.Errorf("unable to parse parquet metadata: %w", err)
	}

	schema := pfile.Schema()
	fields := schema.Fields()
	columns := make([]models.LocalColumnMetadata, 0, len(fields))

	for _, f := range fields {
		node := f
		typeName := "Unknown"
		if node.Type() != nil {
			typeName = node.Type().String()
		}
		columns = append(columns, models.LocalColumnMetadata{
			Name:     f.Name(),
			Type:     typeName,
			Nullable: node.Optional(),
		})
	}

	rowCount := pfile.NumRows()

	// Read sample rows
	sampleRows := make([]map[string]any, 0)
	reader := parquet.NewReader(file)
	defer reader.Close()

	// Read up to 10 rows
	for i := 0; i < 10; i++ {
		rowBuf := make([]parquet.Row, 1)
		n, err := reader.ReadRows(rowBuf)
		if err != nil || n == 0 {
			break
		}
		rowMap := make(map[string]any)
		for colIdx, col := range columns {
			if colIdx < len(rowBuf[0]) {
				val := rowBuf[0][colIdx]
				rowMap[col.Name] = val.String()
			}
		}
		sampleRows = append(sampleRows, rowMap)
	}

	return &models.LocalFileProfile{
		FilePath:      path,
		FileName:      filepath.Base(path),
		Format:        "Parquet",
		FileSizeBytes: size,
		RowCount:      rowCount,
		Columns:       columns,
		SampleRows:    sampleRows,
		RunnerUsed:    "pure-go-engine (parquet-go)",
	}, nil
}

func (fi *FileInspector) inspectCSVPureGo(path string, isTSV bool, size int64) (*models.LocalFileProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	if isTSV {
		reader.Comma = '\t'
	}
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	sampleRows := make([]map[string]any, 0)
	rawSamples := make([][]string, 0)
	var rowCount int64 = 0

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowCount++
		if len(rawSamples) < 15 {
			rawSamples = append(rawSamples, rec)
			item := make(map[string]any)
			for i, h := range headers {
				if i < len(rec) {
					item[h] = rec[i]
				}
			}
			sampleRows = append(sampleRows, item)
		}
	}

	// Infer types from raw sample values
	columns := make([]models.LocalColumnMetadata, len(headers))
	for colIdx, h := range headers {
		inferred := inferColumnType(colIdx, rawSamples)
		columns[colIdx] = models.LocalColumnMetadata{
			Name:     h,
			Type:     inferred,
			Nullable: true,
		}
	}

	fmtName := "CSV"
	if isTSV {
		fmtName = "TSV"
	}

	return &models.LocalFileProfile{
		FilePath:      path,
		FileName:      filepath.Base(path),
		Format:        fmtName,
		FileSizeBytes: size,
		RowCount:      rowCount,
		Columns:       columns,
		SampleRows:    sampleRows,
		RunnerUsed:    "pure-go-engine (csv-sniffer)",
	}, nil
}

func inferColumnType(colIdx int, samples [][]string) string {
	if len(samples) == 0 {
		return "String"
	}

	isInt := true
	isFloat := true
	isDateTime := true

	for _, s := range samples {
		if colIdx >= len(s) {
			continue
		}
		val := strings.TrimSpace(s[colIdx])
		if val == "" || strings.EqualFold(val, "null") {
			continue
		}

		if _, err := strconv.ParseInt(val, 10, 64); err != nil {
			isInt = false
		}
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			isFloat = false
		}
		if _, err := time.Parse(time.RFC3339, val); err != nil {
			if _, err2 := time.Parse("2006-01-02 15:04:05", val); err2 != nil {
				if _, err3 := time.Parse("2006-01-02", val); err3 != nil {
					isDateTime = false
				}
			}
		}
	}

	if isInt {
		return "Int64"
	}
	if isFloat {
		return "Float64"
	}
	if isDateTime {
		return "DateTime"
	}
	return "String"
}
