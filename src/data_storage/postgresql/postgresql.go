package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgreSQLConnection - structure, that contains all data for connecting to PostgreSQL
// and execute database requests.
type PostgreSQLConnection struct {
	Host     string // Host - database host
	UserName string // User - user, that connects to database
	Password string // Password - password, that gives an opportunity to connect to database
	DBName   string // DBName - database name,
	dbConn   *sql.DB
}

// NewPostgreSQLConnection - function for inserting PostgreSQL credentials for connection.
func NewPostgreSQLConnection(host string, userName string, password string, dbName string) *PostgreSQLConnection {
	postgreSQL := &PostgreSQLConnection{Host: host, UserName: userName, Password: password, DBName: dbName}

	return postgreSQL
}

// Connect - function for connecting to PotgreSQL and creating application tables for data storing.
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

	_, err = pg.dbConn.Exec(`CREATE TABLE IF NOT EXISTS Users (ID BIGSERIAL PRIMARY KEY,
												userLogin VARCHAR(100) NOT NULL UNIQUE,
												userPassword VARCHAR(100) NOT NULL,
												userEmail VARCHAR(100) NOT NULL UNIQUE);`)

	if err != nil {
		return fmt.Errorf("error while creating table Users: %w", err)
	}

	_, err = pg.dbConn.Exec(`CREATE TABLE IF NOT EXISTS BankCards (ID BIGSERIAL PRIMARY KEY,
															userID BIGINT REFERENCES Users (id) ON DELETE CASCADE,
															cardNumber VARCHAR(100) NOT NULL UNIQUE, 
															cvcCode VARCHAR(100) NOT NULL,
															date DATE NOT NULL,
															bank VARCHAR(100), 
															metaData TEXT, 
															initVector BYTEA);`)

	if err != nil {
		return fmt.Errorf("error while creating table BankCards: %w", err)
	}

	_, err = pg.dbConn.Exec(`CREATE TABLE IF NOT EXISTS Credentials (ID BIGSERIAL PRIMARY KEY,
																userID BIGINT REFERENCES Users (id) ON DELETE CASCADE,
																password VARCHAR(100) NOT NULL,
																application VARCHAR(100) NOT NULL UNIQUE, 
																metaData TEXT,
																initVector BYTEA);`)

	if err != nil {
		return fmt.Errorf("error while creating table Credentials: %w", err)
	}

	_, err = pg.dbConn.Exec(`CREATE TABLE IF NOT EXISTS UserFiles (ID BIGSERIAL PRIMARY KEY,
																userID BIGINT REFERENCES Users (id) ON DELETE CASCADE,
																fileName VARCHAR(100) NOT NULL UNIQUE,
																metaData TEXT);`)

	if err != nil {
		return fmt.Errorf("error while creating table UserFiles: %w", err)
	}
	return
}
