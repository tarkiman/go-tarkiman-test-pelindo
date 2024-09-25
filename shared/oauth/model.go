package oauth

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/guregu/null"
)

type TokenType string

const (
	Bearer TokenType = "Bearer"
)

type Scope string

const (
	ScopeBrandID Scope = "brandId"
)

var scope = struct {
	User string
}{
	"user",
}

// Credential is
type Credential struct {
	GrantType    GrantType
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

type OauthAccessToken struct {
	AccessToken string      `json:"accessToken" db:"access_token"`
	ClientID    string      `json:"clientId" db:"client_id"`
	UserID      null.String `json:"userId" db:"user_id"`
	Expires     time.Time   `json:"expires" db:"expires"`
	Scope       null.String `json:"scope" db:"scope"`
}

type OauthRefreshToken struct {
	RefreshToken string      `json:"refreshToken" db:"refresh_token"`
	ClientID     string      `json:"clientId" db:"client_id"`
	UserID       null.String `json:"userId" db:"user_id"`
	Expires      time.Time   `json:"expires" db:"expires"`
	Scope        null.String `json:"scope" db:"scope"`
}

func (o *OauthAccessToken) Generate(accessToken string, clientID string, userID string, withScope bool, config *Config) OauthAccessToken {
	if userID != "" {
		o.UserID = null.StringFrom(userID)
	}

	if withScope {
		o.Scope = null.StringFrom(scope.User)
	}

	o.ClientID = clientID
	o.AccessToken = accessToken
	o.Expires = time.Now().Add(time.Second * time.Duration(config.Expiration))

	return *o
}

func (o *OauthAccessToken) VerifyUserId() bool {
	return o.UserID.Valid
}

func (o *OauthAccessToken) VerifyExpireIn() bool {
	now := time.Now()
	if now.After(o.Expires) {
		return false
	}

	return true
}

func (o *OauthAccessToken) VerifyUserLoggedIn() bool {
	if o.UserID.Valid && !o.Scope.Valid {
		return true
	}
	return false
}

func (o *OauthAccessToken) toCreateTokenResponse() *TokenResponse {

	if !o.UserID.Valid {
		o.Scope = null.StringFrom(scope.User)
	}

	return &TokenResponse{
		AccessToken: o.AccessToken,
		// ExpiresIn:   o.Expires,
		TokenType: string(Bearer),
		Scope:     scope.User,
	}
}

type OauthClient struct {
	ClientID     string `json:"clientId" db:"client_id"`
	ClientSecret string `json:"clientSecret" db:"client_secret"`
	RedirectURI  string `json:"redirectUri" db:"redirect_uri"`
	GrantTypes   string `json:"grantTypes" db:"grant_types"`
}

func (o *OauthClient) VerifyClient(credential Credential) bool {
	if o.ClientID != credential.ClientID {
		return false
	}

	if o.ClientSecret != credential.ClientSecret {
		return false
	}

	return true
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

func (u *User) ValidCredential(credential Credential) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(credential.Password))
	if err != nil {
		return false
	}

	return true
}

func (o *OauthAccessTokenRequest) ScopeBrandID() string {
	return fmt.Sprintf("%s:%s", ScopeBrandID, o.BrandID)
}

// func (o *OauthAccessToken) Generate(accessToken string, request OauthAccessTokenRequest, withScope bool, config *Config) OauthAccessToken {
// 	if request.UserID != "" {
// 		o.UserID = null.StringFrom(request.UserID)
// 	}

// 	if withScope {
// 		o.Scope = null.StringFrom(scope.User)
// 	}

// 	if request.BrandID != "" {
// 		o.Scope = null.StringFrom(request.ScopeBrandID())
// 	}

// 	o.ClientID = request.ClientID
// 	o.AccessToken = accessToken

// 	//set static value to handle empty env
// 	if config.AccessTokenLifetime == 0 {
// 		config.AccessTokenLifetime = DefaultAccessLifetime
// 	}

// 	o.Expires = time.Now().Add(time.Second * time.Duration(config.AccessTokenLifetime))

// 	return *o
// }

// OauthAccessTokenRequest is token request used to do access oauth
type OauthAccessTokenRequest struct {
	ClientID    string `json:"clientId" validate:"required"`
	UserID      string `json:"userId" validate:"required"`
	BrandID     string `json:"brandId"`
	LoginMethod string `json:"loginMethod"`
	DeviceInfo  string `json:"-" validate:"omitempty"`
	IpAddress   string `json:"-" validate:"omitempty"`
}
