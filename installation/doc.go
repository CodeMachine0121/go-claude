// Package installation verifies that the Claude Code CLI is installed and recent
// enough for this library.
//
// Most programs never use this package directly: claude.New verifies the
// installation and exposes the result through Client.Installation. Use this
// package to recognise its refusals:
//
//	if errors.Is(err, installation.ErrNotInstalled) { ... }
//	if errors.Is(err, installation.ErrUnsupportedVersion) { ... }
package installation
