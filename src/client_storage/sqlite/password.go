package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
)

// DeletePassword - function, that deletes application, password and other data from application cache.
func (cache *SQLite) DeletePassword(application string, userID int) error {

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to SQLite to delete password for application %s: %w", application, err)
	}

	defer db.Close()

	statement, err := db.Prepare("DELETE from Passwords WHERE application=? AND userID")
	if err != nil {
		return fmt.Errorf("error while making request for deleting password for application %s: %w", application, err)
	}
	_, err = statement.ExecContext(ctxCache, application, userID)
	if err != nil {
		return fmt.Errorf("error while deleting password for application %s: %w", application, err)
	}
	return nil
}

// SavePasswordOperation - function for saving all operations with password for the application, if server is unavailable.
func (cache *SQLite) SavePasswordOperation(application string, operation cs.Operation, fields []string, opTime string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to insert data about operations with passsword for application %s: %w", application, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationID := uuid.New()

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM PasswordOperations WHERE application=$1", application)
	if err != nil {
		return fmt.Errorf("error while getting all operations for application %s: %w", application, err)
	}
	for rows.Next() {
		rows.Scan(&operationName)
		if operationName == string(cs.Delete) {
			return nil
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error after scanning all operations for application %s: %w", application, err)
	}

	if operation == cs.Delete {
		statement, err := db.Prepare("DELETE from PasswordOperations WHERE application=?")
		if err != nil {
			return fmt.Errorf("error while making request for deleting all operations with password for application %s: %w", application, err)
		}
		_, err = statement.ExecContext(ctxCache, application)
		if err != nil {
			return fmt.Errorf("error while deleting all operations with password for application %s: %w", application, err)
		}
	}

	statement, err := db.Prepare("INSERT INTO PasswordOperations (operationID, application, operationName, operationUploadTime) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with password for application %s: %w", operation, application, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, application, operation, opTime)
	if err != nil {
		return fmt.Errorf("error while adding new operation %s with password for application %s: %w", operation, application, err)
	}

	if operation == cs.Update {
		for _, field := range fields {
			fieldID := uuid.New()
			_, err = db.ExecContext(ctxCache, "INSERT INTO OperationPasswordDiff (diffID, operationID, field) VALUES ($1,$2,$3,)"+
				"ON CONFLICT (field) DO "+
				"UPDATE SET operationID = excluded.operationID, diffID = excluded.diffID WHERE PasswordOperations.field = excluded.field", fieldID, operationID, field)
			if err != nil {
				return fmt.Errorf("error while inserting field %s for application %s for update operation: %w", field, application, err)
			}
		}

		_, err = db.ExecContext(ctxCache, "DELETE FROM PasswordOperations "+
			"WHERE NOT EXISTS "+
			"(SELECT 1 FROM fields WHERE OperationPasswordDiff.operationID = PasswordOperations.operationID)")

		if err != nil {
			return fmt.Errorf("error while deleting all operations without fields to update: %w", err)
		}

	}

	return nil
}

// GetPassword - function for getting password and data about it from application cache.
// The function also checks, if the data is up to date.
func (cache *SQLite) GetPassword(application string, userID int) (password string, metadata string, err error) {
	var lastUpdated, uploadTime string
	var accessCount int

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return "", "", fmt.Errorf("error while openning connection to get data about application %s: %w", application, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctxCache, "SELECT password, metadata, lastUpdated, accessCount, uploadTime WHERE application=$1 AND userID=$2", application, userID)

	err = row.Scan(&password, &metadata, &lastUpdated, &accessCount, &uploadTime)
	if err != nil {
		return "", "", fmt.Errorf("error while getting data for password of application %s: %w", application, err)
	}
	// проверка актуальности данных (lastUpdated + accessCount), увеличиваем accessCount

	// логика по рашисфровке пароля
	return
}

// UploadPassword - function for updating existing password and data about it or
// inserting new one.
func (cache *SQLite) UploadPassword(application, password, metadata, uploadTime string, userID int) (err error) {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to update data about application %s or add new one: %w", application, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctxCache, "INSERT INTO Passwords (application, password, metadata, lastUpdated, uploadTime, userID) VALUES ($1,$2,$3,$4,$5,$6) "+
		"ON CONFLICT (application, userID) DO "+
		"UPDATE SET "+
		"password = CASE WHEN excluded.password <> '' THEN excluded.password ELSE password END, "+
		"metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE metadata END, "+
		"lastUpdated = excluded.lastUpdated, "+
		"Passwords.accessCount = Passwords.accessCount + 1 WHERE Passwords.application = excluded.application AND Passwords.userID = excluded.userID", application, password, metadata, uploadTime, uploadTime, userID)

	if err != nil {
		return fmt.Errorf("error while updating existing application %s or inserting new one: %w", application, err)
	}

	return
}
