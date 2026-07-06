# Screenshot Mode (Headless PNG Generation)

gotui can render to a **simulation screen** (in-memory) instead of a real
terminal, and then save the buffer as a PNG. This is invaluable for:

- Documentation screenshots without opening a terminal.
- Visual regression tests.
- Generating README images in CI.

## The `-screenshot` Pattern

Use this convention for screenshot-capable examples and generated docs:

```go
const screenshotW, screenshotH = 120, 40

func main() {
    grid := buildGrid()

    if len(os.Args) > 1 && os.Args[1] == "-screenshot" {
        grid.SetRect(0, 0, screenshotW, screenshotH)
        if err := ui.SaveImage("screenshot.png", screenshotW, screenshotH, grid); err != nil {
            log.Fatal(err)
        }
        fmt.Println("Saved screenshot.png")
        return
    }

    // ... normal interactive run ...
}
```

When the user runs `program -screenshot`, the program:
1. Builds the same UI as in interactive mode.
2. Sets a fixed cell size (120×40 is a common safe choice).
3. Calls `ui.SaveImage(path, w, h, grid)` and exits.

`SaveImage` internally creates a `tcell.SimulationScreen`, calls `Init`
on it, lets you `Render` into it, then captures the screen pixels and
writes PNG.

## Programmatic Capture

If you want a buffer instead of a file:

```go
// Capture to *image.RGBA, then encode yourself.
img, err := ui.Capture(grid)  // grid must already have SetRect called
if err != nil { return err }

f, _ := os.Create("out.png")
defer f.Close()
png.Encode(f, img)
```

Or render to an in-memory `SimulationScreen`:

```go
sim := tcell.NewSimulationScreen("UTF-8")
sim.Init()
sim.SetSize(w, h)

// draw into `sim`...
ui.Render(...)  // if you can route Render to use sim as backend

img, _, _ := sim.GetContents()  // returns *image.RGBA
```

## Using SimulationScreen Without `-screenshot`

For tests / tooling, you can construct a simulation backend directly:

```go
import "github.com/gdamore/tcell/v3"

sim := tcell.NewSimulationScreen("UTF-8")
if err := sim.Init(); err != nil { return err }
defer sim.Close()
sim.SetSize(120, 40)

// Now use sim like a tcell.Screen. Prefer the built-in SaveImage path unless
// you need custom capture behavior.
app, err := ui.NewBackend(&ui.InitConfig{Screen: sim})
```

**Note**: if your installed gotui version exposes a different `InitConfig`
field for injecting a screen, prefer `ui.SaveImage` unless you are writing
tooling that must own the simulation screen.

## Common Sizes for Screenshots

| Use                     | Cells  |
| ----------------------- | ------ |
| README banner           | 120×30 |
| GitHub social preview   | 160×40 |
| Documentation tile      | 80×24  |
| Mobile / narrow display | 60×20  |
| Wide chart              | 200×40 |

## Resolution Note

Each cell renders at `cellWidth × cellHeight` pixels in the output PNG
(default `7×13` via `golang.org/x/image/font/basicfont`). For a 120×40
buffer: **840×520 pixels**. Adjust by passing a custom font to the
internal `screenshot.NewFontFace` if available, or by switching to
`x/image/font/opentype` with a TTF.

## CI Integration

```yaml
# .github/workflows/screenshots.yml
name: Regenerate screenshots
on: [push]
jobs:
  screenshots:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go mod download
      - run: |
          for d in examples/*/; do
            (cd "$d" && go run . -screenshot)
          done
      - uses: actions/upload-artifact@v4
        with:
          name: screenshots
          path: "**/screenshot.png"
```

## Anti-Patterns

- ❌ Capturing at huge sizes (e.g. 400×100) — image gets unreadable.
- ❌ Capturing mid-animation — pick one frame deterministically.
- ❌ Hardcoding paths in the binary — let the user pass `-out path.png`.