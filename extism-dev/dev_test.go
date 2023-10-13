package main

import (
	// "os"
	// "os/exec"
	// "runtime"
	"testing"
)

func TestInit(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"init", "--root", "test-data", "--local"})
	err := cmd.Execute()
	if err != nil {
		t.Error(err)
	}

	cmd = rootCmd()
	cmd.SetArgs([]string{"list", "--root", "test-data"})
	err = cmd.Execute()
	if err != nil {
		t.Error(err)
	}
}
