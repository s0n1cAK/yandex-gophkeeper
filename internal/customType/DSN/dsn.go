package customtype

import (
	"fmt"
	"net/url"
)

type DSN struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (e *DSN) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", e.User, e.Password, e.Host, e.Port, e.Name, e.SSLMode)
}

func (e *DSN) Type() string {
	return "DSN"
}

func formatDSN(value string) (DSN, error) {
	var dsn DSN
	uri, err := url.Parse(value)
	if err != nil {
		return DSN{}, fmt.Errorf("invalid DSN: %w", err)
	}

	dsn.User = uri.User.Username()
	dsn.Password, _ = uri.User.Password()

	dsn.Host = uri.Hostname()
	dsn.Port = uri.Port()

	dbName := uri.Path
	if len(dbName) > 0 && dbName[0] == '/' {
		dsn.Name = dbName[1:]
	}

	dsn.SSLMode = uri.Query().Get("sslmode")

	return dsn, nil
}

func (e *DSN) GetHost() string {
	return e.Host
}
func (e *DSN) GetUser() string {
	return e.User
}
func (e *DSN) GetPassword() string {
	return e.Password
}
func (e *DSN) GetDB() string {
	return e.Name
}
func (e *DSN) GetSSLMode() string {
	return e.SSLMode
}

func (e *DSN) Set(value string) error {
	gValue, err := formatDSN(string(value[:]))
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}

func (e *DSN) UnmarshalText(text []byte) error {
	gValue, err := formatDSN(string(text[:]))
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}
