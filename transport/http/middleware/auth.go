package middleware

import (
	"net/http"

	"github.com/tarkiman/go/infras"
	"github.com/tarkiman/go/shared/oauth"
	"github.com/tarkiman/go/transport/http/response"
)

type Authentication struct {
	db         *infras.OracleConn
	TokenRead  *oauth.Token
	TokenWrite *oauth.Token
}

const (
	HeaderAuthorization = "Authorization"
)

func ProvideAuthentication(db *infras.OracleConn) *Authentication {
	return &Authentication{
		db:         db,
		TokenRead:  oauth.New(db.Read, oauth.Config{}),
		TokenWrite: oauth.New(db.Write, oauth.Config{}),
	}
}

// func (a *Authentication) UserCredential(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		accessToken := r.Header.Get(oauth.HeaderAuthorization)
// 		parseToken, err := a.TokenRead.ParseWithAccessToken(accessToken)

// 		// check to db write when client not found in db read
// 		if err != nil && err.Error() == oauth.ErrorClientNotFound {
// 			parseToken, err = a.TokenWrite.ParseWithAccessToken(accessToken)
// 		}

// 		if err != nil {
// 			response.WithMessage(w, http.StatusUnauthorized, err.Error())
// 			return
// 		}

// 		if !parseToken.VerifyExpireIn() {
// 			response.WithMessage(w, http.StatusUnauthorized, "Token Expired")
// 			return
// 		}

// 		if !parseToken.VerifyUserId() {
// 			response.WithMessage(w, http.StatusUnauthorized, oauth.ErrorInvalidPassword)
// 			return
// 		}

// 		response.WithJSON(w, http.StatusOK, parseToken)

// 	})
// }

func (a *Authentication) ClientCredential(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get(HeaderAuthorization)
		token := oauth.New(a.db.Read, oauth.Config{})

		parseToken, err := token.ParseWithAccessToken(accessToken, r.Method, r.RequestURI)
		if err != nil {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !parseToken.VerifyExpireIn() {
			response.WithMessage(w, http.StatusUnauthorized, "Token Expired")
			return
		}

		headerUserId := r.Header.Get("x-userid")
		if headerUserId == "" {
			r.Header.Set("x-userid", parseToken.UserID.String)
		}

		headerAccessToken := r.Header.Get("x-access-token")
		if headerAccessToken == "" {
			r.Header.Set("x-access-token", parseToken.AccessToken)
		}

		next.ServeHTTP(w, r)
	})
}

// func (a *Authentication) ClientCredentialWithQueryParameter(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		params := r.URL.Query()
// 		token := params.Get("token")
// 		tokenType := params.Get("token_type")
// 		accessToken := tokenType + " " + token

// 		auth := oauth.New(a.db.Read, oauth.Config{})
// 		parseToken, err := auth.ParseWithAccessToken(accessToken)
// 		if err != nil {
// 			response.WithMessage(w, http.StatusUnauthorized, err.Error())
// 			return
// 		}

// 		if !parseToken.VerifyExpireIn() {
// 			response.WithMessage(w, http.StatusUnauthorized, err.Error())
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// func (a *Authentication) Password(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		accessToken := r.Header.Get(HeaderAuthorization)
// 		token := oauth.New(a.db.Read, oauth.Config{})

// 		parseToken, err := token.ParseWithAccessToken(accessToken)
// 		if err != nil {
// 			response.WithMessage(w, http.StatusUnauthorized, err.Error())
// 			return
// 		}

// 		if !parseToken.VerifyExpireIn() {
// 			response.WithMessage(w, http.StatusUnauthorized, err.Error())
// 			return
// 		}

// 		if !parseToken.VerifyUserLoggedIn() {
// 			response.WithMessage(w, http.StatusUnauthorized, oauth.ErrorInvalidPassword)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }
