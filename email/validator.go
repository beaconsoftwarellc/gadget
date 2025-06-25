package email

import (
	"net/mail"
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
	return true, nil
}

func (v Validator) validateForm(email string) bool {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return true
}
