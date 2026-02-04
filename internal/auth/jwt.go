package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtClockLeeway = 5 * time.Second

type jwtTokenHandler struct {
	signKey []byte
	aesgcm  cipher.AEAD
}

func (jwtH *jwtTokenHandler) sign(tokenID string, claims JWTClaims, lifetime time.Duration) (string, error) {
	now := time.Now()
	if lifetime < 0 {
		lifetime = 0
	}

	claims.Subject = tokenID
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(lifetime))
	claims.IssuedAt = jwt.NewNumericDate(now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtH.signKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (jwtH *jwtTokenHandler) encrypt(token string) string {
	cipherText := jwtH.aesgcm.Seal(nil, nil, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(cipherText)
}

func (jwtH *jwtTokenHandler) decrypt(cipherText string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	token, err := jwtH.aesgcm.Open(nil, nil, raw, nil)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (jwtH *jwtTokenHandler) verify(rawToken string, opts ...jwt.ParserOption) (*JWTClaims, error) {
	claims := &JWTClaims{}
	defaultOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(jwtClockLeeway),
	}
	defaultOpts = append(defaultOpts, opts...)

	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		return jwtH.signKey, nil
	}, defaultOpts...)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func newJWTHandler(signKey, encryptKey []byte) (*jwtTokenHandler, error) {
	hashedEncKey := sha256.Sum256(encryptKey)
	block, err := aes.NewCipher(hashedEncKey[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	return &jwtTokenHandler{
		signKey: signKey,
		aesgcm:  aesgcm,
	}, nil
}
