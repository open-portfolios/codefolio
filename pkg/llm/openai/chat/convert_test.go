package chat

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/open-portfolios/codefolio/pkg/llm"
)

func TestDeveloperMessageUsesSystemRoleForOpenAICompatibleAPIs(t *testing.T) {
	converter := &messageConverter{}
	if err := (llm.DeveloperMessage{Content: "workspace instructions"}).Accept(converter); err != nil {
		t.Fatalf("convert developer message: %v", err)
	}
	if len(converter.msgs) != 1 {
		t.Fatalf("expected one converted message, got %d", len(converter.msgs))
	}

	encoded, err := json.Marshal(converter.msgs[0])
	if err != nil {
		t.Fatalf("marshal converted message: %v", err)
	}
	if !strings.Contains(string(encoded), `"role":"system"`) {
		t.Fatalf("expected system role, got %s", encoded)
	}
	if strings.Contains(string(encoded), `"role":"developer"`) {
		t.Fatalf("developer role must not reach OpenAI-compatible APIs: %s", encoded)
	}
}
