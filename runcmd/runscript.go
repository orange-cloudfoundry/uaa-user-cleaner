package runcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-shellwords"
)

func retrieveExecCommand(script string) ([]string, error) {
	var err error
	if strings.HasPrefix(script, "#!") {
		execCmd := strings.SplitN(strings.TrimPrefix(script, "#!"), "\n", 2)[0]
		cmds, err := shellwords.Parse(execCmd)
		if err != nil {
			return []string{}, fmt.Errorf("Error when parsing shebang: %s", err.Error())
		}
		cmds[0], err = exec.LookPath(cmds[0])
		return cmds, err
	}
	bashPath, err := exec.LookPath("bash")
	if err == nil {
		return []string{bashPath}, nil
	}
	shellPath, err := exec.LookPath("sh")
	if err == nil {
		return []string{shellPath}, nil
	}
	return []string{}, fmt.Errorf("Can't found bash or shell.")

}

type ScriptOpts struct {
	Script  string
	Args    []string
	EnvVars []string
	Stdout  io.Writer
	Stderr  io.Writer
	WithPty bool
	Stdin   *os.File
	WorkDir string
}

func sanitizeScript(script string) string {
	if strings.HasPrefix(script, "#!") {
		script = strings.SplitN(strings.TrimPrefix(script, "#!"), "\n", 2)[1]
	}
	return script
}

func RunScript(opts ScriptOpts) error {
	if opts.Script == "" {
		return nil
	}
	execCmd, err := retrieveExecCommand(opts.Script)
	if err != nil {
		return err
	}
	stdin := opts.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	execCmd = append(execCmd, stdin.Name())
	execCmd = append(execCmd, opts.Args...)
	buf := bytes.NewBufferString(sanitizeScript(opts.Script))

	return RunProcess(ProcessOpts{
		Cmd:     execCmd,
		EnvVars: opts.EnvVars,
		Stdout:  opts.Stdout,
		Stderr:  opts.Stderr,
		WithPty: opts.WithPty,
		Stdin:   buf,
		WorkDir: opts.WorkDir,
	})
}
