package common

import "testing"

func TestGenerateHexVerificationCode(t *testing.T) {
	code, err := GenerateHexVerificationCode(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 8 {
		t.Fatalf("expected 8 chars, got %d", len(code))
	}
	for _, r := range code {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') {
			t.Fatalf("expected uppercase hex code, got %q", code)
		}
	}
}

func TestGenerateHexVerificationCodeRejectsInvalidLength(t *testing.T) {
	if _, err := GenerateHexVerificationCode(0); err == nil {
		t.Fatal("expected error for zero length")
	}
}
