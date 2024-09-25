package infras

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/tarkiman/go/shared/failure"

	"github.com/tarkiman/go/configs"
	// use Oracle driver
	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	maxIdleConnection = 10
	maxOpenConnection = 10
)

// Block contains a transaction block
type Block func(db *sqlx.Tx, c chan error)

// OracleConn wraps a pair of read/write Oracle connections.
type OracleConn struct {
	Read  *sqlx.DB
	Write *sqlx.DB
}

// ProvideOracleConn is the provider for OracleConn.
func ProvideOracleConn(config *configs.Config) *OracleConn {
	return &OracleConn{
		Read:  CreateOracleReadConn(*config),
		Write: CreateOracleWriteConn(*config),
	}
}

// CreateOracleWriteConn creates a database connection for write access.
func CreateOracleWriteConn(config configs.Config) *sqlx.DB {
	return CreateDBConnection(
		"write",
		config.DB.Oracle.Write.Username,
		config.DB.Oracle.Write.Password,
		config.DB.Oracle.Write.Host,
		config.DB.Oracle.Write.Port,
		config.DB.Oracle.Write.Name,
		config.DB.Oracle.Write.Timezone)

}

// CreateOracleReadConn creates a database connection for read access.
func CreateOracleReadConn(config configs.Config) *sqlx.DB {
	return CreateDBConnection(
		"read",
		config.DB.Oracle.Read.Username,
		config.DB.Oracle.Read.Password,
		config.DB.Oracle.Read.Host,
		config.DB.Oracle.Read.Port,
		config.DB.Oracle.Read.Name,
		config.DB.Oracle.Read.Timezone)

}

// CreateDBConnection creates a database connection.
func CreateDBConnection(name, username, password, host, port, dbName, timeZone string) *sqlx.DB {
	descriptor := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s&parseTime=true",
		username,
		password,
		host,
		port,
		dbName,
		url.QueryEscape(timeZone))
	db, err := sqlx.Connect("oracle", descriptor)
	if err != nil {
		log.
			Fatal().
			Err(err).
			Str("name", name).
			Str("host", host).
			Str("port", port).
			Str("dbName", dbName).
			Msg("Failed connecting to database")
	} else {
		log.
			Info().
			Str("name", name).
			Str("host", host).
			Str("port", port).
			Str("dbName", dbName).
			Msg("Connected to database")
	}
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetMaxOpenConns(maxOpenConnection)

	return db
}

// OpenMock opens a database connection for mocking purposes.
func OpenMock(db *sql.DB) *OracleConn {
	conn := sqlx.NewDb(db, "oracle")
	return &OracleConn{
		Write: conn,
		Read:  conn,
	}
}

// WithTransaction performs queries with transaction
func (m *OracleConn) WithTransaction(block Block) (err error) {
	e := make(chan error)
	tx, err := m.Write.Beginx()
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
