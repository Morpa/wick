package cmd

import (
	"os"
	"path/filepath"

	"github.com/Morpa/wick/internal/tui"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [cwd]",
	Short: "Live dashboard in the terminal (auto-updates)",
	Long: `Starts a live dashboard that automatically updates every few seconds.
If [cwd] is not provided, the current directory is used.
Press 'q' or Ctrl+C to quit.
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		cwd := "."
		if len(args) > 0 {
			cwd = args[0]
		}

		absCwd, err := filepath.Abs(cwd)
		if err != nil {
			absCwd, _ = os.Getwd()
		}

		tui.RunWatch(absCwd)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
