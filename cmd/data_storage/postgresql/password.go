package postgresql

import (
	"context"
	"fmt"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (pg *PostgreSQLConnection) UploadPassword(ctx context.Context, password, app, md string, initVector []byte) error {

	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO Credentials (userID, password, application, metaData, initVector) VALUES ($1,$2,$3,$4,$5)"+
		" ON CONFLICT (application) DO"+
		" UPDATE SET password = excluded.password, metaData = excluded.metaData, initVector = excluded.initVector WHERE Credentials.application = excluded.application", ctx.Value(ut.IDKey), password, app, md, initVector)

	if err != nil {
		return fmt.Errorf("error while inserting/updating password for application %s : %w", app, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeletePassword(ctx context.Context, application string) (err error) {

	_, err = pg.dbConn.Exec("DELETE FROM Credentials WHERE application=$1", application)

	return
}

func (pg *PostgreSQLConnection) GetPassword(ctx context.Context, application string) (passwordApp pb.PasswordMessage, initVector []byte, err error) {

	row := pg.dbConn.QueryRowContext(ctx, "SELECT password, metaData, initVector FROM Credentials WHERE application=$1", application)

	err = row.Scan(&passwordApp.Password, &passwordApp.MetaData, &initVector)

	return
}

func (pg *PostgreSQLConnection) UpdatePassword(ctx context.Context, password, app, md string, initVector []byte) error {

	_, err := pg.dbConn.ExecContext(ctx,
		"UPDATE Credentials SET "+
			"password = CASE WHEN $1 <> '' THEN $1 ELSE password END, "+
			"metaData = CASE WHEN $2 <> '' THEN $2 ELSE metaData END, "+
			"initVector = CASE WHEN $3::bytea IS NOT NULL THEN $3::bytea ELSE initVector END "+
			"WHERE application=$4;", password, md, initVector, app)

	if err != nil {
		return fmt.Errorf("error while updating password for application %s : %w", app, err)
	}

	return nil
}
