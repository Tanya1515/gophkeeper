package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	"github.com/google/uuid"
)

// DeleteFile - function for deleting file from cache
func (cache *SQLite) DeleteFile(fileName, userName string) error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite to delete file %s: %s\n", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statement, err := db.Prepare("DELETE from Files WHERE fileName=? AND userName=?")
	if err != nil {
		return fmt.Errorf("error while making request for deleting file with name %s: %w", fileName, err)
	}
	_, err = statement.ExecContext(ctxCache, fileName, userName)
	if err != nil {
		return fmt.Errorf("error while deleting file with name %s: %w", fileName, err)
	}
	return nil
}

// SaveFileOperation - function, that saves operations with file if server is unavailable.
func (cache *SQLite) SaveFileOperation(fileName, userName string, operation cs.Operation, fields []string, opTime, filePath string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection for adding data about operations with file %s: %w", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationID := uuid.New()

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM FileOperations WHERE fileName=$1 AND userName=$2", fileName, userName)
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

	statement, err := db.Prepare("INSERT INTO FileOperations (operationID, filePath, userName, fileName, operationName, operationUploadTime) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with password for application %s: %w", operation, fileName, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, filePath, userName, fileName, operation, opTime)
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
func (cache *SQLite) GetFile(fileName, userName string) (metadata, pathFile string, fileContent []byte, err error) {
	var lastUpdated, uploadTime string
	var accessCount int

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return "", "", nil, fmt.Errorf("error while openning connection to update data about application %s or add new one: %w", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctxCache, "SELECT filePath, content, metadata, lastUpdated, accessCount, uploadTime FROM Files WHERE fileName=$1 AND userName=$2", fileName, userName)

	err = row.Scan(&pathFile, &fileContent, &metadata, &lastUpdated, &accessCount, &uploadTime)
	if err != nil {
		return "", "", nil, fmt.Errorf("error while scanning all data about file %s: %w", fileName, err)
	}

	// проверка актуальности данных (lastUpdated + accessCount), увеличиваем accessCount

	return
}

// UploadFile - function for updating existing file or insert new one to application cache.
func (cache *SQLite) UploadFile(fileName, filePath, metadata, uploadTime, userName string) error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to update file %s or add new one: %w", fileName, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	processedFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error while openning file %s: %w", filePath, err)
	}

	fileInfo, err := processedFile.Stat()
	if err != nil {
		return fmt.Errorf("error while getting info about file %s: %w", filePath, err)
	}

	if fileInfo.Size() <= 102400 {
		content := make([]byte, fileInfo.Size())

		_, err = processedFile.Read(content)
		if err != nil {
			return fmt.Errorf("error while reading data from file %s: %w", filePath, err)
		}

		_, err = db.ExecContext(ctxCache, "INSERT INTO Files (fileName, content, metadata, lastUpdated, uploadTime, userName) VALUES ($1,$2,$3,$4,$5,$6) "+
			"ON CONFLICT (fileName, userName) DO"+
			"UPDATE SET "+
			"content = CASE WHEN excluded.content::bytea IS NOT NULL THEN excluded.content::bytea ELSE content END, "+
			"metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE metadata END, "+
			"lastUpdated = excluded.lastUpdated,"+
			"Files.accessCount = Files.accessCount + 1 WHERE Files.fileName = excluded.fileName AND Files.userName = excluded.userName", fileName, content, metadata, uploadTime, uploadTime, userName)

		if err != nil {
			return fmt.Errorf("error while updating existing file %s or inserting new one with content: %w", filePath, err)
		}
	} else {
		err = os.Rename(filePath, "/tmp/"+fileName)
		if err != nil {
			return fmt.Errorf("error while moving file %s to /tmp/%s: %w", filePath, fileName, err)
		}

		_, err = db.ExecContext(ctxCache, "INSERT INTO Files (fileName, filePath, metadata, lastUpdated, uploadTime, userName) VALUES ($1,$2,$3,$4,$5,$6) "+
			"ON CONFLICT (fileName, userName) DO"+
			"UPDATE SET "+
			"filePath = CASE WHEN excluded.filePath <> '' THEN excluded.filePath ELSE filePath END, "+
			"metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE metadata END, "+
			"lastUpdated = excluded.lastUpdated, "+
			"Files.accessCount = Files.accessCount + 1 WHERE Files.fileName = excluded.fileName AND Files.userName = excluded.userName", fileName, filePath, metadata, uploadTime, uploadTime, userName)

		if err != nil {
			return fmt.Errorf("error while updating existing file %s or inserting new one with content: %w", filePath, err)
		}
	}

	return nil
}

func (cache *SQLite) GetAllFileWithOperation() (result map[string][]string, err error) {

	var userName, fileName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return nil, fmt.Errorf("error while openning connection to get all file operations: %w", err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT userName, fileName FROM FileOperations GROUP BY (userName, fileName)")
	if err != nil {
		return nil, fmt.Errorf("error while getting all operations with files: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&userName, &fileName)
		if err != nil {
			return nil, fmt.Errorf("error while scanning operations with files: %w", err)
		}
		_, exists := result[userName]
		if !exists {
			result[userName] = make([]string, 0)
		}
		result[userName] = append(result[userName], fileName)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error after scanning all operations with files: %w", err)
	}
	return
}

func (cache *SQLite) GetFileOpearionsInfo(user, fileName string) (map[string]cs.OperationInfo, error) {
	var field string
	var operationID string
	var operationFilesInfo cs.OperationInfo
	operationsFilesInfo := make(map[string]cs.OperationInfo, 20)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return operationsFilesInfo, fmt.Errorf("error while openning connection to get operations info with passwords: %w", err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT operationID, operationName, operationUploadTime, Field FROM FileOperations JOIN OperationsFileDiff ON "+
		"FileOperations.operationID = OperationsFileDiff.operationID WHERE FileOperations.userName=$1 AND FileOperations.fileName=$2", user, fileName)

	if err != nil {
		return operationsFilesInfo, fmt.Errorf("error while getting operations info from SQLite: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&operationID, &operationFilesInfo.OperationName, &operationFilesInfo.OperationTime, &field)
		if err != nil {
			return operationsFilesInfo, fmt.Errorf("error while getting info about operations of file %s: %w", fileName, err)
		}
		opInfo, exists := operationsFilesInfo[operationID]
		if !exists {
			operationsFilesInfo[operationID] = operationFilesInfo
		}
		opInfo.Fields = append(opInfo.Fields, field)
	}

	err = rows.Err()
	if err != nil {
		return operationsFilesInfo, fmt.Errorf("error while scanning info about operations of file %s: %w", fileName, err)
	}

	return operationsFilesInfo, nil
}
