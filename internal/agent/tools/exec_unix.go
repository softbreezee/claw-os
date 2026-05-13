//go:build !windows

package tools

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts the child into its own process group so a
// subsequent SIGKILL with a negative pid (-pgid) reaps the entire
// subtree. We also wire cmd.Cancel to perform that group-kill when
// the command's context fires; the default behaviour only kills the
// immediate child, which would orphan grandchildren spawned by
// `nohup … &` style commands and leave the inherited stdout / stderr
// pipes open (causing CombinedOutput to hang).
//
// This file is unix-only; the windows counterpart in exec_windows.go
// is a no-op because Windows lacks the unix process-group semantics
// (job objects exist but the additional plumbing isn't worth it for
// the rare windows daemon developer right now — the WaitDelay alone
// still prevents the indefinite hang there).
func setProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true

	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// Negative pid → entire process group. We deliberately ignore
		// the error: the most common failure is ESRCH (group already
		// gone), which is exactly the state we wanted anyway. Falling
		// back to killing just cmd.Process if the group-kill fails is
		// not useful here — exec.CommandContext's default already does
		// that and we explicitly opted out of it.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return nil
	}
}
