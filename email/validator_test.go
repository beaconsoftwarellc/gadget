package email_test

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/email"
	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate_WithFormValidation(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator()
	assert := assert.New(t)

	tests := []struct {
		email    string
		expected bool
	}{
		{
			email:    "test@example.com",
			expected: true,
		},
		{
			email:    "user@domain.co.uk",
			expected: true,
		},
		{
			email:    "invalid-email",
			expected: false,
		},
		{
			email:    "@missinglocal.com",
			expected: false,
		},
		{
			email:    "missingatsign.com",
			expected: false,
		},
		{
			email:    "missingdomain@",
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

	validator := email.NewValidator(email.WithFormValidation(false))
	assert := assert.New(t)

	emails := []string{
		"test@example.com",
		"invalid-email",
		"",
		"@missinglocal.com",
	}

	for _, e := range emails {
		ok, err := validator.Validate(e)
		assert.True(ok)
		assert.NoError(err)
	}
}

func TestValidator_Validate_WithDNSValidation(t *testing.T) {
	t.Parallel()

	validator := email.NewValidator(
		email.WithFormValidation(false),
		email.WithDNSValidation(true),
	)

	assert := assert.New(t)

	tests := []struct {
		email       string
		expected    bool
		expectedErr string
	}{
		{
			email:    "test@example.com",
			expected: true,
		},
		{
			email:    "user@domain.co.uk",
			expected: true,
		},
		{
			email:       "invalid-email",
			expected:    false,
			expectedErr: `'@' not found in address: "invalid-email"`,
		},
		{
			email:       "@missinglocal.com",
			expected:    false,
			expectedErr: "no such host",
		},
		{
			email:       "missingatsign.com",
			expected:    false,
			expectedErr: `'@' not found in address: "missingatsign.com"`,
		},
		{
			email:       "missingdomain@",
			expected:    false,
			expectedErr: "lookup : no such host",
		},
		{
			email:       "",
			expected:    false,
			expectedErr: `'@' not found in address: ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			ok, err := validator.Validate(tt.email)
			assert.Equal(tt.expected, ok)
			if !tt.expected {
				assert.ErrorContains(err, tt.expectedErr)
			}
		})
	}
}
