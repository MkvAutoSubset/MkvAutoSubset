//go:build windows

package mkvlib

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

func newProcess(stdin io.Reader, stdout, stderr io.Writer, dir, prog string, args ...string) (p *os.Process, err error) {
	cmd := exec.Command(prog, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if dir != "" {
		cmd.Dir = dir
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	err = cmd.Start()
	if err == nil {
		p = cmd.Process
	}
	return
}
