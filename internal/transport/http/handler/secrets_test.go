package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"yandex-gophkeeper/internal/domain"
	secretsvc "yandex-gophkeeper/internal/service/secrets"
	"yandex-gophkeeper/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type fakeSecretsService struct {
	createIn  secretsvc.SecretUpsert
	createUID domain.UserID
	createID  domain.SecretID
	createErr error

	listUID domain.UserID
	listOut []domain.SecretMeta
	listErr error

	getUID   domain.UserID
	getID    domain.SecretID
	getSec   domain.Secret
	getPlain []byte
	getErr   error

	updateUID domain.UserID
	updateID  domain.SecretID
	updateIn  secretsvc.SecretUpsert
	updateErr error

	deleteUID domain.UserID
	deleteID  domain.SecretID
	deleteErr error
}

func (f *fakeSecretsService) Create(ctx context.Context, ownerID domain.UserID, in secretsvc.SecretUpsert) (domain.SecretID, error) {
	f.createUID = ownerID
	f.createIn = in
	return f.createID, f.createErr
}
func (f *fakeSecretsService) List(ctx context.Context, ownerID domain.UserID) ([]domain.SecretMeta, error) {
	f.listUID = ownerID
	return f.listOut, f.listErr
}
func (f *fakeSecretsService) Get(ctx context.Context, ownerID domain.UserID, id domain.SecretID) (domain.Secret, []byte, error) {
	f.getUID = ownerID
	f.getID = id
	return f.getSec, f.getPlain, f.getErr
}
func (f *fakeSecretsService) Update(ctx context.Context, ownerID domain.UserID, id domain.SecretID, in secretsvc.SecretUpsert) error {
	f.updateUID = ownerID
	f.updateID = id
	f.updateIn = in
	return f.updateErr
}
func (f *fakeSecretsService) Delete(ctx context.Context, ownerID domain.UserID, id domain.SecretID) error {
	f.deleteUID = ownerID
	f.deleteID = id
	return f.deleteErr
}

func newTestLogger() *zap.Logger {
	core, _ := observer.New(zap.InfoLevel)
	return zap.New(core)
}

func withUser(req *http.Request, uid domain.UserID) *http.Request {
	ctx := middleware.WithUserID(req.Context(), uid)
	return req.WithContext(ctx)
}

func TestSecretsHandler_List_OK(t *testing.T) {
	svc := &fakeSecretsService{
		listOut: []domain.SecretMeta{
			{ID: 1, Type: domain.SecretText, Comment: "a", UpdatedAt: time.Unix(1, 0)},
			{ID: 2, Type: domain.SecretBinary, Comment: "b", UpdatedAt: time.Unix(2, 0)},
		},
	}
	h := NewSecrets(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/secrets/", nil)
	req = withUser(req, 7)

	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Secrets []domain.SecretMeta `json:"secrets"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(out.Secrets) != 2 {
		t.Fatalf("len=%d", len(out.Secrets))
	}
}

func TestSecretsHandler_Create_Binary_DecodesBase64(t *testing.T) {
	raw := []byte("bin\x00data")
	b64 := base64.StdEncoding.EncodeToString(raw)

	svc := &fakeSecretsService{createID: 10}
	h := NewSecrets(newTestLogger(), svc)

	body := `{"type":"binary","comment":"c","data":"` + b64 + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/secrets/", strings.NewReader(body))
	req = withUser(req, 7)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}

	if svc.createUID != 7 {
		t.Fatalf("uid=%d", svc.createUID)
	}
	if svc.createIn.Type != domain.SecretBinary {
		t.Fatalf("type=%s", svc.createIn.Type)
	}
	if string(svc.createIn.Data) != string(raw) {
		t.Fatalf("data mismatch: %q", svc.createIn.Data)
	}
}

func TestSecretsHandler_Create_BankCard_NormalizesNumber(t *testing.T) {
	svc := &fakeSecretsService{createID: 1}
	h := NewSecrets(newTestLogger(), svc)

	body := `{"type":"bank_card","comment":"x","data":"{\"number\":\"4242 4242 4242 4242\",\"expiry\":\"12/30\",\"holder\":\"ALICE\",\"cvc\":\"123\"}"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/secrets/", strings.NewReader(body))
	req = withUser(req, 7)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}

	var p domain.BankCardPayload
	if err := json.Unmarshal(svc.createIn.Data, &p); err != nil {
		t.Fatalf("unmarshal stored payload: %v", err)
	}
	if p.Number != "4242424242424242" {
		t.Fatalf("number=%q", p.Number)
	}
}

func TestSecretsHandler_Get_NotFound_404(t *testing.T) {
	svc := &fakeSecretsService{getErr: domain.ErrSecretNotFound}
	h := NewSecrets(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/secrets/1/", nil)
	req = withUser(req, 7)
	req = withChiURLParam(req, "id", "1")

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSecretsHandler_Get_Binary_EncodesBase64(t *testing.T) {
	raw := []byte("bin\x00data")
	svc := &fakeSecretsService{
		getSec:   domain.Secret{ID: 4, Type: domain.SecretBinary, Comment: "blob"},
		getPlain: raw,
	}
	h := NewSecrets(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/secrets/4/", nil)
	req = withUser(req, 7)
	req = withChiURLParam(req, "id", "4")

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Secret domain.Secret `json:"secret"`
		Data   string        `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if out.Data != base64.StdEncoding.EncodeToString(raw) {
		t.Fatalf("data=%q", out.Data)
	}
}

func TestSecretsHandler_Delete_OK(t *testing.T) {
	svc := &fakeSecretsService{}
	h := NewSecrets(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/secrets/2/", nil)
	req = withUser(req, 7)
	req = withChiURLParam(req, "id", "2")

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSecretsHandler_Delete_NotFound_404(t *testing.T) {
	svc := &fakeSecretsService{deleteErr: domain.ErrSecretNotFound}
	h := NewSecrets(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/secrets/2/", nil)
	req = withUser(req, 7)
	req = withChiURLParam(req, "id", "2")

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func withChiURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}
