package chat

import (
	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/open-portfolios/codefolio/pkg/stdx"
)

var (
	_ llm.Delta = stdx.Zero[completionsDelta]()
)

type completionsDelta struct {
	role    string
	content string
	usage   int64
}

func newCompletionsDelta(role, content string, usage int64) completionsDelta {
	return completionsDelta{
		role:    role,
		content: content,
		usage:   usage,
	}
}

func (c completionsDelta) Role() string    { return c.role }
func (c completionsDelta) Content() string { return c.content }
func (c completionsDelta) Usage() int64    { return c.usage }
