package validate

import (
	"context"
	"testing"
)

func TestNewEnvChecker(t *testing.T) {
	domains := []string{"example.com", "test.org", "UPPERCASE.COM"}
	checker := NewEnvChecker(domains)

	if checker == nil {
		t.Fatalf("NewEnvChecker() returned nil")
	}

	if len(checker.domains) != 3 {
		t.Errorf("expected 3 domains, got %d", len(checker.domains))
	}
}

func TestEnvChecker_DomainExists(t *testing.T) {
	domains := []string{"example.com", "test.org"}
	checker := NewEnvChecker(domains)

	tests := []struct {
		name        string
		domain      string
		expectedOk  bool
		expectedErr bool
	}{
		{"exact_match", "example.com", true, false},
		{"case_insensitive_lower", "EXAMPLE.COM", true, false},
		{"case_insensitive_mixed", "ExAmple.CoM", true, false},
		{"not_found", "nonexistent.com", false, false},
		{"partial_domain", "example", false, false},
		{"empty_domain", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ok, err := checker.DomainExists(ctx, tt.domain)

			if (err != nil) != tt.expectedErr {
				t.Errorf("DomainExists() error = %v, expectedErr %v", err, tt.expectedErr)
			}

			if ok != tt.expectedOk {
				t.Errorf("DomainExists() = %v, expected %v", ok, tt.expectedOk)
			}
		})
	}
}

func TestEnvChecker_EmptyDomains(t *testing.T) {
	checker := NewEnvChecker([]string{})

	ctx := context.Background()
	ok, err := checker.DomainExists(ctx, "example.com")

	if ok {
		t.Errorf("expected false for domain check with empty list, got true")
	}

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestEnvChecker_DuplicateDomains(t *testing.T) {
	domains := []string{"example.com", "example.com", "test.org"}
	checker := NewEnvChecker(domains)

	// Map deduplicates, so we should have 2 unique entries
	if len(checker.domains) != 2 {
		t.Errorf("expected 2 unique domains after deduplication, got %d", len(checker.domains))
	}

	ctx := context.Background()
	ok, err := checker.DomainExists(ctx, "example.com")

	if !ok || err != nil {
		t.Errorf("failed to find example.com")
	}
}
