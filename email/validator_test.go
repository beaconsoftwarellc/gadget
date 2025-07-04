package email_test

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/email"
	"github.com/stretchr/testify/assert"
)

const (
	testEmail          = "test@example.com"
	userDomainEmail    = "user@domain.co.uk"
	invalidEmail       = "invalid-email"
	missingLocalEmail  = "@missinglocal.com"
	missingAtSignEmail = "missingatsign.com"
	missingDomainEmail = "missingdomain@"
)

func TestValidator_EmptyValidatorAlwaysPasses(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator()
	assert := assert.New(t)

	tests := []struct {
		email    string
		expected bool
	}{
		{
			email:    testEmail,
			expected: true,
		},
		{
			email:    userDomainEmail,
			expected: true,
		},
		{
			email:    invalidEmail,
			expected: true,
		},
		{
			email:    missingLocalEmail,
			expected: true,
		},
		{
			email:    missingAtSignEmail,
			expected: true,
		},
		{
			email:    missingDomainEmail,
			expected: true,
		},
		{
			email:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			ok, err := validator.Validate(tt.email)
			assert.Equal(tt.expected, ok)
			assert.NoError(err)
		})
	}
}

func TestValidator_ValidateForm(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator().WithFormValidation()
	assert := assert.New(t)

	tests := []struct {
		email    string
		expected bool
	}{
		{
			email:    testEmail,
			expected: true,
		},
		{
			email:    userDomainEmail,
			expected: true,
		},
		{
			email:    invalidEmail,
			expected: false,
		},
		{
			email:    missingLocalEmail,
			expected: false,
		},
		{
			email:    missingAtSignEmail,
			expected: false,
		},
		{
			email:    missingDomainEmail,
			expected: false,
		},
		{
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			ok, err := validator.Validate(tt.email)
			assert.Equal(tt.expected, ok)
			assert.NoError(err)
		})
	}
}

func TestValidator_Validate_WithoutFormValidation(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator()
	assert := assert.New(t)

	emails := []string{
		testEmail,
		invalidEmail,
		"",
		missingLocalEmail,
	}

	for _, e := range emails {
		ok, err := validator.Validate(e)
		assert.True(ok)
		assert.NoError(err)
	}
}

func TestValidator_Validate_WithDNSValidation(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator().WithDNSValidation()
	assert := assert.New(t)

	tests := []struct {
		email       string
		expected    bool
		expectedErr string
	}{
		{
			email:    testEmail,
			expected: true,
		},
		{
			email:    userDomainEmail,
			expected: true,
		},
		{
			email:       invalidEmail,
			expected:    false,
			expectedErr: "'@' not found in address: \"invalid-email\"",
		},
		{
			email:       missingLocalEmail,
			expected:    false,
			expectedErr: "no such host",
		},
		{
			email:       missingAtSignEmail,
			expected:    false,
			expectedErr: "'@' not found in address: \"missingatsign.com\"",
		},
		{
			email:       missingDomainEmail,
			expected:    false,
			expectedErr: "no such host",
		},
		{
			email:       "",
			expected:    false,
			expectedErr: "'@' not found in address: \"\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			ok, err := validator.Validate(tt.email)
			assert.Equal(tt.expected, ok)
			if !tt.expected {
				assert.ErrorContains(err, tt.expectedErr)
			} else {
				assert.NoError(err)
			}
		})
	}
}
