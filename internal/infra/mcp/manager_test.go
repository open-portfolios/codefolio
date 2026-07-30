package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestManagerDiscoversToolsOnlyOnce(t *testing.T) {
	client := &fakeClient{tools: []RemoteTool{{Name: "search", Description: "Search issues", InputSchema: map[string]any{"type": "object", "properties": map[string]any{}}}}}
	factory := &fakeFactory{client: client}
	manager := NewManagerWithFactory([]conf.MCPServer{{Name: "github", Transport: "stdio", Command: "fake"}}, factory)

	tools, err := manager.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name() != "mcp__github__search" {
		t.Fatalf("unexpected discovered tools: %#v", tools)
	}
	tools, err = manager.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 0 || factory.starts != 1 || client.listCalls != 1 {
		t.Fatalf("discovery repeated: tools=%d starts=%d lists=%d", len(tools), factory.starts, client.listCalls)
	}
}

func TestManagerKeepsOtherServersWhenOneFails(t *testing.T) {
	good := &fakeClient{tools: []RemoteTool{{Name: "search", InputSchema: map[string]any{"type": "object"}}}}
	factory := &fakeFactory{clients: map[string]*fakeClient{"good": good}, errors: map[string]error{"bad": errors.New("unavailable")}}
	manager := NewManagerWithFactory([]conf.MCPServer{{Name: "bad", Transport: "stdio", Command: "bad"}, {Name: "good", Transport: "stdio", Command: "good"}}, factory)
	tools, err := manager.Discover(context.Background())
	if err != nil {
		t.Fatalf("partial discovery should succeed: %v", err)
	}
	if len(tools) != 1 || tools[0].Name() != "mcp__good__search" {
		t.Fatalf("unexpected discovered tools: %#v", tools)
	}
	summary := manager.Summary()
	if summary.Configured != 2 || summary.Ready != 1 || summary.Unavailable != 1 || summary.Tools != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestManagerSummaryCountsEmptyCatalogAsReady(t *testing.T) {
	manager := NewManagerWithFactory([]conf.MCPServer{{Name: "empty", Transport: "stdio", Command: "fake"}}, &fakeFactory{client: &fakeClient{}})
	if _, err := manager.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	summary := manager.Summary()
	if summary.Ready != 1 || summary.Tools != 0 || summary.Unavailable != 0 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestManagerCloseClosesClients(t *testing.T) {
	client := &fakeClient{tools: []RemoteTool{{Name: "search", InputSchema: map[string]any{"type": "object"}}}}
	manager := NewManagerWithFactory([]conf.MCPServer{{Name: "github", Transport: "stdio", Command: "fake"}}, &fakeFactory{client: client})
	if _, err := manager.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if !client.closed {
		t.Fatal("client was not closed")
	}
	if _, err := manager.Discover(context.Background()); err == nil {
		t.Fatal("expected closed manager error")
	}
}

func TestToolMapsErrorsAndTruncatesOutput(t *testing.T) {
	client := &fakeClient{call: CallResult{Output: "server failed", IsError: true}}
	tool := NewTool("github-server", RemoteTool{Name: "search.issues", InputSchema: map[string]any{"type": "object"}}, client)
	if tool.Name() != "mcp__github_server__search_issues" {
		t.Fatalf("unexpected tool name %q", tool.Name())
	}
	result := tool.Execute(context.Background(), map[string]any{"query": "bug"})
	if !result.IsError || result.Outcome != domain.ToolOutcomeFailed || result.Output != "Error: server failed" {
		t.Fatalf("unexpected tool result: %#v", result)
	}
}

type fakeFactory struct {
	client  *fakeClient
	clients map[string]*fakeClient
	errors  map[string]error
	starts  int
}

func (f *fakeFactory) Start(_ context.Context, cfg conf.MCPServer) (Client, error) {
	f.starts++
	if err := f.errors[cfg.Name]; err != nil {
		return nil, err
	}
	if client := f.clients[cfg.Name]; client != nil {
		return client, nil
	}
	return f.client, nil
}

type fakeClient struct {
	tools     []RemoteTool
	call      CallResult
	listCalls int
	closed    bool
}

func (c *fakeClient) ListTools(context.Context) ([]RemoteTool, error) {
	c.listCalls++
	return c.tools, nil
}

func (c *fakeClient) CallTool(context.Context, string, map[string]any) (CallResult, error) {
	return c.call, nil
}

func (c *fakeClient) Close() error {
	c.closed = true
	return nil
}
