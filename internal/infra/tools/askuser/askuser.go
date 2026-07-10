package askuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type Question struct {
	Text        string   `json:"question"`
	Header      string   `json:"header"`
	Options     []Option `json:"options"`
	MultiSelect bool     `json:"multiSelect"`
}

type Request struct {
	Questions  []Question
	ResponseCh chan Response
}

type Response struct {
	Answers map[string]string
}

type Tool struct {
	RequestCh chan<- Request
}

func (t *Tool) ShouldDefer() bool { return false }

func (t *Tool) Name() string        { return "AskUserQuestion" }

func (t *Tool) Description() string {
	return `Ask the user a question with structured multiple-choice options. Use this to:
- Gather user preferences or requirements
- Clarify ambiguous instructions
- Get decisions on implementation choices
- Offer choices about direction to take

Each question has 2-4 options. An "Other" option for custom input is automatically provided.
Use multiSelect: true when choices are not mutually exclusive.`
}

func (t *Tool) Category() domain.ToolCategory { return domain.CategoryRead }

func (t *Tool) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"questions": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"question": map[string]any{
								"type":        "string",
								"description": "The question to ask the user",
							},
							"header": map[string]any{
								"type":        "string",
								"description": "Short label (max 12 chars)",
							},
							"options": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"label":       map[string]any{"type": "string", "description": "Option display text (1-5 words)"},
										"description": map[string]any{"type": "string", "description": "What this option means"},
									},
									"required": []string{"label", "description"},
								},
								"minItems": 2,
								"maxItems": 4,
							},
							"multiSelect": map[string]any{
								"type":    "boolean",
								"default": false,
							},
						},
						"required": []string{"question", "header", "options", "multiSelect"},
					},
					"minItems": 1,
					"maxItems": 4,
				},
			},
			"required": []string{"questions"},
		},
	}
}

func (t *Tool) Execute(ctx context.Context, args map[string]any) domain.ToolResult {
	questionsRaw, ok := args["questions"]
	if !ok {
		return domain.ToolResult{Output: "Error: questions is required", IsError: true}
	}

	questionsJSON, err := json.Marshal(questionsRaw)
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error: invalid questions format: %s", err), IsError: true}
	}

	var questions []Question
	if err := json.Unmarshal(questionsJSON, &questions); err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error: invalid questions format: %s", err), IsError: true}
	}

	if len(questions) == 0 || len(questions) > 4 {
		return domain.ToolResult{Output: "Error: must have 1-4 questions", IsError: true}
	}

	for _, q := range questions {
		if len(q.Options) < 2 || len(q.Options) > 4 {
			return domain.ToolResult{Output: fmt.Sprintf("Error: question '%s' must have 2-4 options", q.Text), IsError: true}
		}
	}

	if t.RequestCh == nil {
		return domain.ToolResult{Output: "Error: AskUserQuestion not available in this context", IsError: true}
	}

	respCh := make(chan Response, 1)
	t.RequestCh <- Request{
		Questions:  questions,
		ResponseCh: respCh,
	}

	select {
	case resp := <-respCh:
		var parts []string
		for q, a := range resp.Answers {
			parts = append(parts, fmt.Sprintf("%q = %q", q, a))
		}
		return domain.ToolResult{
			Output: fmt.Sprintf("User has answered your questions: %s. You can now continue with the user's answers in mind.", strings.Join(parts, ", ")),
		}
	case <-ctx.Done():
		return domain.ToolResult{Output: "Question cancelled", IsError: true}
	}
}
