//go:build !windows

package main

import "os"

func replaceAtomic(source, target string) error {
	return os.Rename(source, target)
}
