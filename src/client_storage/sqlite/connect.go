// SQLite - package, that conatains functions for CRUD operations
// with users' sensetive data in SQLIte.
package sqlite

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct{}

// Connect - function for connecting to SQLite3 and creating
// tables for saving data.
func (cache *SQLite) Connect() error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite: %s\n", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return err
	}

	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS Users (userName TEXT PRIMARY KEY, cache_miss INTEGER)")
	if err != nil {
		log.Println("Error while creating table for users: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Passwords (application TEXT, " +
		"userName TEXT, " +
		"password TEXT, " +
		"metadata TEXT, " +
		"accessCount INTEGER DEFAULT 0, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT, " +
		"PRIMARY KEY (application, userName), " +
		"FOREIGN KEY(userName) REFERENCES Users(userName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for passwords: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Cards (cardNumber TEXT, " +
		"userName TEXT, " +
		"cvcCode TEXT, " +
		"date TEXT, " +
		"bank TEXT, " +
		"metadata TEXT, " +
		"accessCount INTEGER DEFAULT 0, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT," +
		"PRIMARY KEY (userName, cardNumber), " +
		"FOREIGN KEY(userName) REFERENCES Users(userName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for cards: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Files (fileName TEXT, " +
		"userName TEXT, " +
		"filePath TEXT, " +
		"content BLOB, " +
		"metadata TEXT, " +
		"accessCount INTEGER DEFAULT 0, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT, " +
		"PRIMARY KEY (fileName, userName), " +
		"FOREIGN KEY(userName) REFERENCES Users(userName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error whcdile creating table for files: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare(`
    CREATE TABLE IF NOT EXISTS FileOperations (
        operationID TEXT PRIMARY KEY,
        userName TEXT,
        fileName TEXT,
        operationName TEXT,
        operationUploadTime TEXT,
        filePath TEXT,
        FOREIGN KEY (userName) REFERENCES Users(userName) ON DELETE CASCADE
    )
`)
	if err != nil {
		log.Println("Error while creating table for operations with files: ", err)
		return err
	}
	statement.Exec()
	// по идее надо раздрабить - чтобы не было дублирующих полей (ID меньше весят, чем строки).
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS PasswordOperations (operationID TEXT PRIMARY KEY, " +
		"userName TEXT, " +
		"application TEXT, " +
		"operationName TEXT, " +
		"operationUploadTime TEXT, " +
		"FOREIGN KEY (userName) REFERENCES Users(userName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for operations with passwords: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS CardOperations (operationID TEXT PRIMARY KEY, " +
		"userName TEXT, " +
		"cardNumber TEXT, " +
		"operationName TEXT, " +
		"operationUploadTime TEXT, " +
		"FOREIGN KEY(userName) REFERENCES Users(userName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for operations with cards: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS OperationBankCardDiff (diffID TEXT PRIMARY KEY, " +
		"operationID TEXT, " +
		"Field TEXT UNIQUE, " +
		"FOREIGN KEY(operationID) REFERENCES CardOperations(operationID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for saving differnces for operations with bank card: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS OperationPasswordDiff (diffID TEXT PRIMARY KEY, " +
		"operationID TEXT, " +
		"field TEXT UNIQUE, " +
		"FOREIGN KEY(operationID) REFERENCES PasswordOperations(operationID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for saving differnces for operations with bank card: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS OperationsFileDiff (diffID TEXT PRIMARY KEY, " +
		"operationID TEXT, " +
		"Field TEXT UNIQUE, " +
		"FOREIGN KEY(operationID) REFERENCES FileOperations(operationID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for saving differnces for operations with bank card: ", err)
		return err
	}

	statement.Exec()

	return nil
}
