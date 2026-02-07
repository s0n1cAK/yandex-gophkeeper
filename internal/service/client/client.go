package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cfg "yandex-gophkeeper/internal/config/client"
	"yandex-gophkeeper/internal/domain"
)

type Client struct {
	baseURL  string
	hc       *http.Client
	token    string
	retryMax int
	waitMin  time.Duration
	waitMax  time.Duration
}

func New(c cfg.Config, token string) *Client {
	tr := &http.Transport{
		TLSClientConfig: nil,
	}

	if c.InsecureSkipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		baseURL: strings.TrimRight(c.Address, "/"),
		hc: &http.Client{
			Timeout:   c.Timeout.Duration(),
			Transport: tr,
		},
		token:    strings.TrimSpace(token),
		retryMax: c.RetryMax,
		waitMin:  c.RetryWaitMin.Duration(),
		waitMax:  c.RetryWaitMax.Duration(),
	}
}

func (c *Client) SetToken(t string) { c.token = strings.TrimSpace(t) }

type errResp struct {
	Error string `json:"error"`
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any, auth bool) error {
	var payload []byte
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = b
	}

	try := func() (*http.Response, error) {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if auth && c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		return c.hc.Do(req)
	}

	var lastErr error
	attempts := c.retryMax + 1
	if attempts < 1 {
		attempts = 1
	}

	for i := 0; i < attempts; i++ {
		resp, err := try()
		if err != nil {
			lastErr = err
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return ctx.Err()
			}
			if !c.sleep(ctx, i) {
				return ctx.Err()
			}
			continue

		}

		b, readErr := readBody(resp)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %s", resp.Status)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !c.sleep(ctx, i) {
				return ctx.Err()
			}
			continue
		}

		if resp.StatusCode >= 300 {
			msg := errorMessageFromBody(b)
			return fmt.Errorf("http %d: %s", resp.StatusCode, msg)
		}

		if out != nil {
			if err := json.Unmarshal(b, out); err != nil {
				return err
			}
		}
		return nil
	}

	return lastErr
}

func (c *Client) sleep(ctx context.Context, i int) bool {
	if i == c.retryMax {
		return false
	}

	d := c.waitMin
	if d <= 0 {
		d = 200 * time.Millisecond
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *Client) Ping(ctx context.Context) error {
	var out map[string]string
	return c.doJSON(ctx, http.MethodGet, "/ping", nil, &out, false)
}

func (c *Client) Register(ctx context.Context, u, p string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/api/user/register",
		map[string]string{"username": u, "password": p},
		&out, false,
	)
	return out.Token, err
}

func (c *Client) Login(ctx context.Context, u, p string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/api/user/login",
		map[string]string{"username": u, "password": p},
		&out, false,
	)
	return out.Token, err
}

func (c *Client) ListSecrets(ctx context.Context) ([]domain.SecretMeta, error) {
	var out struct {
		Secrets []domain.SecretMeta `json:"secrets"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/api/user/secrets/", nil, &out, true)
	return out.Secrets, err
}

func (c *Client) GetSecret(ctx context.Context, id int64) (domain.Secret, string, error) {
	var out struct {
		Secret domain.Secret `json:"secret"`
		Data   string        `json:"data"`
	}
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/user/secrets/%d/", id), nil, &out, true)
	return out.Secret, out.Data, err
}

func (c *Client) CreateSecret(ctx context.Context, t domain.SecretType, comment, data string) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/api/user/secrets/",
		map[string]any{"type": t, "comment": comment, "data": data},
		&out, true,
	)
	return out.ID, err
}

func (c *Client) UpdateSecret(ctx context.Context, id int64, t domain.SecretType, comment, data string) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/user/secrets/%d/", id),
		map[string]any{"type": t, "comment": comment, "data": data},
		nil, true,
	)
}

func (c *Client) DeleteSecret(ctx context.Context, id int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/user/secrets/%d/", id), nil, nil, true)
}

func (c *Client) Close() {
	if c == nil || c.hc == nil || c.hc.Transport == nil {
		return
	}
	if tr, ok := c.hc.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}

func readBody(resp *http.Response) ([]byte, error) {
	br := bufio.NewReader(resp.Body)

	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		return io.ReadAll(zr)
	}

	if peek, _ := br.Peek(2); len(peek) == 2 && peek[0] == 0x1f && peek[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		return io.ReadAll(zr)
	}

	return io.ReadAll(br)
}

func errorMessageFromBody(b []byte) string {
	var e errResp
	if json.Unmarshal(b, &e) == nil && strings.TrimSpace(e.Error) != "" {
		return strings.TrimSpace(e.Error)
	}
	return strings.TrimSpace(string(b))
}
