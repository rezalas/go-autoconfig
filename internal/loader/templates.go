package loader

import (
	"fmt"
	"path/filepath"
	"text/template"
)

// TemplateRegistry maps a template file path (as declared in ClientConfig.TemplateFile)
// to its parsed *template.Template.
type TemplateRegistry map[string]*template.Template

// LoadTemplates scans dir for all *.tmpl files, parses each one, and returns
// a registry keyed by the path as it would appear in clientConfigs (e.g.
// "/templates/autoconfig.xml.tmpl").
func LoadTemplates(dir string) (TemplateRegistry, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("scanning templates dir: %w", err)
	}

	registry := make(TemplateRegistry, len(entries))
	for _, path := range entries {
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", path, err)
		}

		// Normalise the key to forward-slash relative path matching the format
		// used in clientConfigs templateFile fields (e.g. "/templates/foo.xml.tmpl").
		key := "/" + filepath.ToSlash(path)
		// Trim any leading directory prefix so the key is rooted at "/templates/..."
		// regardless of the absolute path on disk.
		if rel, err := filepath.Rel(filepath.Dir(dir), path); err == nil {
			key = "/" + filepath.ToSlash(rel)
		}
		registry[key] = tmpl
	}

	return registry, nil
}

// Get returns the template for the given templateFile path, or an error if not found.
func (r TemplateRegistry) Get(templateFile string) (*template.Template, error) {
	tmpl, ok := r[templateFile]
	if !ok {
		return nil, fmt.Errorf("template %q not loaded; ensure the file exists in the templates/ directory", templateFile)
	}
	return tmpl, nil
}
