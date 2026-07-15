package infra

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type environmentService struct{}

func NewEnvironmentService() domain.EnvironmentService {
	return &environmentService{}
}

func (s *environmentService) Detect(workDir string) domain.PromptEnvironment {
	env := domain.PromptEnvironment{
		WorkDir: workDir,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Shell:   detectShell(),
		Date:    time.Now().Format("2006-01-02"),
	}

	if out, err := exec.Command("git", "-C", workDir, "rev-parse", "--is-inside-work-tree").Output(); err == nil && strings.TrimSpace(string(out)) == "true" {
		env.IsGitRepo = true
		if branch, err := exec.Command("git", "-C", workDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			env.GitBranch = strings.TrimSpace(string(branch))
		}
	}

	return env
}

func detectShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	if s := os.Getenv("COMSPEC"); s != "" {
		return s
	}
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	return "bash"
}
