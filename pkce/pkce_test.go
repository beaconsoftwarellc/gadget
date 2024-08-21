package pkce

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeVerifier(t *testing.T) {
	var tests = []struct {
		name      string
		length    int
		wantError bool
	}{
		{
			name:      "minimum",
			length:    CodeVerifierMinimumLength,
			wantError: false,
		},
		{
			name:      "maximum",
			length:    CodeVerifierMaximumLength,
			wantError: false,
		},
		{
			name:      "too small",
			length:    10,
			wantError: true,
		},
		{
			name:      "too big",
			length:    500,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeVerifier, err := GenerateCodeVerifier(tt.length)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.length, len(codeVerifier))
			}
		})
	}
	// now check the entire range of valid lengths
	for i := CodeVerifierMinimumLength; i <= CodeVerifierMaximumLength; i++ {
		codeVerifier, err := GenerateCodeVerifier(i)
		require.NoError(t, err)
		require.Equal(t, len(codeVerifier), i)
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	expected, err := GenerateCodeVerifier(50)
	require.NoError(t, err)

	_, err = GenerateCodeChallenge("nonsense", expected)
	require.EqualError(t, err, "transformation must be 'S256' or 'plain'")

	_, err = GenerateCodeChallenge(PlainTransformation, "nonsense")
	require.EqualError(t, err,
		"code verifier is not valid see: RFC7636 Section 4.1",
	)

	actual, err := GenerateCodeChallenge(PlainTransformation, expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	codeVerifier, err := GenerateCodeVerifier(CodeVerifierMaximumLength)
	require.NoError(t, err)

	expected, err = GenerateCodeChallenge(S256Transformation, codeVerifier)
	require.NoError(t, err)

	actual, err = GenerateCodeChallenge(S256Transformation, codeVerifier)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	codeVerifier, err = GenerateCodeVerifier(CodeVerifierMaximumLength)
	require.NoError(t, err)

	actual, err = GenerateCodeChallenge(S256Transformation, codeVerifier)
	require.NoError(t, err)
	require.NotEqual(t, expected, actual)
}

func Test_VerifyCodeVerifier(t *testing.T) {
	codeVerifier, err := GenerateCodeVerifier(CodeVerifierMaximumLength)
	require.NoError(t, err)

	codeChallenge, err := GenerateCodeChallenge(S256Transformation, codeVerifier)
	require.NoError(t, err)

	actual := VerifyCodeVerifier(generator.ID("xfm"), codeVerifier, codeChallenge)
	require.EqualError(t, actual, InvalidGrantError.Error())

	actual = VerifyCodeVerifier(S256Transformation, generator.ID("verify"), codeChallenge)
	require.EqualError(t, actual, InvalidGrantError.Error())

	actual = VerifyCodeVerifier(S256Transformation, codeVerifier, generator.ID("code"))
	require.EqualError(t, actual, InvalidGrantError.Error())

	actual = VerifyCodeVerifier(PlainTransformation, codeVerifier, codeChallenge)
	require.EqualError(t, actual, InvalidGrantError.Error())

	actual = VerifyCodeVerifier(S256Transformation, codeVerifier, codeChallenge)
	require.NoError(t, actual)

	codeVerifier2, err := GenerateCodeVerifier(CodeVerifierMaximumLength)
	require.NoError(t, err)

	actual = VerifyCodeVerifier(S256Transformation, codeVerifier2, codeChallenge)
	require.EqualError(t, actual, InvalidGrantError.Error())
}
