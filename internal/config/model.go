package config

type Web struct {
	ListenAddr string `validate:"listenaddr" default:":8080"`
	BaseURL    string `validate:"required,url" default:"http://localhost:8080"`
	SessionKey string `validate:"required,min=32"`
}

type AuthBasic struct {
	Users []string `validate:"required,min=1,dive,required"`
}

type AuthOIDC struct {
	ClientID     string `validate:"required"`
	ClientSecret string `validate:"required"`
	Issuer       string `validate:"required,url"`
}

type Auth struct {
	Type  string `validate:"required,oneof=basic_auth oidc"`
	Basic *AuthBasic
	OIDC  *AuthOIDC
}

type config struct {
	Web  Web  `validate:"required"`
	Auth Auth `validate:"required"`
}
