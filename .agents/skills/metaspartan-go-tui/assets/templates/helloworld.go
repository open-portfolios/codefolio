// helloworld.go
//
// Minimal runnable scaffold using the Application (OO) API.
//
// Build:  go run helloworld.go
// Quit:   press q, <C-c>, or <Escape> (handled by Application defaults)

package main

import (
	"log"
	"strconv"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/metaspartan/gotui/v5/widgets"
)

// CounterWidget demonstrates the two things every gotui widget needs:
//  1. A *Block (or other Drawable) as the embedded rendering surface
//  2. A HandleEvent method to receive KeyboardEvent / MouseEvent
type CounterWidget struct {
	*widgets.Paragraph
	clickCount int
}

func NewCounterWidget() *CounterWidget {
	p := widgets.NewParagraph()
	p.Title = "Hello, gotui!"
	p.Text = "Press <Space> to bump.\nClick to count.\nPress q or <C-c> to quit."
	p.TitleStyle = ui.NewStyle(ui.ColorCyan, ui.ColorClear, ui.ModifierBold)
	p.BorderStyle.Fg = ui.ColorCyan
	p.BorderRounded = true
	return &CounterWidget{Paragraph: p}
}

func (c *CounterWidget) HandleEvent(e ui.Event) bool {
	switch e.Type {
	case ui.KeyboardEvent:
		if e.ID == "<Space>" || e.ID == " " {
			c.clickCount++
			c.Text = "Bumped " + strconv.Itoa(c.clickCount) + " times"
			return true
		}
	case ui.MouseEvent:
		if e.ID == "<MouseLeft>" {
			c.clickCount++
			c.Text = "Clicked " + strconv.Itoa(c.clickCount) + " times"
			return true
		}
	}
	return false
}

func main() {
	app := ui.NewApp()
	app.SetRoot(NewCounterWidget(), true) // 2nd arg = focus root

	if err := app.Run(); err != nil {
		log.Fatalf("App run failed: %v", err)
	}
}
