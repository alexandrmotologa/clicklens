package inspector

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/clicklens/internal/client"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// ColumnsInspector analyzes storage efficiency and recommends compression codecs.
type ColumnsInspector struct {
	client *client.Client
}

// NewColumnsInspector creates a ColumnsInspector.
func NewColumnsInspector(c *client.Client) *ColumnsInspector {
	return &ColumnsInspector{client: c}
}

// AuditColumns queries system.columns and computes codec recommendations.
func (ci *ColumnsInspector) AuditColumns(ctx context.Context, database, table string) (*models.TableColumnsResponse, error) {
	if ci.client.IsSimulated() {
		res := ci.client.Simulator().GetColumnsAudit(database, table)
		return &res, nil
	}

	conn := ci.client.Conn()
	if conn == nil {
		res := ci.client.Simulator().GetColumnsAudit(database, table)
		return &res, nil
	}

	query := `
		SELECT
			name,
			type,
			default_expression,
			compression_codec,
			data_compressed_bytes,
			data_uncompressed_bytes
		FROM system.columns
		WHERE database = ? AND table = ?
		ORDER BY data_compressed_bytes DESC
	`

	rows, err := conn.Query(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.columns for %s.%s: %w", database, table, err)
	}
	defer rows.Close()

	columns := []models.ColumnAudit{}
	var totalComp uint64 = 0
	var totalUncomp uint64 = 0
	var totalSavings uint64 = 0

	for rows.Next() {
		var (
			colName          string
			colType          string
			defaultExpr      string
			codec            string
			compressedBytes  uint64
			uncompressedBytes uint64
		)

		if err := rows.Scan(
			&colName,
			&colType,
			&defaultExpr,
			&codec,
			&compressedBytes,
			&uncompressedBytes,
		); err != nil {
			continue
		}

		if codec == "" {
			codec = "Default(LZ4)"
		}

		ratio := 1.0
		if compressedBytes > 0 {
			ratio = float64(uncompressedBytes) / float64(compressedBytes)
		}

		recommended, savingsPct, rationale := recommendCodec(colName, colType, codec)
		projectedSavings := uint64(float64(compressedBytes) * (savingsPct / 100.0))

		alterSQL := ""
		if savingsPct > 0 {
			if strings.HasPrefix(recommended, "LowCardinality") {
				alterSQL = fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN %s %s;", database, table, colName, recommended)
			} else {
				alterSQL = fmt.Sprintf("ALTER TABLE %s.%s MODIFY COLUMN %s %s %s;", database, table, colName, colType, recommended)
			}
		}

		columns = append(columns, models.ColumnAudit{
			Database:               database,
			Table:                  table,
			Column:                 colName,
			Type:                   colType,
			DefaultExpression:      defaultExpr,
			CompressionCodec:       codec,
			DataCompressedBytes:    compressedBytes,
			DataUncompressedBytes:  uncompressedBytes,
			CompressionRatio:       ratio,
			RecommendedCodec:       recommended,
			SavingsEstimatePercent: savingsPct,
			ProjectedSavingsBytes:  projectedSavings,
			AlterSQL:               alterSQL,
			Rationale:              rationale,
		})

		totalComp += compressedBytes
		totalUncomp += uncompressedBytes
		totalSavings += projectedSavings
	}

	migrationScript := generateZeroDowntimeMigrationScript(database, table, columns)

	return &models.TableColumnsResponse{
		Database:                    database,
		Table:                       table,
		TotalColumns:                len(columns),
		TotalCompressed:             totalComp,
		TotalUncompressed:           totalUncomp,
		PotentialSavings:            totalSavings,
		ZeroDowntimeMigrationScript: migrationScript,
		Columns:                     columns,
	}, nil
}

// generateZeroDowntimeMigrationScript builds an executable, zero-downtime table migration playbook.
func generateZeroDowntimeMigrationScript(database, table string, columns []models.ColumnAudit) string {
	var alterStatements strings.Builder
	for _, col := range columns {
		if col.AlterSQL != "" {
			// Replace {table} with {table}_new
			newAlter := strings.Replace(col.AlterSQL, fmt.Sprintf("%s.%s", database, table), fmt.Sprintf("%s.%s_new", database, table), 1)
			alterStatements.WriteString(newAlter + "\n")
		}
	}

	return fmt.Sprintf(`-- ==============================================================================
-- Zero-Downtime Codec Migration Playbook for %s.%s
-- Generated automatically by ClickLens Studio
-- ==============================================================================

-- Step 1: Create shadow staging table matching current schema
CREATE TABLE IF NOT EXISTS %s.%s_new AS %s.%s;

-- Step 2: Apply recommended compression codecs to new table definition
%s
-- Step 3: Populate new table with compressed data
-- For large datasets, attach partition by partition or stream in batches:
INSERT INTO %s.%s_new SELECT * FROM %s.%s;

-- Step 4: Atomic zero-downtime swap of tables (transparent to client queries)
EXCHANGE TABLES %s.%s AND %s.%s_new;

-- Step 5: Verification & Cleanup
-- SELECT count() FROM %s.%s;
-- DROP TABLE %s.%s_new;
`, database, table, database, table, database, table, alterStatements.String(), database, table, database, table, database, table, database, table, database, table, database, table)
}

// recommendCodec applies heuristics based on ClickHouse column types.
func recommendCodec(colName, colType, currentCodec string) (recommended string, savingsPct float64, rationale string) {
	upperType := strings.ToUpper(colType)
	upperCodec := strings.ToUpper(currentCodec)

	// 1. Timestamps & Dates
	if strings.Contains(upperType, "DATETIME") || strings.Contains(upperType, "DATE") {
		if !strings.Contains(upperCodec, "DOUBLEDELTA") && !strings.Contains(upperCodec, "GORILLA") {
			return "CODEC(DoubleDelta, ZSTD(1))", 62.0, "Monotonic timestamp series exhibit uniform spacing. DoubleDelta reduces bit representation before ZSTD entropy coding."
		}
		return "Optimal", 0, "Already utilizing specialized time-series encoding."
	}

	// 2. Floating-point metrics (Float32, Float64)
	if strings.Contains(upperType, "FLOAT") {
		if !strings.Contains(upperCodec, "GORILLA") {
			return "CODEC(Gorilla, ZSTD(1))", 58.0, "Floating-point metrics have repetitive XOR mantissa patterns. Gorilla compression significantly outcompresses LZ4."
		}
		return "Optimal", 0, "Already utilizing Gorilla floating-point encoding."
	}

	// 3. Integer IDs and Counters
	if strings.Contains(upperType, "INT") && !strings.Contains(upperType, "POINT") {
		if !strings.Contains(upperCodec, "DOUBLEDELTA") && !strings.Contains(upperCodec, "T64") {
			if strings.HasSuffix(strings.ToLower(colName), "_id") || strings.HasSuffix(strings.ToLower(colName), "id") {
				return "CODEC(T64, ZSTD(1))", 52.0, "Sequential numeric identifiers benefit from T64 transposition encoding."
			}
			return "CODEC(DoubleDelta, ZSTD(1))", 45.0, "Integer counters with consistent increments benefit from DoubleDelta delta encoding."
		}
		return "Optimal", 0, "Already utilizing numeric integer codecs."
	}

	// 4. Low-cardinality candidate strings
	if upperType == "STRING" && !strings.Contains(upperType, "LOWCARDINALITY") {
		lowerName := strings.ToLower(colName)
		isCandidate := strings.HasSuffix(lowerName, "_type") ||
			strings.HasSuffix(lowerName, "status") ||
			strings.HasSuffix(lowerName, "country") ||
			strings.HasSuffix(lowerName, "code") ||
			strings.HasSuffix(lowerName, "category") ||
			strings.HasSuffix(lowerName, "host") ||
			strings.HasSuffix(lowerName, "env")

		if isCandidate {
			return "LowCardinality(String) CODEC(ZSTD(3))", 48.0, "Column name suggests repetitive categorical values. LowCardinality dictionary index yields dramatic memory and storage reductions."
		}

		if !strings.Contains(upperCodec, "ZSTD") {
			return "CODEC(ZSTD(3))", 26.0, "Switching general string text from default LZ4 to ZSTD(3) improves ratio with minimal decompression overhead."
		}
	}

	// 5. High-entropy payload text
	if (strings.Contains(upperType, "STRING") || strings.Contains(upperType, "JSON")) && strings.Contains(upperCodec, "LZ4") {
		return "CODEC(ZSTD(6))", 32.0, "Complex JSON and unstructured text reach superior density under ZSTD level 6."
	}

	return "Optimal", 0, "Current column encoding conforms to recommended best practices."
}
