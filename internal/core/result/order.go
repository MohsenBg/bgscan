package result

// Less compares two scan results by latency, then by IP
func (a IPScanResult) Less(b IPScanResult) bool {
	scoreA := a.Score()
	scoreB := b.Score()

	if scoreA != scoreB {
		return scoreB < scoreA
	}

	// Tie-breaker
	return a.IP < b.IP
}

func (a IPScanResult) Equal(b IPScanResult) bool {
	return a.IP == b.IP
}

// Score calculates a single numeric value representing the quality of the IP.
// Higher score = Better IP.
func (a IPScanResult) Score() float64 {
	latMs := float64(a.Latency.Milliseconds())
	if latMs < 1 {
		latMs = 1
	}

	downloadWeight := float64(a.Download) * 0.60
	uploadWeight := float64(a.Upload) * 0.20
	latencyPenalty := (1000.0 / latMs) * 500.0

	return downloadWeight + uploadWeight + latencyPenalty
}
