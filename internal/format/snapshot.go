package format

import (
	"fmt"
	"strings"

	"github.com/Morpa/wick/internal/session"
)

// FormatSnapshot produces the plain-text snapshot output for the snapshot
// command. Matches the format of the original Node.js implementation.
func FormatSnapshot(t session.Totals, sessionID, projectLabel string, warnings int, topN int) string {
	if topN <= 0 {
		topN = 5
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Active session: %s (%s)\n", sessionID, projectLabel)

	total := t.Usage.InputTokens + t.Usage.OutputTokens +
		t.Usage.CacheCreationInput + t.Usage.CacheReadInputTokens

	fmt.Fprintf(&b,
		"Total: %s tok  (in %s / out %s / cache-w %s / cache-r %s)\n",
		FormatTokenCount(total),
		FormatTokenCount(t.Usage.InputTokens),
		FormatTokenCount(t.Usage.OutputTokens),
		FormatTokenCount(t.Usage.CacheCreationInput),
		FormatTokenCount(t.Usage.CacheReadInputTokens),
	)

	if t.ContextWindow != nil {
		fmt.Fprintf(&b,
			"Context window: %.1f%% (%s / %s)\n",
			t.ContextWindow.UsedPercentage,
			FormatTokenCount(t.ContextWindow.TotalInputTokens),
			FormatTokenCount(t.ContextWindow.WindowSize),
		)
	}

	b.WriteString("\nTop prompts:\n")
	top := t.TopTurns

	if len(top) > topN {
		top = top[:topN]
	}

	if len(top) == 0 {
		b.WriteString(" (no prompts recorded yet)\n")
	} else {
		for i, turn := range top {
			total := turn.Usage.InputTokens + turn.Usage.OutputTokens +
				turn.Usage.CacheCreationInput + turn.Usage.CacheReadInputTokens
			preview := turn.TextPreview
			if preview == "" {
				preview = "(no preview)"
			}
			fmt.Fprintf(&b, " %d. %s  \"%s\"\n",
				i+1, FormatTokenCount(total), TruncatePreview(preview, 60))
		}
	}

	b.WriteString("\nBy skill/agent:\n")

	if len(t.BySkill) == 0 {
		b.WriteString(" (no data)\n")
	} else {
		for _, s := range t.BySkill {
			fmt.Fprintf(&b, " %-12s %s\n", s.Skill, FormatTokenCount(s.Total))
		}
	}

	if warnings > 0 {
		fmt.Fprintf(&b, "\n(%d line(s) skipped due to corrupted log)\n", warnings)
	}

	return b.String()
}
