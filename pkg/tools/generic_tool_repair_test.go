// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package tools

import (
	"context"
	"testing"
)

func TestGenericToolRepairsTrailingComma(t *testing.T) {
	var got map[string]interface{}
	tool := NewGenericTool("echo", "echo args", func(_ context.Context, args map[string]interface{}) (interface{}, error) {
		got = args
		return "ok", nil
	}, map[string]interface{}{"type": "object"})
	_, err := tool.Execute(context.Background(), `{"path":"a.go","content":"x",}`)
	if err != nil {
		t.Fatal(err)
	}
	if got["path"] != "a.go" {
		t.Fatalf("%v", got)
	}
}

func TestGenericToolRepairsSingleQuotes(t *testing.T) {
	var got map[string]interface{}
	tool := NewGenericTool("echo", "echo args", func(_ context.Context, args map[string]interface{}) (interface{}, error) {
		got = args
		return "ok", nil
	}, map[string]interface{}{"type": "object"})
	_, err := tool.Execute(context.Background(), `{'path':'b.go'}`)
	if err != nil {
		t.Fatal(err)
	}
	if got["path"] != "b.go" {
		t.Fatalf("%v", got)
	}
}
