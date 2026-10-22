package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/inspector"
	"github.com/alexandrmotologa/clicklens/internal/models"
	"github.com/alexandrmotologa/clicklens/internal/profiler"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local studio
	},
}

// APIHandler coordinates all REST endpoints and WebSocket live feeds.
type APIHandler struct {
	client         *client.Client
	localRunner    *client.LocalRunner
	fileInspector  *client.FileInspector
	fileConverter  *client.FileConverter
	partsInsp      *inspector.PartsInspector
	mergesInsp     *inspector.MergesInspector
	mutationsInsp  *inspector.MutationsInspector
	columnsInsp    *inspector.ColumnsInspector
	processesInsp  *inspector.ProcessesInspector
	queryLogInsp   *inspector.QueryLogInspector
	benchmarkInsp  *inspector.CodecBenchmarkInspector
	reportGen      *inspector.ReportGenerator
	explainParser  *profiler.ExplainParser
	memoryTracker  *profiler.MemoryTracker

	wsMu      sync.Mutex
	wsClients map[*websocket.Conn]bool
}

// NewAPIHandler initializes the API handler and background WebSocket broadcaster.
func NewAPIHandler(c *client.Client) *APIHandler {
	lr := client.NewLocalRunner()
	fi := client.NewFileInspector(lr)
	fc := client.NewFileConverter(lr)
	pi := inspector.NewPartsInspector(c)
	ci := inspector.NewColumnsInspector(c)

	h := &APIHandler{
		client:        c,
		localRunner:   lr,
		fileInspector: fi,
		fileConverter: fc,
		partsInsp:     pi,
		mergesInsp:    inspector.NewMergesInspector(c),
		mutationsInsp: inspector.NewMutationsInspector(c),
		columnsInsp:   ci,
		processesInsp: inspector.NewProcessesInspector(c),
		queryLogInsp:  inspector.NewQueryLogInspector(c),
		benchmarkInsp: inspector.NewCodecBenchmarkInspector(c),
		reportGen:     inspector.NewReportGenerator(c, pi, ci),
		explainParser: profiler.NewExplainParser(c),
		memoryTracker: profiler.NewMemoryTracker(c),
		wsClients:     make(map[*websocket.Conn]bool),
	}

	go h.startWebSocketBroadcaster()
	return h
}

func (h *APIHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *APIHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

// HandleHealth returns cluster connection status.
func (h *APIHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := h.client.GetStatus(ctx)
	h.writeJSON(w, http.StatusOK, status)
}

// HandleConnect updates the active ClickHouse connection DSN.
func (h *APIHandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DSN string `json:"dsn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.client.Connect(req.DSN); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Connection failed: %v", err))
		return
	}

	status := h.client.GetStatus(r.Context())
	h.writeJSON(w, http.StatusOK, status)
}

// HandleDatabases lists available databases and tables.
func (h *APIHandler) HandleDatabases(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if h.client.IsSimulated() {
		h.writeJSON(w, http.StatusOK, h.client.Simulator().GetDatabases())
		return
	}

	conn := h.client.Conn()
	if conn == nil {
		h.writeJSON(w, http.StatusOK, h.client.Simulator().GetDatabases())
		return
	}

	query := `
		SELECT database, name, engine
		FROM system.tables
		WHERE database NOT IN ('information_schema', 'INFORMATION_SCHEMA')
		ORDER BY database ASC, name ASC
	`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		h.writeJSON(w, http.StatusOK, h.client.Simulator().GetDatabases())
		return
	}
	defer rows.Close()

	dbMap := make(map[string]*models.DatabaseInfo)
	dbOrder := make([]string, 0)

	for rows.Next() {
		var dbName, tblName, engine string
		if err := rows.Scan(&dbName, &tblName, &engine); err != nil {
			continue
		}

		entry, exists := dbMap[dbName]
		if !exists {
			entry = &models.DatabaseInfo{
				Name:   dbName,
				Engine: engine,
				Tables: []string{},
			}
			dbMap[dbName] = entry
			dbOrder = append(dbOrder, dbName)
		}
		entry.Tables = append(entry.Tables, tblName)
		entry.TableCount = len(entry.Tables)
	}

	result := make([]models.DatabaseInfo, 0, len(dbOrder))
	for _, name := range dbOrder {
		result = append(result, *dbMap[name])
	}

	h.writeJSON(w, http.StatusOK, result)
}

// HandleParts returns parts inspection and partition health.
func (h *APIHandler) HandleParts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	db := r.URL.Query().Get("database")
	tbl := r.URL.Query().Get("table")

	res, err := h.partsInsp.GetTableParts(ctx, db, tbl)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, res)
}

// HandleMerges returns currently active background merges.
func (h *APIHandler) HandleMerges(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	merges, err := h.mergesInsp.GetActiveMerges(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, merges)
}

// HandleMutations returns table mutations.
func (h *APIHandler) HandleMutations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	db := r.URL.Query().Get("database")
	tbl := r.URL.Query().Get("table")

	mutations, err := h.mutationsInsp.GetMutations(ctx, db, tbl)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, mutations)
}

// HandleColumns returns column compression audit and recommendations.
func (h *APIHandler) HandleColumns(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	db := r.URL.Query().Get("database")
	tbl := r.URL.Query().Get("table")

	audit, err := h.columnsInsp.AuditColumns(ctx, db, tbl)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, audit)
}

// HandleProcesses returns active queries.
func (h *APIHandler) HandleProcesses(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	procs, err := h.processesInsp.GetActiveProcesses(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, procs)
}

// HandleKill terminates a running query.
func (h *APIHandler) HandleKill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueryID string `json:"query_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.QueryID == "" {
		h.writeError(w, http.StatusBadRequest, "Missing query_id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.processesInsp.KillQuery(ctx, req.QueryID); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "killed", "query_id": req.QueryID})
}

// HandleExplain executes EXPLAIN PIPELINE or EXPLAIN PLAN and returns parsed DAG.
func (h *APIHandler) HandleExplain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
		Mode  string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		h.writeError(w, http.StatusBadRequest, "Missing query in request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	analysis, err := h.explainParser.ParseQueryPipeline(ctx, req.Query, req.Mode)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check for potential memory spill alerts
	spills := h.memoryTracker.DetectSpillRisks(ctx, req.Query)
	analysis.SpillAlerts = append(analysis.SpillAlerts, spills...)

	h.writeJSON(w, http.StatusOK, analysis)
}

// HandleQuery executes a SQL query and returns rows for the Query Console.
func (h *APIHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		h.writeError(w, http.StatusBadRequest, "Missing query")
		return
	}

	start := time.Now()

	if h.client.IsSimulated() {
		// Provide mock results for the simulated demo console
		h.writeJSON(w, http.StatusOK, models.QueryResult{
			Query:           req.Query,
			Columns:         []string{"event_type", "count", "avg_duration_ms"},
			Rows:            [][]any{{"login", 142090, 14.2}, {"page_view", 890420, 8.4}, {"checkout", 28410, 48.9}},
			RowsCount:       3,
			ExecutionTimeMs: 12.4,
			RowsRead:        1060920,
			BytesRead:       38491024,
		})
		return
	}

	conn := h.client.Conn()
	if conn == nil {
		h.writeError(w, http.StatusInternalServerError, "ClickHouse is not connected")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	rows, err := conn.Query(ctx, req.Query)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Query error: %v", err))
		return
	}
	defer rows.Close()

	colTypes := rows.ColumnTypes()
	cols := make([]string, len(colTypes))
	for i, ct := range colTypes {
		cols[i] = ct.Name()
	}

	matrix := make([][]any, 0)
	for rows.Next() {
		vals := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range vals {
			pointers[i] = &vals[i]
		}
		if err := rows.Scan(pointers...); err == nil {
			matrix = append(matrix, vals)
		}
		if len(matrix) >= 500 {
			break // Safety limit for web studio
		}
	}

	elapsed := time.Since(start).Seconds() * 1000.0
	h.writeJSON(w, http.StatusOK, models.QueryResult{
		Query:           req.Query,
		Columns:         cols,
		Rows:            matrix,
		RowsCount:       len(matrix),
		ExecutionTimeMs: elapsed,
	})
}

// HandleLocalInspect inspects a local Parquet, CSV, or TSV file.
func (h *APIHandler) HandleLocalInspect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath string `json:"file_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FilePath == "" {
		h.writeError(w, http.StatusBadRequest, "Missing file_path")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	profile, err := h.fileInspector.InspectFile(ctx, req.FilePath)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, profile)
}

// HandleLocalQuery executes an ad-hoc query against a local file via clickhouse-local.
func (h *APIHandler) HandleLocalQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		h.writeError(w, http.StatusBadRequest, "Missing query")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := h.localRunner.ExecuteQuery(ctx, req.Query)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// HandleQueryLog retrieves historical queries from system.query_log.
func (h *APIHandler) HandleQueryLog(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		var parsed int
		if _, err := fmt.Sscanf(l, "%d", &parsed); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.queryLogInsp.GetQueryLog(ctx, filter, limit)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, res)
}

// HandleBenchmarkCodec executes simulated or live compression benchmarks for a column.
func (h *APIHandler) HandleBenchmarkCodec(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Database string `json:"database"`
		Table    string `json:"table"`
		Column   string `json:"column"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Column == "" {
		h.writeError(w, http.StatusBadRequest, "Missing column in request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	res, err := h.benchmarkInsp.BenchmarkColumn(ctx, req.Database, req.Table, req.Column)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, res)
}

// HandleConvert converts tabular files (CSV/TSV) into Parquet format.
func (h *APIHandler) HandleConvert(w http.ResponseWriter, r *http.Request) {
	var req models.FileConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SourcePath == "" {
		h.writeError(w, http.StatusBadRequest, "Missing source_path in request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	res, err := h.fileConverter.ConvertFile(ctx, req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, res)
}

// HandleReport generates a diagnostic audit report in struct and markdown format.
func (h *APIHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("database")
	tbl := r.URL.Query().Get("table")

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	report, err := h.reportGen.GenerateReport(ctx, db, tbl)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, report)
}

// HandleLocalRunnerStatus returns clickhouse-local availability.
func (h *APIHandler) HandleLocalRunnerStatus(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{
		"available":   h.localRunner.IsAvailable(),
		"binary_path": h.localRunner.BinaryPath(),
	})
}

// HandleWebSocket upgrades connection and registers client for live telemetry.
func (h *APIHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.wsMu.Lock()
	h.wsClients[conn] = true
	h.wsMu.Unlock()

	defer func() {
		h.wsMu.Lock()
		delete(h.wsClients, conn)
		h.wsMu.Unlock()
		_ = conn.Close()
	}()

	// Keep alive reader
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *APIHandler) startWebSocketBroadcaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		h.wsMu.Lock()
		clientCount := len(h.wsClients)
		h.wsMu.Unlock()

		if clientCount == 0 {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		merges, err := h.mergesInsp.GetActiveMerges(ctx)
		cancel()

		if err == nil {
			msg := models.TelemetryWebSocketMessage{
				Type:      "merges_update",
				Timestamp: time.Now(),
				Data:      merges,
			}
			data, err := json.Marshal(msg)
			if err == nil {
				h.broadcastWebSocket(data)
			}
		}
	}
}

func (h *APIHandler) broadcastWebSocket(message []byte) {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()

	for conn := range h.wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			_ = conn.Close()
			delete(h.wsClients, conn)
		}
	}
}
