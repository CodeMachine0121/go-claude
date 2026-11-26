package installation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Verify(t *testing.T) {
	tests := []struct {
		name          string
		observedText  string
		proxyError    error
		expectedDto   InstallationDto
		expectedError error
		expectedText  []string
	}{
		{
			name:         "recent release found on the search path is ready",
			observedText: "2.1.332 (Claude Code)",
			expectedDto:  InstallationDto{Path: "/usr/local/bin/claude", Version: "2.1.332"},
		},
		{
			name:         "exactly the oldest supported release is accepted",
			observedText: "2.1.0 (Claude Code)",
			expectedDto:  InstallationDto{Path: "/usr/local/bin/claude", Version: "2.1.0"},
		},
		{
			name:          "a release older than supported is refused",
			observedText:  "2.0.9 (Claude Code)",
			expectedError: ErrUnsupportedVersion,
			expectedText:  []string{"too old", "2.0.9", "2.1.0"},
		},
		{
			name:          "missing claude code is refused with install instructions",
			proxyError:    ErrNotInstalled,
			expectedError: ErrNotInstalled,
			expectedText:  []string{"not installed", "npm install -g @anthropic-ai/claude-code"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed := InstallationVo{}
			if tt.proxyError == nil {
				version, err := NewVersionVo(tt.observedText)
				require.NoError(t, err)
				observed = InstallationVo{Path: "/usr/local/bin/claude", Version: version}
			}
			proxy := new(mockInstallationProxy)
			proxy.On("Inspect", mock.Anything).Return(observed, tt.proxyError)
			useCase := NewUseCase(proxy)

			dto, err := useCase.Verify(context.Background())

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				for _, text := range tt.expectedText {
					assert.Contains(t, err.Error(), text)
				}
				assert.Equal(t, InstallationDto{}, dto)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDto, dto)
			}
			proxy.AssertExpectations(t)
		})
	}
}

func TestUseCase_Verify_TooOldCarriesBothReleases(t *testing.T) {
	version, err := NewVersionVo("2.0.9")
	require.NoError(t, err)
	proxy := new(mockInstallationProxy)
	proxy.On("Inspect", mock.Anything).Return(InstallationVo{Path: "/bin/claude", Version: version}, nil)

	_, err = NewUseCase(proxy).Verify(context.Background())

	var unsupported *UnsupportedVersionError
	require.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "2.0.9", unsupported.Found)
	assert.Equal(t, "2.1.0", unsupported.Required)
	assert.Equal(t, MinimumSupportedVersion, unsupported.Required)
}
