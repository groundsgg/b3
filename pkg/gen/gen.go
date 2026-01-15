// SPDX-License-Identifier: AGPL-3.0-or-later
package gen

import (
	"crypto/rand"
	"encoding/hex"
)

func Bytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func String(len int) (string, error) {
	buf, err := Bytes(len)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func Hex(len int) (string, error) {
	buf, err := Bytes(len)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
