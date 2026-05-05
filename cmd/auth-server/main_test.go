package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpx "github.com/loaikanou/GoAuth/internal/httpx"
	"github.com/loaikanou/GoAuth/internal/policy"
	"github.com/loaikanou/GoAuth/internal/store"
	"github.com/loaikanou/GoAuth/internal/token"
)

type apiResp struct {
	Success bool                 `json:"success"`
	Data    json.RawMessage      `json:"data"`
	Error   *httpx.ErrorResponse `json:"error"`
}

func newTestServer(t *testing.T) (*store.InMemoryStore, http.Handler) {
	t.Helper()

	repo := store.NewInMemoryStore()
	ts, err := token.NewService([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	p := &policy.CerbosPolicyEngine{Rules: []policy.CerbosRule{
		{Subject: "system", Action: "authorize", Resource: "tenant/tenant1", Allowed: true},
	}}
	app := &Server{Repo: repo, TokenSvc: ts, Policy: p}
	return repo, newMux(app)
}

func TestTokenAndIntrospect(t *testing.T) {
	t.Parallel()

	_, h := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/token?tenant_id=tenant1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var r apiResp
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !r.Success {
		t.Fatalf("expected success=true, got error=%+v", r.Error)
	}

	var tokenPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(r.Data, &tokenPayload); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if tokenPayload.AccessToken == "" {
		t.Fatalf("expected access_token to be set")
	}

	introspectBody, _ := json.Marshal(map[string]string{"token": tokenPayload.AccessToken})
	ireq := httptest.NewRequest(http.MethodPost, "/token/introspect", bytes.NewReader(introspectBody))
	ireq.Header.Set("Content-Type", "application/json")
	irec := httptest.NewRecorder()
	h.ServeHTTP(irec, ireq)

	if irec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", irec.Code, irec.Body.String())
	}

	var ir apiResp
	if err := json.Unmarshal(irec.Body.Bytes(), &ir); err != nil {
		t.Fatalf("decode introspect response: %v", err)
	}
	if !ir.Success {
		t.Fatalf("expected success=true, got error=%+v", ir.Error)
	}

	var introspectPayload struct {
		Active    bool   `json:"active"`
		TenantID  string `json:"tenant_id"`
		TokenType string `json:"token_type"`
	}
	if err := json.Unmarshal(ir.Data, &introspectPayload); err != nil {
		t.Fatalf("decode introspect data: %v", err)
	}
	if !introspectPayload.Active {
		t.Fatalf("expected active=true")
	}
	if introspectPayload.TenantID != "tenant1" {
		t.Fatalf("expected tenant_id=tenant1, got %q", introspectPayload.TenantID)
	}
	if introspectPayload.TokenType != "access" {
		t.Fatalf("expected token_type=access, got %q", introspectPayload.TokenType)
	}
}

func TestCreateUser_ValidatesInputAndPersists(t *testing.T) {
	t.Parallel()

	repo, h := newTestServer(t)

	badReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"tenant_id":"tenant1"}`))
	badReq.Header.Set("Content-Type", "application/json")
	badRec := httptest.NewRecorder()
	h.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", badRec.Code, badRec.Body.String())
	}

	body := bytes.NewBufferString(`{"email":"eve@example.com","tenant_id":"tenant1"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var r apiResp
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	var userPayload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Data, &userPayload); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if userPayload.ID == "" {
		t.Fatalf("expected id")
	}

	u, err := repo.GetUser(userPayload.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.Email != "eve@example.com" {
		t.Fatalf("expected email persisted, got %q", u.Email)
	}
}

func TestTenantInfo_Returns404ForUnknownTenant(t *testing.T) {
	t.Parallel()

	_, h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/tenants/info?tenant_id=does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
