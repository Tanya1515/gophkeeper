package postgresql

import (
	"context"
	"fmt"
)

func (pg *PostgreSQLConnection) LoginUser(ctx context.Context, login, password string) (string, error) {
	ok := false
	var email string
	row := pg.dbConn.QueryRowContext(ctx, `SELECT userEmail, (userPassword = crypt($1, userPassword)) 
								AS password_match
								FROM Users
								WHERE userLogin = $2;`, password, login)

	err := row.Scan(&email, &ok)
	if err != nil {
		return email, err
	}

	if !ok {
		return email, fmt.Errorf("user %s password is incorrect", login)
	}
	return email, nil
}

func (pg *PostgreSQLConnection) RegisterUser(ctx context.Context, login, password, email string) (string, error) {
	var userID string
	row := pg.dbConn.QueryRowContext(ctx, "INSERT INTO Users (userLogin, userPassword, userEmail) VALUES($1,crypt($2, gen_salt('xdes')),$3) RETURNING ID", login, password, email)

	err := row.Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("error while inserting user with login %s: %w", login, err)
	}

	return userID, nil
}

func (pg *PostgreSQLConnection) CheckUserJWT(ctx context.Context, userLogin string) (userID string, err error) {

	row := pg.dbConn.QueryRowContext(ctx, "SELECT ID FROM Users WHERE userLogin=$1", userLogin)

	err = row.Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("error while getting user %s ID: %w", userLogin, err)
	}

	return
}
