package runcmd

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type ProcessOpts struct {
	Cmd     []string
	EnvVars []string
	Stdout  io.Writer
	Stderr  io.Writer
	WithPty bool
	Stdin   io.Reader
	WorkDir string
}

func RunProcess(opts ProcessOpts) error {
	cmd := exec.Command(opts.Cmd[0], opts.Cmd[1:]...)
	var err error
	var fpty *os.File
	var ftty *os.File
	if !opts.WithPty {
		cmd.Stdout = opts.Stdout
		cmd.Stderr = opts.Stderr
	} else {
		fpty, ftty, err = pty.Open()
		if err != nil {
			return err
		}

		defer func() {
			fpty.Close()
			ftty.Close()
		}()
		cmd.Stdout = ftty
		cmd.Stderr = ftty
	}
	cmd.Stdin = opts.Stdin
	cmd.Env = opts.EnvVars
	cmd.Dir = opts.WorkDir
	wg := &sync.WaitGroup{}
	if fpty != nil {
		wg.Add(1)
		go func() {
			io.Copy(opts.Stdout, fpty)
			wg.Done()
		}()
	}
	err = cmd.Start()
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
		return nil
	}
	err = cmd.Wait()
	if fpty != nil {
		ftty.Close()
	}
	wg.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
		return nil
	}
	return err
}
