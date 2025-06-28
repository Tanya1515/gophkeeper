package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (cache *SQLite) DeleteOperaionByID(operationID, dataType string) error {

	var tableName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to delete data related to the operation %s: %w", operationID, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if dataType == "password" {
		tableName = "Password"
	} else if dataType == "file" {
		tableName = "File"
	} else if dataType == "card" {
		tableName = "Card"
	}

	_, err = db.ExecContext(ctxCache, "DELETE FROM "+tableName+"Operations WHERE operationID = $1", operationID)
	if err != nil {
		return fmt.Errorf("error while deleting operation %s: %w", operationID, err)
	}

	return nil

}
