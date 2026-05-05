package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/loaikanou/GoAuth/internal/db"
	httpx "github.com/loaikanou/GoAuth/internal/httpx"
	"github.com/loaikanou/GoAuth/internal/policy"
	"github.com/loaikanou/GoAuth/internal/store"
	"github.com/loaikanou/GoAuth/internal/token"
)

type Repo interface {
	db.TenantRepo
	db.UserRepo
}

// Server aggregates core services for the MVP HTTP API.
type Server struct {
	Repo     Repo
	TokenSvc *token.Service
	Policy   policy.PolicyEngine
}

func main() {
	signingKey, ephemeral := loadJWTSigningKey()
	if ephemeral {
		log.Printf("JWT_SIGNING_KEY not set; using a random ephemeral signing key (tokens will become invalid on restart)")
	}
	ts, err := token.NewService(signingKey)
	if err != nil {
		log.Fatalf("invalid JWT signing configuration: %v", err)
	}

	repo, closeRepo := loadRepo()
	if closeRepo != nil {
		defer closeRepo()
	}

	cerbos := &policy.CerbosPolicyEngine{Rules: []policy.CerbosRule{
		{Subject: "system", Action: "authorize", Resource: "tenant/tenant1", Allowed: true},
		{Subject: "system", Action: "authorize", Resource: "tenant/*", Allowed: false},
	}}
	policyEngine := cerbos
	app := &Server{Repo: repo, TokenSvc: ts, Policy: policyEngine}
	mux := newMux(app)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Auth server listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	case <-sigCh:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func loadRepo() (Repo, func()) {
	if dsn, ok := os.LookupEnv("POSTGRES_DSN"); ok && dsn != "" {
		pg, err := db.Connect(db.Config{DSN: dsn})
		if err != nil {
			log.Fatalf("failed to connect to Postgres: %v", err)
		}
		return pg, func() { _ = pg.Close() }
	}
	return store.NewInMemoryStore(), nil
}

func newMux(s *Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/auth/authorize", s.authorizeHandler)
	mux.HandleFunc("/token", s.tokenHandler)
	mux.HandleFunc("/token/revoke", s.revokeHandler)
	mux.HandleFunc("/token/introspect", s.introspectHandler)
	mux.HandleFunc("/tenants/info", s.tenantInfoHandler)
	mux.HandleFunc("/users", s.createUserHandler)
	return mux
}

func loadJWTSigningKey() (key []byte, ephemeral bool) {
	v, ok := os.LookupEnv("JWT_SIGNING_KEY")
	if ok && v != "" {
		return []byte(v), false
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate random JWT signing key: %v", err)
	}
	return b, true
}

// healthHandler provides a simple liveness probe.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]string{"status": "ok"}})
}

// authorizeHandler returns a placeholder authorization code.
func (s *Server) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "tenant1"
	}

	allowed, reason, err := s.Policy.Evaluate("system", "authorize", "tenant/"+tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "policy_error", "policy evaluation failed")
		return
	}
	if !allowed {
		writeError(w, http.StatusForbidden, "forbidden", reason)
		return
	}

	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]string{"code": "AUTH_CODE_12345"}})
}

// tokenHandler issues access/refresh tokens for MVP flows.
func (s *Server) tokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	// For MVP, accept any form and issue tokens for a system user.
	// In production, validate grant_type, code, client_id, etc.
	// Extract tenant_id from query or default.
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "tenant1"
	}
	access, err := s.TokenSvc.CreateAccessToken("system", tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_creation_failed", "token creation failed")
		return
	}
	refresh, err := s.TokenSvc.CreateRefreshToken("system", tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_creation_failed", "refresh token creation failed")
		return
	}
	resp := map[string]interface{}{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int(s.TokenSvc.AccessTTL.Seconds()),
		"refresh_token": refresh,
		"scope":         "openid profile email",
	}
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: resp})
}

func (s *Server) revokeHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder: in production, revoke token(s) in store/cache.
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]string{"status": "revoked"}})
}

func (s *Server) introspectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	tokenString := ""
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			var payload struct {
				Token string `json:"token"`
			}
			if err := decodeJSON(w, r, &payload); err == nil {
				tokenString = payload.Token
			}
		} else {
			_ = r.ParseForm()
			tokenString = r.FormValue("token")
		}
	} else {
		tokenString = r.URL.Query().Get("token")
	}

	if tokenString == "" {
		w.WriteHeader(http.StatusOK)
		httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]interface{}{"active": false}})
		return
	}

	claims, err := s.TokenSvc.ParseAndValidate(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]interface{}{"active": false}})
		return
	}

	info := map[string]interface{}{
		"active":     true,
		"sub":        claims.Subject,
		"tenant_id":  claims.TenantID,
		"exp":        claims.ExpiresAt.Unix(),
		"token_type": claims.TokenType,
	}
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: info})
}

func (s *Server) tenantInfoHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "tenant_id required")
		return
	}

	t, err := s.Repo.GetTenant(tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "tenant not found")
		return
	}
	if t.DefaultLanguage == "" {
		t.DefaultLanguage = "en"
	}
	if t.Slug == "" {
		t.Slug = t.ID
	}
	if t.Name == "" {
		t.Name = "Sample Tenant"
	}

	info := map[string]interface{}{
		"tenant_id":        t.ID,
		"name":             t.Name,
		"slug":             t.Slug,
		"default_language": t.DefaultLanguage,
	}
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: info})
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	type req struct {
		Email    string `json:"email"`
		TenantID string `json:"tenant_id"`
	}
	var payload req
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "bad request")
		return
	}

	if payload.TenantID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "tenant_id required")
		return
	}
	if payload.Email == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "email required")
		return
	}

	id := "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	u := &store.User{
		ID:        id,
		TenantID:  payload.TenantID,
		Email:     payload.Email,
		UserType:  "end_user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.Repo.CreateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	resp := map[string]interface{}{
		"id":        u.ID,
		"email":     u.Email,
		"tenant_id": u.TenantID,
		"status":    u.Status,
	}
	w.WriteHeader(http.StatusOK)
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: resp})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return err
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	httpx.WriteJSON(w, httpx.APIResponse{
		Success: false,
		Error:   &httpx.ErrorResponse{Code: code, Message: message},
	})
}
