package main

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestLibVersions(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"lib", "versions"})
	err := cmd.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestCall(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"call", "../test/code.wasm", "count_vowels", "-i", "aaa"})
	err := cmd.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestCallBadName(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"call", "../test/code.wasm", "something"})
	err := cmd.Execute()
	if err == nil {
		t.Error(err)
	}
}

func TestInstall(t *testing.T) {
	cmd := rootCmd()
	if err := exec.Command("rm", "-rf", "tmp").Run(); err != nil {
		t.Error(err)
	}
	if err := exec.Command("mkdir", "-p", "tmp/include", "tmp/lib").Run(); err != nil {
		t.Error(err)
	}
	cmd.SetArgs([]string{"lib", "install", "--prefix", "tmp"})
	err := cmd.Execute()
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat("tmp/include/extism.h")
	if err != nil {
		t.Error("Invalid header file", err)
	}

	if runtime.GOOS == "darwin" {
		_, err = os.Stat("tmp/lib/libextism.dylib")
	} else if runtime.GOOS == "windows" {
		_, err = os.Stat("tmp/lib/extism.dll")
	} else {
		_, err = os.Stat("tmp/lib/libextism.so")
	}
	if err != nil {
		t.Error("Invalid lib file", err)
	}

	exec.Command("rm", "-rf", "tmp").Run()
}
