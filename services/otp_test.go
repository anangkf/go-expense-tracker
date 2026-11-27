package services

import (
	"regexp"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	re := regexp.MustCompile(`^\d{6}$`)
	for i := 0; i < 200; i++ {
		code, err := GenerateOTP()
		if err != nil {
			t.Fatal(err)
		}
		if !re.MatchString(code) {
			t.Fatalf("unexpected OTP format: %q", code)
		}
	}
}

func TestHashOTP(t *testing.T) {
	base := HashOTP("secret", 1, "123456")
	if base != HashOTP("secret", 1, "123456") {
		t.Fatal("hash must be deterministic")
	}
	if base == HashOTP("secret", 2, "123456") {
		t.Fatal("hash must depend on user id")
	}
	if base == HashOTP("other", 1, "123456") {
		t.Fatal("hash must depend on secret")
	}
	if base == HashOTP("secret", 1, "654321") {
		t.Fatal("hash must depend on code")
	}
}
