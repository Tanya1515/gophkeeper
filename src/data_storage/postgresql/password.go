package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// DeletePassword - function for deleting password data for current user.
func (pg *PostgreSQLConnection) DeletePassword(ctx context.Context, application string) (err error) {

	_, err = pg.dbConn.Exec("DELETE FROM Credentials WHERE application=$1 AND userID=$2", application, ctx.Value(ut.IDKey))

	return
}

// GetPassword - function for getting password data for current user.
func (pg *PostgreSQLConnection) GetPassword(ctx context.Context, application string) (passwordApp pb.PasswordMessage, initVector []byte, err error) {

	row := pg.dbConn.QueryRowContext(ctx, "SELECT password, metaData, initVector FROM Credentials WHERE application=$1 AND userID=$2", application, ctx.Value(ut.IDKey))

	err = row.Scan(&passwordApp.Password, &passwordApp.MetaData, &initVector)

	return
}

// UpdatePassword - function for updating password data for current user.
func (pg *PostgreSQLConnection) UploadPassword(ctx context.Context, password, app, md string, updatedAt time.Time, initVector []byte) error {

	var passwordGet, metaDataGet string
	var initVectorGet []byte
	var updatedAtGet time.Time

	tx, err := pg.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT password, metaData, initVector, updatedAt FROM Credentials WHERE application=$1 AND userID=$2 FOR UPDATE", app, ctx.Value(ut.IDKey))

	err = row.Scan(&passwordGet, &metaDataGet, &initVectorGet, &updatedAtGet)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		tx.Rollback()
		return fmt.Errorf("error while credentials for application %s: %w", app, err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO Credentials (userID, password, application, metaData, initVector, updatedAt) VALUES ($1,$2,$3,$4,$5,$6)"+
		"ON CONFLICT (application, userID) DO UPDATE SET "+
		"password = CASE WHEN $2 <> '' THEN $2 ELSE Credentials.password END, "+
		"metaData = CASE WHEN $4 <> '' THEN $4 ELSE Credentials.metaData END, "+
		"initVector = CASE WHEN $5::bytea IS NOT NULL THEN $5::bytea ELSE Credentials.initVector END, "+
		"updatedAt = $6 "+
		"WHERE Credentials.updatedAt <= excluded.updatedAt", ctx.Value(ut.IDKey), password, app, md, initVector, updatedAt)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while updating password for application %s : %w", app, err)
	}

	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}

	return nil
}
