//go:build darwin || linux || freebsd

package cli

import "github.com/ebitengine/purego"

func dlopen(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_GLOBAL|purego.RTLD_NOW)
}
