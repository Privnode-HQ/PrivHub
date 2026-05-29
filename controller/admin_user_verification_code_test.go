package controller

import (
	"strings"
	"testing"
)

func TestNormalizeAdminUserVerificationCodePurpose(t *testing.T) {
	purpose, err := normalizeAdminUserVerificationCodePurpose("  账号迁移人工确认  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if purpose != "账号迁移人工确认" {
		t.Fatalf("expected trimmed purpose, got %q", purpose)
	}
}

func TestNormalizeAdminUserVerificationCodePurposeRejectsAuthPurpose(t *testing.T) {
	if _, err := normalizeAdminUserVerificationCodePurpose("登录确认"); err == nil {
		t.Fatal("expected auth purpose to be rejected")
	}
	if _, err := normalizeAdminUserVerificationCodePurpose("password reset"); err == nil {
		t.Fatal("expected password reset purpose to be rejected")
	}
}

func TestNormalizeAdminUserVerificationCodePurposeRejectsControlCharacters(t *testing.T) {
	if _, err := normalizeAdminUserVerificationCodePurpose("账号迁移\n人工确认"); err == nil {
		t.Fatal("expected newline to be rejected")
	}
}

func TestBuildAdminUserVerificationCodeEmailContent(t *testing.T) {
	content := buildAdminUserVerificationCodeEmailContent("PrivHub", "<script>alert(1)</script>", "a1b2c3d4")
	if !strings.Contains(content, "A1B2C3D4") {
		t.Fatalf("expected uppercase code in content, got %q", content)
	}
	if !strings.Contains(content, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped purpose in content, got %q", content)
	}
	if !strings.Contains(content, "不是登录验证码") {
		t.Fatalf("expected non-login disclaimer, got %q", content)
	}
}
