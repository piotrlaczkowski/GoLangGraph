// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// ErrStreamEarlyExit is returned by a StreamCallback (or CollectStream) when a
// complete tool-call / structured JSON result is already formed and the rest of
// the token stream should be cancelled to save decode latency on SLMs.
var ErrStreamEarlyExit = errors.New("llm: stream early exit")

// IsStreamEarlyExit reports whether err is (or wraps) ErrStreamEarlyExit.
func IsStreamEarlyExit(err error) bool {
	return err != nil && errors.Is(err, ErrStreamEarlyExit)
}

// EarlyExitFunc decides whether accumulated stream content is complete enough
// to cancel remaining token generation. Returning true cancels the stream.
type EarlyExitFunc func(content string, toolCalls []ToolCall) bool

// StreamCollector accumulates content + tool-call argument deltas from chunks.
type StreamCollector struct {
	content strings.Builder
	tools   map[int]*ToolCall
	meta    CompletionResponse
	hasMeta bool
}

// NewStreamCollector constructs an empty accumulator.
func NewStreamCollector() *StreamCollector {
	return &StreamCollector{tools: map[int]*ToolCall{}}
}

// Add folds one streaming chunk into the collector.
func (c *StreamCollector) Add(chunk CompletionResponse) {
	if len(chunk.Choices) == 0 {
		return
	}
	c.meta = chunk
	c.hasMeta = true
	delta := chunk.Choices[0].Delta
	if delta.Content != "" {
		c.content.WriteString(delta.Content)
	}
	// Some providers put the full message on the final chunk.
	if delta.Content == "" && chunk.Choices[0].Message.Content != "" && c.content.Len() == 0 {
		c.content.WriteString(chunk.Choices[0].Message.Content)
	}
	for i, tc := range delta.ToolCalls {
		idx := tc.Index
		if idx == 0 && tc.ID == "" && i > 0 {
			idx = i
		}
		existing, ok := c.tools[idx]
		if !ok {
			cp := tc
			cp.Index = idx
			c.tools[idx] = &cp
			continue
		}
		if tc.ID != "" {
			existing.ID = tc.ID
		}
		if tc.Type != "" {
			existing.Type = tc.Type
		}
		if tc.Function.Name != "" {
			existing.Function.Name = tc.Function.Name
		}
		existing.Function.Arguments += tc.Function.Arguments
	}
	for i, tc := range chunk.Choices[0].Message.ToolCalls {
		idx := tc.Index
		if idx == 0 && i > 0 {
			idx = i
		}
		if _, ok := c.tools[idx]; !ok {
			cp := tc
			cp.Index = idx
			c.tools[idx] = &cp
		}
	}
}

// Content returns the accumulated assistant text.
func (c *StreamCollector) Content() string { return c.content.String() }

// ToolCalls returns accumulated tool calls in index order.
func (c *StreamCollector) ToolCalls() []ToolCall {
	if len(c.tools) == 0 {
		return nil
	}
	max := -1
	for idx := range c.tools {
		if idx > max {
			max = idx
		}
	}
	out := make([]ToolCall, 0, len(c.tools))
	for i := 0; i <= max; i++ {
		if tc, ok := c.tools[i]; ok {
			out = append(out, *tc)
		}
	}
	return out
}

// Response builds a non-streaming CompletionResponse from accumulated chunks.
func (c *StreamCollector) Response() *CompletionResponse {
	tools := c.ToolCalls()
	role := "assistant"
	finish := "stop"
	if c.hasMeta && len(c.meta.Choices) > 0 {
		if c.meta.Choices[0].Delta.Role != "" {
			role = c.meta.Choices[0].Delta.Role
		} else if c.meta.Choices[0].Message.Role != "" {
			role = c.meta.Choices[0].Message.Role
		}
		if fr := c.meta.Choices[0].FinishReason; fr != "" {
			finish = fr
		}
	}
	resp := &CompletionResponse{
		Object: "chat.completion",
		Choices: []Choice{{
			Index: 0,
			Message: Message{
				Role:      role,
				Content:   c.Content(),
				ToolCalls: tools,
			},
			FinishReason: finish,
		}},
	}
	if c.hasMeta {
		resp.ID = c.meta.ID
		resp.Created = c.meta.Created
		resp.Model = c.meta.Model
		resp.SystemFingerprint = c.meta.SystemFingerprint
	}
	return resp
}

// CollectStream runs streamFn, accumulates deltas, and cancels when
// req.EarlyExit reports a complete result. Early exit is returned as a
// successful CompletionResponse (not an error).
func CollectStream(
	ctx context.Context,
	streamFn func(context.Context, CompletionRequest, StreamCallback) error,
	req CompletionRequest,
) (*CompletionResponse, error) {
	col := NewStreamCollector()
	err := streamFn(ctx, req, func(chunk CompletionResponse) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		col.Add(chunk)
		if req.EarlyExit != nil && req.EarlyExit(col.Content(), col.ToolCalls()) {
			return ErrStreamEarlyExit
		}
		return nil
	})
	if err != nil && !IsStreamEarlyExit(err) {
		return nil, err
	}
	resp := col.Response()
	if IsStreamEarlyExit(err) {
		resp.Choices[0].FinishReason = "early_exit"
		if resp.Metadata == nil {
			resp.Metadata = map[string]interface{}{}
		}
		resp.Metadata["early_exit"] = true
	}
	// Providers often omit usage on cancelled streams — estimate so callers can account.
	_ = EnsureUsage(resp, req)
	if resp == nil || (resp.Choices[0].Message.Content == "" && len(resp.Choices[0].Message.ToolCalls) == 0) {
		if err != nil && !IsStreamEarlyExit(err) {
			return nil, err
		}
		if !col.hasMeta {
			return nil, errors.New("empty stream response")
		}
	}
	return resp, nil
}

// DefaultEarlyExit returns true when accumulated output is a complete structured
// JSON object (plan/tasks/status/review) or when every tool call has valid JSON args.
// Tuned for local SLMs that keep generating after a usable result is ready.
func DefaultEarlyExit(content string, toolCalls []ToolCall) bool {
	if LooksCompleteToolCalls(toolCalls) {
		return true
	}
	return LooksCompleteStructured(content)
}

// LooksCompleteToolCalls reports whether toolCalls are non-empty and each has a
// name plus parseable JSON arguments (balanced object/array).
func LooksCompleteToolCalls(toolCalls []ToolCall) bool {
	if len(toolCalls) == 0 {
		return false
	}
	for _, tc := range toolCalls {
		if strings.TrimSpace(tc.Function.Name) == "" {
			return false
		}
		args := strings.TrimSpace(tc.Function.Arguments)
		if args == "" {
			return false
		}
		if !json.Valid([]byte(args)) {
			return false
		}
	}
	return true
}

// LooksCompleteStructured reports whether s contains a parseable JSON object
// with a planning / worker / review shape that is safe to early-exit on.
func LooksCompleteStructured(s string) bool {
	raw := extractJSONObject(s)
	if raw == "" {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil || len(m) == 0 {
		return false
	}
	if _, ok := m["tasks"]; ok {
		return true
	}
	if _, ok := m["steps"]; ok {
		return true
	}
	if _, ok := m["approved"]; ok {
		return true
	}
	if status, ok := m["status"].(string); ok {
		st := strings.ToLower(status)
		return st == "done" || st == "blocked"
	}
	if _, ok := m["passed"]; ok {
		return true
	}
	if _, ok := m["summary"]; ok {
		if _, ok2 := m["goals"]; ok2 {
			return true
		}
		if _, ok2 := m["relevant_files"]; ok2 {
			return true
		}
		if _, ok2 := m["actions"]; ok2 {
			return true
		}
	}
	return false
}

func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	// Prefer fenced ```json blocks when present.
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		rest = strings.TrimPrefix(rest, "json")
		rest = strings.TrimPrefix(rest, "JSON")
		rest = strings.TrimLeft(rest, "\n\r ")
		if j := strings.Index(rest, "```"); j >= 0 {
			s = rest[:j]
		}
	}
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		r := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if r == '\\' {
				esc = true
				continue
			}
			if r == '"' {
				inStr = false
			}
			continue
		}
		switch r {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				candidate := s[start : i+1]
				if json.Valid([]byte(candidate)) {
					return candidate
				}
				return ""
			}
		}
	}
	return ""
}
