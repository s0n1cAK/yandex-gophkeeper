package customtype

import (
	"encoding"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrBadDuration = errors.New("bad duration")
)

type Duration time.Duration

var _ encoding.TextUnmarshaler = (*Duration)(nil)

func (d Duration) Duration() time.Duration { return time.Duration(d) }

func (d Duration) String() string { return time.Duration(d).String() }

func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

func (d *Duration) Set(value string) error {
	s := strings.TrimSpace(value)
	if s == "" {
		return ErrBadDuration
	}

	if isDigits(s) {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return ErrBadDuration
		}
		*d = Duration(time.Duration(n) * time.Second)
		return nil
	}

	td, err := time.ParseDuration(s)
	if err != nil || td < 0 {
		return ErrBadDuration
	}
	*d = Duration(td)
	return nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
