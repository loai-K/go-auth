package token

import (
	"testing"
	"time"
)

func TestNewService_RejectsShortKey(t *testing.T) {
	t.Parallel()

	_, err := NewService([]byte("short"))
	if err == nil {
		t.Fatalf("expected error for short key")
	}
}

func TestJWT_RoundTrip(t *testing.T) {
	t.Parallel()

	svc, err := NewService([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	tokenStr, err := svc.CreateAccessToken("user1", "tenant1")
	if err != nil {
		t.Fatalf("CreateAccessToken error: %v", err)
	}

	claims, err := svc.ParseAndValidate(tokenStr)
	if err != nil {
		t.Fatalf("ParseAndValidate error: %v", err)
	}

	if claims.Subject != "user1" {
		t.Fatalf("expected sub=user1, got %q", claims.Subject)
	}
	if claims.TenantID != "tenant1" {
		t.Fatalf("expected tenant_id=tenant1, got %q", claims.TenantID)
	}
	if claims.TokenType != "access" {
		t.Fatalf("expected token_type=access, got %q", claims.TokenType)
	}
	if claims.Issuer != svc.Issuer {
		t.Fatalf("expected iss=%q, got %q", svc.Issuer, claims.Issuer)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		t.Fatalf("expected exp in the future, got %v", claims.ExpiresAt)
	}
}

func TestJWT_RejectsWrongKey(t *testing.T) {
	t.Parallel()

	svcA, err := NewService([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewService A error: %v", err)
	}
	svcB, err := NewService([]byte("abcdef0123456789abcdef0123456789"))
	if err != nil {
		t.Fatalf("NewService B error: %v", err)
	}

	tokenStr, err := svcA.CreateAccessToken("user1", "tenant1")
	if err != nil {
		t.Fatalf("CreateAccessToken error: %v", err)
	}

	if _, err := svcB.ParseAndValidate(tokenStr); err == nil {
		t.Fatalf("expected validation error for token signed with a different key")
	}
}
