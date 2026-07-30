package controller

import (
	"context"

	"github.com/open-portfolios/codefolio/internal/infra/approval"
)

func (c *Controller) approvalLoop() {
	for request := range c.approvalCh {
		c.mu.RLock()
		runtime := c.runtime
		c.mu.RUnlock()
		if runtime == nil {
			request.Resolve(approval.Cancelled)
			continue
		}
		if err := runtime.Dispatch(context.Background(), func() { c.enqueueApproval(request) }); err != nil {
			request.Resolve(approval.Cancelled)
		}
	}
}

func (c *Controller) enqueueApproval(request *approval.Request) {
	if c.approval.Request == nil {
		c.approval.Request = request
		if c.setApprovalOpen != nil {
			c.setApprovalOpen(true)
		}
	} else {
		c.approval.Queue = append(c.approval.Queue, request)
	}
	c.invalidate()
}

func (c *Controller) ApproveOnce() { c.resolveApproval(approval.AllowOnce) }

func (c *Controller) ApproveSession() { c.resolveApproval(approval.AllowSession) }

func (c *Controller) DenyApproval() { c.resolveApproval(approval.Deny) }

func (c *Controller) resolveApproval(decision approval.Decision) {
	request := c.approval.Request
	if request == nil {
		return
	}
	request.Resolve(decision)
	c.advanceApproval()
}

func (c *Controller) dismissApproval() {
	if c.approval.Request == nil {
		return
	}
	c.approval.Request.Resolve(approval.Cancelled)
	c.advanceApproval()
}

func (c *Controller) denyAllApprovals() {
	if c.approval.Request != nil {
		c.approval.Request.Resolve(approval.Cancelled)
	}
	for _, request := range c.approval.Queue {
		request.Resolve(approval.Cancelled)
	}
	c.approval = ApprovalState{}
	if c.setApprovalOpen != nil {
		c.setApprovalOpen(false)
	}
}

func (c *Controller) advanceApproval() {
	if len(c.approval.Queue) > 0 {
		c.approval.Request = c.approval.Queue[0]
		c.approval.Queue = c.approval.Queue[1:]
	} else {
		c.approval.Request = nil
		if c.setApprovalOpen != nil {
			c.setApprovalOpen(false)
		}
	}
	c.invalidate()
}
