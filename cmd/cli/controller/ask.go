package controller

import (
	"context"

	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
)

func (c *Controller) askLoop() {
	for request := range c.askUserCh {
		c.mu.RLock()
		runtime := c.runtime
		c.mu.RUnlock()
		if runtime == nil {
			continue
		}
		_ = runtime.Dispatch(context.Background(), func() {
			c.ask = AskState{Request: &request, Selections: make([]int, len(request.Questions))}
			if c.setAskOpen != nil {
				c.setAskOpen(true)
			}
			c.invalidate()
		})
	}
}

func (c *Controller) MoveAsk(delta int) {
	if c.ask.Request == nil || c.ask.Question >= len(c.ask.Request.Questions) {
		return
	}
	options := c.ask.Request.Questions[c.ask.Question].Options
	if len(options) == 0 {
		return
	}
	c.ask.Selected = max(min(c.ask.Selected+delta, len(options)-1), 0)
	c.invalidate()
}

func (c *Controller) ConfirmAsk() {
	if c.ask.Request == nil || c.ask.Question >= len(c.ask.Request.Questions) {
		return
	}
	c.ask.Selections[c.ask.Question] = c.ask.Selected
	if c.ask.Question < len(c.ask.Request.Questions)-1 {
		c.ask.Question++
		c.ask.Selected = c.ask.Selections[c.ask.Question]
		c.invalidate()
		return
	}
	c.RespondAsk(false)
}

func (c *Controller) RespondAsk(defaults bool) {
	if c.ask.Request == nil {
		return
	}
	response := make(map[string]string)
	for i, question := range c.ask.Request.Questions {
		selected := c.ask.Selections[i]
		if defaults {
			selected = 0
		}
		if selected >= 0 && selected < len(question.Options) {
			response[question.Header] = question.Options[selected].Label
		}
	}
	select {
	case c.ask.Request.ResponseCh <- askuser.Response{Answers: response}:
	default:
	}
	c.ask = AskState{}
	if c.setAskOpen != nil {
		c.setAskOpen(false)
	}
	c.invalidate()
}
