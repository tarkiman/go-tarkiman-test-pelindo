package oauth

import (
	"errors"
	"time"

	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
)

type RefreshCredentialsAuth struct {
	tokenStore TokenStore
	config     *Config
}

func NewRefreshCredentialsAuth(tokenStore TokenStore, config Config) RefreshCredentialsAuth {
	return RefreshCredentialsAuth{
		tokenStore: tokenStore,
		config:     &config,
	}
}

func (c *RefreshCredentialsAuth) CreateWithTx(tx *sqlx.Tx, request OauthAccessTokenRequest) (oauthRefreshToken OauthRefreshToken, err error) {
	refreshToken, err := generateAccessToken()
	if err != nil {
		err = errors.New(ErrorGenerateAccessToken)
		return
	}

	//set static value to handle empty env
	if c.config.RefreshTokenLifetime == 0 {
		c.config.RefreshTokenLifetime = DefaultRefreshLifetime
	}

	oauthRefreshToken = OauthRefreshToken{
		RefreshToken: refreshToken,
		ClientID:     request.ClientID,
		UserID:       null.StringFrom(request.UserID),
		Expires:      time.Now().Add(time.Second * time.Duration(c.config.RefreshTokenLifetime)),
	}

	if request.BrandID != "" {
		oauthRefreshToken.Scope = null.StringFrom(request.ScopeBrandID())
	}

	err = c.tokenStore.createRefreshTokenWithTx(tx, oauthRefreshToken)
	if err != nil {
		return
	}

	return
}

func (c *RefreshCredentialsAuth) ExtendWithTx(tx *sqlx.Tx, refreshToken string) (err error) {
	//set static value to handle empty env
	if c.config.RefreshTokenLifetime == 0 {
		c.config.RefreshTokenLifetime = DefaultRefreshLifetime
	}

	expires := time.Now().Add(time.Second * time.Duration(c.config.RefreshTokenLifetime))
	err = c.tokenStore.extendRefreshTokeLifetimeWithTx(tx, refreshToken, expires)
	if err != nil {
		return
	}

	return
}

func (a *TokenStore) extendRefreshTokeLifetimeWithTx(tx *sqlx.Tx, refreshToken string, expires time.Time) (err error) {
	query := `UPDATE oauth_refresh_tokens SET expires = :expires WHERE refresh_token = :refresh_token`
	_, err = tx.NamedExec(query, map[string]interface{}{
		"refresh_token": refreshToken,
		"expires":       expires,
	})

	return
}
