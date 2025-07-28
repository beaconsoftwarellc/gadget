package email

import (
	_ "embed"
	"slices"
	"strings"
)

type blocklist []string

//go:embed resources/blocklist.txt
var rawDomains string

// NewBlockList creates a new domain blocklist from an embedded list of allowed domains.
// Returns a blocklist that can be used to validate if email domains are allowed.
func NewBlockList() blocklist {
	var bl []string
	for _, domainLine := range strings.Split(rawDomains, "\n") {
		domainStripped := strings.TrimSpace(domainLine)
		if len(domainStripped) == 0 || strings.HasPrefix(domainStripped, "#") {
			continue
		}

		bl = append(bl, domainStripped)
	}
	return blocklist(bl)
}

// Validate checks if the provided domain is present in the blocklist.
// It performs an exact string match against all domains in the blocklist.
// Returns true if the domain is NOT found in the blocklist, false otherwise.
func (w blocklist) Validate(domain string) bool {
	return !slices.Contains(w, domain)
}
