package render

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"go-autoconfig/internal/loader"
)

// Data is the context passed to every template at render time.
type Data struct {
	EmailAddress string
	Domain       string
	// Vars contains merged vendor defaults + TMPL_<VENDOR>_<KEY> env overrides.
	Vars map[string]string
}

// Render executes the template associated with cfg against the given email
// address and writes the result to w.
func Render(w io.Writer, tmpl *template.Template, cfg *loader.ClientConfig, emailAddress string) error {
	domain, err := domainFromEmail(emailAddress)
	if err != nil {
		return err
	}

	data := Data{
		EmailAddress: emailAddress,
		Domain:       domain,
		Vars:         cfg.Vars,
	}

	// Use a buffer so a mid-render error doesn't produce a partial response.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("rendering template for vendor %s: %w", cfg.Vendor, err)
	}

	_, err = io.Copy(w, &buf)
	return err
}

func domainFromEmail(email string) (string, error) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("invalid email address %q", email)
	}
	return strings.ToLower(parts[1]), nil
}
