package messages

import (
	"encoding/json"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type messageConverter struct {
	llm.BaseMessageVisitor

	sys        []anthropic.TextBlockParam
	msgs       []anthropic.MessageParam
	toolBlocks []anthropic.ContentBlockParamUnion
}

func (c *messageConverter) VisitUser(m llm.UserMessage) error {
	c.flushToolBlocks()
	c.msgs = append(c.msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
	return nil
}

func (c *messageConverter) VisitAssistant(m llm.AssistantMessage) error {
	c.flushToolBlocks()
	var blocks []anthropic.ContentBlockParamUnion
	if m.Content != "" {
		blocks = append(blocks, anthropic.NewTextBlock(m.Content))
	}
	for _, tc := range m.ToolCalls {
		var input any
		if tc.Input != "" {
			json.Unmarshal([]byte(tc.Input), &input)
		}
		blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
	}
	if m.Thinking != "" {
		blocks = append(blocks, anthropic.NewThinkingBlock(m.ThinkingSignature, m.Thinking))
	}
	if len(blocks) == 0 {
		blocks = append(blocks, anthropic.NewTextBlock(""))
	}
	c.msgs = append(c.msgs, anthropic.NewAssistantMessage(blocks...))
	return nil
}

func (c *messageConverter) VisitSystem(m llm.SystemMessage) error {
	c.sys = append(c.sys, anthropic.TextBlockParam{Text: m.Content})
	return nil
}

func (c *messageConverter) VisitDeveloper(m llm.DeveloperMessage) error {
	c.sys = append(c.sys, anthropic.TextBlockParam{Text: m.Content})
	return nil
}

func (c *messageConverter) VisitTool(m llm.ToolMessage) error {
	isError := strings.HasPrefix(m.Content, "Error:")
	c.toolBlocks = append(c.toolBlocks, anthropic.NewToolResultBlock(m.ToolCallID, m.Content, isError))
	return nil
}

func (c *messageConverter) VisitFunction(m llm.FunctionMessage) error {
	c.flushToolBlocks()
	c.msgs = append(c.msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
	return nil
}

func (c *messageConverter) flushToolBlocks() {
	if len(c.toolBlocks) == 0 {
		return
	}
	c.msgs = append(c.msgs, anthropic.NewUserMessage(c.toolBlocks...))
	c.toolBlocks = nil
}
