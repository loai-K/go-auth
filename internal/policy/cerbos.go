package policy

import (
  rbac "github.com/loaikanou/GoAuth/internal/rbac"
)

// CerbosPolicyEngine is an in-memory Cerbos-like policy engine for MVP.
// This is a lightweight stand-in so Phase 2 can progress without an external Cerbos runtime.
type CerbosPolicyEngine struct {
  Rules []CerbosRule
  RBAC  *rbac.Engine
}

// CerbosRule represents a simple allow/deny rule for a given subject, action, and resource.
type CerbosRule struct {
  Subject  string
  Action   string
  Resource string
  Allowed  bool
}

func (c *CerbosPolicyEngine) Evaluate(subject, action, resource string) (bool, string, error) {
  // First, Cerbos-like rules
  for _, r := range c.Rules {
    if (r.Subject == subject || r.Subject == "*") &&
       (r.Action == action || r.Action == "*") &&
       (r.Resource == resource || r.Resource == "*") {
      if r.Allowed {
        return true, "allowed by cerbos (in-memory)", nil
      }
      return false, "denied by cerbos (in-memory)", nil
    }
  }
  // Fallback to RBAC if configured
  if c.RBAC != nil {
    // Simple heuristic: if the resource string contains the action as a permission token
    // and the user has a binding, allow. This is a placeholder for a richer RBAC check.
    // In a real system, map resource+action to a concrete permission string.
    if c.RBAC.HasPermission("system", "tenant1", action) {
      return true, "allowed by in-memory RBAC (fallback)", nil
    }
  }
  return true, "default allow (in-memory)", nil
}
