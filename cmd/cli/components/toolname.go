package components

import "strings"

func displayToolName(name string) string {
	if !isMCPTool(name) {
		return name
	}
	parts := strings.Split(name, "__")
	if len(parts) < 3 {
		return name
	}
	return "MCP " + parts[1] + " / " + strings.Join(parts[2:], "__")
}
