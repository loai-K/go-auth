package policy

// PolicyEngine evaluates whether a given action on a resource by a subject is permitted.
type PolicyEngine interface {
  Evaluate(subject string, action string, resource string) (bool, string, error)
}
