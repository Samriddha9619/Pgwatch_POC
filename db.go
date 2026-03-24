package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// FetchMetrics simulates a fetch to a TimescaleDB/pgwatch store based on the extracted intent.
func FetchMetrics(ctx context.Context, intent IntentData) (string, error) {
	dsn := os.Getenv("PGWATCH_DSN")
	if dsn == "" {
		// Mock behavior when no DB is provided to make the PoC easy to test without a real PG instance.
		return mockData(intent), nil
	}

	// If PGWATCH_DSN is provided, connect to the real database and fetch metrics.
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return "", fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Build query based on category (a simplified heuristic for the PoC)
	var query string
	switch intent.Category {
	case "pg_stat_statements":
		query = "SELECT query as metric, mean_exec_time as value FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 5"
	case "locks":
		query = "SELECT mode as metric, count(*) as value FROM pg_locks GROUP BY mode"
	default:
		// Attempt to query pg_stat_activity if unknown
		query = "SELECT state as metric, count(*) as value FROM pg_stat_activity GROUP BY state"
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result string
	for rows.Next() {
		var metric string
		var value float64 // treating generic values as float64 for simplicity
		if err := rows.Scan(&metric, &value); err != nil {
			return "", err
		}
		result += fmt.Sprintf("- %s: %.2f\n", metric, value)
	}
	return result, nil
}

// mockData returns dummy pgwatch metrics if no database connection is configured.
func mockData(intent IntentData) string {
	switch intent.Category {
	case "pg_stat_statements", "slow_queries":
		return "Query: UPDATE accounts SET balance = balance - 100 WHERE id = 1234\nMean Exec Time: 845.2ms\nCalls: 12500\n\nQuery: SELECT sum(amount) FROM transactions WHERE date > '2025-01-01'\nMean Exec Time: 1240.5ms\nCalls: 50"
	case "locks":
		return "Lock Mode: ExclusiveLock\nRelation: accounts\nGranted: false\nWait Time: 4500ms\n\nLock Mode: AccessShareLock\nRelation: transactions\nGranted: true\nWait Time: 1.2ms"
	case "stat_activity", "connections":
		return "State: active\nQuery: autovacuum: vacuum analyze public.accounts\nDuration: 15m 32s\n\nState: idle in transaction\nQuery: BEGIN; UPDATE users SET...\nDuration: 45m 12s"
	case "replication":
		return "Client: replica_1\nState: streaming\nReplay Lag: 45 seconds\nWrite Lag: 12 seconds"
	default:
		return fmt.Sprintf("No anomalous metrics found for category: %s. Everything looks healthy.", intent.Category)
	}
}
