# Pgwatch Copilot PoC

This is a quick lightweight proof-of-concept for my GSoC 2026 proposal about bringing natural language analysis to `pgwatch`.

It's a Go CLI that demonstrates the two-step LLM routing architecture:
1. It takes a natural-language question (e.g. "Why were queries slow yesterday?") and passes it through a local LLM (`Ollama` + `gemma:2b`) using strict JSON schema extraction to grab the metric category and time window.
2. It takes that intent and "fetches" the raw PostgreSQL metrics (the DB rows are mocked in this repo since we don't have a live TimescaleDB hookup).
3. It pipes those raw `pg_stat` rows right back into the LLM to generate a human-readable DBA analysis report.

### How to run it locally
1. Make sure you have Ollama running in the background with the Gemma model: `ollama run gemma:2b`
2. Run the code:
```bash
go run . "Why were my queries slow yesterday between 2-4pm?"
```
