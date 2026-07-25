package format

import (
	"fmt"
	"strings"
)

// FormatTokenCount formats a token count into a human-readable string.
//
//	< 1k      → raw number (e.g. "512")
//	< 1_000_000 → "X.Xk" (e.g. "47.5k")
//	else       → "X.XM" (e.g. "7.7M")
func FormatTokenCount(n int) string {
	if n < 1_000 {
		return fmt.Sprintf("%d", n)
	}

	if n < 1_000_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}

	return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
}

// TruncatePreview truncates text to maxLen, adding an ellipsis if cut.
// It trims whitespace and collapses multiple spaces into one.
func TruncatePreview(text string, maxLen int) string {
	if text == "" {
		return ""
	}

	trimmed := strings.Join(strings.Fields(text), " ")
	if len(trimmed) <= maxLen {
		return trimmed
	}

	sliced := trimmed[:maxLen]
	lastSpace := strings.LastIndex(sliced, " ")
	cut := sliced

	if lastSpace > 0 {
		cut = sliced[:lastSpace]
	}

	return cut + "…"
}
