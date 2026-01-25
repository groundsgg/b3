package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenHandler struct {
	secretKey []byte
}

func (jwtH *jwtTokenHandler) sign(tokenID, username string, pl int) (string, error) {
	claims := jwt.MapClaims{
		"sub":              tokenID,
		"exp":              time.Now().Add(time.Hour * 24).Unix(),
		"iat":              time.Now().Unix(),
		"username":         username,
		"permission_level": pl,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtH.secretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (jwtH *jwtTokenHandler) verify(rawToken string) (*UserInfo, error) {
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
		return jwtH.secretKey, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithJSONNumber())

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("invalid username claim")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, fmt.Errorf("invalid subject claim")
	}

	plVal, ok := claims["permission_level"]
	if !ok {
		return nil, fmt.Errorf("missing permission level claim")
	}

	plN, ok := plVal.(json.Number)
	if !ok {
		return nil, fmt.Errorf("invalid permission level claim: %T", plVal)
	}

	pl, err := plN.Int64()
	if err != nil || pl > 255 {
		return nil, fmt.Errorf("invalid permission level claim: %w", err)
	}

	return &UserInfo{
		Username:        username,
		PermissionLevel: uint8(pl),
		SessionID:       sub,
	}, nil
}
