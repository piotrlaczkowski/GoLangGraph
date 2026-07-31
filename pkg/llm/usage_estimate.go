// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package llm

import "strings"

// EstimateTokens approximates token count with a tiktoken-lite heuristic:
// ~4 chars/token for ASCII-heavy text, with a small message overhead.
// Used when providers omit usage (common on stream early_exit).
func EstimateTokens(text string) int {
	n := len(text)
	if n == 0 {
		return 0
	}
	r := len([]rune(text))
	if r > n {
		n = r
	}
	tok := (n + 3) / 4
	if tok < 1 {
		tok = 1
	}
	return tok
}

// EstimateMessagesTokens estimates prompt tokens for a chat transcript.
func EstimateMessagesTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateTokens(msg.Content) + 4
		if msg.Name != "" {
			total += EstimateTokens(msg.Name)
		}
		for _, tc := range msg.ToolCalls {
			total += EstimateTokens(tc.Function.Name) + EstimateTokens(tc.Function.Arguments) + 8
		}
		if strings.TrimSpace(msg.ToolCallID) != "" {
			total += 4
		}
	}
	return total
}

// EstimateCompletionTokens estimates tokens for assistant content + tool calls.
func EstimateCompletionTokens(content string, toolCalls []ToolCall) int {
	total := EstimateTokens(content)
	for _, tc := range toolCalls {
		total += EstimateTokens(tc.Function.Name) + EstimateTokens(tc.Function.Arguments) + 8
	}
	if total == 0 {
		total = 1
	}
	return total
}

// EnsureUsage fills Usage when the provider omitted counts (early_exit streams).
// Returns true when values were estimated.
func EnsureUsage(resp *CompletionResponse, req CompletionRequest) bool {
	if resp == nil {
		return false
	}
	if resp.Usage.TotalTokens > 0 || resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 {
		if resp.Usage.TotalTokens == 0 {
			resp.Usage.TotalTokens = resp.Usage.PromptTokens + resp.Usage.CompletionTokens
		}
		return false
	}
	prompt := EstimateMessagesTokens(req.Messages)
	for _, t := range req.Tools {
		prompt += EstimateTokens(t.Function.Name) + EstimateTokens(t.Function.Description) + 16
	}
	var content string
	var tools []ToolCall
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
		tools = resp.Choices[0].Message.ToolCalls
	}
	completion := EstimateCompletionTokens(content, tools)
	resp.Usage = Usage{
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
	}
	if resp.Metadata == nil {
		resp.Metadata = map[string]interface{}{}
	}
	resp.Metadata["usage_estimated"] = true
	return true
}
