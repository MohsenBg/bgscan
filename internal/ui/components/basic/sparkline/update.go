package sparkline

// Push adds a sample, applying EMA smoothing when enabled.
func (m *Model) Push(v float64) {
	if n := len(m.samples); n > 0 && m.smoothing > 0 {
		v = m.smoothing*v + (1-m.smoothing)*m.samples[n-1]
	}
	m.samples = append(m.samples, v)
	m.trim()
	m.dirty = true
}

// SetSamples replaces all samples without smoothing.
func (m *Model) SetSamples(samples []float64) {
	m.samples = append(m.samples[:0], samples...)
	m.trim()
	m.dirty = true
}

// Reset clears all samples.
func (m *Model) Reset() {
	m.samples = m.samples[:0]
	m.dirty = true
}

// SetWidth changes the chart width. Values below 1 are ignored.
func (m *Model) SetWidth(width int) {
	if width < 1 || width == m.width {
		return
	}
	m.width = width
	m.trim()
	m.dirty = true
}

// trim removes the oldest samples beyond capacity.
func (m *Model) trim() {
	if extra := len(m.samples) - m.capacity(); extra > 0 {
		m.samples = append(m.samples[:0], m.samples[extra:]...)
	}
}
