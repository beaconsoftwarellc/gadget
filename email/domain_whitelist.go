package email

import (
	_ "embed"
	"strings"
)

type whitelist []string

//go:embed resources/whitelist.txt
var rawDomains string

// NewWhitelist creates a new domain whitelist from an embedded list of allowed domains.
// Returns a whitelist that can be used to validate if email domains are allowed.
func NewWhitelist() whitelist {
	var wl []string
	for _, domainLine := range strings.Split(rawDomains, "\n") {
		domainStripped := strings.TrimSpace(domainLine)
		if len(domainStripped) == 0 || strings.HasPrefix(domainStripped, "#") {
			continue
		}

		wl = append(wl, domainStripped)
	}
	return whitelist(wl)
}

// Validate checks if the provided domain is present in the whitelist.
// It performs an exact string match against all domains in the whitelist.
// Returns true if the domain is found in the whitelist, false otherwise.
func (w whitelist) Validate(domain string) bool {
	for _, validDomain := range w {
		if domain == validDomain {
			return true
		}
	}

	return false
}
