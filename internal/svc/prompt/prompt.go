package prompt

import (
	"embed"
	"strings"
	"text/template"

	"github.com/open-portfolios/codefolio/internal/domain"
)

//go:embed templates/*
var templateFS embed.FS

const planModeReminderInterval = 5

type promptService struct {
	sections           []string
	envTmpl            *template.Template
	modelInfoTmpl      *template.Template
	planFileExistsTmpl *template.Template
	planFileNewTmpl    *template.Template
	planFullTmpl       *template.Template
	planSparseTmpl     *template.Template
}

func NewPromptService() domain.PromptService {
	s := &promptService{}

	sectionNames := []string{
		"identity", "system", "doing_tasks", "executing_actions",
		"using_tools", "tone_style", "output_efficiency",
	}
	for _, name := range sectionNames {
		content, err := templateFS.ReadFile("templates/" + name + ".txt")
		if err != nil {
			panic(err)
		}
		s.sections = append(s.sections, strings.TrimSpace(string(content)))
	}

	s.envTmpl = loadTemplate("templates/environment.txt")
	s.modelInfoTmpl = loadTemplate("templates/model_info.txt")
	s.planFileExistsTmpl = loadTemplate("templates/plan_file_info_exists.txt")
	s.planFileNewTmpl = loadTemplate("templates/plan_file_info_new.txt")
	s.planFullTmpl = loadTemplate("templates/plan_mode_full.txt")
	s.planSparseTmpl = loadTemplate("templates/plan_mode_sparse.txt")

	return s
}

func loadTemplate(path string) *template.Template {
	content, err := templateFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New(path).Parse(string(content))
	if err != nil {
		panic(err)
	}
	return tmpl
}

func (s *promptService) BuildSystemPrompt(env domain.PromptEnvironment) string {
	var parts []string
	for _, sec := range s.sections {
		parts = append(parts, sec)
	}

	parts = append(parts, renderTemplate(s.envTmpl, env))

	if env.Model != "" {
		parts = append(parts, renderTemplate(s.modelInfoTmpl, env))
	}

	return strings.Join(parts, "\n\n")
}

func (s *promptService) BuildPlanModeReminder(planFilePath string, planExists bool, iteration int) string {
	data := planFileData{PlanFilePath: planFilePath}

	var planFileInfo string
	if planExists {
		planFileInfo = renderTemplate(s.planFileExistsTmpl, data)
	} else {
		planFileInfo = renderTemplate(s.planFileNewTmpl, data)
	}

	fullData := planReminderData{PlanFileInfo: planFileInfo}

	if iteration == 1 {
		return renderTemplate(s.planFullTmpl, fullData)
	}

	attachmentIndex := (iteration - 1) / planModeReminderInterval
	if attachmentIndex%planModeReminderInterval == 0 {
		return renderTemplate(s.planFullTmpl, fullData)
	}

	return renderTemplate(s.planSparseTmpl, data)
}

type planFileData struct {
	PlanFilePath string
}

type planReminderData struct {
	PlanFileInfo string
}

func renderTemplate(tmpl *template.Template, data any) string {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}
