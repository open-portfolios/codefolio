# Spinner Frame Sets

> **Usage**: Pass any frame set to `widgets.NewSpinner()` via its `Frames` field, or use the
> pre-defined exported `widgets.Spinner*` package variables.

All frame sets are pure Unicode / ASCII strings — no dependencies, no licensing concerns.
They are safe to copy into your own projects.

## Quick Start

```go
sp := widgets.NewSpinner()
sp.Frames = widgets.SpinnerDots
sp.Label = "Loading…"

// Drive animation
go func() {
    ticker := time.NewTicker(80 * time.Millisecond)
    defer ticker.Stop()
    for range ticker.C {
        sp.Advance()
        app.Backend.Render(app.GetRoot())
        // or: ui.Render(sp) when using the functional API
    }
}()
```

## Frame Sets

| Name | Frames | Notes |
|---|---|---|
| `SpinnerLine` | `\|` `/` `-` `\` | Classic ASCII. Always works. |
| `SpinnerDots` | `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` | 10 frames, braille dots. |
| `SpinnerMiniDots` | `⠋ ⠙ ⠚ ⠞ ⠖ ⠦ ⠴ ⠲ ⠳ ⠓` | 10 frames, denser. |
| `SpinnerPulse` | `█ ▓ ▒ ░` | 4 frames, block fade. |
| `SpinnerPoints` | `∙∙∙ ●∙∙ ∙●∙ ∙∙● ∙∙∙` | 5 frames, dot chase. |
| `SpinnerGlobe` | `🌍 🌎 🌏` | 3 frames, requires emoji font. |
| `SpinnerMoon` | `🌑 🌒 🌓 🌔 🌕 🌖 🌗 🌘` | 8 frames, requires emoji font. |
| `SpinnerClock` | `🕛 🕐 🕑 🕒 🕓 🕔 🕕 🕖 🕗 🕘 🕙 🕚` | 12 frames, hour hand sweep. |
| `SpinnerMonkey` | `🙈 🙉 🙊` | 3 frames, "see no evil" cycle. |
| `SpinnerStar` | `✶ ✸ ✹ ✺ ✹ ✸` | 6 frames, pulse. |
| `SpinnerHamburger` | `☱ ☲ ☴` | 3 frames, trigram morph. |
| `SpinnerGrowVertical` | `space ▃ ▄ ▅ ▆ ▇ █ ▇ ▆ ▅ ▄ ▃` | 12 frames, vertical bar grow/shrink. |
| `SpinnerGrowHorizontal` | `▉ ▊ ▋ ▌ ▍ ▎ ▏ ▎ ▍ ▌ ▋ ▊ ▉` | 13 frames, horizontal bar grow/shrink. |
| `SpinnerArrow` | `← ↖ ↑ ↗ → ↘ ↓ ↙` | 8 frames, full rotation. |
| `SpinnerTriangle` | `◢ ◣ ◤ ◥` | 4 frames, spinning triangle. |
| `SpinnerCircleHalves` | `◐ ◓ ◑ ◒` | 4 frames, half-circle rotation. |
| `SpinnerBouncingBall` | `⠁ ⠂ ⠄ ⡀⢀ ⠠ ⠐ ⠈` | 8 frames, braille ball bounce. |

## Fallback Recommendations

| Terminal | Recommended set |
|---|---|
| Headless / CI / unknown | `SpinnerLine` (ASCII only) |
| Modern terminal emulator (iTerm2, WezTerm, Windows Terminal) | `SpinnerDots`, `SpinnerGrowVertical`, `SpinnerArrow` |
| Emoji-capable terminal | `SpinnerMoon`, `SpinnerGlobe`, `SpinnerClock` |

## Animation Cadence

| Frame count | Recommended tick |
|---|---|
| 3–4 frames | 120–150 ms |
| 8–10 frames | 80–100 ms |
| 12+ frames | 60–80 ms |

Higher tick rates cause flicker on SSH / slow terminals. Start at 100 ms and tune.