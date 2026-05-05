package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	httpx "github.com/loaikanou/GoAuth/internal/httpx"
	"github.com/loaikanou/GoAuth/internal/store"
	"github.com/loaikanou/GoAuth/internal/token"
	"github.com/loaikanou/GoAuth/internal/policy"
)

// Server aggregates core services for the MVP HTTP API.
type Server struct {
    Store    *store.InMemoryStore
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

    st := store.NewInMemoryStore()
    policyEngine := &policy.MockPolicyEngine{}
    srv := &Server{Store: st, TokenSvc: ts, Policy: policyEngine}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)
	mux.HandleFunc("/auth/authorize", srv.authorizeHandler)
	mux.HandleFunc("/token", srv.tokenHandler)
	mux.HandleFunc("/token/revoke", srv.revokeHandler)
	mux.HandleFunc("/token/introspect", srv.introspectHandler)
	mux.HandleFunc("/tenants/info", srv.tenantInfoHandler) // uses query param tenant_id
	mux.HandleFunc("/users", srv.createUserHandler)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Auth server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
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
    // Basic authorization decision via policy engine (Phase 2): evaluate access
    tenantID := r.URL.Query().Get("tenant_id")
    if tenantID == "" {
        tenantID = "tenant1"
    }
    allowed, reason, err := s.Policy.Evaluate("system", "authorize", "tenant/"+tenantID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if !allowed {
        http.Error(w, reason, http.StatusForbidden)
        return
    }
    resp := map[string]string{"code": "AUTH_CODE_12345"}
    httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: resp})
}

// tokenHandler issues access/refresh tokens for MVP flows.
func (s *Server) tokenHandler(w http.ResponseWriter, r *http.Request) {
	// For MVP, accept any form and issue tokens for a system user.
	// In production, validate grant_type, code, client_id, etc.
	// Extract tenant_id from query or default.
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "tenant1"
	}
	access, err := s.TokenSvc.CreateAccessToken("system", tenantID)
	if err != nil {
		http.Error(w, "token creation failed", http.StatusInternalServerError)
		return
	}
	refresh, err := s.TokenSvc.CreateRefreshToken("system", tenantID)
	if err != nil {
		http.Error(w, "refresh token creation failed", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int(s.TokenSvc.AccessTTL.Seconds()),
		"refresh_token": refresh,
		"scope":         "openid profile email",
	}
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: resp})
}

func (s *Server) revokeHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder: in production, revoke token(s) in store/cache.
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]string{"status": "revoked"}})
}

func (s *Server) introspectHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := ""
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			var payload struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
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
		httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: map[string]interface{}{"active": false}})
		return
	}

	claims, err := s.TokenSvc.ParseAndValidate(tokenString)
	if err != nil {
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
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: info})
}

func (s *Server) tenantInfoHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id required", http.StatusBadRequest)
		return
	}
	// Fetch from store if available; otherwise return a minimal payload.
	info := map[string]interface{}{
		"tenant_id":        tenantID,
		"name":             "Sample Tenant",
		"slug":             tenantID,
		"default_language": "en",
	}
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: info})
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email    string `json:"email"`
		TenantID string `json:"tenant_id"`
	}
	var payload req
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Create a simple user in memory (no password for MVP).
	id := "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	user := map[string]interface{}{
		"id":        id,
		"email":     payload.Email,
		"tenant_id": payload.TenantID,
		"status":    "active",
	}
	// In a real implementation, we'd persist the user to the DB. Here we just return the blob.
	httpx.WriteJSON(w, httpx.APIResponse{Success: true, Data: user})
}
