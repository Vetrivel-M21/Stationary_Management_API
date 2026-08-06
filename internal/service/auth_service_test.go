package service_test

import (
	"stationery-management/pkg/hash"
	"testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
	password := "Admin@123"

	hashed, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !hash.CheckPasswordHash(password, hashed) {
		t.Errorf("Expected password hash verification to pass, but failed")
	}

	if hash.CheckPasswordHash("WrongPassword", hashed) {
		t.Errorf("Expected password hash verification to fail for wrong password, but passed")
	}
}
