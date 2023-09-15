package cli

import "golang.org/x/sys/windows"

func dlopen(name string) (uintptr, error) {
	handle, err := windows.LoadLibrary(name)
	return uintptr(handle), err
}
