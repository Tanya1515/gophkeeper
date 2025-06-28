package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CreateUser creates user and data about him in local storage for client.
func (cache *SQLite) CreateUser(userName string) error {
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite: %s\n", err)
		return err
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctxCache, "INSERT INTO Users (userName, cache_miss) VALUES ($1, $2) ON CONFLICT(userName) DO NOTHING", userName, 0)
	if err != nil {
		fmt.Printf("Error while creating user %s in local storage: %s\n", userName, err)
		return err
	}

	return nil
}
