package format

import (
	"fmt"

	"github.com/Morpa/wick/internal/session"
)

// TopRow is one entry in the top prompts ranking.
type TopRow struct {
	Rank    int
	Tokens  string
	Preview string
}

// SkillRow is one entry in the by-skill breakdown.
type SkillRow struct {
	Skill  string
	Tokens string
}

// ViewModel is the data model ready for rendering in either text snapshot or
// TUI dashboard. All values are pre-formatted strings.
type ViewModel struct {
	Header            string
	TotalTokens       string
	Breakdown         string
	ContextPercentage float64 // -1 means "no data"
	ContextSeverity   string  // "ok", "warn", "danger", or ""
	ContextDetail     string
	TopRows           []TopRow
	SkillRows         []SkillRow
	WarningsLine      string
}

// BuildViewModel converts computed totals into a display-ready ViewModel.
func BuildViewModel(t session.Totals, sessionID, projectLabel string, warnings int) ViewModel {
	header := fmt.Sprintf("wick — %s (%s)", projectLabel, sessionID)
	totalTokens := FormatTokenCount(t.Usage.InputTokens + t.Usage.OutputTokens +
		t.Usage.CacheCreationInput + t.Usage.CacheReadInputTokens)
	breakdown := fmt.Sprintf("in %s · out %s · cache-w %s · cache-r %s",
		FormatTokenCount(t.Usage.InputTokens),
		FormatTokenCount(t.Usage.OutputTokens),
		FormatTokenCount(t.Usage.CacheCreationInput),
		FormatTokenCount(t.Usage.CacheReadInputTokens),
	)
	var ctxPct float64 = -1
	var ctxSeverity string
	var ctxDetail string
	if t.ContextWindow != nil {
		ctxPct = t.ContextWindow.UsedPercentage
		switch {
		case ctxPct >= 80:
			ctxSeverity = "danger"
		case ctxPct >= 50:
			ctxSeverity = "warn"
		default:
			ctxSeverity = "ok"
		}
		ctxDetail = fmt.Sprintf("%s / %s",
			FormatTokenCount(t.ContextWindow.TotalInputTokens),
			FormatTokenCount(t.ContextWindow.WindowSize),
		)
	}
	topN := min(8, len(t.TopTurns))
	topRows := make([]TopRow, 0, topN)
	for i, turn := range t.TopTurns[:topN] {
		total := turn.Usage.InputTokens + turn.Usage.OutputTokens +
			turn.Usage.CacheCreationInput + turn.Usage.CacheReadInputTokens
		topRows = append(topRows, TopRow{
			Rank:    i + 1,
			Tokens:  FormatTokenCount(total),
			Preview: turn.TextPreview,
		})
	}
	skillRows := make([]SkillRow, 0, len(t.BySkill))
	for _, s := range t.BySkill {
		skillRows = append(skillRows, SkillRow{
			Skill:  s.Skill,
			Tokens: FormatTokenCount(s.Total),
		})
	}
	var warnLine string
	if warnings > 0 {
		warnLine = fmt.Sprintf("%d linha(s) ignorada(s) (log corrompido)", warnings)
	}
	return ViewModel{
		Header:            header,
		TotalTokens:       totalTokens,
		Breakdown:         breakdown,
		ContextPercentage: ctxPct,
		ContextSeverity:   ctxSeverity,
		ContextDetail:     ctxDetail,
		TopRows:           topRows,
		SkillRows:         skillRows,
		WarningsLine:      warnLine,
	}
}
