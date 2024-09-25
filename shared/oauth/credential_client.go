package oauth

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

type ClientCredentialsAuth struct {
	tokenStore TokenStore
	config     *Config
}

func NewClientCredentialAuth(tokenStore TokenStore, config Config) ClientCredentialsAuth {
	return ClientCredentialsAuth{
		tokenStore: tokenStore,
		config:     &config,
	}
}

func (c *ClientCredentialsAuth) Create(credential Credential) (oauthAccessToken OauthAccessToken, err error) {
	client, err := c.tokenStore.resolveClientByClientID(credential.ClientID)
	if err != nil {
		return
	}

	if !client.VerifyClient(credential) {
		err = errors.New(ErrorInvalidClient)
		return
	}

	accessToken, err := generateAccessToken()
	if err != nil {
		err = errors.New(ErrorGenerateAccessToken)
		return
	}

	oauthAccessToken = new(OauthAccessToken).Generate(accessToken, credential.ClientID, "", true, c.config)
	err = c.tokenStore.createAccessToken(oauthAccessToken)
	if err != nil {
		return
	}

	return
}

func (c *ClientCredentialsAuth) CreateWithTx(tx *sqlx.Tx, request OauthAccessTokenRequest) (oauthAccessToken OauthAccessToken, err error) {
	accessToken, err := generateAccessToken()
	if err != nil {
		err = errors.New(ErrorGenerateAccessToken)
		return
	}

	oauthAccessToken = new(OauthAccessToken).Generate(accessToken, request.ClientID, request.UserID, false, c.config)

	err = c.tokenStore.createAccessTokenWithTx(tx, oauthAccessToken)
	if err != nil {
		return
	}

	return
}
