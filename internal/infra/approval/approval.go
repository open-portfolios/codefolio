package approval

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type Decision uint8

const (
	AllowOnce Decision = iota + 1
	AllowSession
	Deny
	Cancelled
)

type Request struct {
	ID       string
	Key      string
	ToolID   string
	ToolName string
	Category domain.ToolCategory
	Summary  string
	Detail   string

	response chan Decision
	once     sync.Once
}

func (r *Request) Resolve(decision Decision) {
	r.once.Do(func() { r.response <- decision })
}

type Broker struct {
	requests chan<- *Request

	mu     sync.Mutex
	grants map[string]struct{}
}

func NewBroker(requests chan *Request) *Broker {
	return &Broker{requests: requests, grants: make(map[string]struct{})}
}

func (b *Broker) Authorize(ctx context.Context, invocation domain.ToolInvocation) domain.PermissionDecision {
	if err := ctx.Err(); err != nil {
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: err.Error()}
	}
	if isProtectedPath(invocation) {
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "protected Codefolio configuration"}
	}
	if invocation.Name == "Bash" && isDangerousCommand(stringArg(invocation.Args, "command")) {
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "dangerous command blocked"}
	}

	key := approvalKey(invocation)
	if b.isGranted(key) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "allowed for this session"}
	}
	if invocation.Category == domain.CategoryRead && !isOutsideWorkspace(invocation) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "read-only tool"}
	}

	request := &Request{
		ID:       invocation.ID,
		Key:      key,
		ToolID:   invocation.ID,
		ToolName: invocation.Name,
		Category: invocation.Category,
		Summary:  summarize(invocation),
		Detail:   detail(invocation),
		response: make(chan Decision, 1),
	}
	select {
	case b.requests <- request:
	case <-ctx.Done():
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "approval cancelled"}
	}

	select {
	case decision := <-request.response:
		switch decision {
		case AllowOnce:
			return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "approved once"}
		case AllowSession:
			b.mu.Lock()
			b.grants[key] = struct{}{}
			b.mu.Unlock()
			return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "approved for this session"}
		case Deny:
			return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "user denied approval"}
		default:
			return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "approval cancelled"}
		}
	case <-ctx.Done():
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "approval cancelled"}
	}
}

func (b *Broker) isGranted(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.grants[key]
	return ok
}

func approvalKey(invocation domain.ToolInvocation) string {
	if path := invocationPath(invocation); path != "" {
		return invocation.Name + ":" + path
	}
	if command := stringArg(invocation.Args, "command"); command != "" {
		return invocation.Name + ":" + command
	}
	return invocation.Name
}

func summarize(invocation domain.ToolInvocation) string {
	if command := stringArg(invocation.Args, "command"); command != "" {
		return command
	}
	if path := invocationPath(invocation); path != "" {
		return path
	}
	return invocation.Name
}

func detail(invocation domain.ToolInvocation) string {
	switch invocation.Category {
	case domain.CategoryWrite:
		return "This tool can change files."
	case domain.CategoryCommand:
		return "This command can change files or affect your environment."
	default:
		return "This action needs your approval."
	}
}

func invocationPath(invocation domain.ToolInvocation) string {
	path := stringArg(invocation.Args, "file_path")
	if path == "" {
		path = stringArg(invocation.Args, "path")
	}
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(invocation.WorkDir, path)
	}
	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return ""
	}
	return path
}

func isOutsideWorkspace(invocation domain.ToolInvocation) bool {
	path := invocationPath(invocation)
	if path == "" {
		return false
	}
	workDir, err := filepath.Abs(filepath.Clean(invocation.WorkDir))
	if err != nil {
		return true
	}
	rel, err := filepath.Rel(workDir, path)
	return err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func isProtectedPath(invocation domain.ToolInvocation) bool {
	if invocation.Category != domain.CategoryWrite {
		return false
	}
	path := invocationPath(invocation)
	if path == "" {
		return false
	}
	workDir, err := filepath.Abs(filepath.Clean(invocation.WorkDir))
	if err != nil {
		return true
	}
	config := filepath.Join(workDir, ".codefolio", "config.json")
	return path == config
}

func stringArg(args map[string]any, key string) string {
	value, _ := args[key].(string)
	return value
}

var dangerousCommandPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+-[^\n]*r[^\n]*f\s+/(?:\s|$)`),
	regexp.MustCompile(`(?i)\bmkfs(?:\.[\w-]+)?\b`),
	regexp.MustCompile(`(?i)\bdd\s+.*\bof=/dev/`),
	regexp.MustCompile(`(?i)\bchmod\s+-R\s+777\s+/`),
	regexp.MustCompile(`(?i)\b(?:curl|wget)\b[^\n]*\|\s*(?:sh|bash)\b`),
	regexp.MustCompile(`(?i)\bgit\s+push\b[^\n]*--force`),
	regexp.MustCompile(`(?i)\bgit\s+reset\s+--hard\b`),
	regexp.MustCompile(`(?i)\bgit\s+clean\s+-[^\n]*f`),
	regexp.MustCompile(`(?i)\bgit\s+checkout\s+\.(?:\s|$)`),
	regexp.MustCompile(`(?i)\bgit\s+branch\s+-D\b`),
}

func isDangerousCommand(command string) bool {
	for _, pattern := range dangerousCommandPatterns {
		if pattern.MatchString(command) {
			return true
		}
	}
	return false
}

var _ domain.Authorizer = (*Broker)(nil)
