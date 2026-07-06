# Logging Reference

Use `charm.land/log/v2` for structured, colorful logging in Go terminal apps. Bubble Tea owns terminal output while running, so uncontrolled stdout prints can corrupt the UI.

## Basic Usage

```go
import log "charm.land/log/v2"

log.Info("starting", "screen", "home")
log.Debug("loaded rows", "count", len(rows))
log.Error("save failed", "err", err)
```

Set level:

```go
log.SetLevel(log.DebugLevel)
```

## Custom Logger

```go
logger := log.NewWithOptions(os.Stderr, log.Options{
	Level:           log.DebugLevel,
	ReportTimestamp: true,
	ReportCaller:    true,
})

logger.Info("ready", "version", version)
```

## File Logging for Bubble Tea

During a Bubble Tea session, prefer a file for debug logs:

```go
f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
if err != nil {
	return err
}
defer f.Close()

logger := log.NewWithOptions(f, log.Options{
	Level:           log.DebugLevel,
	ReportTimestamp: true,
})
```

Bubble Tea also provides `tea.LogToFile` for simple file logging:

```go
f, err := tea.LogToFile("debug.log", "debug")
if err != nil {
	return err
}
defer f.Close()
```

## Formatters

- `log.TextFormatter` for human-readable terminal logs.
- `log.JSONFormatter` for structured machine-readable logs.
- `log.LogfmtFormatter` for logfmt output.

```go
logger.SetFormatter(log.JSONFormatter)
```

## slog Compatibility

`*log.Logger` implements `slog.Handler`, so it can back standard Go `log/slog`:

```go
handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.InfoLevel})
logger := slog.New(handler)
logger.Info("hello", "key", "value")
```

## TUI Rules

- Do not use `fmt.Println` for debug output while Bubble Tea is rendering.
- Use file logs, stderr before startup, or Bubble Tea's `Program.Println` only when intentional.
- Log errors from commands as structured values; also surface user-facing errors in the model/view.
- Avoid logging inside `View` unless diagnosing a rendering bug.
