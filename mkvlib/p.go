//go:build !windows

package mkvlib

import (
	"io"
	"os"
	"os/exec"
)

func newProcess(stdin io.Reader, stdout, stderr io.Writer, dir, prog string, args ...string) (p *os.Process, err error) {
	cmd := exec.Command(prog, args...)

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
