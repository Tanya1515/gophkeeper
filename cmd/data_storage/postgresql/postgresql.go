package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLConnection struct {
	Host     string
	UserName string
	Password string
	DBName   string
	dbConn   *sql.DB
}

func NewPostgreSQLConnection(host string, userName string, password string, dbName string) *PostgreSQLConnection {
	postgreSQL := &PostgreSQLConnection{Host: host, UserName: userName, Password: password, DBName: dbName}

	return postgreSQL
}

func (pg *PostgreSQLConnection) Connect() (err error) {
	ps := fmt.Sprintf("host=%s user=%s password=%s database=%s sslmode=disable",
		pg.Host, pg.UserName, pg.Password, pg.DBName)

	pg.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return
	}

	_, err = pg.dbConn.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if err != nil {
		return fmt.Errorf("error while creating extension pgcrypto: %w", err)
	}

	_, err = pg.dbConn.Exec(`CREATE TABLE IF NOT EXISTS Users (Id BIGSERIAL PRIMARY KEY,
												userLogin VARCHAR(100) NOT NULL UNIQUE,
												userPassword VARCHAR(100) NOT NULL,
												userEmail VARCHAR(100) NOT NULL UNIQUE);`)
	
	if err != nil {
		return fmt.Errorf("error while creating table Users: %w", err)
	}

	return
}
