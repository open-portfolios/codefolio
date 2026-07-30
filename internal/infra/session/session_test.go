package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestSessionCopiesMessagesAndPersistsCheckpoints(t *testing.T) {
	value := New().(*Session)
	root := t.TempDir()
	value.ConfigurePersistence(root)
	value.AddUserMessage("remember this turn")
	value.StartAssistantMessage()
	value.AddToolCallToAssistant(domain.ToolCall{ID: "tool-1", Name: "ReadFile"})
	value.UpdateToolCallInput("tool-1", `{"file_path":"README.md"}`)
	value.FinishAssistantMessage()

	messages := value.Messages()
	messages[0].Content = "mutated copy"
	if got := value.Messages()[0].Content; got != "remember this turn" {
		t.Fatalf("session exposed mutable message data: %q", got)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".codefolio", "sessions"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one session log, entries=%d err=%v", len(entries), err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".codefolio", "sessions", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "tool-1") || !strings.Contains(string(content), "README.md") {
		t.Fatalf("expected complete tool checkpoint, got %s", content)
	}
}

func TestLoadRestoresLatestMessageCheckpoint(t *testing.T) {
	value := New().(*Session)
	root := t.TempDir()
	value.ConfigurePersistence(root)
	value.AddUserMessage("continue this session")
	value.StartAssistantMessage()
	value.AppendDelta("partial")
	value.FinishAssistantMessage()
	entries, err := os.ReadDir(filepath.Join(root, ".codefolio", "sessions"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected session log: entries=%d err=%v", len(entries), err)
	}
	loaded, err := Load(filepath.Join(root, ".codefolio", "sessions", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	messages := loaded.Messages()
	if len(messages) != 2 || messages[1].Content != "partial" || messages[1].Streaming {
		t.Fatalf("unexpected restored messages: %#v", messages)
	}
}

func TestFinishAssistantMessageDoesNotDuplicateFinalCheckpoint(t *testing.T) {
	value := New().(*Session)
	root := t.TempDir()
	value.ConfigurePersistence(root)
	value.StartAssistantMessage()
	value.AppendDelta("final answer")
	value.FinishAssistantMessage()
	value.FinishAssistantMessage()

	entries, err := os.ReadDir(filepath.Join(root, ".codefolio", "sessions"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected session log: entries=%d err=%v", len(entries), err)
	}
	file, err := os.Open(filepath.Join(root, ".codefolio", "sessions", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	finalCheckpoints := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record logRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		if record.Message.ID == "a-1" && !record.Message.Streaming {
			finalCheckpoints++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if finalCheckpoints != 1 {
		t.Fatalf("expected one final assistant checkpoint, got %d", finalCheckpoints)
	}
}
