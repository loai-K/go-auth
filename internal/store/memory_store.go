package store

import (
  "time"
  st "github.com/loaikanou/GoAuth/internal"
)

// InMemoryStore provides a minimal, in‑memory data layer for MVP development.
type InMemoryStore struct {
  Tenants  map[string]st.Tenant
  Users    map[string]st.User
  Sessions map[string]st.Session
  Tokens   map[string]st.RefreshToken
}

// NewInMemoryStore initializes a tiny dataset for MVP testing.
func NewInMemoryStore() *InMemoryStore {
  now := time.Now()
  t := st.Tenant{
    ID:              "tenant1",
    Name:            "Acme Corp",
    Slug:            "acme",
    DefaultLanguage: "en",
    SupportedLanguages: []string{"en"},
    Settings:        map[string]interface{}{},
    Branding:        map[string]interface{}{},
    CreatedAt:       now,
    UpdatedAt:       now,
  }
  u := st.User{
    ID:                  "user1",
    TenantID:            t.ID,
    UserType:            "end_user",
    Email:               "alice@example.com",
    EmailVerified:       true,
    PasswordHash:        "",
    PasswordLastChanged: now,
    PreferredLanguage:   "en",
    Profile:             map[string]interface{}{},
    Status:              "active",
    CreatedAt:           now,
    UpdatedAt:           now,
  }
  s := st.Session{
    ID:         "sess1",
    UserID:     u.ID,
    DeviceID:   "dev1",
    IPAddress:  "127.0.0.1",
    UserAgent:  "GoClient",
    LoginTime:  now,
    ExpiresAt:  now.Add(24 * time.Hour),
    LastSeen:   now,
    RefreshTokenID: "rt1",
  }
  rt := st.RefreshToken{
    ID:        "rt1",
    UserID:    u.ID,
    TenantID:  t.ID,
    TokenHash: "hash",
    ExpiresAt: now.Add(30 * 24 * time.Hour),
    CreatedAt: now,
  }
  return &InMemoryStore{
    Tenants:  map[string]st.Tenant{t.ID: t},
    Users:    map[string]st.User{u.ID: u},
    Sessions: map[string]st.Session{s.ID: s},
    Tokens:   map[string]st.RefreshToken{rt.ID: rt},
  }
}
