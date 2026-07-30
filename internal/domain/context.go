package domain

// ContextMetrics describes the prompt budget used for the next model request.
// EstimatedInputTokens is available before a request; ActualInputTokens is
// populated when a provider reports final usage for that request.
type ContextMetrics struct {
	WindowTokens         int64
	ReservedOutputTokens int64
	EstimatedInputTokens int64
	ActualInputTokens    int64
}

func (m ContextMetrics) UsableInputTokens() int64 {
	usable := m.WindowTokens - m.ReservedOutputTokens
	if usable < 0 {
		return 0
	}
	return usable
}

func (m ContextMetrics) UsedInputTokens() int64 {
	if m.ActualInputTokens > 0 {
		return m.ActualInputTokens
	}
	return m.EstimatedInputTokens
}

func (m ContextMetrics) UsagePercent() int64 {
	usable := m.UsableInputTokens()
	if usable == 0 {
		return 0
	}
	return m.UsedInputTokens() * 100 / usable
}
