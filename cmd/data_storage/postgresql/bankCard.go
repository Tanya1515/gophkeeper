package postgresql

import (
	"context"
	"fmt"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
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

func (pg *PostgreSQLConnection) GetBankCardCredentials(ctx context.Context, cardNumber string) (cardCreds pb.BankCardMessage, err error) {

	row := pg.dbConn.QueryRowContext(ctx, "SELECT cvcCode, date, bank, metadata FROM BankCards WHERE cardNumber=$1", cardNumber)

	err = row.Scan(&cardCreds.CvcCode, &cardCreds.Data, &cardCreds.Bank, &cardCreds.Metadata)

	return
}
