package oauth

import "github.com/jmoiron/sqlx"

type AuthorizationMethod interface {
	Create(credential Credential) (OauthAccessToken, error)
}

type Grant struct {
	TokenStore TokenStore
	Config     *Config
	UserCredentialsAuth
	ClientCredentialsAuth
	RefreshCredentialsAuth
}

func NewGrant(tokenStore TokenStore, config Config) *Grant {
	return &Grant{
		TokenStore:             tokenStore,
		Config:                 &config,
		UserCredentialsAuth:    NewUserCredentialAuth(tokenStore, config),
		ClientCredentialsAuth:  NewClientCredentialAuth(tokenStore, config),
		RefreshCredentialsAuth: NewRefreshCredentialsAuth(tokenStore, config),
	}
}

func (g *Grant) Create(credential Credential) (OauthAccessToken, error) {
	authMap := make(map[GrantType]AuthorizationMethod)
	authMap[ClientCredentials] = &ClientCredentialsAuth{tokenStore: g.TokenStore, config: g.Config}
	authMap[Password] = &PasswordAuth{tokenStore: g.TokenStore, config: g.Config}

	return authMap[credential.GrantType].Create(credential)
}

func (g *Grant) CreateUserToken(tx *sqlx.Tx, request OauthAccessTokenRequest) (OauthAccessToken, error) {
	return g.UserCredentialsAuth.CreateWithTx(tx, request)
}

func (g *Grant) CreateRefreshToken(tx *sqlx.Tx, request OauthAccessTokenRequest) (OauthRefreshToken, error) {
	return g.RefreshCredentialsAuth.CreateWithTx(tx, request)
}

func (g *Grant) ExtendUserToken(tx *sqlx.Tx, accessToken string) (err error) {
	return g.UserCredentialsAuth.ExtendWithTx(tx, accessToken)
}

func (g *Grant) ExtendRefreshToken(tx *sqlx.Tx, refreshToken string) (err error) {
	return g.RefreshCredentialsAuth.ExtendWithTx(tx, refreshToken)
}
