package postgresql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (pg *PostgreSQLConnection) UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string) error {

	_, err := pg.dbConn.ExecContext(ctx, "INSERT INTO BankCards (userID, cardNumber, cvcCode, date, bank, metaData) VALUES ($1,$2,$3,TO_DATE($4, 'MM/YY'),$5,$6) "+
		" ON CONFLICT (cardNumber) DO"+
		" UPDATE SET date = excluded.date, metaData = excluded.metaData WHERE BankCards.cardNumber = excluded.cardNumber", ctx.Value(ut.IDKey), cardNumber, cvc, date, bank, md)

	if err != nil {
		return fmt.Errorf("error while inserting/updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeleteBankCard(ctx context.Context, cardNumber string) (err error) {

	_, err = pg.dbConn.ExecContext(ctx, "DELETE FROM BankCards WHERE cardNumber=$1", cardNumber)

	return
}

func (pg *PostgreSQLConnection) GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, error) {
	var date string
	var err error
	var cardCreds pb.BankCardMessage
	row := pg.dbConn.QueryRowContext(ctx, "SELECT cvcCode, date, bank, metadata FROM BankCards WHERE cardNumber=$1", cardNumber)

	err = row.Scan(&cardCreds.CvcCode, &date, &cardCreds.Bank, &cardCreds.Metadata)
	if err != nil {
		return &cardCreds, err
	}

	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return &cardCreds, fmt.Errorf("error while parsing date to format MM/YY: %w", err)
	}

	cardCreds.Data = t.Format("01/06")

	return &cardCreds, err
}

func (pg *PostgreSQLConnection) UpdateBankCardCreds(ctx context.Context, cardNumber, cvc, date, bank, md string) error {
	_, err := pg.dbConn.ExecContext(ctx,
		"UPDATE BankCards SET cvc=CASE WHEN NULLIF(TRIM($1), '') IS NOT NULL THEN $1 ELSE cvc END, "+
			"SET date=CASE WHEN NULLIF(TRIM($2), '') IS NOT NULL THEN $2 ELSE date END, "+
			"SET bank=CASE WHEN NULLIF(TRIM($3), '') IS NOT NULL THEN $3 ELSE bank END, "+
			"SET metaData=CASE WHEN NULLIF(TRIM($4), '') IS NOT NULL THEN $4 ELSE metaData END "+
			"WHERE cardNumber=&5", cvc, date, bank, md, cardNumber)

	if err != nil {
		return fmt.Errorf("error while updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}
