package postgresql

import (
	"context"
	"fmt"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

func (pg *PostgreSQLConnection) UploadPassword(ctx context.Context, password, app, md string) error {

	_, err := pg.dbConn.Exec("INSERT INTO Credentials (userID, password, application, metaData) VALUES ($1,crypt($2, gen_salt('xdes')),$3,$4)"+
		" ON CONFLICT (application) DO"+
		" UPDATE SET password = excluded.password, metaData = excluded.metaData WHERE Credentials.application = excluded.application", ctx.Value(ut.IDKey), password, app, md)

	if err != nil {
		return fmt.Errorf("error while inserting/updating password for application %s : %w", app, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeletePassword(ctx context.Context, application string) (err error) {

	_, err = pg.dbConn.Exec("DELETE FROM Credentials WHERE application=$1", application)

	return 
}

func (pg *PostgreSQLConnection) GetPassword(ctx context.Context, application string) (passwordApp pb.PasswordMessage, err error) {
	
	row := pg.dbConn.QueryRowContext(ctx, "SELECT password, metaData FROM Credentials WHERE application=$1", application)

	err = row.Scan(&passwordApp.Password, &passwordApp.MetaData)

	return
}
