// Package randutil generates cryptographically random tokens (refresh tokens)
// and numeric OTPs.
package randutil

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// Hex returns n random bytes hex-encoded (2n characters).
func Hex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Digits returns a numeric OTP of the given length (default 6).
func Digits(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	out := make([]byte, length)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		out[i] = byte('0' + n.Int64())
	}
	return string(out), nil
}
