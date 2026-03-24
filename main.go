package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run . <question>\nExample: go run . \"Why were my queries slow yesterday?\"")
	}
	prompt := os.Args[1]

	ctx := context.Background()

	fmt.Println("🤖 Extracting intent from query...")
	intent, err := ExtractIntent(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to extract intent: %v", err)
	}
	fmt.Printf("🔍 Intent Extracted:\n   Category:  %s\n   StartTime: %s\n   EndTime:   %s\n\n", intent.Category, intent.StartTime, intent.EndTime)

	// Safety check to prevent analyzing generic off-topic chat
	if intent.Category == "unknown" {
		fmt.Println("❌ Error: The AI determined this query is not related to database performance.")
		fmt.Println("   Please ask a database-specific question (e.g., 'Are my queries slow?', 'Are there replication lags?').")
		os.Exit(1)
	}

	fmt.Println("📊 Fetching metrics from database...")
	metrics, err := FetchMetrics(ctx, intent)
	if err != nil {
		log.Fatalf("Failed to fetch metrics: %v", err)
	}
	fmt.Printf("📈 Retrieved Metrics:\n%s\n", metrics)

	fmt.Println("🧠 Analyzing metrics with LLM...")
	analysis, err := AnalyzeMetrics(ctx, prompt, metrics)
	if err != nil {
		log.Fatalf("Failed to analyze metrics: %v", err)
	}

	fmt.Println("\n===========================")
	fmt.Println("   Analysis Result")
	fmt.Println("===========================")
	fmt.Println(analysis)
}
