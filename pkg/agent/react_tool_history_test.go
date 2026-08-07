// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func TestSanitizeToolPairingInjectsMissingTools(t *testing.T) {
	msgs := []llm.Message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "c1", Type: "function"}}},
		{Role: "user", Content: "finalize"},
	}
	out := sanitizeToolPairing(msgs)
	if len(out) < 3 || out[2].Role != "tool" || out[2].ToolCallID != "c1" {
		t.Fatalf("expected stub tool after unpaired assistant, got %+v", out)
	}
}

func TestShouldActRoutingTargets(t *testing.T) {
	a := &BaseAgent{config: &AgentConfig{MaxIterations: 3}}
	st := core.NewBaseState()
	st.Set("pending_tool_calls", []llm.ToolCall{{ID: "c1"}})
	next, err := a.shouldAct(context.Background(), st)
	if err != nil || next != "act" {
		t.Fatalf("with tools → act, got %q err=%v", next, err)
	}
	st.Set("pending_tool_calls", []llm.ToolCall{})
	next, err = a.shouldAct(context.Background(), st)
	if err != nil || next != "finalize" {
		t.Fatalf("no tools → finalize, got %q err=%v", next, err)
	}
}

func TestActObserveAppendsToolMessages(t *testing.T) {
	reg := tools.NewToolRegistry()
	echo := tools.NewGenericTool(
		"echo",
		"echo text",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			if v, ok := args["text"]; ok {
				return v, nil
			}
			return "ok", nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{"type": "string"},
			},
		},
	)
	reg.RegisterTool(echo)

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	a := &BaseAgent{
		config: &AgentConfig{
			ID:       "t",
			Name:     "t",
			Type:     AgentTypeReAct,
			Model:    "m",
			Provider: "p",
		},
		conversation: llm.NewConversationHistory(),
		toolNode:     NewToolNode(reg),
		logger:       log,
	}
	a.conversation.AddMessage(llm.Message{Role: "user", Content: "read greet.go"})
	a.conversation.AddMessage(llm.Message{
		Role: "assistant",
		ToolCalls: []llm.ToolCall{{
			ID:   "call_1",
			Type: "function",
			Function: llm.FunctionCall{
				Name:      "echo",
				Arguments: `{"text":"hello"}`,
			},
		}},
	})

	state := core.NewBaseState()
	state.Set("pending_tool_calls", []llm.ToolCall{{
		ID:   "call_1",
		Type: "function",
		Function: llm.FunctionCall{
			Name:      "echo",
			Arguments: `{"text":"hello"}`,
		},
	}})
	state.Set("reasoning", "")
	state.Set("iteration", 0)

	ctx := context.Background()
	var err error
	state, err = a.actNode(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	state, err = a.observeNode(ctx, state)
	if err != nil {
		t.Fatal(err)
	}

	msgs := a.conversation.GetMessages()
	if len(msgs) < 3 {
		t.Fatalf("expected user+assistant+tool, got %d: %+v", len(msgs), msgs)
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) == 0 {
		t.Fatalf("assistant tool_calls missing: %+v", msgs[1])
	}
	if msgs[2].Role != "tool" || msgs[2].ToolCallID != "call_1" {
		t.Fatalf("expected role=tool with call_1, got %+v", msgs[2])
	}
	for _, m := range msgs[2:] {
		if m.Role == "assistant" && strings.HasPrefix(m.Content, "Observation:") {
			t.Fatal("must not append assistant Observation after tool messages")
		}
	}
}
