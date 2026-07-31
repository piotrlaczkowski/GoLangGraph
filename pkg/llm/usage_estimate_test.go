// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package llm

import (
	"strings"
	"testing"
)

func TestEstimateTokensPrefersTiktoken(t *testing.T) {
	if !TokenizerAvailable() {
		t.Skip("tiktoken cl100k_base unavailable")
	}
	if EstimateTokens("") != 0 {
		t.Fatal("empty")
	}
	text := "please plan a careful change across several modules with enough detail"
	got := EstimateTokens(text)
	heur := estimateTokensHeuristic(text)
	if got <= 0 {
		t.Fatalf("tiktoken count=%d", got)
	}
	// Real tokenizer should differ from naive chars/4 on typical prose.
	if got == heur {
		t.Fatalf("expected tiktoken (%d) != heuristic (%d) for prose", got, heur)
	}
	// Sanity: longer text → more tokens.
	long := EstimateTokens(strings.Repeat(text+" ", 8))
	if long <= got {
		t.Fatalf("long=%d short=%d", long, got)
	}
}

func TestEnsureUsageUsesTokenizerWhenEmpty(t *testing.T) {
	if !TokenizerAvailable() {
		t.Skip("tiktoken cl100k_base unavailable")
	}
	resp := &CompletionResponse{
		Choices: []Choice{{Message: Message{Content: `{"ok":true}`}, FinishReason: "early_exit"}},
	}
	req := CompletionRequest{Messages: []Message{
		{Role: "user", Content: "return strict json please"},
	}}
	if !EnsureUsage(resp, req) {
		t.Fatal("expected estimate")
	}
	if resp.Usage.PromptTokens <= 0 || resp.Usage.CompletionTokens <= 0 {
		t.Fatalf("usage=%+v", resp.Usage)
	}
	if resp.Metadata["usage_tokenizer"] != "cl100k_base" {
		t.Fatalf("tokenizer meta=%v", resp.Metadata["usage_tokenizer"])
	}
}
