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
func (cache *SQLite) DeletePassword(application, userName string) error {

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to SQLite to delete password for application %s: %w", application, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return err
	}

	statement, err := db.Prepare("DELETE from Passwords WHERE application=? AND userName = ?")
	if err != nil {
		return fmt.Errorf("error while making request for deleting password for application %s: %w", application, err)
	}
	_, err = statement.ExecContext(ctxCache, application, userName)
	if err != nil {
		return fmt.Errorf("error while deleting password for application %s: %w", application, err)
	}
	return nil
}

// SavePasswordOperation - function for saving all operations with password for the application, if server is unavailable.
func (cache *SQLite) SavePasswordOperation(application, userName string, operation cs.Operation, fields []string, opTime string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to insert data about operations with passsword for application %s: %w", application, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationID := uuid.New()

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM PasswordOperations WHERE application=$1 AND userName = $2", application, userName)
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
		ctxCacheDelete, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctxCacheDelete, "DELETE FROM PasswordOperations WHERE application=$1 AND userName=$2", application, userName)
		if err != nil {
			return fmt.Errorf("error while delete all operations for application %s: %w", application, err)
		}
	}

	statement, err := db.Prepare("INSERT INTO PasswordOperations (operationID, userName, application, operationName, operationUploadTime) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with password for application %s: %w", operation, application, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, userName, application, operation, opTime)
	if err != nil {
		return fmt.Errorf("error while adding new operation %s with password for application %s: %w", operation, application, err)
	}

	if operation == cs.Update {
		for _, field := range fields {
			fieldID := uuid.New()
			_, err = db.ExecContext(ctxCache, "INSERT INTO OperationPasswordDiff (diffID, operationID, field) VALUES ($1,$2,$3)"+
				"ON CONFLICT (field) DO "+
				"UPDATE SET operationID = excluded.operationID, diffID = excluded.diffID ", fieldID, operationID, field)
			if err != nil {
				return fmt.Errorf("error while inserting field %s for application %s for update operation: %w", field, application, err)
			}
		}

		_, err = db.ExecContext(ctxCache,
			"DELETE FROM PasswordOperations "+
				"WHERE NOT EXISTS "+
				"(SELECT 1 FROM OperationPasswordDiff WHERE OperationPasswordDiff.operationID = PasswordOperations.operationID) "+
				"AND PasswordOperations.application = $1 AND PasswordOperations.userName = $2 AND PasswordOperations.operationName = $3",
			application, userName, cs.Update)
		if err != nil {
			return fmt.Errorf("error while deleting all operations without fields to update: %w", err)
		}

	}

	return nil
}

// GetPassword - function for getting password and data about it from application cache.
// The function also checks, if the data is up to date.
func (cache *SQLite) GetPassword(application, userName string) (password, uploadTime, metadata string, initVector []byte, exists bool, err error) {
	var lastUpdated string
	var accessCount int

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return "", "", "", nil, false, fmt.Errorf("error while openning connection to get data about application %s: %w", application, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return "", "", "", nil, false, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctxCache, "SELECT password, metadata, lastUpdated, initVector, accessCount, uploadTime FROM Passwords WHERE application=$1 AND userName=$2", application, userName)

	err = row.Scan(&password, &metadata, &lastUpdated, &initVector, &accessCount, &uploadTime)
	if err != nil {
		return "", "", "", nil, false, fmt.Errorf("error while getting data for password of application %s: %w", application, err)
	}

	exists = true
	// проверка актуальности данных (lastUpdated + accessCount), увеличиваем accessCount

	// логика по рашисфровке пароля
	return
}

// UploadPassword - function for updating existing password and data about it or
// inserting new one.
func (cache *SQLite) UploadPassword(application, password, metadata, uploadTime, userName string, initVector []byte) (err error) {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to update data about application %s or add new one: %w", application, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctxCache, `
        INSERT INTO Passwords (application, password, metadata, lastUpdated, uploadTime, userName, initVector)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (application, userName) DO UPDATE SET
            password = CASE WHEN excluded.password <> '' THEN excluded.password ELSE Passwords.password END,
            metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE Passwords.metadata END,
			initVector = CASE WHEN excluded.initVector <> '' THEN excluded.initVector ELSE Passwords.initVector END,
            lastUpdated = excluded.lastUpdated,
            accessCount = Passwords.accessCount + 1
    `, application, password, metadata, uploadTime, uploadTime, userName, initVector)
	if err != nil {
		return fmt.Errorf("error while updating existing application %s or inserting new one: %w", application, err)
	}

	return
}

func (cache *SQLite) GetAllPasswordWithOperation() (result map[string][]string, err error) {
	var userName, application string
	result = make(map[string][]string, 10)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return nil, fmt.Errorf("error while openning connection to get all operations with passwords: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return nil, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT userName, application FROM PasswordOperations GROUP BY userName, application")
	if err != nil {
		return nil, fmt.Errorf("error while getting all operations with passwords: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&userName, &application)
		if err != nil {
			return nil, fmt.Errorf("error while scanning operations with passwords: %w", err)
		}
		_, exists := result[userName]
		if !exists {
			result[userName] = make([]string, 0)
		}
		result[userName] = append(result[userName], application)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error after scanning all operations with passwords: %w", err)
	}
	return
}

func (cache *SQLite) GetPasswordOperationsInfo(user, application string) (map[string]cs.OperationInfo, error) {

	var field sql.NullString
	var operationID string

	var operationPasswordsInfo cs.OperationInfo
	operationsPasswordsInfo := make(map[string]cs.OperationInfo, 20)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return operationsPasswordsInfo, fmt.Errorf("error while openning connection to get operations info with passwords: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return operationsPasswordsInfo, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT PasswordOperations.operationID, PasswordOperations.operationName, PasswordOperations.operationUploadTime, OperationPasswordDiff.Field FROM PasswordOperations LEFT JOIN OperationPasswordDiff ON "+
		"PasswordOperations.operationID = OperationPasswordDiff.operationID WHERE PasswordOperations.userName=$1 AND PasswordOperations.application=$2 ORDER BY PasswordOperations.operationUploadTime", user, application)
	if err != nil {
		return operationsPasswordsInfo, fmt.Errorf("error while getting all data about password operations: %w", err)
	}
	for rows.Next() {
		err = rows.Scan(&operationID, &operationPasswordsInfo.OperationName, &operationPasswordsInfo.OperationTime, &field)
		if err != nil {
			return operationsPasswordsInfo, fmt.Errorf("error while getting info about operations of application %s: %w", application, err)
		}
		opInfo, exists := operationsPasswordsInfo[operationID]
		if !exists {
			operationsPasswordsInfo[operationID] = operationPasswordsInfo
		}
		if field.Valid {
			opInfo.Fields = append(opInfo.Fields, field.String)
		}

	}

	err = rows.Err()
	if err != nil {
		return operationsPasswordsInfo, fmt.Errorf("error while scanning password data: %w", err)
	}

	return operationsPasswordsInfo, nil

}
