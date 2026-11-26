//go:build unix

package claudecode

import (
	"os/exec"
	"syscall"
)

// stopTogether makes giving up stop the commands Claude started, too.
//
// Claude Code runs each tool command in a session of its own, which no signal to
// Claude Code's process group reaches. Only Claude Code knows those commands, so
// giving up first asks it to terminate (it then stops its commands), and the
// command's WaitDelay kills the whole group if it has not exited in time.
//
// Interactive runs stay in the terminal's group: a background group reading the
// terminal would be stopped by the operating system.
func (run invocation) stopTogether(command *exec.Cmd) {
	if run.terminal {
		return
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGTERM)
	}
}
