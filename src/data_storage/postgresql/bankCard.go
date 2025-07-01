package postgresql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// UploadBankCard - function for uploading credentials of new bank card.
func (pg *PostgreSQLConnection) UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error {

	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO BankCards (userID, cardNumber, cvcCode, date, bank, metaData, initVector, updatedAt) VALUES ($1,$2,$3,TO_DATE($4, 'MM/YY'),$5,$6,$7,$8) "+
		" ON CONFLICT (cardNumber) DO"+
		" UPDATE SET date = excluded.date, metaData = excluded.metaData, initVector = excluded.initVector, updatedAt = excluded.updatedAt WHERE updatedAt =< excluded.updatedAt", ctx.Value(ut.IDKey), cardNumber, cvc, date, bank, md, initVector, updatedAt)

	if err != nil {
		return fmt.Errorf("error while inserting/updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeleteBankCard(ctx context.Context, cardNumber string) (err error) {

	_, err = pg.dbConn.ExecContext(ctx, "DELETE FROM BankCards WHERE cardNumber=$1 AND userID=$2", cardNumber, ctx.Value(ut.IDKey))

	return
}

// GetBankCardCredentials - function for getting bank card credentials for specified user.
func (pg *PostgreSQLConnection) GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, []byte, error) {
	var date string
	var err error
	var cardCreds pb.BankCardMessage
	var initVector []byte
	row := pg.dbConn.QueryRowContext(ctx, "SELECT cvcCode, date, bank, metadata, initVector FROM BankCards WHERE cardNumber=$1 AND userID=$2", cardNumber, ctx.Value(ut.IDKey))

	err = row.Scan(&cardCreds.CvcCode, &date, &cardCreds.Bank, &cardCreds.Metadata, &initVector)
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
func (pg *PostgreSQLConnection) UpdateBankCardCreds(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error {
	_, err := pg.dbConn.ExecContext(ctx,
		"UPDATE BankCards SET cvcCode=CASE WHEN $1 <> '' THEN $1 ELSE cvcCode END, "+
			"date=CASE WHEN $2 <> '' THEN TO_DATE($2, 'MM/YY') ELSE date END, "+
			"bank=CASE WHEN $3 <> '' THEN $3 ELSE bank END, "+
			"metaData=CASE WHEN $4 <> '' THEN $4 ELSE metaData END, "+
			"initVector = CASE WHEN $5::bytea IS NOT NULL THEN $5::bytea ELSE initVector END, "+
			"updatedAt = $6 "+
			"WHERE cardNumber=$7 AND userID=$8 AND updatedAt <= $4", cvc, date, bank, md, initVector, updatedAt, cardNumber, ctx.Value(ut.IDKey))

	if err != nil {
		return fmt.Errorf("error while updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}
