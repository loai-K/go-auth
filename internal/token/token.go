package token

// Service encapsulates token signing mechanics (minimal MVP stub).
type Service struct {
  SigningKey []byte
}

// CreateAccessToken returns a simple, signed-like token for MVP (no external deps).
func (s *Service) CreateAccessToken(userID, tenantID string) (string, error) {
  // In a full implementation this would be a JWT; here we return a deterministic string.
  return "access-" + userID + "-" + tenantID, nil
}

// CreateRefreshToken returns a simple refresh token for MVP.
func (s *Service) CreateRefreshToken(userID, tenantID string) (string, error) {
  return "refresh-" + userID + "-" + tenantID, nil
}
