package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"yandex-gophkeeper/internal/domain"
)

type CLI struct {
	Client    *Client
	TokenPath string
	Stdout    io.Writer
	Stderr    io.Writer
}

func NewCLI(с CLI) *CLI {
	stdout := с.Stdout
	stderr := с.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &CLI{
		Client:    с.Client,
		TokenPath: с.TokenPath,
		Stdout:    stdout,
		Stderr:    stderr,
	}
}

func (c *CLI) Run(ctx context.Context, args []string) int {
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}

	if len(args) == 0 {
		fmt.Fprintln(c.Stderr, "need command: register|login|ping|list|get|create|update|delete")
		return 2
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "ping":
		return fail(c.Stderr, c.Client.Ping(ctx))

	case "register":
		u, p, code := mustUP(c.Stderr, cmd, rest)
		if code != 0 {
			return code
		}

		tok, err := c.Client.Register(ctx, u, p)
		if code := fail(c.Stderr, err); code != 0 {
			return code
		}

		c.Client.SetToken(tok)

		if code := fail(c.Stderr, Write(c.TokenPath, tok)); code != 0 {
			return code
		}

		fmt.Fprintln(c.Stdout, "ok")
		return 0

	case "login":
		u, p, code := mustUP(c.Stderr, cmd, rest)
		if code != 0 {
			return code
		}

		tok, err := c.Client.Login(ctx, u, p)
		if code := fail(c.Stderr, err); code != 0 {
			return code
		}

		c.Client.SetToken(tok)

		if code := fail(c.Stderr, Write(c.TokenPath, tok)); code != 0 {
			return code
		}

		fmt.Fprintln(c.Stdout, "ok")
		return 0

	case "list":
		items, err := c.Client.ListSecrets(ctx)
		if code := fail(c.Stderr, err); code != 0 {
			return code
		}

		const layout = "2006-01-02 15:04:05"
		for _, it := range items {
			fmt.Fprintf(c.Stdout, "%d\t%s\t%s\t%s\n",
				it.ID, it.Type, it.Comment, it.UpdatedAt.Format(layout),
			)
		}
		return 0

	case "get":
		fs := flag.NewFlagSet("get", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)

		var id int64
		fs.Int64Var(&id, "id", 0, "secret id")

		if err := fs.Parse(rest); err != nil {
			return 2
		}
		if id <= 0 {
			fmt.Fprintln(c.Stderr, "bad id")
			return 2
		}

		sec, data, err := c.Client.GetSecret(ctx, id)
		if code := fail(c.Stderr, err); code != 0 {
			return code
		}

		fmt.Fprintf(c.Stdout, "id=%d type=%s comment=%q updated=%s\n",
			sec.ID, sec.Type, sec.Comment, sec.UpdatedAt,
		)
		fmt.Fprintln(c.Stdout, "data:", data)
		return 0

	case "create":
		t, comment, data, code := mustTCD(c.Stderr, cmd, rest)
		if code != 0 {
			return code
		}

		id, err := c.Client.CreateSecret(ctx, t, comment, data)
		if code := fail(c.Stderr, err); code != 0 {
			return code
		}

		fmt.Fprintln(c.Stdout, "id:", id)
		return 0

	case "update":
		fs := flag.NewFlagSet("update", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)

		var (
			id      int64
			typ     string
			comment string
			data    string
		)

		fs.Int64Var(&id, "id", 0, "secret id")
		fs.StringVar(&typ, "type", "", "password|text|bank_card|binary")
		fs.StringVar(&comment, "comment", "", "comment")
		fs.StringVar(&data, "data", "", "data")

		if err := fs.Parse(rest); err != nil {
			return 2
		}
		if id <= 0 {
			fmt.Fprintln(c.Stderr, "bad id")
			return 2
		}

		t := domain.SecretType(typ)
		if err := domain.ValidateSecretType(t); err != nil {
			fmt.Fprintln(c.Stderr, err)
			return 2
		}

		if code := fail(c.Stderr, c.Client.UpdateSecret(ctx, id, t, comment, data)); code != 0 {
			return code
		}

		fmt.Fprintln(c.Stdout, "ok")
		return 0

	case "delete":
		fs := flag.NewFlagSet("delete", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)

		var id int64
		fs.Int64Var(&id, "id", 0, "secret id")

		if err := fs.Parse(rest); err != nil {
			return 2
		}
		if id <= 0 {
			fmt.Fprintln(c.Stderr, "bad id")
			return 2
		}

		if code := fail(c.Stderr, c.Client.DeleteSecret(ctx, id)); code != 0 {
			return code
		}

		fmt.Fprintln(c.Stdout, "ok")
		return 0

	default:
		fmt.Fprintln(c.Stderr, "unknown command:", cmd)
		return 2
	}
}

func fail(stderr io.Writer, err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		fmt.Fprintln(stderr, "shutting down...")
		return 130
	}
	fmt.Fprintln(stderr, "error:", err)
	return 1
}
