package store

import "time"

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

// User represents a user within a tenant.
type User struct {
  ID                   string
  TenantID             string
  UserType             string
  Email                string
  EmailVerified        bool
  PasswordHash         string
  PasswordLastChanged  time.Time
  PreferredLanguage    string
  Profile              map[string]interface{}
  Status               string
  CreatedAt            time.Time
  UpdatedAt            time.Time
}

// Session tracks a user login session.
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
