package session

// Turn represents one user request and all assistant responses that follow
// (including sidechain / subagent work triggered by it).
type Turn struct {
	ID          string
	Ts          string
	TextPreview string
	Usage       Usage
	BySkill     map[string]int // skill name → tokens
}

// GroupIntoTurns groups a flat event list into conversational turns.
//
// A new turn starts when we see a real user text message (not a tool result,
// not a sidechain). Every subsequent assistant/sidechain event accumulates
// its usage into that turn until the next real user message.
//
// Sidechain events are bucketed under their AttributionSkill (or "subagent"
// if nil) inside the turn's BySkill map.
func GroupIntoTurns(events []NormalizedEvent) []Turn {
	var turns []Turn
	var current *Turn

	for _, evt := range events {
		if evt.Role == "user" && evt.IsRealUserText && !evt.IsSidechain {
			current = &Turn{
				ID:          evt.UUID,
				Ts:          evt.Ts,
				TextPreview: evt.TextPreview,
				Usage:       Usage{},
				BySkill:     map[string]int{},
			}
			turns = append(turns, *current)
			continue
		}

		if current == nil || evt.Usage == nil {
			continue
		}

		bucket := "main"
		if evt.IsSidechain {
			if evt.AttributionSkill != "" {
				bucket = evt.AttributionSkill
			} else {
				bucket = "subagent"
			}
		}

		turns[len(turns)-1].Usage.InputTokens += evt.Usage.InputTokens
		turns[len(turns)-1].Usage.OutputTokens += evt.Usage.OutputTokens
		turns[len(turns)-1].Usage.CacheCreationInput += evt.Usage.CacheCreationInput
		turns[len(turns)-1].Usage.CacheReadInputTokens += evt.Usage.CacheReadInputTokens

		total := evt.Usage.InputTokens + evt.Usage.OutputTokens +
			evt.Usage.CacheCreationInput + evt.Usage.CacheReadInputTokens

		// Safety: if the current pointer got detached by a copy, re-index.
		turns[len(turns)-1].BySkill[bucket] += total
	}

	return turns
}
