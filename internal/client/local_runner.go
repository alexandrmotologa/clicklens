package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/models"
)

// LocalRunner wraps clickhouse-local CLI commands.
type LocalRunner struct {
	binaryPath string
	useWSL     bool
	available  bool
}

// NewLocalRunner discovers clickhouse-local in host PATH or WSL.
func NewLocalRunner() *LocalRunner {
	lr := &LocalRunner{}
	lr.discoverBinary()
	return lr
}

func (lr *LocalRunner) discoverBinary() {
	// 1. Check direct binary in PATH
	if p, err := exec.LookPath("clickhouse-local"); err == nil {
		lr.binaryPath = p
		lr.available = true
		return
	}
	if p, err := exec.LookPath("clickhouse"); err == nil {
		lr.binaryPath = p
		lr.available = true
		return
	}

	// 2. Check local ./bin directory
	candidatePaths := []string{
		filepath.Join(".", "bin", "clickhouse-local.exe"),
		filepath.Join(".", "bin", "clickhouse-local"),
		filepath.Join(".", "bin", "clickhouse.exe"),
		filepath.Join(".", "bin", "clickhouse"),
	}
	for _, cand := range candidatePaths {
		if _, err := exec.LookPath(cand); err == nil {
			lr.binaryPath = cand
			lr.available = true
			return
		}
	}

	// 3. On Windows, check if WSL is available and has clickhouse-local
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("wsl"); err == nil {
			cmd := exec.Command("wsl", "which", "clickhouse-local")
			if out, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
				lr.binaryPath = strings.TrimSpace(string(out))
				lr.useWSL = true
				lr.available = true
				return
			}
			cmd2 := exec.Command("wsl", "which", "clickhouse")
			if out, err := cmd2.Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
				lr.binaryPath = strings.TrimSpace(string(out))
				lr.useWSL = true
				lr.available = true
				return
			}
		}
	}

	lr.available = false
}

// IsAvailable reports whether clickhouse-local binary was found.
func (lr *LocalRunner) IsAvailable() bool {
	return lr.available
}

// BinaryPath returns discovered binary location.
func (lr *LocalRunner) BinaryPath() string {
	if lr.useWSL {
		return "wsl " + lr.binaryPath
	}
	return lr.binaryPath
}

// ExecuteQuery runs an ad-hoc query against a local dataset using clickhouse-local.
func (lr *LocalRunner) ExecuteQuery(ctx context.Context, query string) (*models.QueryResult, error) {
	if !lr.available {
		return nil, fmt.Errorf("clickhouse-local is not installed on this system. You can install it via 'curl https://clickhouse.com/ | sh' or use the built-in pure Go file inspector")
	}

	start := time.Now()
	formatQuery := query
	// Ensure JSON output format for parsing
	if !strings.Contains(strings.ToUpper(query), "FORMAT") {
		formatQuery += " FORMAT JSON"
	}

	var cmd *exec.Cmd
	if lr.useWSL {
		cmd = exec.CommandContext(ctx, "wsl", lr.binaryPath, "--query", formatQuery)
	} else if strings.HasSuffix(lr.binaryPath, "clickhouse") || strings.HasSuffix(lr.binaryPath, "clickhouse.exe") {
		cmd = exec.CommandContext(ctx, lr.binaryPath, "local", "--query", formatQuery)
	} else {
		cmd = exec.CommandContext(ctx, lr.binaryPath, "--query", formatQuery)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	elapsed := time.Since(start).Seconds() * 1000.0

	if err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = err.Error()
		}
		return &models.QueryResult{
			Query:           query,
			ExecutionTimeMs: elapsed,
			Error:           errText,
		}, fmt.Errorf("clickhouse-local execution failed: %s", errText)
	}

	// Parse JSON output from ClickHouse
	type chJSONResponse struct {
		Meta []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"meta"`
		Data []map[string]any `json:"data"`
		Rows int              `json:"rows"`
		Statistics struct {
			Elapsed   float64 `json:"elapsed"`
			RowsRead  uint64  `json:"rows_read"`
			BytesRead uint64  `json:"bytes_read"`
		} `json:"statistics"`
	}

	var parsed chJSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		// Output might be tabular text if format was overridden
		lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		return &models.QueryResult{
			Query:           query,
			Columns:         []string{"output"},
			Rows:            [][]any{{strings.Join(lines, "\n")}},
			RowsCount:       len(lines),
			ExecutionTimeMs: elapsed,
		}, nil
	}

	cols := make([]string, len(parsed.Meta))
	for i, m := range parsed.Meta {
		cols[i] = m.Name
	}

	matrix := make([][]any, len(parsed.Data))
	for i, rowMap := range parsed.Data {
		rowArr := make([]any, len(cols))
		for j, colName := range cols {
			rowArr[j] = rowMap[colName]
		}
		matrix[i] = rowArr
	}

	return &models.QueryResult{
		Query:           query,
		Columns:         cols,
		Rows:            matrix,
		RowsCount:       len(matrix),
		ExecutionTimeMs: elapsed,
		BytesRead:       parsed.Statistics.BytesRead,
		RowsRead:        parsed.Statistics.RowsRead,
	}, nil
}
