package validate

import (
	"context"
	"strings"
)

// DomainChecker is satisfied by both the DB-backed checker and the env-var checker.
type DomainChecker interface {
	DomainExists(ctx context.Context, domain string) (bool, error)
	UserExists(ctx context.Context, email string) (bool, error)
}

// EnvChecker validates domains against a static allow-list loaded from
// SUPPORTED_DOMAINS. UserExists always returns true when using this checker
// because env-var mode does not enumerate individual users.
type EnvChecker struct {
	domains map[string]struct{}
}

// NewEnvChecker creates an EnvChecker from the pre-parsed domain slice in config.
func NewEnvChecker(domains []string) *EnvChecker {
	m := make(map[string]struct{}, len(domains))
	for _, d := range domains {
		m[strings.ToLower(d)] = struct{}{}
	}
	return &EnvChecker{domains: m}
}

func (e *EnvChecker) DomainExists(_ context.Context, domain string) (bool, error) {
	_, ok := e.domains[strings.ToLower(domain)]
	return ok, nil
}

// UserExists always returns true in env-var mode — individual users are not
// enumerated, only the domain is checked.
func (e *EnvChecker) UserExists(_ context.Context, _ string) (bool, error) {
	return true, nil
}
