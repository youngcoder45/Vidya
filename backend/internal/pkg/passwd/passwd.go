// Package passwd hashes and verifies passwords with bcrypt.
package passwd

import "golang.org/x/crypto/bcrypt"

// Hash returns a bcrypt hash of the password.
func Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// Verify reports whether password matches the stored hash.
func Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
