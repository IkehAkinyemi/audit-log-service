package utils

import (
	"strings"
	"testing"
)

func TestRandomServiceID(t *testing.T) {
	serviceID := RandomServiceID()
	if len(serviceID) < 6 {
		t.Errorf("Expected length of service ID to be at least 6, but got %d", len(serviceID))
	}

	if !strings.Contains(alphabets, serviceID[:6]) {
		t.Errorf("Expected service ID to contain only characters from alphabets, but got %s", serviceID[:6])
	}
}
