package profiler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// ExplainParser parses EXPLAIN PIPELINE and EXPLAIN PLAN output into a visual DAG.
type ExplainParser struct {
	client *client.Client
}

// NewExplainParser creates an ExplainParser.
func NewExplainParser(c *client.Client) *ExplainParser {
	return &ExplainParser{client: c}
}

// ParseQueryPipeline executes EXPLAIN PIPELINE and constructs the execution DAG.
func (ep *ExplainParser) ParseQueryPipeline(ctx context.Context, query string, mode string) (*models.PipelineAnalysis, error) {
	if mode == "" {
		mode = "PIPELINE"
	}

	cleanQuery := strings.TrimSpace(query)
	cleanQuery = strings.TrimSuffix(cleanQuery, ";")

	if ep.client.IsSimulated() {
		analysis := ep.client.Simulator().GenerateSamplePipelineAnalysis(cleanQuery)
		analysis.Mode = mode
		return &analysis, nil
	}

	conn := ep.client.Conn()
	if conn == nil {
		analysis := ep.client.Simulator().GenerateSamplePipelineAnalysis(cleanQuery)
		analysis.Mode = mode
		return &analysis, nil
	}

	explainCmd := fmt.Sprintf("EXPLAIN %s %s", mode, cleanQuery)
	rows, err := conn.Query(ctx, explainCmd)
	if err != nil {
		// If real explain fails (e.g. invalid query or syntax), return simulated graph with error message
		analysis := ep.client.Simulator().GenerateSamplePipelineAnalysis(cleanQuery)
		analysis.RawText = fmt.Sprintf("Query execution warning: %v\n\nFalling back to reference execution graph.", err)
		return &analysis, nil
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err == nil {
			lines = append(lines, line)
		}
	}

	rawText := strings.Join(lines, "\n")
	analysis := ep.buildDAGFromExplainLines(cleanQuery, mode, rawText, lines)
	return &analysis, nil
}

var (
	multiplierRe = regexp.MustCompile(`[×x]\s*(\d+)`)
	transitionRe = regexp.MustCompile(`(\d+)\s*(?:→|->)\s*(\d+)`)
	stepHeaderRe = regexp.MustCompile(`^\s*\(([^)]+)\)`)
)

func (ep *ExplainParser) buildDAGFromExplainLines(query, mode, rawText string, lines []string) models.PipelineAnalysis {
	nodes := []models.DAGNode{}
	edges := []models.DAGEdge{}
	bottlenecks := []models.PipelineBottleneck{}
	maxConcurrency := 1

	type parsedLine struct {
		indent   int
		text     string
		streams  int
		outStreams int
		category string
	}

	parsed := []parsedLine{}

	currentCategory := "Execution"
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}

		if match := stepHeaderRe.FindStringSubmatch(trimmed); len(match) > 1 {
			currentCategory = match[1]
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		streams := 1
		outStreams := 1

		if m := transitionRe.FindStringSubmatch(trimmed); len(m) > 2 {
			inCount, _ := strconv.Atoi(m[1])
			outCount, _ := strconv.Atoi(m[2])
			streams = inCount
			outStreams = outCount
		} else if m := multiplierRe.FindStringSubmatch(trimmed); len(m) > 1 {
			c, _ := strconv.Atoi(m[1])
			streams = c
			outStreams = c
		}

		if streams > maxConcurrency {
			maxConcurrency = streams
		}

		parsed = append(parsed, parsedLine{
			indent:     indent,
			text:       trimmed,
			streams:    streams,
			outStreams: outStreams,
			category:   currentCategory,
		})
	}

	// If no valid lines parsed, fallback
	if len(parsed) == 0 {
		return ep.client.Simulator().GenerateSamplePipelineAnalysis(query)
	}

	// ClickHouse EXPLAIN PIPELINE is often printed bottom-up (sources at bottom) or top-down.
	// We reverse order so data flows from Read -> Filter -> Aggregate -> Sort -> Output
	reversed := make([]parsedLine, len(parsed))
	for i, p := range parsed {
		reversed[len(parsed)-1-i] = p
	}

	prevNodeID := ""
	prevStreams := 1

	for i, p := range reversed {
		nodeID := fmt.Sprintf("node-%d", i+1)
		label := p.text
		stepType := p.category

		parts := strings.Fields(p.text)
		if len(parts) > 0 {
			stepType = parts[0]
		}

		isBottleneck := false
		if (prevStreams >= 8 && p.streams == 1) || (p.streams >= 8 && p.outStreams == 1) {
			isBottleneck = true
			fromCount := prevStreams
			if p.streams >= 8 && p.outStreams == 1 {
				fromCount = p.streams
			}
			bottlenecks = append(bottlenecks, models.PipelineBottleneck{
				NodeID:          nodeID,
				StepType:        stepType,
				PreviousStreams: fromCount,
				CurrentStreams:  p.outStreams,
				Severity:        "WARNING",
				Description:     fmt.Sprintf("Pipeline concurrency collapse: stream fan-in from %d threads to %d thread.", fromCount, p.outStreams),
				Recommendation: "Examine whether aggregation or sorting can preserve multi-stream parallelism or utilize external spill settings.",
			})
		}

		node := models.DAGNode{
			ID:           nodeID,
			Label:        label,
			StepType:     stepType,
			StreamCount:  p.streams,
			Details:      fmt.Sprintf("Category: %s | Active Streams: %d", p.category, p.streams),
			IsBottleneck: isBottleneck,
		}

		if prevNodeID != "" {
			node.Inputs = []string{prevNodeID}
			edgeLabel := fmt.Sprintf("%d streams", prevStreams)
			if prevStreams != p.streams {
				edgeLabel = fmt.Sprintf("%d → %d", prevStreams, p.streams)
			}
			edges = append(edges, models.DAGEdge{
				From:        prevNodeID,
				To:          nodeID,
				StreamCount: p.streams,
				Label:       edgeLabel,
			})
		}

		nodes = append(nodes, node)
		prevNodeID = nodeID
		prevStreams = p.outStreams
	}

	return models.PipelineAnalysis{
		Query:          query,
		Mode:           mode,
		RawText:        rawText,
		TotalStreams:   len(nodes),
		MaxConcurrency: maxConcurrency,
		Nodes:          nodes,
		Edges:          edges,
		Bottlenecks:    bottlenecks,
		SpillAlerts:    []models.SpillAlert{},
	}
}
