package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
	"github.com/google/uuid"
)

// DeleteFile - function for deleting file from cache
func (cache *SQLite) DeleteFile(fileName, userName string) (string, error) {
	var filePath string
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite to delete file %s: %s\n", fileName, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return "", fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statement, err := db.Prepare("DELETE from Files WHERE fileName=? AND userName=? RETURNING filePath")
	if err != nil {
		return "", fmt.Errorf("error while making request for deleting file with name %s: %w", fileName, err)
	}
	res := statement.QueryRowContext(ctxCache, fileName, userName)

	err = res.Scan(&filePath)
	if err != nil {
		return "", fmt.Errorf("error while scanning file path %s from delete request: %w", filePath, err)
	}
	return filePath, nil
}

// SaveFileOperation - function, that saves operations with file if server is unavailable.
func (cache *SQLite) SaveFileOperation(fileName, userName string, operation ut.Operation, fields []string, opTime, filePath string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection for adding data about operations with file %s: %w", fileName, err)
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

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM FileOperations WHERE fileName=$1 AND userName=$2", fileName, userName)
	if err != nil {
		return fmt.Errorf("error while getting all operations for file %s: %w", fileName, err)
	}
	for rows.Next() {
		rows.Scan(&operationName)
		if operationName == string(ut.Delete) {
			return nil
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error after scanning all operations for file %s: %w", fileName, err)
	}

	if operation == ut.Delete {
		ctxCacheDelete, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctxCacheDelete, "DELETE FROM FileOperations WHERE fileName=$1 AND userName=$2", fileName, userName)
		if err != nil {
			return fmt.Errorf("error while delete all operations for file %s: %w", fileName, err)
		}
	}

	statement, err := db.Prepare("INSERT INTO FileOperations (operationID, filePath, userName, fileName, operationName, operationUploadTime) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with file %s: %w", operation, fileName, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, filePath, userName, fileName, operation, opTime)
	if err != nil {
		return fmt.Errorf("error while adding new operation %s with file %s: %w", operation, fileName, err)
	}

	if operation == ut.Update {
		for _, field := range fields {
			fieldID := uuid.New()
			_, err = db.ExecContext(ctxCache, "INSERT INTO OperationsFileDiff (diffID, operationID, field) VALUES ($1,$2,$3)"+
				"ON CONFLICT (field) DO "+
				"UPDATE SET operationID = excluded.operationID, diffID = excluded.diffID", fieldID, operationID, field)
			if err != nil {
				return fmt.Errorf("error while inserting field %s for file %s for update operation: %w", field, fileName, err)
			}
		}

		_, err = db.ExecContext(ctxCache, "DELETE FROM FileOperations "+
			"WHERE NOT EXISTS "+
			"(SELECT 1 FROM OperationsFileDiff WHERE OperationsFileDiff.operationID = FileOperations.operationID) "+
			"AND fileName = $1 AND userName = $2 AND operationName = $3", fileName, userName, ut.Update)

		if err != nil {
			return fmt.Errorf("error while deleting all operations without fields to update: %w", err)
		}

	}

	return nil
}

// GetFile - function for getting file from application cache or return path to local file.
// The function is also checks if data is up to date.
func (cache *SQLite) GetFile(fileName, userName string) (metadata, pathFile string, exists bool, fileContent []byte, err error) {
	var lastUpdated, uploadTime string
	var pathToFile sql.NullString

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return "", "", false, nil, fmt.Errorf("error while openning connection to update data about application %s or add new one: %w", fileName, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return "", "", false, nil, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctxCache, "SELECT filePath, content, metadata, lastUpdated, uploadTime FROM Files WHERE fileName=$1 AND userName=$2", fileName, userName)

	err = row.Scan(&pathToFile, &fileContent, &metadata, &lastUpdated, &uploadTime)
	if err != nil {
		return "", "", false, nil, fmt.Errorf("error while scanning all data about file %s: %w", fileName, err)
	}

	exists = true
	pathFile = pathToFile.String

	return
}

// UploadFile - function for updating existing file or insert new one to application cache.
func (cache *SQLite) UploadFile(fileName, filePath, metadata, uploadTime, lastUpdated, userName string) error {

	content := make([]byte, 0)
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to update file %s or add new one: %w", fileName, err)
	}
	defer db.Close()
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	if filePath != "" {

		processedFile, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error while openning file %s: %w", filePath, err)
		}

		fileInfo, err := processedFile.Stat()
		if err != nil {
			return fmt.Errorf("error while getting info about file %s: %w", filePath, err)
		}

		if fileInfo.Size() <= 102400 {
			content = make([]byte, fileInfo.Size())

			_, err := processedFile.Read(content)
			if err != nil && err != io.EOF {
				return fmt.Errorf("error while reading data from file %s: %w", filePath, err)
			}
			os.Remove(filePath)
			filePath = ""

		} else {
			if !strings.Contains(filePath, "/tmp/") {
				destFile, err := os.Create("/tmp/" + fileName)
				if err != nil {
					return fmt.Errorf("error while creating destination file with path /tmp/%s: %w", fileName, err)
				}

				_, err = io.Copy(destFile, processedFile)
				if err != nil {
					return fmt.Errorf("error while copying data from source file %s to destination file /tmp/%s: %w", filePath, fileName, err)
				}

				os.Remove(filePath)
				filePath = "/tmp/" + fileName
			}

		}
		processedFile.Close()
	}

	_, err = db.Exec("INSERT INTO Files (fileName, filePath, content, metadata, lastUpdated, uploadTime, userName) VALUES ($1,$2,$3,$4,$5,$6,$7) "+
		"ON CONFLICT (fileName, userName) DO "+
		"UPDATE SET "+
		"filePath = CASE WHEN LENGTH(excluded.content) > 0 OR excluded.filePath <> '' THEN excluded.filePath ELSE filePath END, "+
		"content = CASE WHEN LENGTH(excluded.content) > 0 OR excluded.filePath <> '' THEN excluded.content ELSE content END, "+
		"metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE metadata END, "+
		"lastUpdated = CASE WHEN excluded.lastUpdated <> '' THEN excluded.lastUpdated ELSE Files.lastUpdated END,"+
		"uploadTime = CASE WHEN excluded.uploadTime <> '' THEN excluded.uploadTime ELSE Files.uploadTime END ", fileName, filePath, content, metadata, lastUpdated, uploadTime, userName)

	if err != nil {
		return fmt.Errorf("error while updating existing file %s or inserting new one with content: %w", filePath, err)
	}

	return nil
}

// GetAllFileWithOperation - function for getting list of files, with which some of operations were not performed.
func (cache *SQLite) GetAllFileWithOperation() (result map[string][]string, err error) {

	var userName, fileName string
	result = make(map[string][]string, 10)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return nil, fmt.Errorf("error while openning connection to get all file operations: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return nil, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT userName, fileName FROM FileOperations GROUP BY userName, fileName")
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

// GetFileOpearionsInfo - function for getting info about operations with files.
func (cache *SQLite) GetFileOpearionsInfo(user, fileName string) ([]ut.OperationInfo, error) {
	var field sql.NullString
	var operationFilesInfo, operationCardsInfoTemp ut.OperationInfo
	operationsFilesInfo := make([]ut.OperationInfo, 0)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return nil, fmt.Errorf("error while openning connection to get operations info with passwords: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return nil, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctxCache, "SELECT FileOperations.operationID, FileOperations.operationName, FileOperations.operationUploadTime, OperationsFileDiff.Field FROM FileOperations LEFT JOIN OperationsFileDiff ON "+
		"FileOperations.operationID = OperationsFileDiff.operationID WHERE FileOperations.userName=$1 AND FileOperations.fileName=$2 ", user, fileName)

	if err != nil {
		return nil, fmt.Errorf("error while getting operations info from SQLite: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&operationCardsInfoTemp.OperationID, &operationCardsInfoTemp.OperationName, &operationCardsInfoTemp.OperationTime, &field)
		if err != nil {
			return nil, fmt.Errorf("error while getting info about operations of file %s: %w", fileName, err)
		}

		if operationFilesInfo.OperationID == operationCardsInfoTemp.OperationID {
			if field.Valid {
				operationFilesInfo.Fields = append(operationFilesInfo.Fields, field.String)
			}
		} else {
			operationsFilesInfo = append(operationsFilesInfo, operationFilesInfo)
			operationFilesInfo = operationCardsInfoTemp
			if field.Valid {
				operationFilesInfo.Fields = append(operationFilesInfo.Fields, field.String)
			}
		}

	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while scanning info about operations of file %s: %w", fileName, err)
	}

	operationsFilesInfo = append(operationsFilesInfo, operationFilesInfo)

	return operationsFilesInfo, nil
}

// ClearFilesByDate - function for clear files info by date.
func (cache *SQLite) ClearFilesByDate(userName string) ([]string, error) {
	var filePath sql.NullString
	filesPath := make([]string, 0)

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return nil, fmt.Errorf("error while openning connection to clear data about files for user %s: %w", userName, err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return nil, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currantTime := (time.Now()).UTC()
	currantTimeStr := currantTime.Format(time.RFC3339)

	rows, err := db.QueryContext(ctxCache, "DELETE FROM Files WHERE DATETIME(uploadTime) < DATETIME($1, '-1 month') AND "+
		"DATETIME(lastUpdated) < DATETIME($1, '-14 days') AND NOT EXISTS "+
		"(SELECT 1 FROM FileOperations WHERE FileOperations.fileName = Files.fileName) RETURNING filePath", currantTimeStr)

	if err != nil {
		return nil, fmt.Errorf("error while removing all sensetive data for files for user %s: %w", userName, err)
	}

	for rows.Next() {
		err = rows.Scan(&filePath)
		if err != nil {
			return nil, fmt.Errorf("error while getting info about deleted files : %w", err)
		}
		if filePath.Valid {
			filesPath = append(filesPath, filePath.String)
		}
	}

	return filesPath, nil
}
