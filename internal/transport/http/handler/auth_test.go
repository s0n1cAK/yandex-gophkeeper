package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type fakeAuthSvc struct {
	regUser string
	regPass string
	logUser string
	logPass string
	regTok  string
	logTok  string
	regErr  error
	logErr  error
}

func (s *fakeAuthSvc) Register(ctx context.Context, u, p string) (string, error) {
	s.regUser, s.regPass = u, p
	return s.regTok, s.regErr
}
func (s *fakeAuthSvc) Login(ctx context.Context, u, p string) (string, error) {
	s.logUser, s.logPass = u, p
	return s.logTok, s.logErr
}

func TestAuth_Register_TrimsAndReturnsToken(t *testing.T) {
	svc := &fakeAuthSvc{regTok: "T"}
	h := NewAuth(zap.NewNop(), svc)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"username":" alice ","password":" pass "}`))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if svc.regUser != "alice" || svc.regPass != "pass" {
		t.Fatalf("trim mismatch: u=%q p=%q", svc.regUser, svc.regPass)
	}
	var out map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["token"] != "T" {
		t.Fatalf("out=%v", out)
	}
}

func TestAuth_Login_ErrorGives401(t *testing.T) {
	svc := &fakeAuthSvc{logErr: context.Canceled}
	h := NewAuth(zap.NewNop(), svc)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"username":"a","password":"b"}`))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
