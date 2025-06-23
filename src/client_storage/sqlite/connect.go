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

	db, err := sql.Open("sqlite3", "./data_cahce.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite: %s\n", err)
	}

	defer db.Close()

	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS Users (userID INTEGER PRIMARY KEY, userName TEXT, cache_miss INTEGER)")
	if err != nil {
		log.Println("Error while creating table for users: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Passwords (application TEXT, " +
		"userID INTEGER, " +
		"password TEXT, " +
		"metadata TEXT, " +
		"accessCount INTEGER, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT, " +
		"PRIMARY KEY (application, userID)" +
		"FOREIGN KEY(userID) REFERENCES Users(userID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for passwords: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Cards (cardNumber TEXT, " +
		"userID INTEGER, " +
		"cvcCode TEXT, " +
		"date TEXT, " +
		"bank TEXT, " +
		"metadata TEXT, " +
		"accessCount INTEGER, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT," +
		"PRIMARY KEY (userID, cardNumber) " +
		"FOREIGN KEY(userID) REFERENCES Users(userID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for cards: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS Files (fileName TEXT, " +
		"filePath TEXT, " +
		"content BLOB, " +
		"metadata TEXT, " +
		"accessCount INTEGER, " +
		"lastUpdated TEXT, " +
		"uploadTime TEXT, " +
		"PRIMARY KEY (fileName, userID)" +
		"FOREIGN KEY(userID) REFERENCES Users(userID) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for files: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS FileOperations (operationID TEXT PRIMARY KEY, " +
		"fileName TEXT, " +
		"operationName TEXT, " +
		"operationUploadTime TEXT, " +
		"FOREIGN KEY(fileName) REFERENCES Files(fileName) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for operations with files: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS PasswordOperations (operationID TEXT PRIMARY KEY, " +
		"application TEXT, " +
		"operationName TEXT, " +
		"operationUploadTime TEXT, " +
		"FOREIGN KEY(application) REFERENCES Passwords(application) ON DELETE CASCADE)")
	if err != nil {
		log.Println("Error while creating table for operations with passwords: ", err)
		return err
	}

	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS CardOperations (operationID TEXT PRIMARY KEY, " +
		"cardNumber TEXT, " +
		"operationName TEXT, " +
		"operationUploadTime TEXT, " +
		"FOREIGN KEY(cardNumber) REFERENCES Cards(cardNumber) ON DELETE CASCADE)")
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
