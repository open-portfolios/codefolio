package tools

import (
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/infra/tools/bash"
	"github.com/open-portfolios/codefolio/internal/infra/tools/file"
	"github.com/open-portfolios/codefolio/internal/infra/tools/search"
	"github.com/open-portfolios/codefolio/internal/infra/tools/toolsearch"
)

type Registry struct {
	tools           map[string]domain.Tool
	discoveredTools map[string]bool
}

func newRegistry() *Registry {
	return &Registry{
		tools:           make(map[string]domain.Tool),
		discoveredTools: make(map[string]bool),
	}
}

func NewRegistry(askUserCh chan askuser.Request) domain.ToolRegistry {
	fsc := file.NewStateCache()
	reg := newRegistry()
	reg.Register(&file.Reader{StateCache: fsc})
	reg.Register(&file.Writer{StateCache: fsc})
	reg.Register(&file.Editor{StateCache: fsc})
	reg.Register(&bash.Tool{})
	reg.Register(&search.Glob{})
	reg.Register(&search.Grep{})
	reg.Register(&askuser.Tool{RequestCh: askUserCh})
	reg.Register(&toolsearch.ToolSearch{Registry: reg})
	return reg
}

func (r *Registry) MarkDiscovered(name string) {
	r.discoveredTools[name] = true
}

func (r *Registry) IsDiscovered(name string) bool {
	return r.discoveredTools[name]
}

func (r *Registry) Register(t domain.Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) domain.Tool {
	return r.tools[name]
}

func (r *Registry) ListTools() []domain.Tool {
	result := make([]domain.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

func isDeferred(t domain.Tool) bool {
	if dt, ok := t.(domain.DeferrableTool); ok {
		return dt.ShouldDefer()
	}
	return false
}

func (r *Registry) GetAllSchemas() []map[string]any {
	schemas := make([]map[string]any, 0, len(r.tools))
	for _, t := range r.tools {
		if isDeferred(t) && !r.discoveredTools[t.Name()] {
			continue
		}
		schemas = append(schemas, t.Schema())
	}
	return schemas
}

func (r *Registry) GetDeferredToolNames() []string {
	var names []string
	for _, t := range r.tools {
		if isDeferred(t) && !r.discoveredTools[t.Name()] {
			names = append(names, t.Name())
		}
	}
	return names
}

func (r *Registry) GetDeferredTools() []domain.Tool {
	var result []domain.Tool
	for _, t := range r.tools {
		if isDeferred(t) {
			result = append(result, t)
		}
	}
	return result
}

func (r *Registry) SearchDeferred(query string, maxResults int) []map[string]any {
	query = strings.ToLower(query)
	var matches []map[string]any
	for _, t := range r.tools {
		if !isDeferred(t) {
			continue
		}
		name := strings.ToLower(t.Name())
		desc := strings.ToLower(t.Description())
		if strings.Contains(name, query) || strings.Contains(desc, query) {
			matches = append(matches, t.Schema())
			if len(matches) >= maxResults {
				break
			}
		}
	}
	return matches
}

func (r *Registry) FindDeferredByNames(names []string) []map[string]any {
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[strings.ToLower(n)] = true
	}
	var matches []map[string]any
	for _, t := range r.tools {
		if nameSet[strings.ToLower(t.Name())] {
			matches = append(matches, t.Schema())
		}
	}
	return matches
}
