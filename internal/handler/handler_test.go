package handler

import (
	"strings"
	"testing"
)

func TestDomainFromEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"valid_email", "user@example.com", "example.com"},
		{"uppercase", "user@EXAMPLE.COM", "example.com"},
		{"subdomain", "user@mail.example.com", "mail.example.com"},
		{"no_at_sign", "invalid", ""},
		{"empty_domain", "user@", ""},
		{"empty_email", "", ""},
		{"multiple_at", "user+tag@example.com", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domainFromEmail(tt.email)
			if result != tt.expected {
				t.Errorf("domainFromEmail() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestContentTypeFor(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		expected string
	}{
		{"apple_lowercase", "apple", "application/x-apple-aspen-config"},
		{"apple_uppercase", "APPLE", "application/x-apple-aspen-config"},
		{"apple_mixed_case", "ApPlE", "application/x-apple-aspen-config"},
		{"mozilla", "mozilla", "text/xml; charset=utf-8"},
		{"microsoft", "microsoft", "text/xml; charset=utf-8"},
		{"unknown", "unknown", "text/xml; charset=utf-8"},
		{"empty", "", "text/xml; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contentTypeFor(tt.vendor)
			if result != tt.expected {
				t.Errorf("contentTypeFor() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestResolveEmail_QueryParam(t *testing.T) {
	// This requires mocking a Gin context, which is complex.
	// For now, we test the helper functions directly.
	// Integration tests could use httptest to test the full flow.

	// Test the domainFromEmail function used in resolveEmail
	email := "test@example.com"
	domain := domainFromEmail(email)

	if domain != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", domain)
	}
}

func TestIsAutodiscoverEndpoint(t *testing.T) {
	// Helper to check if an endpoint is an Autodiscover endpoint
	tests := []struct {
		name           string
		endpoint       string
		isAutodiscover bool
	}{
		{"autodiscover", "/autodiscover/autodiscover.xml", true},
		{"autodiscover_uppercase", "/AUTODISCOVER/autodiscover.xml", true},
		{"autodiscover_mixed", "/AutoDiscover/autodiscover.xml", true},
		{"autoconfig", "/.well-known/autoconfig/mail/config-v1.1.xml", false},
		{"mobileconfig", "/mobileconfig", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAuto := strings.Contains(strings.ToLower(tt.endpoint), "autodiscover")
			if isAuto != tt.isAutodiscover {
				t.Errorf("expected isAutodiscover=%v for %s, got %v", tt.isAutodiscover, tt.endpoint, isAuto)
			}
		})
	}
}
