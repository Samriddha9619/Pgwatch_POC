package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const ollamaURL = "http://localhost:11434/api/generate"
// We default to gemma:2b based on your local Ollama setup
const modelName = "gemma:2b"

type IntentData struct {
	Category  string `json:"category"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format string `json:"format,omitempty"` // "json" forces JSON output formatting
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// ExtractIntent calls the Routing LLM to convert a prompt into structured JSON.
func ExtractIntent(ctx context.Context, prompt string) (IntentData, error) {
	systemPrompt := `You are an intent extraction agent mapping user questions to PostgreSQL metric categories.
Categories available:
- "pg_stat_statements": For anything about queries, slow queries, performance of queries, or what queries are running.
- "locks": For locking or blocking issues.
- "stat_activity": For connections, idle transactions, or active session issues.
- "replication": For replication lag.
- "unknown": Only if the question is completely unrelated to databases or query performance (e.g. "hello", "weather").

Respond with ONLY valid JSON with exactly these keys: category, start_time, end_time. If times are not specified, use "now" and "24h ago".`

	fullPrompt := systemPrompt + "\n\nUser prompt: " + prompt

	reqBody := OllamaRequest{
		Model:  modelName,
		Prompt: fullPrompt,
		Format: "json",
		Stream: false,
	}

	respBody, err := callOllama(ctx, reqBody)
	if err != nil {
		return IntentData{}, fmt.Errorf("failed to contact local AI. Please make sure Ollama is running in another terminal (`ollama run llama3`): %v", err)
	}


	var intent IntentData
	err = json.Unmarshal(respBody, &intent)
	if err != nil {
		return IntentData{}, fmt.Errorf("failed to parse JSON from LLM: %v (raw LLM response: %s)", err, string(respBody))
	}

	return intent, nil
}

// AnalyzeMetrics calls the Heavy LLM to interpret the raw database rows.
func AnalyzeMetrics(ctx context.Context, originalPrompt, metricsData string) (string, error) {
	systemPrompt := `You are a PostgreSQL database performance analyst. 
You are analyzing metrics collected by pgwatch. 
Provide a concise analysis based ONLY on the provided metrics. Include:
1. What the metrics show
2. Likely root cause
3. Recommended actions`

	fullPrompt := fmt.Sprintf("%s\n\nMetrics Data:\n%s\n\nUser question: %s", systemPrompt, metricsData, originalPrompt)

	reqBody := OllamaRequest{
		Model:  modelName,
		Prompt: fullPrompt,
		Stream: false,
	}

	respBody, err := callOllama(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to contact local AI. Please make sure Ollama is running in another terminal (`ollama run llama3`): %v", err)
	}

	return string(respBody), nil
}

// callOllama is a generic helper to make HTTP requests to the local Ollama instance.
func callOllama(ctx context.Context, reqBody OllamaRequest) ([]byte, error) {
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Ollama at %s: %v. Is Ollama running?", ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama returned HTTP status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return nil, err
	}

	return []byte(ollamaResp.Response), nil
}
