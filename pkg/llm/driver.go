package llm

import "context"

type Driver interface {
	Stream(ctx context.Context, messages []Message) (<-chan Delta, <-chan error)
}
