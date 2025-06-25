package email

import (
	_ "embed"
	"strings"
)

type whitelist []string

//go:embed resources/whitelist.txt
var rawDomains string

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

func (w whitelist) Validate(domain string) bool {
	for _, validDomain := range w {
		if domain == validDomain {
			return true
		}
	}

	return false
}
