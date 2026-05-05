package rbac

import "sync"

type Role struct {
  ID          string
  Name        string
  TenantID    string
  Permissions []string
}

type UserRoleBinding struct {
  UserID   string
  RoleID   string
  TenantID string
}

// Engine is a minimal in-memory RBAC store.
type Engine struct {
  mu       sync.RWMutex
  roles    map[string]Role
  bindings []UserRoleBinding
}

func NewEngine() *Engine {
  return &Engine{roles: make(map[string]Role)}
}

func (e *Engine) AddRole(r Role) {
  e.mu.Lock()
  defer e.mu.Unlock()
  e.roles[r.ID] = r
}

func (e *Engine) BindUser(userID, roleID, tenantID string) {
  e.mu.Lock()
  defer e.mu.Unlock()
  e.bindings = append(e.bindings, UserRoleBinding{UserID: userID, RoleID: roleID, TenantID: tenantID})
}

// HasPermission checks if the user has the given permission via their roles.
func (e *Engine) HasPermission(userID, tenantID, permission string) bool {
  e.mu.RLock()
  defer e.mu.RUnlock()
  for _, b := range e.bindings {
    if b.UserID == userID && b.TenantID == tenantID {
      if r, ok := e.roles[b.RoleID]; ok {
        for _, p := range r.Permissions {
          if p == permission {
            return true
          }
        }
      }
    }
  }
  return false
}
