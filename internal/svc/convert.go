package svc

import (
	"fmt"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

func ChatMessagesToLLM(messages []domain.ChatMessage) []llm.Message {
	msgs := make([]llm.Message, 0, len(messages))
	for _, m := range messages {
		if m.Streaming {
			continue
		}
		msgs = append(msgs, messageAdapter{m})
	}
	return msgs
}

type messageAdapter struct{ msg domain.ChatMessage }

func (a messageAdapter) Accept(v llm.MessageVisitor) error {
	switch a.msg.Role {
	case llm.RoleUser:
		return v.VisitUser(llm.UserMessage{Content: a.msg.Content})
	case llm.RoleAssistant:
		tcs := make([]llm.ToolCallInfo, len(a.msg.ToolCalls))
		for i, tc := range a.msg.ToolCalls {
			tcs[i] = llm.ToolCallInfo{ID: tc.ID, Name: tc.Name, Input: tc.Input}
		}
		return v.VisitAssistant(llm.AssistantMessage{
			Content:           a.msg.Content,
			Thinking:          a.msg.Thinking,
			ThinkingSignature: a.msg.ThinkingSignature,
			ToolCalls:         tcs,
		})
	case llm.RoleSystem:
		return v.VisitSystem(llm.SystemMessage{Content: a.msg.Content})
	case llm.RoleDeveloper:
		return v.VisitDeveloper(llm.DeveloperMessage{Content: a.msg.Content})
	case llm.RoleTool:
		return v.VisitTool(llm.ToolMessage{
			Content:    a.msg.Content,
			ToolCallID: a.msg.ToolCallID,
		})
	case llm.RoleFunction:
		return v.VisitFunction(llm.FunctionMessage{Content: a.msg.Content})
	default:
		return fmt.Errorf("unknown message role: %s", a.msg.Role)
	}
}
