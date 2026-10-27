package main

import "strings"

type targetFlag []string

func (values *targetFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func (values *targetFlag) String() string {
	return strings.Join(*values, ",")
}
