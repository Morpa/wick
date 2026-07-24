package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wick",
	Short: "Claude Code Token Usage Monitor",
	Long: `Claude Code Token Usage Monitor is a command-line tool that allows you to monitor and track the usage of tokens in your codebase. It provides insights into token usage patterns, helping you optimize your code and improve performance.

Reads the JSONL transcript that Claude Code itself already saves per session
(~/.claude/projects/<project>/<session>.jsonl) and displays:
  • Total tokens spent
  • Ranking of most expensive prompts
  • Breakdown by skill/subagent
  • % of context window used

Available commands:
  setup [--force]   Configures integration with Claude Code (skill, /tokens, statusLine)
  snapshot [cwd]    Text snapshot with total, ranking, and context window
  watch [cwd]       Live dashboard in the terminal (updates automatically)
  statusline        StatusLine mode — reads JSON from stdin (internal use)
  help              Displays this help

Translated with DeepL.com (free version)
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fmt.Errorf("wick: unknown command %q\n See 'wick help'", args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.EnableCommandSorting = false
}
