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

func (cache *SQLite) UpdateCacheMiss(userName string, cacheMiss int) error {
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite: %s\n", err)
		return fmt.Errorf("error while openning connection to SQLite: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctxCache, "UPDATE Users SET "+
		"cache_miss = CASE WHEN $2 = 0 THEN 0 ELSE Users.cache_miss + 1 END "+
		"WHERE userName=$1", userName, cacheMiss)

	if err != nil {
		return fmt.Errorf("error while updating cache miss for user %s: %w", userName, err)
	}

	return nil

}

func (cache *SQLite) GetCacheMiss(userName string) (int, error) {
	var cacheMiss int
	db, err := sql.Open("sqlite3", "./data_cache.db")
	if err != nil {
		fmt.Printf("Error while openning connection to SQLite: %s\n", err)
		return 0, fmt.Errorf("error while openning connection to SQLite: %w", err)
	}

	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Println("Error while adding an opportunity to use foreign keys: ", err)
		return 0, fmt.Errorf("error while adding foreign_key extension: %w", err)
	}

	ctxCache, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := db.QueryContext(ctxCache, "SELECT cache_miss FROM Users WHERE userName=$1", userName)
	if err != nil {
		return 0, fmt.Errorf("error while selecting cache miss for user %s: %w", userName, err)
	}

	err = row.Scan(&cacheMiss)
	if err != nil {
		return 0, fmt.Errorf("error while scanning cache miss for user %s: %w", userName, err)
	}

	return cacheMiss, nil

}
