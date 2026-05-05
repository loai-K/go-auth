package store

import "time"

// InMemoryStore provides a minimal, in‑memory data layer for MVP development.
type InMemoryStore struct {
  Tenants  map[string]Tenant
  Users    map[string]User
  Sessions map[string]Session
  Tokens   map[string]RefreshToken
}

// Tenant represents a tenant in the system.
type Tenant struct {
  ID                 string
  Name               string
  Slug               string
  DefaultLanguage    string
  SupportedLanguages []string
  Settings           map[string]interface{}
  Branding           map[string]interface{}
  CreatedAt          time.Time
  UpdatedAt          time.Time
}

// User represents a user in a tenant.
type User struct {
  ID                  string
  TenantID            string
  UserType            string
  Email               string
  EmailVerified       bool
  PasswordHash        string
  PasswordLastChanged time.Time
  PreferredLanguage   string
  Profile             map[string]interface{}
  Status              string
  CreatedAt           time.Time
  UpdatedAt           time.Time
}

// Session tracks a login session for a user.
type Session struct {
  ID              string
  UserID          string
  DeviceID        string
  IPAddress       string
  UserAgent       string
  LoginTime       time.Time
  ExpiresAt       time.Time
  LastSeen        time.Time
  RefreshTokenID  string
}

// RefreshToken represents a persistent refresh token.
type RefreshToken struct {
  ID        string
  UserID    string
  TenantID  string
  TokenHash string
  ExpiresAt time.Time
  RevokedAt *time.Time
  CreatedAt time.Time
}

// NewInMemoryStore initializes a tiny dataset for MVP testing.
func NewInMemoryStore() *InMemoryStore {
  now := time.Now()
  t := Tenant{
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
  u := User{
    ID:            "user1",
    TenantID:      t.ID,
    UserType:      "end_user",
    Email:         "alice@example.com",
    EmailVerified: true,
    PasswordHash:  "",
    PasswordLastChanged: now,
    PreferredLanguage: "en",
    Profile:       map[string]interface{}{},
    Status:        "active",
    CreatedAt:     now,
    UpdatedAt:     now,
  }
  s := Session{
    ID:        "sess1",
    UserID:    u.ID,
    DeviceID:  "dev1",
    IPAddress: "127.0.0.1",
    UserAgent: "GoClient",
    LoginTime: now,
    ExpiresAt: now.Add(24 * time.Hour),
    LastSeen:  now,
    RefreshTokenID: "rt1",
  }
  rt := RefreshToken{
    ID:        "rt1",
    UserID:    u.ID,
    TenantID:  t.ID,
    TokenHash: "hash",
    ExpiresAt: now.Add(30 * 24 * time.Hour),
    CreatedAt: now,
  }
  return &InMemoryStore{
    Tenants:  map[string]Tenant{t.ID: t},
    Users:    map[string]User{u.ID: u},
    Sessions: map[string]Session{s.ID: s},
    Tokens:   map[string]RefreshToken{rt.ID: rt},
  }
}
