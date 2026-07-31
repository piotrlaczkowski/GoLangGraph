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
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

func init() {
	// Prefer offline BPE dictionaries so estimation works without network.
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
}

var (
	encOnce sync.Once
	enc     *tiktoken.Tiktoken
	encOK   bool
)

func tokenizer() (*tiktoken.Tiktoken, bool) {
	encOnce.Do(func() {
		tke, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return
		}
		enc = tke
		encOK = true
	})
	return enc, encOK
}

// TokenizerAvailable reports whether a real tiktoken encoding is loaded.
func TokenizerAvailable() bool {
	_, ok := tokenizer()
	return ok
}

// EstimateTokens counts tokens with cl100k_base (tiktoken) when available,
// otherwise falls back to a ~4 chars/token heuristic. Used when providers
// omit usage (common on stream early_exit).
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	if tke, ok := tokenizer(); ok {
		return len(tke.Encode(text, nil, nil))
	}
	return estimateTokensHeuristic(text)
}

// estimateTokensHeuristic is the chars/4 fallback when tiktoken is unavailable.
func estimateTokensHeuristic(text string) int {
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
	if TokenizerAvailable() {
		resp.Metadata["usage_tokenizer"] = "cl100k_base"
	} else {
		resp.Metadata["usage_tokenizer"] = "heuristic"
	}
	return true
}
