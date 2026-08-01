package engine

import (
	"testing"
)

func TestParsePipelineMode(t *testing.T) {
	cases := []struct {
		input string
		want  PipelineMode
	}{
		{"sequential", ModeSequential},
		{"SEQUENTIAL", ModeSequential},
		{"simple", ModeSequential},
		{"streaming", ModeStreaming},
		{"STREAMING", ModeStreaming},
		{"parallel", ModeStreaming},
		{"batch", ModeBatch},
		{"BATCH", ModeBatch},
		{"pipeline", ModeBatch},
		{"", ModeSequential},
		{"unknown", ModeSequential},
		{"  streaming  ", ModeStreaming},
	}

	for _, tc := range cases {
		got := ParsePipelineMode(tc.input)
		if got != tc.want {
			t.Errorf("ParsePipelineMode(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
