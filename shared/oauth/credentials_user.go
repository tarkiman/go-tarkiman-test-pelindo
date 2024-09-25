package oauth

import (
	"errors"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type UserCredentialsAuth struct {
	tokenStore TokenStore
	config     *Config
}

func NewUserCredentialAuth(tokenStore TokenStore, config Config) UserCredentialsAuth {
	return UserCredentialsAuth{
		tokenStore: tokenStore,
		config:     &config,
	}
}

func (c *UserCredentialsAuth) Create(credential Credential) (oauthAccessToken OauthAccessToken, err error) {
	client, err := c.tokenStore.resolveClientByClientID(credential.ClientID)
	if err != nil {
		return
	}

	if !client.VerifyClient(credential) {
		err = errors.New(ErrorInvalidClient)
		return
	}

	user, err := c.tokenStore.resolveByTelephoneOrEmail(credential.Username)
	if err != nil {
		return
	}

	if !user.ValidCredential(credential) {
		err = errors.New(ErrorInvalidPassword)
		return
	}

	accessToken, err := generateAccessToken()
	if err != nil {
		err = errors.New(ErrorGenerateAccessToken)
		return
	}

	request := OauthAccessTokenRequest{
		ClientID: credential.ClientID,
		UserID:   strconv.Itoa(user.ID),
	}

	oauthAccessToken = new(OauthAccessToken).Generate(accessToken, request.ClientID, request.UserID, false, c.config)

	err = c.tokenStore.createAccessToken(oauthAccessToken)
	if err != nil {
		return
	}

	return
}

func (c *UserCredentialsAuth) CreateWithTx(tx *sqlx.Tx, request OauthAccessTokenRequest) (oauthAccessToken OauthAccessToken, err error) {
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

func (c *UserCredentialsAuth) ExtendWithTx(tx *sqlx.Tx, accessToken string) (err error) {
	//set static value to handle empty env
	if c.config.AccessTokenLifetime == 0 {
		c.config.AccessTokenLifetime = DefaultAccessLifetime
	}

	expires := time.Now().Add(time.Second * time.Duration(c.config.AccessTokenLifetime))
	err = c.tokenStore.extendAccessTokenLifetimeWithTx(tx, accessToken, expires)
	if err != nil {
		return
	}

	return
}
