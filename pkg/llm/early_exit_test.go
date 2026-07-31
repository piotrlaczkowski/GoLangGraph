// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLooksCompleteStructured(t *testing.T) {
	if !LooksCompleteStructured(`{"tasks":[{"id":"T1"}]}`) {
		t.Fatal("tasks")
	}
	if !LooksCompleteStructured("here\n```json\n{\"status\":\"done\",\"summary\":\"x\"}\n```\n") {
		t.Fatal("fenced status")
	}
	if LooksCompleteStructured(`not json`) {
		t.Fatal("plain")
	}
	if LooksCompleteStructured(`{"foo":1}`) {
		t.Fatal("unrelated")
	}
	// Trailing prose after a complete object should still early-exit.
	if !LooksCompleteStructured(`{"steps":["a"],"summary":"ok"} Thanks!`) {
		t.Fatal("trailing prose")
	}
}

func TestLooksCompleteToolCalls(t *testing.T) {
	if LooksCompleteToolCalls(nil) {
		t.Fatal("nil")
	}
	if LooksCompleteToolCalls([]ToolCall{{Function: FunctionCall{Name: "ws_edit", Arguments: `{"path":`}}}) {
		t.Fatal("incomplete args")
	}
	if !LooksCompleteToolCalls([]ToolCall{{
		Function: FunctionCall{Name: "ws_edit", Arguments: `{"path":"a.go","old":"x","new":"y"}`},
	}}) {
		t.Fatal("complete tool")
	}
}

func TestCollectStreamEarlyExit(t *testing.T) {
	chunks := []string{
		`{"status":`,
		`"done","summary":"ok"}`,
		` and lots of wasted tokens the model would keep generating`,
	}
	var sent int
	streamFn := func(_ context.Context, _ CompletionRequest, cb StreamCallback) error {
		for _, part := range chunks {
			sent++
			if err := cb(CompletionResponse{
				Choices: []Choice{{Delta: Message{Role: "assistant", Content: part}}},
			}); err != nil {
				return err
			}
		}
		return nil
	}
	resp, err := CollectStream(context.Background(), streamFn, CompletionRequest{
		EarlyExit: DefaultEarlyExit,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sent != 2 {
		t.Fatalf("expected early cancel after 2 chunks, sent=%d", sent)
	}
	if resp.Choices[0].FinishReason != "early_exit" {
		t.Fatalf("finish=%s", resp.Choices[0].FinishReason)
	}
	if !strings.Contains(resp.Choices[0].Message.Content, `"done"`) {
		t.Fatalf("content=%s", resp.Choices[0].Message.Content)
	}
	if v, _ := resp.Metadata["early_exit"].(bool); !v {
		t.Fatal("metadata early_exit missing")
	}
}

func TestCollectStreamToolCallEarlyExit(t *testing.T) {
	var sent int
	streamFn := func(_ context.Context, _ CompletionRequest, cb StreamCallback) error {
		parts := []ToolCall{
			{Index: 0, ID: "c1", Type: "function", Function: FunctionCall{Name: "ws_write", Arguments: `{"path":"a.go",`}},
			{Index: 0, Function: FunctionCall{Arguments: `"content":"package main\n"}`}},
			{Index: 0, Function: FunctionCall{Arguments: `,"extra":"waste"}`}}, // would not be sent
		}
		for _, tc := range parts {
			sent++
			if err := cb(CompletionResponse{
				Choices: []Choice{{Delta: Message{ToolCalls: []ToolCall{tc}}}},
			}); err != nil {
				return err
			}
		}
		return nil
	}
	resp, err := CollectStream(context.Background(), streamFn, CompletionRequest{
		EarlyExit: DefaultEarlyExit,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sent != 2 {
		t.Fatalf("sent=%d want 2", sent)
	}
	tc := resp.Choices[0].Message.ToolCalls
	if len(tc) != 1 || tc[0].Function.Name != "ws_write" {
		t.Fatalf("tools=%+v", tc)
	}
	if !LooksCompleteToolCalls(tc) {
		t.Fatalf("args incomplete: %s", tc[0].Function.Arguments)
	}
}

func TestIsStreamEarlyExit(t *testing.T) {
	if !IsStreamEarlyExit(ErrStreamEarlyExit) {
		t.Fatal("direct")
	}
	if !IsStreamEarlyExit(errors.Join(ErrStreamEarlyExit, errors.New("x"))) {
		t.Fatal("joined")
	}
	if IsStreamEarlyExit(errors.New("nope")) {
		t.Fatal("other")
	}
}

func TestStreamCollectorToolMerge(t *testing.T) {
	c := NewStreamCollector()
	c.Add(CompletionResponse{Choices: []Choice{{Delta: Message{
		ToolCalls: []ToolCall{{Index: 0, ID: "1", Type: "function", Function: FunctionCall{Name: "ws_edit", Arguments: `{"a":`}}},
	}}}})
	c.Add(CompletionResponse{Choices: []Choice{{Delta: Message{
		ToolCalls: []ToolCall{{Index: 0, Function: FunctionCall{Arguments: `1}`}}},
	}}}})
	got := c.ToolCalls()
	if len(got) != 1 || got[0].Function.Arguments != `{"a":1}` {
		t.Fatalf("%+v", got)
	}
}

func TestCollectStreamEstimatesUsageWhenEmpty(t *testing.T) {
	streamFn := func(_ context.Context, _ CompletionRequest, cb StreamCallback) error {
		return cb(CompletionResponse{
			Choices: []Choice{{Delta: Message{Role: "assistant", Content: `{"status":"done","summary":"ok"}`}}},
		})
	}
	prompt := "plan the work carefully with enough detail"
	resp, err := CollectStream(context.Background(), streamFn, CompletionRequest{
		Messages:  []Message{{Role: "user", Content: prompt}},
		EarlyExit: DefaultEarlyExit,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Usage.PromptTokens <= 0 || resp.Usage.CompletionTokens <= 0 {
		t.Fatalf("expected estimated usage, got %+v", resp.Usage)
	}
	if resp.Usage.TotalTokens != resp.Usage.PromptTokens+resp.Usage.CompletionTokens {
		t.Fatalf("total mismatch: %+v", resp.Usage)
	}
	if v, _ := resp.Metadata["usage_estimated"].(bool); !v {
		t.Fatal("expected usage_estimated metadata")
	}
	r := &CompletionResponse{
		Choices: []Choice{{Message: Message{Content: "hi"}, FinishReason: "stop"}},
		Usage:   Usage{PromptTokens: 11, CompletionTokens: 7, TotalTokens: 18},
	}
	if EnsureUsage(r, CompletionRequest{Messages: []Message{{Content: "x"}}}) {
		t.Fatal("should not estimate when usage present")
	}
	if r.Usage.TotalTokens != 18 {
		t.Fatalf("usage mutated: %+v", r.Usage)
	}
}

func TestEstimateTokensHeuristic(t *testing.T) {
	if EstimateTokens("") != 0 {
		t.Fatal("empty")
	}
	if EstimateTokens("abcd") != 1 {
		t.Fatalf("got %d", EstimateTokens("abcd"))
	}
	if EstimateTokens(strings.Repeat("x", 40)) != 10 {
		t.Fatalf("got %d", EstimateTokens(strings.Repeat("x", 40)))
	}
}
