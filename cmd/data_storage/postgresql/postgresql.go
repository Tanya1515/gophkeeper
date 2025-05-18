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
	return
}
