package conf

import "testing"

func TestValidateMCPServers(t *testing.T) {
	tests := []struct {
		name    string
		servers []MCPServer
		wantErr bool
	}{
		{name: "stdio", servers: []MCPServer{{Name: "context7", Transport: "stdio", Command: "npx"}}},
		{name: "streamable", servers: []MCPServer{{Name: "remote", Transport: "streamable", URL: "https://example.com/mcp"}}},
		{name: "sse", servers: []MCPServer{{Name: "legacy", Transport: "sse", URL: "https://example.com/sse"}}},
		{name: "duplicate", servers: []MCPServer{{Name: "same", Transport: "stdio", Command: "one"}, {Name: "same", Transport: "stdio", Command: "two"}}, wantErr: true},
		{name: "missing command", servers: []MCPServer{{Name: "stdio", Transport: "stdio"}}, wantErr: true},
		{name: "invalid url", servers: []MCPServer{{Name: "remote", Transport: "streamable", URL: "file:///tmp/mcp"}}, wantErr: true},
		{name: "unknown transport", servers: []MCPServer{{Name: "remote", Transport: "http", URL: "https://example.com/mcp"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMCPServers(tt.servers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMCPServers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandFieldsExpandsMCPMaps(t *testing.T) {
	t.Setenv("MCP_TOKEN", "secret")
	config := &Struct{MCPServers: []MCPServer{{Env: map[string]string{"TOKEN": "${MCP_TOKEN}"}, Headers: map[string]string{"Authorization": "Bearer ${MCP_TOKEN}"}}}}
	expandFields(config)
	if config.MCPServers[0].Env["TOKEN"] != "secret" {
		t.Fatalf("env was not expanded: %#v", config.MCPServers[0].Env)
	}
	if config.MCPServers[0].Headers["Authorization"] != "Bearer secret" {
		t.Fatalf("headers were not expanded: %#v", config.MCPServers[0].Headers)
	}
}
