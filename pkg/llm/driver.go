package llm

import "context"

type Driver interface {
	Stream(ctx context.Context, messages []Message, options ...StreamOption) (<-chan Delta, <-chan error)
}

type StreamOption func(*streamConf)

func WithModel(name string) StreamOption {
	return func(s *streamConf) { s.Model = name }
}

type streamConf struct {
	Model string
}

func CollectStreamOptions(options ...StreamOption) *streamConf {
	s := &streamConf{}
	for _, opt := range options {
		opt(s)
	}
	return s
}
