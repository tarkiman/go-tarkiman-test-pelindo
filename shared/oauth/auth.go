package oauth

import (
	"github.com/jmoiron/sqlx"
	"github.com/tarkiman/go/shared/failure"
)

type GrantType string

const (
	ClientCredentials GrantType = "client_credentials"
	Password          GrantType = "password"

	// HeaderAuthorization Request Header supplied for authorization
	HeaderAuthorization = "Authorization"

	DefaultAccessLifetime  = 3600
	DefaultRefreshLifetime = 1209600
)

type Token struct {
	config          Config
	tokenRepository TokenStore
}

func New(db *sqlx.DB, config Config) *Token {
	return &Token{
		config:          config,
		tokenRepository: NewTokenStore(db),
	}
}

type Config struct {
	AccessTokenLifetime  int
	RefreshTokenLifetime int
	Expiration           int64
	ClientScope          []string
}

// Create is function to store NewToken into database
func (t *Token) Create(credential Credential) (*TokenResponse, error) {
	grant, err := NewGrant(t.tokenRepository, t.config).Create(credential)
	if err != nil {
		return &TokenResponse{}, err
	}

	return grant.toCreateTokenResponse(), nil
}

// ParseWithAccessToken is function to exchange valid token into token info
func (t *Token) ParseWithAccessToken(accessToken string, method string, endpoint string) (OauthAccessToken, error) {
	return NewParser(t.tokenRepository).Parse(accessToken, method, endpoint)
}

// ClientScopeAllowed is function that is used to limit the client
// set * to allowed all client example in confing, ex : ClientScope: ["*"] or keep it empty
// set clientId to limit scope, ex : ClientScope: ["client_web"]
func (t *Token) ClientScopeAllowed(clientID string) bool {
	if len(t.config.ClientScope) == 0 {
		return true
	}

	if len(t.config.ClientScope) > 0 {
		if t.config.ClientScope[0] == "*" {
			return true
		}
	}

	for _, c := range t.config.ClientScope {
		if c == clientID {
			return true
		}
	}

	return false
}

// CreateByTokenRequest is function to generate & store access & refresh token into database
func (t *Token) CreateByTokenRequest(request OauthAccessTokenRequest) (res *TokenResponse, err error) {
	grant := NewGrant(t.tokenRepository, t.config)

	if _, err = t.tokenRepository.resolveClientByClientID(request.ClientID); err != nil {
		if err.Error() == ErrorClientNotFound {
			err = failure.NotFound(err.Error())
		}
		return
	}

	return res, t.tokenRepository.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		accessToken, err := grant.CreateUserToken(tx, request)
		if err != nil {
			e <- failure.InternalError(err)
			return
		}

		refreshToken, err := grant.CreateRefreshToken(tx, request)
		if err != nil {
			e <- failure.InternalError(err)
			return
		}

		res = accessToken.toCreateTokenResponse()
		res.ExpiresIn = grant.Config.AccessTokenLifetime
		res.RefreshToken = refreshToken.RefreshToken

		e <- nil
	})
}
