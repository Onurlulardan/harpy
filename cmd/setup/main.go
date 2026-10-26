package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := findModuleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup: %s\n", err)
		os.Exit(1)
	}
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, root); err != nil {
		if errors.Is(err, errHelp) {
			os.Exit(0)
		}
		if !errors.Is(err, errPrinted) {
			fmt.Fprintf(os.Stderr, "setup: %s\n", err)
		}
		os.Exit(1)
	}
}

type options struct {
	defaults bool
	rotate   bool
	profile  string
}

func run(args []string, in io.Reader, stdout, stderr io.Writer, root string) error {
	opts, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}
	var answers answers
	var prev map[string]string
	if opts.defaults {
		answers = defaultAnswers()
		if opts.profile != "" {
			answers.Profile = opts.profile
		}
		if err := validProfile(answers.Profile); err != nil {
			return err
		}
		prev, err = loadPrev(root, answers.Profile)
		if err != nil {
			return err
		}
	} else {
		answers, prev, err = prompt(in, stderr, root, opts)
		if err != nil {
			return err
		}
	}
	answers, err = fillSecrets(answers, prev, opts.rotate)
	if err != nil {
		return err
	}
	if err := answers.validate(); err != nil {
		return err
	}
	if err := writeProfile(root, answers); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Wrote deploy/%s/.env\n%s", answers.Profile, nextCommands(answers.Profile))
	return err
}

func parseArgs(args []string, stderr io.Writer) (options, error) {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: go run ./cmd/setup [--defaults] [--rotate-secrets] [--profile single-tenant|multi-tenant]")
	}
	defaults := fs.Bool("defaults", false, "answer with defaults and do not prompt")
	rotate := fs.Bool("rotate-secrets", false, "generate new local secrets")
	profile := fs.String("profile", "", "single-tenant or multi-tenant")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return options{}, errHelp
		}
		return options{}, errPrinted
	}
	if fs.NArg() != 0 {
		return options{}, errors.New("unexpected arguments")
	}
	opts := options{defaults: *defaults, rotate: *rotate, profile: *profile}
	if opts.profile != "" {
		if err := validProfile(opts.profile); err != nil {
			return options{}, err
		}
	}
	return opts, nil
}

var (
	errHelp    = errors.New("help")
	errPrinted = errors.New("printed")
)

func nextCommands(profile string) string {
	return fmt.Sprintf("docker compose -f deploy/%s/compose.yml up -d\n", profile) +
		"go run ./core/cmd/migrator\n" +
		"go run ./core/cmd/api\n" +
		"go run ./core/cmd/worker\n"
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		data, readErr := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readErr == nil && strings.Contains(string(data), "module github.com/Onurlulardan/harpy\n") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run go run ./cmd/setup from the harpy module")
		}
		dir = parent
	}
}
