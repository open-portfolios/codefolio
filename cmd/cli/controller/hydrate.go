package controller

import (
	"fmt"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

func HydrateMessages(messages []domain.ChatMessage, profile string) []*Message {
	result := make([]*Message, 0, len(messages))
	tools := map[string]*Tool{}
	for _, message := range messages {
		switch message.Role {
		case llm.RoleUser:
			result = append(result, &Message{ID: message.ID, Role: "user", Profile: profile, Content: message.Content})
		case llm.RoleAssistant:
			view := &Message{ID: message.ID, Role: "assistant", Content: message.Content, Thinking: message.Thinking, Streaming: false}
			for _, call := range message.ToolCalls {
				tool := &Tool{ID: call.ID, Name: call.Name, Input: call.Input}
				view.Tools = append(view.Tools, tool)
				tools[call.ID] = tool
			}
			result = append(result, view)
		case llm.RoleTool:
			if tool := tools[message.ToolCallID]; tool != nil {
				tool.Done, tool.Output, tool.IsError = true, message.Content, message.IsError
			} else {
				result = append(result, &Message{ID: fmt.Sprintf("notice-%s", message.ID), Role: "notice", Content: message.Content})
			}
		}
	}
	return result
}
