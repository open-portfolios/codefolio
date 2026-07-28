# AGENTS.md - codefolio

codefolio — an AI-powered coding tool in the vein of Claude Code or Codex, with both TUI and desktop app interfaces. Tech stack: Go (CLI with the local TUX framework), Wails v2 + React + TypeScript (desktop app).

The project is a work in progress.

## Directory layout

| Directory   | Purpose                                                                            |
| ----------- | ---------------------------------------------------------------------------------- |
| `cmd/cli/`  | TUI entry point, built with the local TUX framework                               |
| `cmd/app/`  | Desktop app entry point, Wails WebView; frontend at `cmd/app/web/`                 |
| `pkg/`      | Reusable, loosely coupled public packages (LLM interface, generic utilities, etc.) |
| `internal/` | Core agent logic shared by TUI and desktop app                                     |

`internal/` serves as the kernel of codefolio — it could be compiled as a DLL and consumed by other tools as well.

When making changes, respect these boundaries: `pkg/` for reusable, project-agnostic code; `internal/` for shared agent logic; `cmd/` for entrypoints and UI only.

## Dev commands

The agent only writes files and runs static checks, **never** compiles or runs the project.

| Command      | Purpose                                                                            |
| ------------ | ---------------------------------------------------------------------------------- |
| `task check` | Correctness check (`go build ./...` + `go vet ./...` + `pnpm build` + `pnpm lint`) |

Other commands (`task cli`, `task app`, `task build`) are for humans — the agent must not invoke them.

## Skills

Installed skills under `.agents/skills/`. Use the skill tool to load them when relevant.

| Skill                                       | Purpose                                                      |
| ------------------------------------------- | ------------------------------------------------------------ |
| `charmland-go-tui`                          | Bubble Tea v2, Lip Gloss, Glamour, Bubbles, Huh TUI patterns |
| `tdd`                                       | Test-driven development workflow                             |
| `domain-modeling`                           | Domain terminology, ADRs, ubiquitous language                |
| `design-pattern`                            | GoF patterns, SOLID, architectural layering                  |
| `grilling` / `grill-me` / `grill-with-docs` | Stress-test plans and designs                                |
| `handoff`                                   | Compact session into a handoff document                      |

## Sub-directory AGENTS.md

Some subdirectories have their own `AGENTS.md` with local conventions. Follow them as well.
