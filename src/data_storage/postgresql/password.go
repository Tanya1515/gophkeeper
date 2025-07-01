package postgresql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// UploadPassword - function for uploading new password and data for it for current user.
func (pg *PostgreSQLConnection) UploadPassword(ctx context.Context, password, app, md string, updatedAt time.Time, initVector []byte) error {

	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO Credentials (userID, password, application, metaData, initVector, updatedAt) VALUES ($1,$2,$3,$4,$5,$6)"+
		" ON CONFLICT (application) DO"+
		" UPDATE SET password = excluded.password, metaData = excluded.metaData, initVector = excluded.initVector, updatedAt = excluded.updatedAt WHERE updatedAt =< excluded.updatedAt ", ctx.Value(ut.IDKey), password, app, md, initVector, updatedAt)

	if err != nil {
		return fmt.Errorf("error while inserting/updating password for application %s : %w", app, err)
	}

	return nil
}

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
func (pg *PostgreSQLConnection) UpdatePassword(ctx context.Context, password, app, md string, updatedAt time.Time, initVector []byte) error {

	_, err := pg.dbConn.ExecContext(ctx,
		"UPDATE Credentials SET "+
			"password = CASE WHEN $1 <> '' THEN $1 ELSE password END, "+
			"metaData = CASE WHEN $2 <> '' THEN $2 ELSE metaData END, "+
			"initVector = CASE WHEN $3::bytea IS NOT NULL THEN $3::bytea ELSE initVector END, "+
			"updatedAt = $4 "+
			"WHERE application=$5 AND userID=$6 AND updatedAt <= $4", password, md, initVector, updatedAt, app, ctx.Value(ut.IDKey))

	if err != nil {
		return fmt.Errorf("error while updating password for application %s : %w", app, err)
	}

	return nil
}
