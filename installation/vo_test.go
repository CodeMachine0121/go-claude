package installation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionVo(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		expected      string
		expectedError error
	}{
		{name: "release with product suffix", text: "2.1.332 (Claude Code)\n", expected: "2.1.332"},
		{name: "release with leading v", text: "v2.1.0", expected: "2.1.0"},
		{name: "text without a release", text: "command not found", expectedError: ErrUnrecognizedVersion},
		{name: "incomplete release", text: "2.1", expectedError: ErrUnrecognizedVersion},
		{name: "release part too large to be real", text: "99999999999999999999.1.0", expectedError: ErrUnrecognizedVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := NewVersionVo(tt.text)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				assert.Contains(t, err.Error(), tt.text)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, version.String())
			}
		})
	}
}

func TestVersionVo_AtLeast(t *testing.T) {
	minimum := VersionVo{major: 2, minor: 1, patch: 0}
	tests := []struct {
		name     string
		version  VersionVo
		expected bool
	}{
		{name: "same release", version: VersionVo{major: 2, minor: 1, patch: 0}, expected: true},
		{name: "newer patch", version: VersionVo{major: 2, minor: 1, patch: 332}, expected: true},
		{name: "older patch line", version: VersionVo{major: 2, minor: 0, patch: 9}, expected: false},
		{name: "newer minor with lower patch", version: VersionVo{major: 2, minor: 2, patch: 0}, expected: true},
		{name: "newer major with lower minor", version: VersionVo{major: 3, minor: 0, patch: 0}, expected: true},
		{name: "older major with higher minor", version: VersionVo{major: 1, minor: 9, patch: 99}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.version.AtLeast(minimum))
		})
	}
}
