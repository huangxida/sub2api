package service

import (
	"testing"

	openaiwsv2 "github.com/Wei-Shaw/sub2api/internal/service/openai_ws_v2"
	"github.com/stretchr/testify/require"
)

func TestShouldTreatPassthroughClientDisconnectAsSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		relayExit   *openaiwsv2.RelayExit
		relayResult openaiwsv2.RelayResult
		want        bool
	}{
		{
			name: "client disconnect after terminal event",
			relayExit: &openaiwsv2.RelayExit{
				Stage: "client_disconnected",
			},
			relayResult: openaiwsv2.RelayResult{
				TerminalEventType: "response.completed",
			},
			want: true,
		},
		{
			name: "client disconnect before terminal event",
			relayExit: &openaiwsv2.RelayExit{
				Stage: "client_disconnected",
			},
			relayResult: openaiwsv2.RelayResult{},
			want:        false,
		},
		{
			name: "other relay stage with terminal event",
			relayExit: &openaiwsv2.RelayExit{
				Stage: "read_upstream",
			},
			relayResult: openaiwsv2.RelayResult{
				TerminalEventType: "response.completed",
			},
			want: false,
		},
		{
			name:      "nil relay exit",
			relayExit: nil,
			relayResult: openaiwsv2.RelayResult{
				TerminalEventType: "response.completed",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, shouldTreatPassthroughClientDisconnectAsSuccess(tt.relayExit, tt.relayResult))
		})
	}
}
