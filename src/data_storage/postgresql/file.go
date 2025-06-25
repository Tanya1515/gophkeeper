package postgresql

import (
	"context"
	"fmt"

	_ "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// UploadFile - function for uploading file info for current user.
func (pg *PostgreSQLConnection) UploadFile(ctx context.Context, fileName, metaData string) error {
	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO UserFiles (userID, fileName, metaData) VALUES ($1, $2, $3)"+
		" ON CONFLICT (fileName) DO"+
		" UPDATE SET metaData = excluded.metaData WHERE UserFiles.fileName = excluded.fileName AND UserFiles.userID = excluded.userID", ctx.Value(ut.IDKey), fileName, metaData)

	if err != nil {
		return fmt.Errorf("error while inserting/updating file %s : %s", fileName, err)
	}

	return nil
}

// DeleteFile - function for deleting file info for current user.
func (pg *PostgreSQLConnection) DeleteFile(ctx context.Context, fileName string) (err error) {
	_, err = pg.dbConn.Exec("DELETE FROM UserFiles WHERE fileName=$1 AND userID=$2", fileName, ctx.Value(ut.IDKey))

	return
}

// UpdateFile - function for updating file info for current user.
func (pg *PostgreSQLConnection) UpdateFile(ctx context.Context, fileName, metaData string) error {
	_, err := pg.dbConn.ExecContext(ctx,
		"UPDATE UserFiles SET metaData= $1"+
			" WHERE fileName=$2 AND userID=$3", metaData, fileName, ctx.Value(ut.IDKey))

	if err != nil {
		return fmt.Errorf("error while updating file %s : %s", fileName, err)
	}

	return nil
}
