package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/orange-cloudfoundry/uaa-user-cleaner/runcmd"
)

// Hook -
type Hook struct {
	path string
	args []*template.Template
}

func newHooks() ([]Hook, error) {
	hooks := make([]Hook, 0)
	for _, hookConfig := range gConfig.Hooks {
		if _, err := os.Stat(hookConfig.Path); os.IsNotExist(err) {
			return nil, err
		}
		hook := Hook{path: hookConfig.Path, args: make([]*template.Template, 0)}
		for _, arg := range hookConfig.Args {
			argTmpl, err := template.New("hookTemplate").Parse(arg)
			if err != nil {
				return nil, err
			}
			hook.args = append(hook.args, argTmpl)
		}
		hooks = append(hooks, hook)
	}
	return hooks, nil
}

func (h *Hook) deleteUser(userID, userName string) error {
	parsedArgs := make([]string, 0)
	var tplResult bytes.Buffer
	for _, argsTmpl := range h.args {
		err := argsTmpl.Execute(&tplResult, struct {
			UserID   string
			UserName string
		}{
			userID,
			userName,
		})
		if err != nil {
			return err
		}
		parsedArgs = append(parsedArgs, tplResult.String())
		tplResult.Reset()
	}
	return h.executeHook(parsedArgs)
}

func (h *Hook) executeHook(args []string) error {
	script, err := ioutil.ReadFile(h.path)
	if err == nil {
		return runcmd.RunScript(runcmd.ScriptOpts{
			Script:  string(script),
			Args:    args,
			EnvVars: os.Environ(),
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Stdin:   os.Stdin,
			WorkDir: "",
		})
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}
