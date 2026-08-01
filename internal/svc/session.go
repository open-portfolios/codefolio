package svc

import (
	"context"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type SessionService struct {
	store   domain.SessionStore
	factory domain.SessionFactory
}

func NewSessionService(store domain.SessionStore, factory domain.SessionFactory) *SessionService {
	return &SessionService{store: store, factory: factory}
}

func (s *SessionService) New(systemPrompt string) domain.Session {
	value := s.factory.New()
	value.AddSystemMessage(systemPrompt)
	return value
}

func (s *SessionService) List(ctx context.Context, workDir string) ([]domain.SessionInfo, error) {
	return s.store.List(ctx, workDir)
}

func (s *SessionService) Load(ctx context.Context, workDir, id string) (domain.Session, domain.SessionInfo, error) {
	return s.store.Load(ctx, workDir, id)
}

func (s *SessionService) Resume(ctx context.Context, workDir string) (domain.Session, domain.SessionInfo, error) {
	return s.store.Latest(ctx, workDir)
}
