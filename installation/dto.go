package installation

// InstallationDto describes the Claude Code CLI the library uses.
type InstallationDto struct {
	// Path is the location of the Claude Code executable.
	Path string
	// Version is the release in use, such as "2.1.332".
	Version string
}
