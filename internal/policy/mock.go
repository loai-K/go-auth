package policy

// MockPolicyEngine is a simple in-memory policy engine used for MVP/testing.
type MockPolicyEngine struct{}

// Evaluate always allows for MVP until a real policy is wired.
func (m *MockPolicyEngine) Evaluate(subject, action, resource string) (bool, string, error) {
  return true, "allowed by mock policy", nil
}
