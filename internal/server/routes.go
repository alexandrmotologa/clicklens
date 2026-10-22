package server

import (
	"net/http"
	"strings"
)

// SetupRoutes registers all REST and WebSocket routes with CORS handling.
func SetupRoutes(h *APIHandler) http.Handler {
	mux := http.NewServeMux()

	// REST Endpoints
	mux.HandleFunc("/api/health", h.HandleHealth)
	mux.HandleFunc("/api/connect", h.HandleConnect)
	mux.HandleFunc("/api/databases", h.HandleDatabases)
	mux.HandleFunc("/api/parts", h.HandleParts)
	mux.HandleFunc("/api/merges", h.HandleMerges)
	mux.HandleFunc("/api/mutations", h.HandleMutations)
	mux.HandleFunc("/api/columns", h.HandleColumns)
	mux.HandleFunc("/api/processes", h.HandleProcesses)
	mux.HandleFunc("/api/kill", h.HandleKill)
	mux.HandleFunc("/api/explain", h.HandleExplain)
	mux.HandleFunc("/api/query", h.HandleQuery)
	mux.HandleFunc("/api/local/inspect", h.HandleLocalInspect)
	mux.HandleFunc("/api/local/query", h.HandleLocalQuery)
	mux.HandleFunc("/api/local/status", h.HandleLocalRunnerStatus)

	// WebSocket Live Feed
	mux.HandleFunc("/api/ws", h.HandleWebSocket)

	// SPA Static Asset Serving
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		ServeStaticSPA(w, r)
	})

	// Wrap in CORS middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})
}
