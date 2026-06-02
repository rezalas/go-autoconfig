package render

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"go-autoconfig/internal/loader"
)

func TestDomainFromEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		expected  string
		wantError bool
	}{
		{"valid_email", "user@example.com", "example.com", false},
		{"uppercase_domain", "user@EXAMPLE.COM", "example.com", false},
		{"subdomain", "user@mail.example.com", "mail.example.com", false},
		{"no_at_sign", "invalid", "", true},
		{"empty_domain", "user@", "", true},
		{"multiple_at_signs", "user+tag@example.com", "example.com", false},
		{"empty_string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := domainFromEmail(tt.email)

			if (err != nil) != tt.wantError {
				t.Errorf("domainFromEmail() error = %v, wantError %v", err, tt.wantError)
			}

			if domain != tt.expected {
				t.Errorf("domainFromEmail() = %s, expected %s", domain, tt.expected)
			}
		})
	}
}

func TestRender(t *testing.T) {
	// Create a simple test template
	tmpl, err := template.New("test").Parse(`
Email: {{.EmailAddress}}
Domain: {{.Domain}}
Hostname: {{.Vars.HOSTNAME}}
`)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := &loader.ClientConfig{
		Vendor:       "TEST",
		TemplateFile: "/templates/test.tmpl",
		Vars: map[string]string{
			"HOSTNAME": "mail.example.com",
		},
	}

	var buf bytes.Buffer
	err = Render(&buf, tmpl, cfg, "user@example.com")
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	result := buf.String()

	if !strings.Contains(result, "user@example.com") {
		t.Errorf("expected email address in output, got: %s", result)
	}

	if !strings.Contains(result, "example.com") {
		t.Errorf("expected domain in output, got: %s", result)
	}

	if !strings.Contains(result, "mail.example.com") {
		t.Errorf("expected hostname variable in output, got: %s", result)
	}
}

func TestRender_InvalidEmail(t *testing.T) {
	tmpl, err := template.New("test").Parse("Email: {{.EmailAddress}}")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := &loader.ClientConfig{
		Vendor: "TEST",
		Vars:   make(map[string]string),
	}

	var buf bytes.Buffer
	err = Render(&buf, tmpl, cfg, "invalid-email")

	if err == nil {
		t.Errorf("expected error for invalid email, got nil")
	}
}

func TestRender_TemplateError(t *testing.T) {
	// Create a template that will fail during execution
	tmpl, err := template.New("test").Parse("{{.UndefinedField.Method}}")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := &loader.ClientConfig{
		Vendor: "TEST",
		Vars:   make(map[string]string),
	}

	var buf bytes.Buffer
	err = Render(&buf, tmpl, cfg, "user@example.com")

	if err == nil {
		t.Errorf("expected error for template execution failure, got nil")
	}
}

func TestRender_EmptyEmail(t *testing.T) {
	tmpl, err := template.New("test").Parse("Email: {{.EmailAddress}}")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := &loader.ClientConfig{
		Vendor: "TEST",
		Vars:   make(map[string]string),
	}

	var buf bytes.Buffer
	err = Render(&buf, tmpl, cfg, "")

	if err == nil {
		t.Errorf("expected error for empty email, got nil")
	}
}

func TestRender_VariablesAvailableInTemplate(t *testing.T) {
	tmpl, err := template.New("test").Parse("{{.Vars.IMAP_PORT}}")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := &loader.ClientConfig{
		Vendor: "TEST",
		Vars: map[string]string{
			"IMAP_PORT": "993",
		},
	}

	var buf bytes.Buffer
	err = Render(&buf, tmpl, cfg, "user@example.com")
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	result := buf.String()
	if result != "993" {
		t.Errorf("expected '993', got '%s'", result)
	}
}
