package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	modelcontextprotocol "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/open-portfolios/codefolio/internal/conf"
)

type RemoteTool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type CallResult struct {
	Output  string
	IsError bool
}

type Client interface {
	ListTools(context.Context) ([]RemoteTool, error)
	CallTool(context.Context, string, map[string]any) (CallResult, error)
	Close() error
}

type ClientFactory interface {
	Start(context.Context, conf.MCPServer) (Client, error)
}

type sdkClientFactory struct{}

func NewSDKClientFactory() ClientFactory { return sdkClientFactory{} }

func (sdkClientFactory) Start(ctx context.Context, cfg conf.MCPServer) (Client, error) {
	client := modelcontextprotocol.NewClient(&modelcontextprotocol.Implementation{Name: "codefolio", Version: "0.1.0"}, nil)
	transport, err := newTransport(cfg)
	if err != nil {
		return nil, err
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connect MCP server %q: %w", cfg.Name, err)
	}
	return &sdkClient{session: session}, nil
}

func newTransport(cfg conf.MCPServer) (modelcontextprotocol.Transport, error) {
	if cfg.Transport == "stdio" {
		cmd := exec.Command(cfg.Command, cfg.Args...)
		cmd.Env = os.Environ()
		for key, value := range cfg.Env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
		// Node-based MCP servers may emit terminal control queries to a TTY stderr.
		// Keep their diagnostics out of the TUI input path.
		cmd.Stderr = io.Discard
		return &modelcontextprotocol.CommandTransport{Command: cmd}, nil
	}

	httpClient := &http.Client{Transport: headerRoundTripper{base: http.DefaultTransport, headers: cfg.Headers}}
	switch cfg.Transport {
	case "streamable":
		return &modelcontextprotocol.StreamableClientTransport{Endpoint: cfg.URL, HTTPClient: httpClient}, nil
	case "sse":
		return &modelcontextprotocol.SSEClientTransport{Endpoint: cfg.URL, HTTPClient: httpClient}, nil
	default:
		return nil, fmt.Errorf("unsupported MCP transport %q", cfg.Transport)
	}
}

type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (h headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for key, value := range h.headers {
		clone.Header.Set(key, value)
	}
	return h.base.RoundTrip(clone)
}

type sdkClient struct {
	session *modelcontextprotocol.ClientSession
}

func (c *sdkClient) ListTools(ctx context.Context) ([]RemoteTool, error) {
	result, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	tools := make([]RemoteTool, 0, len(result.Tools))
	for _, tool := range result.Tools {
		schema, err := inputSchema(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("tool %q: %w", tool.Name, err)
		}
		tools = append(tools, RemoteTool{Name: tool.Name, Description: tool.Description, InputSchema: schema})
	}
	return tools, nil
}

func inputSchema(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}, nil
	}
	schema, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input schema is not an object")
	}
	if schema["type"] != nil && schema["type"] != "object" {
		return nil, fmt.Errorf("input schema must be an object")
	}
	return schema, nil
}

func (c *sdkClient) CallTool(ctx context.Context, name string, args map[string]any) (CallResult, error) {
	result, err := c.session.CallTool(ctx, &modelcontextprotocol.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return CallResult{}, err
	}
	parts := make([]string, 0, len(result.Content))
	for _, content := range result.Content {
		if text, ok := content.(*modelcontextprotocol.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	output := strings.Join(parts, "\n")
	if output == "" {
		output = "(MCP tool returned no text output)"
	}
	return CallResult{Output: output, IsError: result.IsError}, nil
}

func (c *sdkClient) Close() error { return c.session.Close() }
