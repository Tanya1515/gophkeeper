package postgresql

import (
	"context"
	"fmt"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
	_ "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

func (pg *PostgreSQLConnection) UploadFile(ctx context.Context, fileName, metaData string) error {
	_, err := pg.dbConn.Exec("INSERT INTO UserFiles (userID, fileName, metaData) VALUES ($1, $2, $3)"+
		" ON CONFLICT (fileName) DO"+
		" UPDATE SET metaData = excluded.metaData WHERE UserFiles.fileName = excluded.fileName", ctx.Value(ut.IDKey), fileName, metaData)

	if err != nil {
		return fmt.Errorf("error while inserting/updating file %s : %s", fileName, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeleteFile(ctx context.Context, fileName string) (err error) {
	_, err = pg.dbConn.Exec("DELETE FROM UserFiles WHERE fileName=$1", fileName)

	return
}
