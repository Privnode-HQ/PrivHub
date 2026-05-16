package helper

import "testing"

func TestIsTerminalStreamData(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{name: "openai done", data: "[DONE]", want: true},
		{name: "anthropic message stop", data: `{"type":"message_stop"}`, want: true},
		{name: "anthropic message delta", data: `{"type":"message_delta"}`, want: false},
		{name: "non json", data: "not-json", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTerminalStreamData(tt.data); got != tt.want {
				t.Fatalf("isTerminalStreamData(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}
