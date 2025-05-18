package postgresql

import (
	"context"
	"fmt"
)

func (pg *PostgreSQLConnection) LoginUser(ctx context.Context, login, password string) (string, error) {
	ok := false
	var email string
	row := pg.dbConn.QueryRowContext(ctx, `SELECT email, (password = crypt($1, password)) 
								AS password_match
								FROM Users
								WHERE login = $2;`, password, login)

	err := row.Scan(&email, &ok)
	if err != nil {
		return email, err
	}

	if !ok {
		return email, fmt.Errorf("user %s password is incorrect", login)
	}
	return email, nil
}

func (pg *PostgreSQLConnection) RegisterUser(ctx context.Context, login, password, email string) error {

	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO Users (userLogin, userPassword, userEmail) VALUES($1,crypt($2, gen_salt('xdes')),$3)", login, password, email)

	if err != nil {
		return fmt.Errorf("error while inserting user with login %s: %w", login, err)
	}

	return nil
}
