package main

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	localManagementListenAddress     = "127.0.0.1:9090"
	containerManagementListenAddress = "0.0.0.0:9000"
)

// requestedManagementListen is the --management-listen value for one setup run.
var requestedManagementListen string

func resolveManagementListenAddress(deployment, override, preserved string) (string, error) {
	candidate := override
	if candidate == "" {
		candidate = preserved
	}
	if candidate == "" {
		switch deployment {
		case "local":
			return localManagementListenAddress, nil
		case "container":
			return "", errors.New("container management listen address must be set explicitly")
		default:
			return "", errors.New("management deployment is unknown")
		}
	}
	if err := validListenAddress(candidate); err != nil {
		return "", err
	}
	return candidate, nil
}

func validListenAddress(raw string) error {
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n") {
		return errors.New("management listen address must be host:port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 || strconv.Itoa(number) != port {
		return errors.New("management listen port must be between 1 and 65535")
	}
	return nil
}
