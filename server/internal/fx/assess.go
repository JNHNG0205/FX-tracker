package fx

// Assess classifies current within [min,max]. Higher MyrUsd is better for the
// buyer, so the top third of the range is "good". The range only spans the
// process lifetime for now.
// TODO(phase2): use persisted rate_samples for a real 30/90-day window.
func Assess(current, min, max float64) string {
	span := max - min
	if span <= 0 {
		return "unknown"
	}
	position := (current - min) / span
	switch {
	case position >= 2.0/3.0:
		return "good"
	case position <= 1.0/3.0:
		return "poor"
	default:
		return "middling"
	}
}
