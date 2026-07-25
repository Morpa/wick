package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Morpa/wick/internal/format"
	"github.com/Morpa/wick/internal/session"
	"github.com/spf13/cobra"
)

type statusLineInput struct {
	TranscriptPath string `json:"transcript_path"`
	ContextWindow  *struct {
		UsedPercentage float64 `json:"used_percentage"`
	} `json:"context_window"`
}

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "statusLine mode — reads JSON from stdin (internal use, configured in ~/.claude/settings.json)",
	Long: `statusLine mode for integration with Claude Code's statusLine field.
Reads a JSON from stdin with the format:
  {"transcript_path": "/path/to/session.jsonl", "context_window": {"used_percentage": 70.1}}
Produces a compact line like:
  🔥 session: 7.7M tok · context: 70% used · most expensive prompt: 6.0M tok
`,

	Run: func(_ *cobra.Command, _ []string) {
		var input statusLineInput
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			fmt.Println("token-monitor: no session")
			return
		}

		if input.TranscriptPath == "" {
			fmt.Println("token-monitor: no session")
			return
		}

		events, warnings := session.ParseSession(input.TranscriptPath)
		if len(events) == 0 {
			fmt.Println("token-monitor: no session")
			return
		}

		turns := session.GroupIntoTurns(events)
		totals := session.ComputeTotals(events, turns)
		var ctxOverride *float64
		if input.ContextWindow != nil {
			ctxOverride = &input.ContextWindow.UsedPercentage
		}

		_ = warnings // not shown in status line
		fmt.Println(format.FormatStatusLine(totals, ctxOverride))
	},
}

func init() {
	rootCmd.AddCommand(statuslineCmd)
}
