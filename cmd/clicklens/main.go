package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/inspector"
	"github.com/alexandrmotologa/clicklens/internal/models"
	"github.com/alexandrmotologa/clicklens/internal/profiler"
	"github.com/alexandrmotologa/clicklens/internal/server"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		runServe([]string{})
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "serve":
		runServe(args)
	case "parts":
		runParts(args)
	case "merges":
		runMerges(args)
	case "audit":
		runAudit(args)
	case "explain":
		runExplain(args)
	case "profile":
		runProfile(args)
	case "version", "-v", "--version":
		fmt.Printf("ClickLens version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		// If first argument starts with a flag like --addr, assume 'serve'
		if strings.HasPrefix(command, "-") {
			runServe(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ClickLens - ClickHouse Observability Studio & Diagnostics Engine

Usage:
  clicklens [command] [options]

Commands:
  serve              Launch the web studio and telemetry server (default)
  parts <table>      Inspect MergeTree active/inactive parts and partition health
  merges             Monitor active background merges
  audit <table>      Audit column compression ratios and recommended codecs
  explain "<query>"  Profile query pipeline execution stream DAG
  profile <file>     Inspect local Parquet or CSV file schema and preview
  version            Print ClickLens version

Options for serve:
  --addr string      HTTP listen address (default ":8080")
  --dsn string       ClickHouse connection DSN (default "demo")
  --demo             Run in standalone demo simulation mode`)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "HTTP listen address")
	dsn := fs.String("dsn", "demo", "ClickHouse DSN")
	demo := fs.Bool("demo", false, "Force demo simulation mode")
	_ = fs.Parse(args)

	cfg := models.ConnectionConfig{
		DSN:    *dsn,
		IsDemo: *demo,
	}
	if *dsn == "" || *dsn == "demo" {
		cfg.IsDemo = true
	}

	cli, err := client.NewClient(cfg)
	if err != nil {
		fmt.Printf("Warning: connecting to ClickHouse (%s) failed: %v. Initializing with simulation engine.\n", *dsn, err)
	}

	handler := server.NewAPIHandler(cli)
	httpHandler := server.SetupRoutes(handler)

	fmt.Println("================================================================")
	fmt.Printf("  ClickLens Observability Studio v%s\n", version)
	fmt.Println("================================================================")
	if cli.IsSimulated() {
		fmt.Println("  Mode:       DEMO SIMULATION (all telemetry emulated in-memory)")
	} else {
		fmt.Printf("  Mode:       LIVE CLICKHOUSE (%s)\n", *dsn)
	}
	fmt.Printf("  Dashboard:  http://localhost%s\n", *addr)
	fmt.Printf("  REST API:   http://localhost%s/api/health\n", *addr)
	fmt.Printf("  WebSocket:  ws://localhost%s/api/ws\n", *addr)
	fmt.Println("================================================================")
	fmt.Println("Press Ctrl+C to stop.")

	srv := &http.Server{
		Addr:         *addr,
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}

func runParts(args []string) {
	fs := flag.NewFlagSet("parts", flag.ExitOnError)
	dsn := fs.String("dsn", "demo", "ClickHouse DSN")
	_ = fs.Parse(args)

	targetTable := "telemetry.events"
	if len(fs.Args()) > 0 {
		targetTable = fs.Arg(0)
	}

	parts := strings.Split(targetTable, ".")
	db := "default"
	tbl := targetTable
	if len(parts) == 2 {
		db = parts[0]
		tbl = parts[1]
	}

	cli, _ := client.NewClient(models.ConnectionConfig{DSN: *dsn, IsDemo: *dsn == "demo"})
	insp := inspector.NewPartsInspector(cli)

	res, err := insp.GetTableParts(context.Background(), db, tbl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting parts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTable: %s.%s (%s)\n", res.Database, res.Table, res.Engine)
	fmt.Printf("Active Parts: %d | Inactive Parts: %d | Overall Ratio: %.2fx | Health: %s\n\n",
		res.TotalActiveParts, res.TotalInactiveParts, res.OverallRatio, res.HealthRating)

	fmt.Println("PARTITION      ACTIVE  INACTIVE  ROWS         DISK SIZE    RATIO   STATUS     FRAGMENTATION")
	fmt.Println("-----------------------------------------------------------------------------------------")
	for _, p := range res.Partitions {
		diskMB := float64(p.TotalBytesOnDisk) / (1024 * 1024)
		fmt.Printf("%-14s %-7d %-9d %-12d %-9.1f MB  %-6.2fx %-10s %.1f%%\n",
			p.Partition, p.ActiveParts, p.InactiveParts, p.TotalRows, diskMB, p.CompressionRatio, p.HealthStatus, p.FragmentationScore)
	}
	fmt.Println()
}

func runMerges(args []string) {
	fs := flag.NewFlagSet("merges", flag.ExitOnError)
	dsn := fs.String("dsn", "demo", "ClickHouse DSN")
	_ = fs.Parse(args)

	cli, _ := client.NewClient(models.ConnectionConfig{DSN: *dsn, IsDemo: *dsn == "demo"})
	insp := inspector.NewMergesInspector(cli)

	merges, err := insp.GetActiveMerges(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying merges: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nActive Background Merges: %d\n\n", len(merges))
	if len(merges) == 0 {
		fmt.Println("No active background merges at this moment.")
		return
	}

	fmt.Println("DATABASE     TABLE     PROGRESS  PARTS  RESULT PART          ELAPSED   READ/s     WRITE/s")
	fmt.Println("------------------------------------------------------------------------------------------")
	for _, m := range merges {
		readMB := m.ReadThroughput / (1024 * 1024)
		writeMB := m.WriteThroughput / (1024 * 1024)
		fmt.Printf("%-12s %-9s %-8.1f%% %-6d %-20s %-8.1fs %-8.1f MB %-8.1f MB\n",
			m.Database, m.Table, m.Progress, m.NumParts, m.ResultPartName, m.ElapsedSeconds, readMB, writeMB)
	}
	fmt.Println()
}

func runAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	dsn := fs.String("dsn", "demo", "ClickHouse DSN")
	_ = fs.Parse(args)

	targetTable := "telemetry.events"
	if len(fs.Args()) > 0 {
		targetTable = fs.Arg(0)
	}

	parts := strings.Split(targetTable, ".")
	db := "default"
	tbl := targetTable
	if len(parts) == 2 {
		db = parts[0]
		tbl = parts[1]
	}

	cli, _ := client.NewClient(models.ConnectionConfig{DSN: *dsn, IsDemo: *dsn == "demo"})
	insp := inspector.NewColumnsInspector(cli)

	res, err := insp.AuditColumns(context.Background(), db, tbl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error auditing columns: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nColumn Compression Audit: %s.%s (%d columns)\n", res.Database, res.Table, res.TotalColumns)
	fmt.Printf("Total Compressed: %.1f MB | Total Uncompressed: %.1f MB | Potential Savings: %.1f MB\n\n",
		float64(res.TotalCompressed)/(1024*1024),
		float64(res.TotalUncompressed)/(1024*1024),
		float64(res.PotentialSavings)/(1024*1024))

	for _, col := range res.Columns {
		compMB := float64(col.DataCompressedBytes) / (1024 * 1024)
		fmt.Printf("Column: %s (%s)\n", col.Column, col.Type)
		fmt.Printf("  Current Codec: %s | Size: %.1f MB | Ratio: %.2fx\n", col.CompressionCodec, compMB, col.CompressionRatio)
		if col.SavingsEstimatePercent > 0 {
			fmt.Printf("  -> Recommendation: %s (Est. savings: %.1f%%)\n", col.RecommendedCodec, col.SavingsEstimatePercent)
			fmt.Printf("  -> Rationale:      %s\n", col.Rationale)
			fmt.Printf("  -> SQL Statement:  %s\n", col.AlterSQL)
		} else {
			fmt.Println("  -> Status:         Optimal encoding")
		}
		fmt.Println()
	}
}

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	dsn := fs.String("dsn", "demo", "ClickHouse DSN")
	mode := fs.String("mode", "PIPELINE", "Explain mode: PIPELINE or PLAN")
	_ = fs.Parse(args)

	query := "SELECT event_type, host, count() FROM telemetry.events GROUP BY event_type, host"
	if len(fs.Args()) > 0 {
		query = fs.Arg(0)
	}

	cli, _ := client.NewClient(models.ConnectionConfig{DSN: *dsn, IsDemo: *dsn == "demo"})
	parser := profiler.NewExplainParser(cli)

	analysis, err := parser.ParseQueryPipeline(context.Background(), query, *mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error profiling query: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nQuery Pipeline Profile (%s mode)\n", analysis.Mode)
	fmt.Printf("Query: %s\n", analysis.Query)
	fmt.Printf("Total Steps: %d | Max Thread Concurrency: %d\n\n", len(analysis.Nodes), analysis.MaxConcurrency)

	fmt.Println("EXECUTION DAG STEPS:")
	for _, n := range analysis.Nodes {
		bottleneckTag := ""
		if n.IsBottleneck {
			bottleneckTag = " [BOTTLENECK: CONCURRENCY COLLAPSE]"
		}
		fmt.Printf("  [%s] %-25s (Streams: %d)%s\n", n.ID, n.Label, n.StreamCount, bottleneckTag)
	}

	if len(analysis.Bottlenecks) > 0 {
		fmt.Println("\nIDENTIFIED BOTTLENECKS:")
		for _, b := range analysis.Bottlenecks {
			fmt.Printf("  - [%s] %s: %s\n", b.Severity, b.StepType, b.Description)
			fmt.Printf("    Recommendation: %s\n", b.Recommendation)
		}
	}
	fmt.Println()
}

func runProfile(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: clicklens profile <file.parquet | file.csv>")
		os.Exit(1)
	}

	path := args[0]
	lr := client.NewLocalRunner()
	inspector := client.NewFileInspector(lr)

	prof, err := inspector.InspectFile(context.Background(), path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFile Profile: %s\n", prof.FileName)
	fmt.Printf("Format: %s | Size: %.2f KB | Estimated Rows: %d | Inspector: %s\n\n",
		prof.Format, float64(prof.FileSizeBytes)/1024.0, prof.RowCount, prof.RunnerUsed)

	fmt.Println("COLUMNS SCHEMA:")
	for _, c := range prof.Columns {
		nullableStr := ""
		if c.Nullable {
			nullableStr = "(Nullable)"
		}
		fmt.Printf("  - %-25s %-15s %s\n", c.Name, c.Type, nullableStr)
	}

	if len(prof.SampleRows) > 0 {
		fmt.Printf("\nSAMPLE DATA (First %d rows):\n", len(prof.SampleRows))
		for i, r := range prof.SampleRows {
			fmt.Printf("Row #%d: %+v\n", i+1, r)
		}
	}
	fmt.Println()
}
