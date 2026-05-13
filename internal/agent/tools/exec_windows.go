//go:build windows

package tools

import "os/exec"

// setProcessGroup is a no-op on Windows — the unix process-group
// semantics it relies on don't translate cleanly. We still benefit
// from cmd.WaitDelay (set in exec.go) for the canonical "exec hangs
// on backgrounded subprocess" scenario, which is the bulk of the
// real-world pain. Properly reaping grandchildren on Windows would
// require Job Objects + additional Win32 plumbing; not worth the
// complexity until a Windows user actually hits it.
func setProcessGroup(cmd *exec.Cmd) {}
