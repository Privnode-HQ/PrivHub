package ratio_setting

import "testing"

func TestCheckGroupCaptureRate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty defaults to zero", input: "", wantErr: false},
		{name: "blank defaults to zero", input: "   ", wantErr: false},
		{name: "zero", input: `{"default":0}`, wantErr: false},
		{name: "one", input: `{"default":1}`, wantErr: false},
		{name: "fraction", input: `{"default":0.25}`, wantErr: false},
		{name: "negative", input: `{"default":-0.01}`, wantErr: true},
		{name: "over one", input: `{"default":1.01}`, wantErr: true},
		{name: "invalid json", input: `{`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckGroupCaptureRate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckGroupCaptureRate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateGroupCaptureRateByJSONStringEmptyClearsRates(t *testing.T) {
	original := GroupCaptureRate2JSONString()
	t.Cleanup(func() {
		if err := UpdateGroupCaptureRateByJSONString(original); err != nil {
			t.Fatalf("restore group capture rate: %v", err)
		}
	})

	if err := UpdateGroupCaptureRateByJSONString(`{"captured":0.5}`); err != nil {
		t.Fatalf("set group capture rate: %v", err)
	}
	if got := GetGroupCaptureRate("captured"); got != 0.5 {
		t.Fatalf("GetGroupCaptureRate() = %v, want 0.5", got)
	}
	if err := UpdateGroupCaptureRateByJSONString(""); err != nil {
		t.Fatalf("clear group capture rate: %v", err)
	}
	if got := GetGroupCaptureRate("captured"); got != 0 {
		t.Fatalf("GetGroupCaptureRate() after clear = %v, want 0", got)
	}
}
