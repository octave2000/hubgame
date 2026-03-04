package controller

import "testing"

func TestAuthorizerEnforce(t *testing.T) {
	a := NewAuthorizer()

	if err := a.Enforce(&Claims{Role: "player"}, ActionEntityDelete); err == nil {
		t.Fatalf("expected player to be denied entity.delete")
	}

	if err := a.Enforce(&Claims{Role: "developer"}, ActionEntityDelete); err != nil {
		t.Fatalf("expected developer to be allowed entity.delete: %v", err)
	}
}
