package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	"github.com/google/uuid"
)

// DeleteBankCard - function for deleting bank card from application cache.
func (cache *SQLite) DeleteBankCard(cardNumber string) error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite to delete bank card %s: %s\n", cardNumber, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statement, err := db.Prepare("DELETE from Cards WHERE cardNumber=?")
	if err != nil {
		return fmt.Errorf("error while making request for deleting bank card credentials with number %s: %w", cardNumber, err)
	}
	_, err = statement.ExecContext(ctxCache, cardNumber)
	if err != nil {
		return fmt.Errorf("error while deleting bank card credentials with number %s: %w", cardNumber, err)
	}

	return nil
}

// SaveCardOperation - function for saving all new operations for bank card to application cache,
// because server is unavailable. 
func (cache *SQLite) SaveCardOperation(cardNumber string, operation cs.Operation, fields []string, opTime string) error {
	var operationName string

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to SQLite to update data about operations on bank card %s: %w", cardNumber, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationID := uuid.New()

	rows, err := db.QueryContext(ctxCache, "SELECT operationName FROM CardOperations WHERE cardNumber=$1", cardNumber)
	if err != nil {
		return fmt.Errorf("error while getting all operations for bank card with number %s: %w", cardNumber, err)
	}
	for rows.Next() {
		rows.Scan(&operationName)
		if operationName == string(cs.Delete) {
			return nil
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error after scanning all operations for bank card with number %s: %w", cardNumber, err)
	}

	if operation == cs.Delete {
		statement, err := db.Prepare("DELETE from CardOperations WHERE cardNumber=?")
		if err != nil {
			return fmt.Errorf("error while making request for deleting all operations with bank card %s: %w", cardNumber, err)
		}
		_, err = statement.ExecContext(ctxCache, cardNumber)
		if err != nil {
			return fmt.Errorf("error while deleting all operations with bank card %s: %w", cardNumber, err)
		}
	}

	statement, err := db.Prepare("INSERT INTO CardOperations (operationID, cardNumber, operationName, operationUploadTime) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating request for adding new operation %s with bank card %s: %w", operation, cardNumber, err)
	}
	_, err = statement.ExecContext(ctxCache, operationID, cardNumber, operation, opTime)
	if err != nil {
		return fmt.Errorf("error while adding new operation %s with bank card %s: %w", operation, cardNumber, err)
	}

	if operation == cs.Update {
		for _, field := range fields {
			fieldID := uuid.New()
			_, err = db.ExecContext(ctxCache, "INSERT INTO OperationBankCardDiff (diffID, operationID, field) VALUES ($1,$2,$3,)"+
				"ON CONFLICT (field) DO "+
				"UPDATE SET operationID = excluded.operationID, diffID = excluded.diffID WHERE CardOperations.field = excluded.field", fieldID, operationID, field)
			if err != nil {
				return fmt.Errorf("error while inserting field %s for bank card %s for update operation: %w", field, cardNumber, err)
			}
		}

		_, err = db.ExecContext(ctxCache, "DELETE FROM CardOperations "+
			"WHERE NOT EXISTS "+
			"(SELECT 1 FROM fields WHERE OperationBankCardDiff.operationID=CardOperations.operationID)")

		if err != nil {
			return fmt.Errorf("error while deleting all operations without fields to update: %w", err)
		}

	}

	return nil
}

// GetBankCard - function for getting bank card credentials from application cache. 
// The function also checks if data is up to date.
func (cache *SQLite) GetBankCard(cardNumber string) error {
	return nil
}

// UploadBankCard - function, that update existing bank card or insert new one. 
func (cache *SQLite) UploadBankCard(cardNumber string) error {
	return nil
}
