package result

import "time"

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
	downloadMs := float64(a.Download) / float64(time.Millisecond)
	uploadMs := float64(a.Upload) / float64(time.Millisecond)
	latencyMs := float64(a.Latency) / float64(time.Millisecond)

	if downloadMs < 1 {
		downloadMs = 1
	}
	if uploadMs < 1 {
		uploadMs = 1
	}
	if latencyMs < 1 {
		latencyMs = 1
	}

	return 0.60*(1000.0/downloadMs) +
		0.20*(1000.0/uploadMs) +
		0.20*(1000.0/latencyMs)
}
