package postgresql

import (
	"context"
	"fmt"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (pg *PostgreSQLConnection) UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string) error {

	_, err := pg.dbConn.Exec("INSERT INTO BankCards (userID, cardNumber, cvcCode, date, bank, metaData) VALUES ($1,$2,crypt($3, gen_salt('xdes')),TO_DATE($4, 'MM/YY'),$5,$6) "+
		" ON CONFLICT (cardNumber) DO"+
		" UPDATE SET date = excluded.date, metaData = excluded.metaData WHERE BankCards.cardNumber = excluded.cardNumber", ctx.Value(ut.LogginKey), cardNumber, cvc, date, bank, md)

	if err != nil {
		return fmt.Errorf("error while inserting/updating bank card credentials for card number %s: %w", cardNumber, err)
	}

	return nil
}

func (pg *PostgreSQLConnection) UploadPassword(ctx context.Context, password, app, md string) error {

	_, err := pg.dbConn.Exec("INSERT INTO Credentials (userID, password, application, metaData) VALUES ($1,crypt($2, gen_salt('xdes')),$3,$4)"+
		" ON CONFLICT (application) DO"+
		" UPDATE SET password = excluded.password, metaData = excluded.metaData WHERE Credentials.application = excluded.application", ctx.Value(ut.LogginKey), password, app, md)

	if err != nil {
		return fmt.Errorf("error while inserting/updating password for application %s : %w", app, err)
	}

	return nil
}
