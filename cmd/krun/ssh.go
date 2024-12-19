package main

import (
	"io"
	"os/exec"
)

type Runner interface {
	Command(addr string, command []string) *exec.Cmd
}

type SSHRunner struct {
}

func (s *SSHRunner) Command(addr string, command []string) *exec.Cmd {
	cmd := exec.Command("ssh", append([]string{"-o", "StrictHostKeyChecking=no", addr}, command...)...)
	return cmd
}

var sshRunner Runner = &SSHRunner{}

func waitGetOutput(cmd *exec.Cmd) (out string, err error) {
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}
	defer cmd.Wait()

	o, err := io.ReadAll(outPipe)
	if err != nil {
		return "", err
	}

	return string(o), nil
}
