package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type Store struct{}

func NewStore() *Store { return &Store{} }

func (s *Store) List(ctx context.Context, workDir string) ([]domain.SessionInfo, error) {
	dir := filepath.Join(workDir, ".codefolio", "sessions")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []domain.SessionInfo
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		loaded, err := Load(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, metadata(strings.TrimSuffix(entry.Name(), ".jsonl"), loaded, fileInfo.ModTime()))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	return result, nil
}

func (s *Store) Load(ctx context.Context, workDir, id string) (domain.Session, domain.SessionInfo, error) {
	if id == "" || strings.ContainsAny(id, `\\/`) {
		return nil, domain.SessionInfo{}, fmt.Errorf("invalid session ID")
	}
	items, err := s.List(ctx, workDir)
	if err != nil {
		return nil, domain.SessionInfo{}, err
	}
	var match *domain.SessionInfo
	for i := range items {
		if items[i].ID == id || strings.HasPrefix(items[i].ID, id) {
			if match != nil {
				return nil, domain.SessionInfo{}, fmt.Errorf("session prefix %q is ambiguous", id)
			}
			match = &items[i]
		}
	}
	if match == nil {
		return nil, domain.SessionInfo{}, fmt.Errorf("session %q was not found", id)
	}
	loaded, err := Load(filepath.Join(workDir, ".codefolio", "sessions", match.ID+".jsonl"))
	return loaded, *match, err
}

func (s *Store) Latest(ctx context.Context, workDir string) (domain.Session, domain.SessionInfo, error) {
	items, err := s.List(ctx, workDir)
	if err != nil {
		return nil, domain.SessionInfo{}, err
	}
	if len(items) == 0 {
		return nil, domain.SessionInfo{}, fmt.Errorf("no resumable sessions")
	}
	return s.Load(ctx, workDir, items[0].ID)
}

func metadata(id string, value domain.Session, updatedAt time.Time) domain.SessionInfo {
	result := domain.SessionInfo{ID: id, UpdatedAt: updatedAt, MessageCount: value.MessageCount()}
	for _, message := range value.Messages() {
		if message.Role == llm.RoleUser && strings.TrimSpace(message.Content) != "" {
			result.Title = strings.TrimSpace(strings.Split(message.Content, "\n")[0])
			break
		}
	}
	if result.Title == "" {
		result.Title = "New session"
	}
	if len(result.Title) > 56 {
		result.Title = result.Title[:56] + "..."
	}
	return result
}

var _ domain.SessionStore = (*Store)(nil)
