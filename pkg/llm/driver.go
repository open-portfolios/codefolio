package llm

import "context"

type Driver interface {
	Stream(ctx context.Context, messages []Message, options ...StreamOption) (<-chan Delta, <-chan error)
}

type StreamConf struct {
	Model        string
	ChanCapacity int
	MaxTokens    int64
}

func CollectStreamOptions(options ...StreamOption) *StreamConf {
	s := &StreamConf{
		ChanCapacity: 64,
		MaxTokens:    4096,
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

type StreamOption func(*StreamConf)

func WithModel(name string) StreamOption {
	return func(s *StreamConf) { s.Model = name }
}

func WithChanCapacity(capacity int) StreamOption {
	return func(s *StreamConf) { s.ChanCapacity = capacity }
}

func WithMaxTokens(tokens int64) StreamOption {
	return func(s *StreamConf) { s.MaxTokens = tokens }
}
