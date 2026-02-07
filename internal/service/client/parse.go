package client

import (
	"flag"
	"fmt"
	"io"

	"yandex-gophkeeper/internal/domain"
)

func mustUP(stderr io.Writer, name string, args []string) (string, string, int) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var u, p string
	fs.StringVar(&u, "u", "", "username")
	fs.StringVar(&p, "p", "", "password")

	if err := fs.Parse(args); err != nil {
		return "", "", 2
	}
	if u == "" || p == "" {
		fmt.Fprintln(stderr, "username/password required")
		return "", "", 2
	}
	return u, p, 0
}

func mustTCD(stderr io.Writer, name string, args []string) (domain.SecretType, string, string, int) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var typ, comment, data string
	fs.StringVar(&typ, "type", "", "password|text|bank_card|binary")
	fs.StringVar(&comment, "comment", "", "comment")
	fs.StringVar(&data, "data", "", "data")

	if err := fs.Parse(args); err != nil {
		return "", "", "", 2
	}

	t := domain.SecretType(typ)
	if err := domain.ValidateSecretType(t); err != nil {
		fmt.Fprintln(stderr, err)
		return "", "", "", 2
	}

	return t, comment, data, 0
}
