# wick

Claude Code token consumption monitor, written in Go, using [Cobra](https://github.com/spf13/cobra) (CLI framework), [Lipgloss](https://github.com/charmbracelet/lipgloss) (styling), and [Bubble Tea](https://github.com/charmbracelet/bubbletea) (live TUI).

Reads the JSONL transcript that Claude Code itself already writes per session (`~/.claude/projects/<project>/<session>.jsonl`) and shows: total tokens spent, ranking of which prompt cost the most, breakdown by skill/subagent, and % of context window used. No hooks, no external API, no auth — it just reads the file that already exists on disk.

## Install

```bash
# Clone and build
git clone https://github.com/Morpa/wick.git
cd wick
go build -o wick .

# Install to PATH (optional)
mv wick /usr/local/bin/
```

## Usage

```bash
wick snapshot        # text snapshot of the active session in the current directory
wick snapshot ~/dev  # snapshot of another directory
wick watch           # live dashboard, auto-updates (Ctrl+C or 'q' to quit)
wick help            # help
```

### Snapshot

Captures a snapshot of the tokens from the active session in a directory:

```bash
wick snapshot ~/Dev/my-project
```

Output:

```
Active session: abc123-def456 (my-project)
Total: 12.2M tok  (in 5.5M / out 104.4k / cache-w 0 / cache-r 6.6M)
Context window: 64.5% (129.1k / 200.0k)

Top prompts:
 1. 11.4M  "refactor the authentication module..."
 2. 770.1k  "explain why this test fails..."

By skill/agent:
 main         12.2M
```

### Watch (live dashboard)

Starts a live dashboard in the terminal, updating every 2 seconds:

```bash
wick watch
```

The dashboard shows:

- 🔥 Total tokens consumed
- 🧠 Context window usage (green/yellow/red as usage rises)
- 🏆 Ranking of the most expensive prompts
- 🧩 Breakdown by skill/agent

Press `q` or `Ctrl+C` to quit.

### Statusline

Compact mode for integration with Claude Code's `statusLine`:

```bash
echo '{"transcript_path": "/path/to/session.jsonl", "context_window": {"used_percentage": 70.1}}' | wick statusline
```

Output:

```
🔥 session: 12.2M tok · context: 70% used · priciest prompt: 11.4M tok
```

### Setup

Configures the integration with Claude Code (skill, `/tokens` command, statusline):

```bash
wick setup           # configure everything
wick setup --force   # overwrite existing statusLine
```

This:

1. Creates the skill directory at `~/.claude/skills/wick`
2. Copies the `/tokens` command to `~/.claude/commands/tokens.md`
3. Registers the `statusLine` in `~/.claude/settings.json`

After that, a compact line appears automatically at the bottom of the Claude Code terminal, and you can use the `/tokens` slash command inside Claude Code.

## Available commands

| Command                | Description                                                       |
| ---------------------- | ----------------------------------------------------------------- |
| `wick snapshot [cwd]`  | Text snapshot: total, ranking, breakdown by skill, context window |
| `wick watch [cwd]`     | Live dashboard in the terminal (auto-updates)                     |
| `wick setup [--force]` | Configures Claude Code integration (skill, /tokens, statusLine)   |
| `wick statusline`      | statusLine mode — reads JSON from stdin (internal use)            |
| `wick help`            | Shows help                                                        |

## Stack

- **Go** — language
- **[Cobra](https://github.com/spf13/cobra)** — CLI framework (commands, flags, help)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — terminal styling (colors, borders, bold fonts)
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework for the live dashboard

## Privacy

Runs 100% locally. Makes no network calls, sends no telemetry, has no API key. The only thing it reads is the `.jsonl` file that Claude Code itself already writes to your disk (`~/.claude/projects/**/*.jsonl`) — nothing leaves your machine.

## Project structure

```
main.go                      # entry point
cmd/
├── root.go                  # root command (Cobra)
├── snapshot.go              # wick snapshot
├── watch.go                 # wick watch (Bubble Tea)
├── setup.go                 # wick setup (Claude Code integration)
└── statusline.go            # wick statusline (stdin JSON)
internal/
├── session/
│   ├── discovery.go         # session file discovery
│   ├── parse.go             # JSONL parsing and event normalization
│   ├── turn.go              # grouping into conversation turns
│   └── totals.go            # token aggregation, ranking, context window
├── format/
│   ├── misc.go              # helpers (FormatTokenCount, TruncatePreview)
│   ├── model.go             # ViewModel for display
│   ├── snapshot.go          # text snapshot formatting
│   └── statusline.go        # compact statusline formatting
└── tui/
    └── watch.go             # live dashboard (Bubble Tea + Lipgloss)
```
