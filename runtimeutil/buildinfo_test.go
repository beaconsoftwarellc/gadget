package runtimeutil

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInfo(t *testing.T) {
	b := NewBuildInfo()
	assert.NotEmpty(t, b.GoVersion)

	b.BuildInfo.Settings = append(b.BuildInfo.Settings,
		debug.BuildSetting{Key: string(BuildInfoRevision), Value: "hash"},
		debug.BuildSetting{Key: string(BuildInfoTime), Value: ""},
	)

	tcs := []struct {
		name  string
		key   BuildInfoSetting
		value string
		found bool
	}{
		{
			name:  "missing key",
			key:   BuildInfoSetting("missing"),
			value: "",
			found: false,
		},
		{
			name:  "found key",
			key:   BuildInfoRevision,
			value: "hash",
			found: true,
		},
		{
			name:  "found key empty value",
			key:   BuildInfoTime,
			value: "",
			found: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			value, found := b.GetStringSetting(tc.key)

			assert.Equal(t, tc.value, value)
			assert.Equal(t, tc.found, found)
		})
	}

}
