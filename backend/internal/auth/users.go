// Package auth provides minimal credential checking and JWT issuance for the
// signaling server. It intentionally uses two hardcoded dummy users — there is
// no database in this project.
package auth

import "crypto/subtle"

// dummyUsers maps username -> password. These are demo credentials only.
var dummyUsers = map[string]string{
	"alice": "alice123",
	"bob":   "bob123",
}

// CheckCredentials reports whether the username/password pair is valid. The
// password comparison is constant-time to avoid leaking timing information.
func CheckCredentials(username, password string) bool {
	want, ok := dummyUsers[username]
	if !ok {
		// Still perform a comparison against a fixed string so the response
		// time does not reveal whether the username exists.
		subtle.ConstantTimeCompare([]byte(password), []byte("dummy-placeholder"))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(want)) == 1
}
