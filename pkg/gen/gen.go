package gen

import (
	"crypto/rand"
	"encoding/hex"
)

// Bytes returns a slice of cryptographically random bytes of the given length.
func Bytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// String returns a raw string made from cryptographically random bytes.
func String(len int) (string, error) {
	buf, err := Bytes(len)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// Hex returns a hex-encoded string from cryptographically random bytes.
func Hex(len int) (string, error) {
	buf, err := Bytes(len)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
