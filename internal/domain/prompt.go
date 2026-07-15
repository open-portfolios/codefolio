package domain

type PromptEnvironment struct {
	WorkDir   string
	OS        string
	Arch      string
	Shell     string
	IsGitRepo bool
	GitBranch string
	Model     string
	Date      string
}

type PromptService interface {
	BuildSystemPrompt(env PromptEnvironment) string
	BuildPlanModeReminder(planFilePath string, planExists bool, iteration int) string
}

type EnvironmentService interface {
	Detect(workDir string) PromptEnvironment
}
