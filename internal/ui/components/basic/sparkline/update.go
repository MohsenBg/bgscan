package sparkline

// Push adds the newest sample (for example the IP/s of the last tick).
// If smoothing is enabled, it applies an exponential moving average.
func (m *Model) Push(v float64) {
	if n := len(m.samples); n > 0 && m.smoothing > 0 {
		v = m.smoothing*v + (1-m.smoothing)*m.samples[n-1]
	}
	m.samples = append(m.samples, v)
	m.trim()
}

// SetSamples replaces all samples at once. The values are copied.
// Smoothing is not applied; pass pre-smoothed data if needed.
func (m *Model) SetSamples(samples []float64) {
	m.samples = append(m.samples[:0], samples...)
	m.trim()
}

// Reset clears all samples, for example when a new scan starts.
func (m *Model) Reset() {
	m.samples = m.samples[:0]
}

// SetWidth changes the chart width in cells. Values < 1 are ignored.
func (m *Model) SetWidth(width int) {
	if width < 1 {
		return
	}
	m.width = width
	m.trim()
}

// trim keeps only the newest width*2 samples.
func (m *Model) trim() {
	if extra := len(m.samples) - m.capacity(); extra > 0 {
		m.samples = append(m.samples[:0], m.samples[extra:]...)
	}
}
