// dashboard.go
//
// Multi-widget dashboard using the Functional API + Grid layout.
//
// Build:  go run dashboard.go
// Quit:   press q or <C-c>
// Flags:  -screenshot  -> write screenshot.png and exit (120x40 cells)
//
// Layout (4 rows × variable cols, ratio-based):
//
//   ┌─────────────────────────────────────────────┐
//   │ Header (Paragraph)                          │   1/10
//   ├──────────────────────────┬──────────────────┤
//   │ SparklineGroup           │ Memory / Load    │   2/10
//   │                          │ LineGauge × 2    │
//   ├────────┬────────┬────────┤                  │
//   │ Bar    │ Pie    │ Plot   │                  │   3.5/10
//   ├────────┴────────┴────────┴──────────────────┤
//   │ System Logs (List)       │ Download Gauge   │   3.5/10
//   └─────────────────────────────────────────────┘

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/metaspartan/gotui/v5/widgets"
)

func buildGrid() *ui.Grid {
	// 1. Header
	p := widgets.NewParagraph()
	p.Title = "Dashboard"
	p.Text = "PRESS q TO QUIT | Periodic updates every 100ms"
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorLightCyan
	p.TitleStyle = ui.NewStyle(ui.ColorLightCyan, ui.ColorClear, ui.ModifierBold)
	p.TitleAlignment = ui.AlignCenter
	p.TitleRight = "v1"

	// 2. Sparklines (CPU)
	slData := make([]float64, 200)
	sl := widgets.NewSparkline()
	sl.Data = slData
	sl.LineColor = ui.ColorGreen
	sl.TitleStyle.Fg = ui.ColorWhite
	sl.MaxVal = 100

	sl2 := widgets.NewSparkline()
	sl2.Data = slData
	sl2.LineColor = ui.ColorMagenta
	sl2.TitleStyle.Fg = ui.ColorWhite
	sl2.MaxVal = 100

	slg := widgets.NewSparklineGroup(sl, sl2)
	slg.Title = "CPU Usage"
	slg.TitleStyle.Fg = ui.ColorGreen
	slg.BorderStyle.Fg = ui.ColorGreen
	slg.TitleRight = "Core 0 & 1"
	slg.BorderRounded = true

	// 3. Line Gauges
	lg1 := widgets.NewLineGauge()
	lg1.Title = "Memory"
	lg1.Percent = 45
	lg1.BarRune = '■'
	lg1.LineColor = ui.ColorYellow
	lg1.TitleStyle.Fg = ui.ColorYellow
	lg1.BorderRounded = true

	lg2 := widgets.NewLineGauge()
	lg2.Title = "Load"
	lg2.Percent = 60
	lg2.BarRune = '▰'
	lg2.BarRuneEmpty = '▱'
	lg2.LineColor = ui.ColorRed
	lg2.TitleStyle.Fg = ui.ColorRed
	lg2.BorderRounded = true

	// 4. Bar Chart
	bc := widgets.NewBarChart()
	bc.Title = "Network"
	bc.TitleBottom = "MB/s"
	bc.TitleBottomAlignment = ui.AlignRight
	bc.Labels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	bc.BarColors = []ui.Color{ui.ColorBlue, ui.ColorLightCyan}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.Data = []float64{3, 2, 5, 3, 9, 5}
	bc.TitleStyle.Fg = ui.ColorBlue
	bc.BorderStyle.Fg = ui.ColorBlue
	bc.BorderRounded = true
	bc.BarWidth = 0
	bc.BarGap = 1
	bc.MaxVal = 10

	// 5. Pie Chart
	pc := widgets.NewPieChart()
	pc.Title = "Disk"
	pc.Data = []float64{10, 20, 30, 40}
	pc.Colors = []ui.Color{ui.ColorRed, ui.ColorYellow, ui.ColorGreen, ui.ColorBlue}
	pc.LabelFormatter = func(i int, v float64) string {
		return fmt.Sprintf("%.0f%%", v)
	}
	pc.TitleStyle.Fg = ui.ColorMagenta
	pc.BorderStyle.Fg = ui.ColorMagenta
	pc.BorderRounded = true

	// 6. Plot (Response Time)
	plotData := make([][]float64, 2)
	plotData[0] = make([]float64, 50)
	plotData[1] = make([]float64, 50)

	plot := widgets.NewPlot()
	plot.Title = "Latency"
	plot.TitleBottom = "(ms)"
	plot.Data = plotData
	plot.AxesColor = ui.ColorWhite
	plot.LineColors[0] = ui.ColorGreen
	plot.LineColors[1] = ui.ColorYellow
	plot.Marker = widgets.MarkerDot
	plot.TitleStyle.Fg = ui.ColorLightCyan
	plot.BorderStyle.Fg = ui.ColorLightCyan
	plot.BorderRounded = true

	// 7. List (System Logs)
	l := widgets.NewList()
	l.Title = "System Logs"
	l.Rows = []string{
		"[INFO] System started",
		"[INFO] Service A initialized",
		"[WARN] Connection timeout (retrying)",
		"[INFO] Cache cleared",
		"[ERROR] Auth failed for user x",
		"[INFO] Worker pool scaled up",
		"[INFO] Health check passed",
		"[WARN] High memory usage detected",
		"[INFO] GC triggered",
		"[INFO] New client connected",
		"[ERROR] Database connection lost",
		"[INFO] Reconnecting to DB...",
		"[INFO] DB Connected",
	}
	l.TextStyle.Fg = ui.ColorYellow
	l.SelectedStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)
	l.TitleStyle.Fg = ui.ColorYellow
	l.BorderStyle.Fg = ui.ColorYellow
	l.TitleBottom = "Wheel to scroll"
	l.TitleBottomAlignment = ui.AlignRight
	l.BorderRounded = true

	// 8. Gauge (Download)
	g := widgets.NewGauge()
	g.Title = "Download"
	g.Percent = 50
	g.BarColor = ui.ColorGreen
	g.BorderStyle.Fg = ui.ColorGreen
	g.TitleStyle.Fg = ui.ColorGreen
	g.BorderRounded = true

	// 9. Grid (ratio-based nested layout)
	grid := ui.NewGrid()
	grid.Set(
		ui.NewRow(1.0/10,
			ui.NewCol(1.0, p),
		),
		ui.NewRow(2.0/10,
			ui.NewCol(1.0/2, slg),
			ui.NewCol(1.0/2,
				ui.NewRow(1.0/2, lg1),
				ui.NewRow(1.0/2, lg2),
			),
		),
		ui.NewRow(3.5/10,
			ui.NewCol(1.0/3, bc),
			ui.NewCol(1.0/3, pc),
			ui.NewCol(1.0/3, plot),
		),
		ui.NewRow(3.5/10,
			ui.NewCol(2.0/3, l),
			ui.NewCol(1.0/3, g),
		),
	)

	return grid
}

func main() {
	grid := buildGrid()

	// Optional: -screenshot flag writes a PNG and exits.
	// Uses SimulationScreen under the hood; no real terminal required.
	if len(os.Args) > 1 && os.Args[1] == "-screenshot" {
		const w, h = 120, 40
		grid.SetRect(0, 0, w, h)
		if err := ui.SaveImage("screenshot.png", w, h, grid); err != nil {
			log.Fatalf("failed to take screenshot: %v", err)
		}
		fmt.Println("Screenshot saved to screenshot.png")
		return
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize gotui: %v", err)
	}
	defer ui.Close()

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(100 * time.Millisecond).C // 10 FPS
	tick := 0

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<MouseWheelUp>":
				ui.Render(grid)
			case "<MouseWheelDown>":
				ui.Render(grid)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		case <-ticker:
			tick++
			// Periodically mutate widget fields and re-render
			ui.Render(grid)
			_ = math.Sin(float64(tick) / 10.0)
			_ = rand.Intn(100)
		}
	}
}
