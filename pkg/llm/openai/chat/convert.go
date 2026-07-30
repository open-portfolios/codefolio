package chat

import (
	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

type messageConverter struct {
	llm.BaseMessageVisitor

	msgs []openai.ChatCompletionMessageParamUnion
}

func (c *messageConverter) VisitUser(m llm.UserMessage) error {
	c.msgs = append(c.msgs, openai.UserMessage(m.Content))
	return nil
}

func (c *messageConverter) VisitAssistant(m llm.AssistantMessage) error {
	if len(m.ToolCalls) > 0 {
		var tcs []openai.ChatCompletionMessageToolCallUnionParam
		for _, tc := range m.ToolCalls {
			tcs = append(tcs, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: tc.ID,
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Name,
						Arguments: tc.Input,
					},
				},
			})
		}
		assistant := &openai.ChatCompletionAssistantMessageParam{
			ToolCalls: tcs,
		}
		if m.Content != "" {
			assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: param.NewOpt(m.Content),
			}
		}
		c.msgs = append(c.msgs, openai.ChatCompletionMessageParamUnion{
			OfAssistant: assistant,
		})
	} else {
		c.msgs = append(c.msgs, openai.AssistantMessage(m.Content))
	}
	return nil
}

func (c *messageConverter) VisitSystem(m llm.SystemMessage) error {
	c.msgs = append(c.msgs, openai.SystemMessage(m.Content))
	return nil
}

func (c *messageConverter) VisitDeveloper(m llm.DeveloperMessage) error {
	// Some OpenAI-compatible APIs, including DeepSeek Chat Completions, do not
	// accept the newer developer role. Preserve instruction precedence through
	// the broadly supported system role at this provider boundary.
	c.msgs = append(c.msgs, openai.SystemMessage(m.Content))
	return nil
}

func (c *messageConverter) VisitTool(m llm.ToolMessage) error {
	c.msgs = append(c.msgs, openai.ToolMessage(m.Content, m.ToolCallID))
	return nil
}

func (c *messageConverter) VisitFunction(m llm.FunctionMessage) error {
	c.msgs = append(c.msgs, openai.ChatCompletionMessageParamOfFunction(m.Content, m.Name))
	return nil
}
