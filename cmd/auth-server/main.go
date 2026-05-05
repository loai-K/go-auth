package main

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/loaikanou/GoAuth/internal/store"
  "github.com/loaikanou/GoAuth/internal/token"
)

// Server aggregates core services for the MVP HTTP API.
type Server struct {
  Store    *store.InMemoryStore
  TokenSvc *token.Service
}

func main() {
  signingKey := []byte(getEnv("JWT_SIGNING_KEY", "default-secret-please-change"))
  ts := &token.Service{SigningKey: signingKey}

  st := store.NewInMemoryStore()
  srv := &Server{Store: st, TokenSvc: ts}

  mux := http.NewServeMux()
  mux.HandleFunc("/healthz", srv.healthHandler)
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

func getEnv(key, def string) string {
  if v := os.Getenv(key); v != "" {
    return v
  }
  return def
}

// healthHandler provides a simple liveness probe.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"ok"}`))
}

// authorizeHandler returns a placeholder authorization code.
func (s *Server) authorizeHandler(w http.ResponseWriter, r *http.Request) {
  // In a real system, this would redirect to a login page or consent screen.
  resp := map[string]string{"code": "AUTH_CODE_12345"}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)
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
    "expires_in":    3600,
    "refresh_token": refresh,
    "scope":         "openid profile email",
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)
}

func (s *Server) revokeHandler(w http.ResponseWriter, r *http.Request) {
  // Placeholder: in production, revoke token(s) in store/cache.
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
}

func (s *Server) introspectHandler(w http.ResponseWriter, r *http.Request) {
  // Simple introspection response for MVP.
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]interface{}{
    "active":     true,
    "sub":        "system",
    "tenant_id":  "tenant1",
    "exp":        time.Now().Add(1 * time.Hour).Unix(),
    "token_type": "access",
  })
}

func (s *Server) tenantInfoHandler(w http.ResponseWriter, r *http.Request) {
  tenantID := r.URL.Query().Get("tenant_id")
  if tenantID == "" {
    http.Error(w, "tenant_id required", http.StatusBadRequest)
    return
  }
  // Fetch from store if available; otherwise return a minimal payload.
  info := map[string]interface{}{
    "tenant_id": tenantID,
    "name":      "Sample Tenant",
    "slug":      tenantID,
    "default_language": "en",
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(info)
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
  type req struct {
    Email string `json:"email"`
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
    "id": id,
    "email": payload.Email,
    "tenant_id": payload.TenantID,
    "status": "active",
  }
  // In a real implementation, we'd persist the user to the DB. Here we just return the blob.
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(user)
}
