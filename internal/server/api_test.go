package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

func TestAPIEndpoints(t *testing.T) {
	cli, err := client.NewClient(models.ConnectionConfig{IsDemo: true})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handler := NewAPIHandler(cli)
	routes := SetupRoutes(handler)

	// 1. Test /api/health
	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/health, got %d", rr.Code)
	}

	var status models.ServerStatus
	if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode /api/health response: %v", err)
	}
	if !status.Connected || !status.IsDemo {
		t.Errorf("expected connected demo status, got connected=%v, isDemo=%v", status.Connected, status.IsDemo)
	}

	// 2. Test /api/databases
	req = httptest.NewRequest("GET", "/api/databases", nil)
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/databases, got %d", rr.Code)
	}

	var dbs []models.DatabaseInfo
	if err := json.NewDecoder(rr.Body).Decode(&dbs); err != nil || len(dbs) == 0 {
		t.Fatalf("expected database list from /api/databases, got %v", dbs)
	}

	// 3. Test /api/parts
	req = httptest.NewRequest("GET", "/api/parts?database=telemetry&table=events", nil)
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/parts, got %d", rr.Code)
	}

	var partsResp models.TablePartsResponse
	if err := json.NewDecoder(rr.Body).Decode(&partsResp); err != nil {
		t.Fatalf("failed to decode /api/parts response: %v", err)
	}
	if partsResp.TotalActiveParts == 0 || len(partsResp.Partitions) == 0 {
		t.Errorf("expected non-empty active parts and partitions, got %d parts", partsResp.TotalActiveParts)
	}

	// 4. Test /api/merges
	req = httptest.NewRequest("GET", "/api/merges", nil)
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/merges, got %d", rr.Code)
	}

	var merges []models.MergeInfo
	if err := json.NewDecoder(rr.Body).Decode(&merges); err != nil {
		t.Fatalf("failed to decode /api/merges: %v", err)
	}

	// 5. Test /api/explain
	explainBody := `{"query": "SELECT count() FROM telemetry.events", "mode": "PIPELINE"}`
	req = httptest.NewRequest("POST", "/api/explain", strings.NewReader(explainBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/explain, got %d", rr.Code)
	}

	var analysis models.PipelineAnalysis
	if err := json.NewDecoder(rr.Body).Decode(&analysis); err != nil || len(analysis.Nodes) == 0 {
		t.Fatalf("failed to decode /api/explain DAG: %v", err)
	}

	// 6. Test /api/query-log
	req = httptest.NewRequest("GET", "/api/query-log?filter=slow&limit=5", nil)
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/query-log, got %d", rr.Code)
	}
	var qLog models.QueryLogResponse
	if err := json.NewDecoder(rr.Body).Decode(&qLog); err != nil || len(qLog.Entries) == 0 {
		t.Fatalf("failed to decode /api/query-log: %v", err)
	}

	// 7. Test /api/benchmark-codec
	benchBody := `{"database": "telemetry", "table": "events", "column": "event_time"}`
	req = httptest.NewRequest("POST", "/api/benchmark-codec", strings.NewReader(benchBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/benchmark-codec, got %d", rr.Code)
	}
	var bResp models.CodecBenchmarkResponse
	if err := json.NewDecoder(rr.Body).Decode(&bResp); err != nil || len(bResp.Results) == 0 {
		t.Fatalf("failed to decode /api/benchmark-codec: %v", err)
	}

	// 8. Test /api/report
	req = httptest.NewRequest("GET", "/api/report?database=telemetry&table=events", nil)
	rr = httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/report, got %d", rr.Code)
	}
	var rep models.HealthAuditReport
	if err := json.NewDecoder(rr.Body).Decode(&rep); err != nil || len(rep.MarkdownContent) == 0 {
		t.Fatalf("failed to decode /api/report: %v", err)
	}
}
