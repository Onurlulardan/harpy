package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	errRequired          = errors.New("value is required")
	errDefaultCredential = errors.New("refusing rustfsadmin credential")
)

func prompt(in io.Reader, out io.Writer, root string, opts options) (answers, map[string]string, error) {
	reader := bufio.NewReader(in)
	answers := defaultAnswers()
	var err error
	if opts.profile != "" {
		answers.Profile = opts.profile
	} else {
		answers.Profile, err = ask(reader, out, "Profile (single-tenant or multi-tenant)", answers.Profile, validProfile)
		if err != nil {
			return answers, nil, err
		}
	}
	prev, err := loadPrev(root, answers.Profile)
	if err != nil {
		return answers, nil, err
	}
	if answers.HTTPPort, err = askPort(reader, out, "HTTP port", answers.HTTPPort); err != nil {
		return answers, nil, err
	}
	if answers.PostgresPort, err = askPort(reader, out, "Postgres port", answers.PostgresPort); err != nil {
		return answers, nil, err
	}
	answers.PostgresDB, err = ask(reader, out, "Postgres database", answers.PostgresDB, func(value string) error {
		return validIdent("database", value)
	})
	if err != nil {
		return answers, nil, err
	}
	answers.PostgresUser, err = ask(reader, out, "Postgres user", answers.PostgresUser, func(value string) error {
		return validIdent("user", value)
	})
	if err != nil {
		return answers, nil, err
	}
	storage, err := ask(reader, out, "Object storage (rustfs or external)", "rustfs", func(value string) error {
		return validChoice(value, "rustfs", "external")
	})
	if err != nil {
		return answers, nil, err
	}
	if storage == "rustfs" {
		answers.S3Provider = "rustfs"
		return answers, prev, nil
	}
	answers.S3Endpoint, err = ask(reader, out, "S3 endpoint", "", validEndpoint)
	if err != nil {
		return answers, nil, err
	}
	answers.S3Provider, err = ask(reader, out, "S3 provider (aws or minio)", "", func(value string) error {
		return validChoice(value, "aws", "minio")
	})
	if err != nil {
		return answers, nil, err
	}
	answers.S3Region, err = ask(reader, out, "S3 region", answers.S3Region, validRegion)
	if err != nil {
		return answers, nil, err
	}
	answers.S3Bucket, err = ask(reader, out, "S3 bucket", answers.S3Bucket, validBucket)
	if err != nil {
		return answers, nil, err
	}
	if prev["S3_PROVIDER"] == answers.S3Provider && prev["S3_ACCESS_KEY"] != "" && prev["S3_SECRET_KEY"] != "" {
		return answers, prev, nil
	}
	answers.S3AccessKey, err = ask(reader, out, "S3 access key", "", rejectDefaultCredential)
	if err != nil {
		return answers, nil, err
	}
	answers.S3SecretKey, err = ask(reader, out, "S3 secret key", "", rejectDefaultCredential)
	return answers, prev, err
}

func askPort(in *bufio.Reader, out io.Writer, label string, fallback int) (int, error) {
	value, err := ask(in, out, label, strconv.Itoa(fallback), validPort)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func ask(in *bufio.Reader, out io.Writer, label, fallback string, valid func(string) error) (string, error) {
	for {
		if fallback == "" {
			fmt.Fprintf(out, "%s: ", label)
		} else {
			fmt.Fprintf(out, "%s [%s]: ", label, fallback)
		}
		line, err := in.ReadString('\n')
		if len(line) == 0 && err != nil {
			return "", err
		}
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" {
			line = fallback
		}
		if validErr := valid(line); validErr != nil {
			fmt.Fprintf(out, "%s\n", validErr.Error())
			continue
		}
		return line, nil
	}
}

func rejectDefaultCredential(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}
	if strings.EqualFold(value, "rustfsadmin") {
		return errDefaultCredential
	}
	return nil
}
