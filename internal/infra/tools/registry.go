package tools

import (
	"strings"

	"github.com/open-portfolios/codefolio/internal/conf"
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
	protocol        string
}

func newRegistry() *Registry {
	return &Registry{
		tools:           make(map[string]domain.Tool),
		discoveredTools: make(map[string]bool),
	}
}

func NewRegistry(askUserCh chan askuser.Request, cfg *conf.Global) domain.ToolRegistry {
	fsc := file.NewStateCache()
	reg := newRegistry()
	reg.protocol = cfg.Protocol
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
		base := t.Schema()
		if r.protocol == "openai" {
			schemas = append(schemas, map[string]any{
				"type":        "function",
				"name":        base["name"],
				"description": base["description"],
				"parameters":  base["input_schema"],
			})
		} else {
			schemas = append(schemas, base)
		}
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
			base := t.Schema()
			if r.protocol == "openai" {
				matches = append(matches, map[string]any{
					"type":        "function",
					"name":        base["name"],
					"description": base["description"],
					"parameters":  base["input_schema"],
				})
			} else {
				matches = append(matches, base)
			}
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
			base := t.Schema()
			if r.protocol == "openai" {
				matches = append(matches, map[string]any{
					"type":        "function",
					"name":        base["name"],
					"description": base["description"],
					"parameters":  base["input_schema"],
				})
			} else {
				matches = append(matches, base)
			}
		}
	}
	return matches
}
