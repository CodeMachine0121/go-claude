//go:build !unix

package claudecode

import "os/exec"

// stopTogether keeps the default behaviour on platforms without process groups:
// giving up kills Claude Code itself.
func (run invocation) stopTogether(command *exec.Cmd) {}
