package email

import (
	"net/mail"
)

// Validator provides email validation using a set of validation functions.
type Validator struct {
	validateFuncs []func(email string) (bool, error)
}

// NewValidator returns a new Validator with no validation functions.
func NewValidator() Validator {
	return Validator{}
}

// WithFormValidation adds form-based email validation to the Validator.
// It returns a new Validator with the validateForm function appended.
func (v Validator) WithFormValidation() Validator {
	v.validateFuncs = append(v.validateFuncs, validateForm)
	return v
}

// Validate runs all validation functions on the provided email string.
// It returns true if all validations pass, or false and an error if any fail.
func (v Validator) Validate(email string) (bool, error) {
	for _, validateFn := range v.validateFuncs {
		ok, err := validateFn(email)
		if !ok || err != nil {
			return ok, err
		}
	}

	return true, nil
}

func validateForm(email string) (bool, error) {
	_, err := mail.ParseAddress(email)
	return err == nil, nil
}
