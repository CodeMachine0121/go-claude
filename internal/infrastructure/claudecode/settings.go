package claudecode

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/CodeMachine0121/go-claude/installation"
)

const (
	executableName  = "claude"
	profileVariable = "CLAUDE_CONFIG_DIR"
)

// Settings is everything the composition root decides about running Claude Code.
type Settings struct {
	// CLIPath is a chosen Claude Code executable; empty means search the path.
	CLIPath string
	// ProfileDirectory isolates sign-in and history; empty means the everyday profile.
	ProfileDirectory string
	// Credential is a supplied secret; zero means use whatever Claude Code is signed in with.
	Credential Credential
	// Stdin, Stdout and Stderr are attached to the interactive sign-in.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// executablePath finds Claude Code. A chosen executable is never replaced by
// another one; a missing one fails with installation.ErrNotInstalled.
func (s Settings) executablePath() (string, error) {
	if s.CLIPath == "" {
		found, err := exec.LookPath(executableName)
		if err != nil {
			return "", fmt.Errorf("%w: %q was not found on the search path", installation.ErrNotInstalled, executableName)
		}
		return found, nil
	}
	if _, err := os.Stat(s.CLIPath); err != nil {
		return "", fmt.Errorf("%w: nothing at %q", installation.ErrNotInstalled, s.CLIPath)
	}
	return s.CLIPath, nil
}

// environment applies the profile and the credential on top of the inherited
// environment. A supplied credential replaces both kinds of inherited
// credential so Claude Code cannot pick the other one.
func (s Settings) environment(inherited []string) []string {
	replaced := map[string]bool{}
	if s.ProfileDirectory != "" {
		replaced[profileVariable] = true
	}
	if !s.Credential.IsZero() {
		replaced[oauthTokenVariable] = true
		replaced[apiKeyVariable] = true
	}
	environment := make([]string, 0, len(inherited)+2)
	for _, entry := range inherited {
		name, _, _ := strings.Cut(entry, "=")
		if !replaced[name] {
			environment = append(environment, entry)
		}
	}
	if s.ProfileDirectory != "" {
		environment = append(environment, profileVariable+"="+s.ProfileDirectory)
	}
	if !s.Credential.IsZero() {
		environment = append(environment, s.Credential.variable+"="+s.Credential.secret)
	}
	return environment
}
