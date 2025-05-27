package postgresql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (pg *PostgreSQLConnection) UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string) error {

	_, err := pg.dbConn.Exec("INSERT INTO BankCards (userID, cardNumber, cvcCode, date, bank, metaData) VALUES ($1,$2,crypt($3, gen_salt('xdes')),TO_DATE($4, 'MM/YY'),$5,$6) "+
		" ON CONFLICT (cardNumber) DO"+
		" UPDATE SET date = excluded.date, metaData = excluded.metaData WHERE BankCards.cardNumber = excluded.cardNumber", ctx.Value(ut.IDKey), cardNumber, cvc, date, bank, md)

	if err != nil {
		return fmt.Errorf("error while inserting/updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) DeleteBankCard(ctx context.Context, cardNumber string) (err error) {

	_, err = pg.dbConn.Exec("DELETE FROM BankCards WHERE cardNumber=$1", cardNumber)

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
