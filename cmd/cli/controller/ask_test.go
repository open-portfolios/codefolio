package controller

import (
	"testing"

	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
)

func TestRespondAskUsesSelectionsAndDefaults(t *testing.T) {
	responses := make(chan askuser.Response, 1)
	controller := &Controller{ask: AskState{Request: &askuser.Request{Questions: []askuser.Question{
		{Header: "Color", Options: []askuser.Option{{Label: "Blue"}, {Label: "Green"}}},
		{Header: "Mode", Options: []askuser.Option{{Label: "Safe"}, {Label: "Fast"}}},
	}, ResponseCh: responses}, Selections: []int{1, 0}}, invalidate: func() {}, setAskOpen: func(bool) {}}
	controller.RespondAsk(false)
	response := <-responses
	if response.Answers["Color"] != "Green" || response.Answers["Mode"] != "Safe" {
		t.Fatalf("selected response = %#v", response.Answers)
	}
	if controller.ask.Request != nil {
		t.Fatal("response should clear the request")
	}
}
