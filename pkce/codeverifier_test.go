package pkce

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_codeVerifierIsValid(t *testing.T) {
	var tests = []struct {
		name string
		code string
		want bool
	}{
		{
			name: "empty",
			code: "",
			want: false,
		},
		{
			name: "too short",
			code: "abacadaba",
			want: false,
		},
		{
			name: "too long",
			code: "01234567890" + // 10
				"1234567890" +
				"1234567890" +
				"1234567890" +
				"1234567890" + // 50
				"1234567890" +
				"1234567890" +
				"1234567890" +
				"1234567890" +
				"1234567890" + // 100
				"1234567890" +
				"1234567890" + // 120
				"12345678", // 129
			want: false,
		},
		{
			name: "valid",
			code: "ThisIsAValidCodeVerifier0" +
				"12345678901234567890123456789-_~.",
			want: true,
		},
		{
			name: "invalid character",
			code: "ThisIsAValidCodeVerifier0" +
				"12345678901234567890123456789-_~.[",
			want: false,
		},
		{
			name: "invalid characters",
			code: "]ThisIsAValidCodeVerifier0" +
				"1234567890123%^&4567890123456789-_~.[",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codeVerifierIsValid(tt.code); got != tt.want {
				t.Errorf("codeVerifierIsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_s256(t *testing.T) {
	// we are just verifying it does not explode
	codeVerifier, err := GenerateCodeVerifier(50)
	require.NoError(t, err)
	actual := s256(codeVerifier)
	require.NotEmpty(t, actual)

	codeVerifier2, err := GenerateCodeVerifier(50)
	require.NoError(t, err)
	actual2 := s256(codeVerifier2)
	require.NotEmpty(t, actual2)

	// sanity check
	require.NotEqual(t, codeVerifier, codeVerifier2)
	// should be different
	require.NotEqual(t, actual, actual2)

	// hash should be stable
	actual3 := s256(codeVerifier2)
	require.Equal(t, actual2, actual3)
}
