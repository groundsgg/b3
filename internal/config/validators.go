package config

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateListenAddr(fl validator.FieldLevel) bool {
	addr := fl.Field().String()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return false
	}

	if host == "" {
		return true
	}

	if net.ParseIP(host) != nil {
		return true
	}

	return isValidHostname(host)
}

var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+$`)

func isValidHostname(h string) bool {
	if !hostnameRegex.MatchString(h) {
		return false
	}

	labels := strings.Split(h, ".")
	for _, label := range labels {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
	}

	return true
}
