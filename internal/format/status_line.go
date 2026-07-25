package format

import (
	"fmt"
	"strings"

	"github.com/Morpa/wick/internal/session"
)

// FormatStatusLine produces a compact one-line status string suitable for
// the Claude Code statusLine field.
func FormatStatusLine(t session.Totals, contextWindowOverride *float64) string {
	total := t.Usage.InputTokens + t.Usage.OutputTokens +
		t.Usage.CacheCreationInput + t.Usage.CacheReadInputTokens

	parts := []string{fmt.Sprintf("🔥 session: %s tok", FormatTokenCount(total))}

	var pct float64
	if contextWindowOverride != nil {
		pct = *contextWindowOverride
	} else if t.ContextWindow != nil {
		pct = t.ContextWindow.UsedPercentage
	}

	if pct > 0 {
		parts = append(parts, fmt.Sprintf("context: %.0f%% used", pct))
	}

	if len(t.TopTurns) > 0 {
		top := t.TopTurns[0]
		topTotal := top.Usage.InputTokens + top.Usage.OutputTokens +
			top.Usage.CacheCreationInput + top.Usage.CacheReadInputTokens
		parts = append(parts, fmt.Sprintf("most expensive prompt: %s tok", FormatTokenCount(topTotal)))
	}

	return strings.Join(parts, " · ")
}
