package domain

import (
	"context"
	"time"
)

type SessionInfo struct {
	ID           string
	Title        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MessageCount int
}

type SessionStore interface {
	List(ctx context.Context, workDir string) ([]SessionInfo, error)
	Load(ctx context.Context, workDir, id string) (Session, SessionInfo, error)
	Latest(ctx context.Context, workDir string) (Session, SessionInfo, error)
}

type SessionFactory interface {
	New() Session
}
