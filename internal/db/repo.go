package db

import st "github.com/loaikanou/GoAuth/internal"

// TenantRepo defines minimal operations for tenants.
type TenantRepo interface {
  CreateTenant(t *st.Tenant) error
  GetTenant(id string) (*st.Tenant, error)
}

// UserRepo defines minimal operations for users.
type UserRepo interface {
  CreateUser(u *st.User) error
  GetUser(id string) (*st.User, error)
}
