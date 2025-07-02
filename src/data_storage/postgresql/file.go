package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// DeleteFile - function for deleting file info for current user.
func (pg *PostgreSQLConnection) DeleteFile(ctx context.Context, fileName string) (err error) {
	_, err = pg.dbConn.Exec("DELETE FROM UserFiles WHERE fileName=$1 AND userID=$2", fileName, ctx.Value(ut.IDKey))

	return
}

// UpdateFile - function for updating file info for current user.
func (pg *PostgreSQLConnection) UploadFile(ctx context.Context, fileName, metaData string, updatedAt time.Time) (bool, error) {

	var fileNameGet, metaDataGet string
	var updatedAtGet time.Time

	tx, err := pg.dbConn.Begin()

	if err != nil {
		return false, fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT fileName, metaData, updatedAt FROM UserFiles WHERE fileName=$1 AND userID=$2 FOR UPDATE", fileName, ctx.Value(ut.IDKey))

	err = row.Scan(&fileNameGet, &metaDataGet, &updatedAtGet)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		tx.Rollback()
		return false, fmt.Errorf("error while getting sensetive data for file %s: %w", fileName, err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO UserFiles (userID, fileName, metaData, updatedAt) VALUES ($1, $2, $3, $4)"+
		" ON CONFLICT (fileName, userID) DO UPDATE SET metaData= $3,"+
		" updatedAt = $4 "+
		" WHERE UserFiles.updatedAt <= excluded.updatedAt", ctx.Value(ut.IDKey), fileName, metaData, updatedAt)

	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("error while updating file %s : %s", fileName, err)
	}

	err = tx.Commit()

	if err != nil {
		return false, fmt.Errorf("error while closing transaction: %w", err)
	}

	if updatedAt.Before(updatedAtGet) {
		return false, nil
	}

	return true, nil
}
