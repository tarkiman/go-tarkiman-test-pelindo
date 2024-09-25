package oauth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tarkiman/go/shared/failure"
)

type TokenStore struct {
	db *sqlx.DB
}

const (
	queryInsertAccessToken = `INSERT INTO oauth_access_tokens (
			access_token,
			client_id,
			user_id,
			expires,
			scope
		) VALUES (
			:access_token,
			:client_id,
			:user_id,
			:expires,
			:scope
		)`
	queryInsertRefreshToken = `INSERT INTO oauth_refresh_tokens (
			refresh_token,
			client_id,
			user_id,
			expires,
			scope
		) VALUES (
			:refresh_token,
			:client_id,
			:user_id,
			:expires,
			:scope
		)`

	querySelectAccessToken = `SELECT 
			access_token,
			client_id,
			user_id,
			expires,
			scope
		FROM
			oauth_access_tokens`

	querySelectAccessTokenAndEndpoint = `SELECT
											oat.access_token,
											oat.client_id,
											oat.user_id,
											oat.expires,
											oat.scope
										FROM oauth_access_tokens oat
										JOIN users u ON u.id=oat.user_id
										JOIN user_role ur ON  ur.id_user=u.id
										JOIN role r ON r.id=ur.id_role
										JOIN role_permission rp ON rp.id_role=r.id
										JOIN permission p ON p.id=rp.id_permission`

	querySelectClients = `SELECT
			client_id,
			client_secret,
			redirect_uri,
			grant_types
		FROM 
			oauth_clients`

	querySelectUser = `
			SELECT
				id,
				username,
				password
			FROM
				user`
)

func NewTokenStore(db *sqlx.DB) TokenStore {
	return TokenStore{
		db: db,
	}
}

func (a *TokenStore) createAccessToken(accessToken OauthAccessToken) error {
	stmt, err := a.db.PrepareNamed(queryInsertAccessToken)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (a *TokenStore) resolveAccessTokenByAccessToken(accessToken string) (oauthAccessToken OauthAccessToken, err error) {
	err = a.db.Get(&oauthAccessToken, querySelectAccessToken+" WHERE access_token = ?", accessToken)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New(ErrorClientNotFound)
		return
	case err != nil:
		return
	}

	return
}

func (a *TokenStore) resolveAccessTokenByAccessTokenAndEndpoint(accessToken string, method string, endpoint string) (oauthAccessToken OauthAccessToken, err error) {
	err = a.db.Get(&oauthAccessToken, querySelectAccessTokenAndEndpoint+" WHERE oat.access_token=? AND p.method=? AND ? LIKE CONCAT(p.endpoint,'%') ORDER BY oat.access_token DESC LIMIT 1", accessToken, method, endpoint)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New(ErrorClientNotFound)
		return
	case err != nil:
		return
	}

	return
}

func (a *TokenStore) resolveAllClients(db *sqlx.DB) ([]OauthClient, error) {
	var clients []OauthClient

	err := a.db.Get(&clients, querySelectClients)
	if err != nil {
		return []OauthClient{}, err
	}

	return clients, nil
}

func (a *TokenStore) resolveClientByClientID(clientID string) (client OauthClient, err error) {
	err = a.db.Get(&client, querySelectClients+" WHERE client_id = ?", clientID)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New(ErrorClientNotFound)
		return
	case err != nil:
		return
	}

	return
}

func (a *TokenStore) resolveByTelephoneOrEmail(username string) (User, error) {
	var user User

	err := a.db.Get(&user, querySelectUser+" WHERE telephone = ? OR  email = ?", username, username)
	switch {
	case err == sql.ErrNoRows:
		return User{}, errors.New(ErrorClientNotFound)
	case err != nil:
		return User{}, err
	}

	return user, nil
}

// WithTransaction performs queries with transaction
func (a *TokenStore) WithTransaction(block func(db *sqlx.Tx, c chan error)) (err error) {
	e := make(chan error)
	tx, err := a.db.Beginx()
	if err != nil {
		return
	}
	go block(tx, e)
	err = <-e
	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			err = failure.InternalError(errTx)
		}
		return
	}
	err = tx.Commit()
	return
}

func (a *TokenStore) createRefreshTokenWithTx(tx *sqlx.Tx, refreshToken OauthRefreshToken) error {
	stmt, err := tx.PrepareNamed(queryInsertRefreshToken)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (a *TokenStore) createAccessTokenWithTx(tx *sqlx.Tx, accessToken OauthAccessToken) error {
	stmt, err := tx.PrepareNamed(queryInsertAccessToken)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (a *TokenStore) extendAccessTokenLifetimeWithTx(tx *sqlx.Tx, accessToken string, expires time.Time) (err error) {
	query := `UPDATE oauth_access_tokens SET expires = :expires WHERE access_token = :access_token`
	_, err = tx.NamedExec(query, map[string]interface{}{
		"access_token": accessToken,
		"expires":      expires,
	})

	return
}
