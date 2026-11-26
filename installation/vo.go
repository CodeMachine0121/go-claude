package installation

import (
	"fmt"
	"regexp"
	"strconv"
)

var versionPattern = regexp.MustCompile(`^\s*v?(\d+)\.(\d+)\.(\d+)`)

// VersionVo is a Claude Code release number (major.minor.patch).
type VersionVo struct {
	major int
	minor int
	patch int
}

// NewVersionVo reads a release number from the text printed by the Claude Code
// CLI, such as "2.1.332 (Claude Code)".
func NewVersionVo(text string) (VersionVo, error) {
	match := versionPattern.FindStringSubmatch(text)
	if match == nil {
		return VersionVo{}, fmt.Errorf("%w: %q", ErrUnrecognizedVersion, text)
	}
	// The pattern only captures digit runs, so these conversions cannot fail
	// unless a part overflows an int, which is not a real release number.
	major, majorErr := strconv.Atoi(match[1])
	minor, minorErr := strconv.Atoi(match[2])
	patch, patchErr := strconv.Atoi(match[3])
	if majorErr != nil || minorErr != nil || patchErr != nil {
		return VersionVo{}, fmt.Errorf("%w: %q", ErrUnrecognizedVersion, text)
	}
	return VersionVo{major: major, minor: minor, patch: patch}, nil
}

// AtLeast reports whether this release is the same as or newer than other.
func (v VersionVo) AtLeast(other VersionVo) bool {
	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	return v.patch >= other.patch
}

// String returns the release as "major.minor.patch".
func (v VersionVo) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// InstallationVo is what the infrastructure observed about the Claude Code CLI.
type InstallationVo struct {
	Path    string
	Version VersionVo
}

// ToDomain converts the observation into the domain model that enforces the
// supported release.
func (vo InstallationVo) ToDomain() InstallationDomain {
	return NewInstallationDomain(vo.Path, vo.Version)
}
