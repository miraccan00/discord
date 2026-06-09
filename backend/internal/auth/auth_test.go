package auth

import "testing"

func TestCheckCredentials(t *testing.T) {
	// Arrange
	cases := []struct {
		name     string
		user     string
		pass     string
		expected bool
	}{
		{"valid alice", "alice", "alice123", true},
		{"valid bob", "bob", "bob123", true},
		{"wrong password", "alice", "nope", false},
		{"unknown user", "charlie", "whatever", false},
		{"empty", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := CheckCredentials(tc.user, tc.pass)

			// Assert
			if got != tc.expected {
				t.Fatalf("CheckCredentials(%q, %q) = %v, want %v", tc.user, tc.pass, got, tc.expected)
			}
		})
	}
}

func TestIssuerRoundTrip(t *testing.T) {
	// Arrange
	issuer := NewIssuer("test-secret", 60)

	// Act
	token, err := issuer.Issue("alice")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	subject, err := issuer.Verify(token)

	// Assert
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if subject != "alice" {
		t.Fatalf("subject = %q, want %q", subject, "alice")
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	// Arrange: a token that expired one minute ago.
	issuer := NewIssuer("test-secret", -1)
	token, err := issuer.Issue("alice")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	// Act
	_, err = issuer.Verify(token)

	// Assert
	if err == nil {
		t.Fatal("expected expired token to be rejected, got nil error")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	// Arrange: sign with one secret, verify with another.
	signer := NewIssuer("secret-a", 60)
	verifier := NewIssuer("secret-b", 60)
	token, err := signer.Issue("alice")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	// Act
	_, err = verifier.Verify(token)

	// Assert
	if err == nil {
		t.Fatal("expected tampered/foreign token to be rejected, got nil error")
	}
}
