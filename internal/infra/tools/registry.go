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

func NewRegistry() *Registry {
	return &Registry{
		tools:           make(map[string]domain.Tool),
		discoveredTools: make(map[string]bool),
	}
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

func (r *Registry) GetAllSchemas(protocol string) []map[string]any {
	schemas := make([]map[string]any, 0, len(r.tools))
	for _, t := range r.tools {
		if isDeferred(t) && !r.discoveredTools[t.Name()] {
			continue
		}
		base := t.Schema()
		if protocol == "openai" {
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

func (r *Registry) SearchDeferred(query string, maxResults int, protocol string) []map[string]any {
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
			if protocol == "openai" {
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

func (r *Registry) FindDeferredByNames(names []string, protocol string) []map[string]any {
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[strings.ToLower(n)] = true
	}
	var matches []map[string]any
	for _, t := range r.tools {
		if nameSet[strings.ToLower(t.Name())] {
			base := t.Schema()
			if protocol == "openai" {
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

type DefaultTools struct {
	Registry *Registry
	Writer   *file.Writer
	Editor   *file.Editor
}

func CreateDefaultRegistry(askUserCh chan askuser.Request, protocol string) *Registry {
	dt := CreateDefaultTools(askUserCh, protocol)
	return dt.Registry
}

func CreateDefaultTools(askUserCh chan askuser.Request, protocol string) DefaultTools {
	fsc := file.NewStateCache()
	wf := &file.Writer{StateCache: fsc}
	ef := &file.Editor{StateCache: fsc}
	reg := NewRegistry()
	reg.Register(&file.Reader{StateCache: fsc})
	reg.Register(wf)
	reg.Register(ef)
	reg.Register(&bash.Tool{})
	reg.Register(&search.Glob{})
	reg.Register(&search.Grep{})
	reg.Register(&askuser.Tool{RequestCh: askUserCh})
	reg.Register(&toolsearch.ToolSearch{Registry: reg, Protocol: protocol})
	return DefaultTools{Registry: reg, Writer: wf, Editor: ef}
}
