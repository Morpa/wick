package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var setupForce bool

var setupCmd = &cobra.Command{
	Use:   "setup [--force]",
	Short: "Configures integration with Claude Code (skill, /tokens, statusLine)",
	Long: `Configures wick's integration with Claude Code:

1. Creates a symlink for the skill at ~/.claude/skills/token-monitor
2. Copies the /tokens command to ~/.claude/commands/tokens.md
3. Registers the statusLine in ~/.claude/settings.json

Use --force to overwrite an existing statusLine.
`,
	Run: func(_ *cobra.Command, _ []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not determine home directory\n")
			os.Exit(1)
		}

		claudeHome := filepath.Join(home, ".claude")
		skillDir := filepath.Join(claudeHome, "skills", "token-monitor")
		commandFile := filepath.Join(claudeHome, "commands", "tokens.md")
		settingsFile := filepath.Join(claudeHome, "settings.json")

		// Find the binary path so we can reference it.
		binPath, err := os.Executable()
		if err != nil {
			binPath = "wick"
		}
		binAbs, _ := filepath.Abs(binPath)

		// Create directories
		os.MkdirAll(filepath.Join(claudeHome, "skills"), 0755)
		os.MkdirAll(filepath.Join(claudeHome, "commands"), 0755)

		// 1. Skill directory (we just create the directory — it's not a symlink
		//    since the Go binary doesn't have a "skill folder" like the Node version.
		//    We create a minimal SKILL.md instead.)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not create %s: %v\n", skillDir, err)
			os.Exit(1)
		}
		skillMDFile := filepath.Join(skillDir, "SKILL.md")
		skillMD := fmt.Sprintf(`---
name: token-monitor
description: Shows token consumption for the current Claude Code session — total, most expensive prompt ranking, breakdown by skill/subagent, and %% of context window.
---

# token-monitor

Claude Code skill for monitoring token consumption per session.

## Usage

Run:
`+"```bash"+`
%s snapshot "$(pwd)"
`+"```"+`
and show the output to the user — it's already formatted, no need to reformat it.
`, binAbs)
		if err := os.WriteFile(skillMDFile, []byte(skillMD), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not write %s: %v\n", skillMDFile, err)
			os.Exit(1)
		}
		fmt.Printf("✓ skill configured at %s\n", skillDir)

		// 2. Command file
		cmdContent := fmt.Sprintf(`---
description: Shows a snapshot of token consumption for the current session (total, prompt ranking, breakdown by skill, %% context window).
---

Run:
`+"```bash"+`
%s snapshot "$(pwd)"
`+"```"+`
Show the command's output to the user exactly as it came out (already formatted), without summarizing or reformatting it.
`, binAbs)
		if err := os.WriteFile(commandFile, []byte(cmdContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not write %s: %v\n", commandFile, err)
			os.Exit(1)
		}
		fmt.Printf("✓ /tokens command copied to %s\n", commandFile)

		// 3. StatusLine in settings.json
		settings := make(map[string]any)
		data, err := os.ReadFile(settingsFile)
		if err == nil {
			if err := json.Unmarshal(data, &settings); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s is not valid JSON — fix it manually before running setup.\n", settingsFile)
				os.Exit(1)
			}
		}

		statusLineConfig := map[string]any{
			"type":    "command",
			"command": fmt.Sprintf("cat /dev/stdin | %s statusline", binAbs),
		}

		if existing, ok := settings["statusLine"]; ok {
			existingJSON, _ := json.Marshal(existing)
			newJSON, _ := json.Marshal(statusLineConfig)
			if string(existingJSON) == string(newJSON) {
				fmt.Println("✓ statusLine already configured correctly")
			} else if !setupForce {
				fmt.Println("⚠ statusLine not changed: statusLine already configured with a different command — run `wick setup --force` to overwrite")
			} else {
				settings["statusLine"] = statusLineConfig
				writeSettings(settingsFile, settings)
				fmt.Printf("✓ statusLine configured at %s\n", settingsFile)
			}
		} else {
			settings["statusLine"] = statusLineConfig
			writeSettings(settingsFile, settings)
			fmt.Printf("✓ statusLine configured at %s\n", settingsFile)
		}

		fmt.Println("\nDone. Open a new Claude Code session to see the statusline, and run `wick watch` or `wick snapshot` whenever you want.")
	},
}

func writeSettings(path string, settings map[string]any) {
	data, _ := json.MarshalIndent(settings, "", "  ")
	data = append(data, '\n')
	os.WriteFile(path, data, 0644)
}

func init() {
	setupCmd.Flags().BoolVarP(&setupForce, "force", "", false, "Overwrite existing statusLine")
	rootCmd.AddCommand(setupCmd)
}
