package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockData(t *testing.T) {
	intent := IntentData{Category: "pg_stat_statements"}
	data := mockData(intent)
	assert.Contains(t, data, "UPDATE accounts")
	assert.Contains(t, data, "845.2ms")
}

func TestLLM_ExtractIntent(t *testing.T) {
	t.Skip("Skipping LLM integration test by default// TestLLM_ExtractIntent requires a local Ollama instance running the \"gemma:2b\" model.")
	ctx := context.Background()
	prompt := "Why were my queries slow yesterday?"
	
	intent, err := ExtractIntent(ctx, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "pg_stat_statements", intent.Category) 
}
