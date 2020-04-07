package main

import (
	"os"
	"testing"
)

func TestHooksBadConfig(t *testing.T) {
	gConfig.Hooks = []HookConfig{
		HookConfig{
			Path: "./fixtures/hooks/missing.sh",
			Args: []string{},
		},
	}

	t.Run("Initialize Hooks", func(t *testing.T) {
		if _, err := newHooks(); !os.IsNotExist(err) {
			t.Errorf(err.Error())
		}
	})
}

func TestHooksNoConfig(t *testing.T) {
	gConfig.Hooks = nil

	t.Run("Initialize Hooks", func(t *testing.T) {
		if _, err := newHooks(); err != nil {
			t.Errorf(err.Error())
		}
	})
}

func TestHooks(t *testing.T) {
	gConfig.Hooks = []HookConfig{
		HookConfig{
			Path: "./fixtures/hooks/test_success.sh",
			Args: []string{"--user", "{{.UserName}}", "--id", "{{.UserID}}"},
		},
		HookConfig{
			Path: "./fixtures/hooks/test_failed.sh",
			Args: []string{"--user", "{{.UserName}}", "--id", "{{.UserID}}"},
		},
	}

	t.Run("Initialize Hooks", func(t *testing.T) {
		if _, err := newHooks(); err != nil {
			t.Errorf(err.Error())
		}
	})

	t.Run("successful deleteUser", func(t *testing.T) {
		mocks, err := newHooks()
		if err != nil {
			t.Errorf(err.Error())
		}
		if err := mocks[0].deleteUser("2960B56A-90F1-43EC-B628-89932314F0F2", "demo"); err != nil {
			t.Errorf(err.Error())
		}
	})

	t.Run("Failed deleteUser", func(t *testing.T) {
		mocks, err := newHooks()
		if err != nil {
			t.Errorf(err.Error())
		}
		if err := mocks[1].deleteUser("aeazrar", "demo"); err == nil {
			t.Errorf(err.Error())
		}
	})

}
