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
func (cache *SQLite) DeleteBankCard(cardNumber string, userID int) error {

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite to delete bank card %s: %s\n", cardNumber, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statement, err := db.Prepare("DELETE from Cards WHERE cardNumber=? AND userID=?")
	if err != nil {
		return fmt.Errorf("error while making request for deleting bank card credentials with number %s: %w", cardNumber, err)
	}
	_, err = statement.ExecContext(ctxCache, cardNumber, userID)
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
func (cache *SQLite) GetBankCard(cardNumber string, userID int) (cvc string, date string, bankName string, metadatabankCard string, err error) {
	var lastUpdated, uploadTime string
	var accessCount int

	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return "", "", "", "", fmt.Errorf("error while openning connection to get data about bank card %s: %w", cardNumber, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctxCache, "SELECT cvcCode, date, bank, metadata, lastUpdated, accessCount, uploadTime FROM Cards WHERE cardNumber=$1 AND userID=$2", cardNumber, userID)

	err = row.Scan(&cvc, &date, &bankName, &metadatabankCard, &lastUpdated, &accessCount, &uploadTime)
	if err != nil {
		return "", "", "", "", fmt.Errorf("error while getting data for bank Card %s: %w", cardNumber, err)
	}

	// проверка актуальности данных (lastUpdated + accessCount), увеличиваем accessCount

	// логика по рашисфровке cvc

	return
}

// UploadBankCard - function, that update existing bank card or insert new one.
func (cache *SQLite) UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime string, userID int) (err error) {
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		return fmt.Errorf("error while openning connection to update data about bank card %s or add new one: %w", cardNumber, err)
	}

	defer db.Close()

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctxCache, "INSERT INTO Cards (cardNumber, cvcCode, date, bank, metadata, lastUpdated, uploadTime, userID) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) "+
		"ON CONFLICT (cardNumber, userID) DO "+
		"UPDATE SET "+
		"cvcCode = CASE WHEN excluded.cvcCode <> '' THEN excluded.cvcCode ELSE cvcCode END, "+
		"date = CASE WHEN excluded.date <> '' THEN excluded.date ELSE cvcCode END, "+
		"bank = CASE WHEN excluded.bank <> '' THEN excluded.bank ELSE bank END, "+
		"metadata = CASE WHEN excluded.metadata <> '' THEN excluded.metadata ELSE metadata END, "+
		"lastUpdated = excluded.lastUpdated, "+
		"Cards.accessCount = Cards.accessCount + 1 WHERE Cards.cardNumber = Cards.cardNumber AND Cards.userID = Cards.userID", cardNumber, cvc, date, bankName, metadatabankCard, uploadTime, uploadTime, userID)

	if err != nil {
		return fmt.Errorf("error while updating existing bank card %s or inserting new one: %w", cardNumber, err)
	}

	return nil
}
