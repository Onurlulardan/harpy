package config

import "time"

type HTTP struct {
	readHeaderTimeout time.Duration
	idleTimeout       time.Duration
	unaryTimeout      time.Duration
	maxHeaderBytes    int
	maxJSONBodyBytes  int64
}

func loadHTTP(values environmentValues) (HTTP, error) {
	readHeader, err := parseDuration(values, "HARPY_HTTP_READ_HEADER_TIMEOUT", "5s", time.Second, 30*time.Second)
	if err != nil {
		return HTTP{}, err
	}
	idle, err := parseDuration(values, "HARPY_HTTP_IDLE_TIMEOUT", "60s", 5*time.Second, 5*time.Minute)
	if err != nil {
		return HTTP{}, err
	}
	unary, err := parseDuration(values, "HARPY_HTTP_UNARY_TIMEOUT", "30s", time.Second, 2*time.Minute)
	if err != nil {
		return HTTP{}, err
	}
	maxHeader, err := parseInteger(values, "HARPY_HTTP_MAX_HEADER_BYTES", "65536", 8192, 1048576)
	if err != nil {
		return HTTP{}, err
	}
	maxBody, err := parseInteger(values, "HARPY_HTTP_MAX_JSON_BODY_BYTES", "1048576", 4096, 16777216)
	if err != nil {
		return HTTP{}, err
	}
	return HTTP{
		readHeaderTimeout: readHeader, idleTimeout: idle, unaryTimeout: unary,
		maxHeaderBytes: int(maxHeader), maxJSONBodyBytes: maxBody,
	}, nil
}

func (h HTTP) ReadHeaderTimeout() time.Duration { return h.readHeaderTimeout }
func (h HTTP) IdleTimeout() time.Duration       { return h.idleTimeout }
func (h HTTP) UnaryTimeout() time.Duration      { return h.unaryTimeout }
func (h HTTP) MaxHeaderBytes() int              { return h.maxHeaderBytes }
func (h HTTP) MaxJSONBodyBytes() int64          { return h.maxJSONBodyBytes }
