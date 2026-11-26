package installation

// MinimumSupportedVersion is the oldest Claude Code release this library accepts.
const MinimumSupportedVersion = "2.1.0"

// InstallationDomain decides whether a Claude Code installation can be used.
type InstallationDomain struct {
	path    string
	version VersionVo
	minimum VersionVo
}

// NewInstallationDomain builds the domain model for an installation found at path.
func NewInstallationDomain(path string, version VersionVo) InstallationDomain {
	return InstallationDomain{
		path:    path,
		version: version,
		minimum: VersionVo{major: 2, minor: 1, patch: 0},
	}
}

// Verify refuses releases older than MinimumSupportedVersion.
func (d InstallationDomain) Verify() error {
	if !d.version.AtLeast(d.minimum) {
		return &UnsupportedVersionError{Found: d.version.String(), Required: d.minimum.String()}
	}
	return nil
}

// ToDto converts the installation into its public shape.
func (d InstallationDomain) ToDto() InstallationDto {
	return InstallationDto{Path: d.path, Version: d.version.String()}
}
