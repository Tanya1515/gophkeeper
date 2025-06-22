package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	"github.com/google/uuid"
)

// DeleteFile - function for deleting file from cache 
func (cache *SQLite) DeleteFile(fileName string) error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite to delete file %s: %s\n", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statement, err := db.Prepare("DELETE from Files WHERE fileName=?")
	if err != nil {
		return fmt.Errorf("error while making request for deleting file with name %s: %w", fileName, err)
	}
	_, err = statement.ExecContext(ctxCache, fileName)
	if err != nil {
		return fmt.Errorf("error while deleting file with name %s: %w", fileName, err)
	}
	return nil
}

// SaveFileOperation - function, that saves operations with file if server is unavailable.
func (cache *SQLite) SaveFileOperation(fileName string, operation cs.Operation, fields []string, opTime string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection for adding data about operations with file %s: %w", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationID := uuid.New()

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM FileOperations WHERE fileName=$1", fileName)
	if err != nil {
		return fmt.Errorf("error while getting all operations for file %s: %w", fileName, err)
	}
	for rows.Next() {
		rows.Scan(&operationName)
		if operationName == string(cs.Delete) {
			return nil
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error after scanning all operations for file %s: %w", fileName, err)
	}

	if operation == cs.Delete {
		statement, err := db.Prepare("DELETE from FileOperations WHERE fileName=?")
		if err != nil {
			return fmt.Errorf("error while making request for deleting all operations with file %s: %w", fileName, err)
		}
		_, err = statement.ExecContext(ctxCache, fileName)
		if err != nil {
			return fmt.Errorf("error while deleting all operations with file %s: %w", fileName, err)
		}
	}

	statement, err := db.Prepare("INSERT INTO FileOperations (operationID, fileName, operationName, operationUploadTime) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with password for application %s: %w", operation, fileName, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, fileName, operation, opTime)
	if err != nil {
		return fmt.Errorf("error while adding new operation %s with file %s: %w", operation, fileName, err)
	}

	if operation == cs.Update {
		for _, field := range fields {
			fieldID := uuid.New()
			_, err = db.ExecContext(ctxCache, "INSERT INTO OperationsFileDiff (diffID, operationID, field) VALUES ($1,$2,$3,)"+
				"ON CONFLICT (field) DO "+
				"UPDATE SET operationID = excluded.operationID, diffID = excluded.diffID WHERE PasswordOperations.field = excluded.field", fieldID, operationID, field)
			if err != nil {
				return fmt.Errorf("error while inserting field %s for file %s for update operation: %w", field, fileName, err)
			}
		}

		_, err = db.ExecContext(ctxCache, "DELETE FROM OperationsFileDiff "+
			"WHERE NOT EXISTS "+
			"(SELECT 1 FROM fields WHERE OperationFileDiff.operationID = FileOperations.operationID)")

		if err != nil {
			return fmt.Errorf("error while deleting all operations without fields to update: %w", err)
		}

	}

	return nil
}
// GetFile - function for getting file from application cache or return path to local file.
// The function is also checks if data is up to date.
func (cache *SQLite) GetFile(fileName string) error {
	return nil
}

// UploadFile - function for updating existing file or insert new one to application cache.
func (cache *SQLite) UploadFile(fileName, filePath, metadata string) error {
	return nil
}
