package cmd

import (
	"fmt"
	"github.com/Morpa/wick/internal/format"
	"github.com/Morpa/wick/internal/session"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [cwd]",
	Short: "Text snapshot: total, prompt ranking, breakdown by skill, context window",
	Long: `Shows a text snapshot with the token consumption of the active session.
If [cwd] is not provided, the current directory is used.
The command reads the active session's JSONL transcript from ~/.claude/projects/.
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		cwd := "."
		if len(args) > 0 {
			cwd = args[0]
		}

		absCwd, err := filepath.Abs(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid path %q\n", cwd)
			os.Exit(1)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not determine home directory\n")
			os.Exit(1)
		}

		encoded := session.EncodeProjectDir(absCwd)
		projectDir := filepath.Join(home, ".claude", "projects", encoded)
		sessionFile, ok := session.FindActiveSession(projectDir)
		if !ok {
			fmt.Printf("wick: no session found in %s\n", projectDir)
			fmt.Println("Run this command from inside a project with an active Claude Code session.")
			return
		}

		events, warnings := session.ParseSession(sessionFile)
		turns := session.GroupIntoTurns(events)
		totals := session.ComputeTotals(events, turns)
		sessionID := filepath.Base(sessionFile)
		// strip .jsonl extension — the original Node code uses a hard slice
		sessionID = sessionID[:len(sessionID)-6] // ".jsonl" is 6 chars
		projectLabel := filepath.Base(absCwd)
		output := format.FormatSnapshot(totals, sessionID, projectLabel, warnings, 5)
		fmt.Println(output)
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
