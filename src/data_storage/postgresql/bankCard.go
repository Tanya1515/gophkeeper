package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// UploadBankCard - function for saving bank card credentials.
func (pg *PostgreSQLConnection) UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, initVector []byte, updatedAt time.Time) error {

	var cvcCodeGet, dateGet, bankGet, metaDataGet string
	var initVectorGet []byte
	var updatedAtGet time.Time

	tx, err := pg.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT cvcCode, date, bank, metaData, initVector, updatedAt FROM BankCards WHERE cardNumber=$1 AND userID=$2 FOR UPDATE", cardNumber, ctx.Value(ut.IDKey))

	err = row.Scan(&cvcCodeGet, &dateGet, &bankGet, &metaDataGet, &initVectorGet, &updatedAtGet)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		tx.Rollback()
		return fmt.Errorf("error while sensetive data for bank card %s: %w", cardNumber, err)
	}

	res, err := tx.ExecContext(ctx, "INSERT INTO BankCards (userID, cardNumber, cvcCode, date, bank, metaData, initVector, updatedAt) VALUES ($1,$2,$3,TO_DATE($4, 'MM/YY'),$5,$6,$7,$8) "+
		" ON CONFLICT (cardNumber, userID) DO "+
		"UPDATE SET cvcCode= $3, "+
		"date= TO_DATE($4, 'MM/YY'), "+
		"bank=$5, "+
		"metaData=$6, "+
		"initVector =$7::bytea, "+
		"updatedAt = $8 "+
		"WHERE BankCards.updatedAt <= excluded.updatedAt RETURNING cardNumber", ctx.Value(ut.IDKey), cardNumber, cvc, date, bank, md, initVector, updatedAt)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while updating/inserting bank card credentials for card number %s: %w", cardNumber, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.New("no rows with bank card " + cardNumber + " has been affected.")
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return errors.New("no rows with bank card " + cardNumber + " has been affected.")
	}

	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}

	return nil
}

// DeleteBankCard - function for deleting bank card credentials.
func (pg *PostgreSQLConnection) DeleteBankCard(ctx context.Context, cardNumber string) (err error) {

	res, err := pg.dbConn.ExecContext(ctx, "DELETE FROM BankCards WHERE cardNumber=$1 AND userID=$2 RETURNING cardNumber", cardNumber, ctx.Value(ut.IDKey))
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAffected == 0 {
		return errors.New("no rows with bank card " + cardNumber + " has been found.")
	}

	return
}

// GetBankCardCredentials - function for getting bank card credentials for specified user.
func (pg *PostgreSQLConnection) GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, []byte, error) {
	var date string
	var err error
	var cardCreds pb.BankCardMessage
	var initVector []byte
	row := pg.dbConn.QueryRowContext(ctx, "SELECT cvcCode, date, bank, metadata, initVector, updatedAt FROM BankCards WHERE cardNumber=$1 AND userID=$2", cardNumber, ctx.Value(ut.IDKey))

	err = row.Scan(&cardCreds.CvcCode, &date, &cardCreds.Bank, &cardCreds.Metadata, &initVector, &cardCreds.UploadTime)
	if err != nil {
		return &cardCreds, initVector, err
	}

	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return &cardCreds, initVector, fmt.Errorf("error while parsing date to format MM/YY: %w", err)
	}

	cardCreds.Data = t.Format("01/06")

	return &cardCreds, initVector, err
}

// UpdateBankCardCreds - function for updating bank card credentials.
func (pg *PostgreSQLConnection) UpdateBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error {

	var cvcCodeGet, dateGet, bankGet, metaDataGet string
	var initVectorGet []byte
	var updatedAtGet time.Time

	tx, err := pg.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT cvcCode, date, bank, metaData, initVector, updatedAt FROM BankCards WHERE cardNumber=$1 AND userID=$2 FOR UPDATE", cardNumber, ctx.Value(ut.IDKey))

	err = row.Scan(&cvcCodeGet, &dateGet, &bankGet, &metaDataGet, &initVectorGet, &updatedAtGet)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while sensetive data for bank card %s: %w", cardNumber, err)
	}

	res, err := tx.ExecContext(ctx, "UPDATE BankCards SET cvcCode=CASE WHEN $3 <> '' THEN $3 ELSE BankCards.cvcCode END, "+
		"date=CASE WHEN $4 <> '' THEN TO_DATE($4, 'MM/YY') ELSE BankCards.date END, "+
		"bank=CASE WHEN $5 <> '' THEN $5 ELSE BankCards.bank END, "+
		"metaData=CASE WHEN $6 <> '' THEN $6 ELSE BankCards.metaData END, "+
		"initVector = CASE WHEN $7::bytea IS NOT NULL THEN $7::bytea ELSE BankCards.initVector END, "+
		"updatedAt = $8 "+
		"WHERE BankCards.updatedAt <= $8 AND cardNumber=$2 AND userID=$1 RETURNING cardNumber", ctx.Value(ut.IDKey), cardNumber, cvc, date, bank, md, initVector, updatedAt)

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.New("no rows with bank card " + cardNumber + " has been affected.")
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return errors.New("no rows with bank card " + cardNumber + " has been affected.")
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}

	return nil
}
