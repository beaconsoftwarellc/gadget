package email

import (
	"fmt"
	"net"
	"net/mail"
	"strings"
)

type Validator struct {
	config Config
}

type optFunc func(Config) Config

func NewValidator(opts ...optFunc) Validator {
	cfg := newConfig()

	for _, optFn := range opts {
		cfg = optFn(cfg)
	}

	return Validator{
		config: cfg,
	}
}

func (v Validator) Validate(email string) (bool, error) {
	if v.config.ValidateForm {
		ok := v.validateForm(email)
		if !ok {
			return ok, nil
		}
	}

	if v.config.ValidateDNS {
		ok, err := v.validateDNS(email)
		if !ok {
			return ok, err
		}
	}

	if v.config.ValidateDisposable {
		ok, err := v.validateDisposable(email)
		if !ok {
			return ok, err
		}
	}

	return true, nil
}

func (v Validator) validateForm(email string) bool {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return true
}

func (v Validator) validateDNS(email string) (bool, error) {
	atIdx := strings.LastIndex(email, "@")
	if atIdx == -1 {
		return false, fmt.Errorf("'@' not found in address: %q", email)
	}

	domain := email[atIdx+1:]

	nameServers, err := net.LookupNS(domain)
	if err != nil {
		return false, err
	}

	return len(nameServers) > 0, nil
}

func (v Validator) validateDisposable(email string) (bool, error) {
	atIdx := strings.LastIndex(email, "@")
	if atIdx == -1 {
		return false, fmt.Errorf("'@' not found in address: %q", email)
	}

	domain := email[atIdx+1:]

	return NewWhitelist().Validate(domain), nil
}
