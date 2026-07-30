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
	if invocation.Name == "Bash" && isDangerousCommand(stringArg(invocation.Args, "command")) {
		return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "dangerous command blocked"}
	}
	if invocation.Profile == domain.ProfilePlan {
		return b.planDecision(ctx, invocation)
	}

	if isOutsideWorkspace(invocation) {
		return b.requestApproval(ctx, invocation, "external workspace access requires approval")
	}
	if isSensitiveInvocation(invocation) {
		return b.requestApproval(ctx, invocation, "sensitive Codefolio configuration requires approval")
	}
	if invocation.Category == domain.CategoryRead || invocation.Category == domain.CategoryWrite || isMCPTool(invocation) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "allowed by build profile"}
	}
	if invocation.Name == "Bash" && isFileModifyingCommand(stringArg(invocation.Args, "command")) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "file-modifying command allowed by build profile"}
	}

	key := approvalKey(invocation)
	if b.isGranted(key) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "allowed for this session"}
	}
	return b.requestApproval(ctx, invocation, "")
}

func (b *Broker) requestApproval(ctx context.Context, invocation domain.ToolInvocation, reason string) domain.PermissionDecision {
	key := approvalKey(invocation)
	if b.isGranted(key) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "allowed for this session"}
	}
	request := &Request{
		ID:       invocation.ID,
		Key:      key,
		ToolID:   invocation.ID,
		ToolName: invocation.Name,
		Category: invocation.Category,
		Summary:  summarize(invocation),
		Detail:   detail(invocation, reason),
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
	prefix := string(invocation.Profile) + ":"
	if path := invocationPath(invocation); path != "" {
		return prefix + invocation.Name + ":" + path
	}
	if command := stringArg(invocation.Args, "command"); command != "" {
		return prefix + invocation.Name + ":" + command
	}
	return prefix + invocation.Name
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

func detail(invocation domain.ToolInvocation, reason string) string {
	if reason != "" {
		return reason
	}
	switch invocation.Category {
	case domain.CategoryWrite:
		return "This tool can change files."
	case domain.CategoryCommand:
		return "This command can change files or affect your environment."
	default:
		return "This action needs your approval."
	}
}

func (b *Broker) planDecision(ctx context.Context, invocation domain.ToolInvocation) domain.PermissionDecision {
	if invocation.Category == domain.CategoryRead {
		if isOutsideWorkspace(invocation) {
			return b.requestApproval(ctx, invocation, "external workspace access requires approval")
		}
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "allowed by plan profile"}
	}
	if isPlanFile(invocation) {
		return domain.PermissionDecision{Effect: domain.PermissionAllow, Reason: "plan artifact write allowed"}
	}
	return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "plan profile blocks this action"}
}

func isPlanFile(invocation domain.ToolInvocation) bool {
	if invocation.Category != domain.CategoryWrite || invocation.PlanFile == "" {
		return false
	}
	path := invocationPath(invocation)
	planFile, err := filepath.Abs(filepath.Clean(invocation.PlanFile))
	return err == nil && path != "" && path == planFile
}

func isMCPTool(invocation domain.ToolInvocation) bool {
	return strings.HasPrefix(invocation.Name, "mcp__")
}

func isFileModifyingCommand(command string) bool {
	command = strings.ToLower(command)
	for _, token := range []string{
		"gofmt -w",
		"go fmt",
		"prettier --write",
		"prettier -w",
		"eslint --fix",
		"sed -i",
		"touch ",
		"mkdir ",
		"cp ",
		"mv ",
		"rm ",
		">",
	} {
		if strings.Contains(command, token) {
			return true
		}
	}
	return false
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
	if path == "" && invocation.Name == "Bash" {
		return bashReferencesExternalPath(stringArg(invocation.Args, "command"), invocation.WorkDir)
	}
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

func bashReferencesExternalPath(command, workDir string) bool {
	for field := range strings.FieldsSeq(command) {
		path := strings.Trim(field, "'\";|&()")
		if !filepath.IsAbs(path) {
			continue
		}
		invocation := domain.ToolInvocation{Args: map[string]any{"path": path}, WorkDir: workDir}
		if isOutsideWorkspace(invocation) {
			return true
		}
	}
	return false
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

func isSensitiveInvocation(invocation domain.ToolInvocation) bool {
	if isProtectedPath(invocation) {
		return true
	}
	if invocation.Name != "Bash" {
		return false
	}
	return strings.Contains(strings.ReplaceAll(stringArg(invocation.Args, "command"), "\\", "/"), ".codefolio/config.json")
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
