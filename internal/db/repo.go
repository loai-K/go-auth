package db

import storepkg "github.com/loaikanou/GoAuth/internal/store"

// TenantRepo defines minimal operations for tenants.
type TenantRepo interface {
  CreateTenant(t *storepkg.Tenant) error
  GetTenant(id string) (*storepkg.Tenant, error)
}

// UserRepo defines minimal operations for users.
type UserRepo interface {
  CreateUser(u *storepkg.User) error
  GetUser(id string) (*storepkg.User, error)
}
