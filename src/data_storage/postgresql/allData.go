// PostgreSQL - package, that contains functions for managing user sensetive data,
// such as passwords, files and card credentials. Supported operations are create,
// update, get and delete for every data type. Also getting all sensetive data for
// the current user is supported.
package postgresql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// GetAllPasswords - function for getting all passwords for user.
func (pg *PostgreSQLConnection) GetAllPasswords(ctx context.Context, passwords *[]*pb.PasswordMessage) (map[string][]byte, error) {

	passwordVector := make(map[string][]byte, 100)

	rows, err := pg.dbConn.QueryContext(ctx, "SELECT password, metaData, initVector, application FROM Credentials WHERE userID=$1", ctx.Value(ut.IDKey))
	if err != nil {
		return nil, fmt.Errorf("error while getting data about all passwords for user with id %s: %w", ctx.Value(ut.IDKey), err)
	}

	defer rows.Close()
	for rows.Next() {
		var initVector []byte
		var passwordInfo pb.PasswordMessage
		err = rows.Scan(&passwordInfo.Password, &passwordInfo.MetaData, &initVector, &passwordInfo.Application)
		if err != nil {
			return nil, fmt.Errorf("error while scanning password data for user with id %s: %w", ctx.Value(ut.IDKey), err)
		}
		*passwords = append(*passwords, &passwordInfo)
		passwordVector[passwordInfo.Password] = initVector
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while saving all password data for user id %s: %w", ctx.Value(ut.IDKey), err)
	}

	return passwordVector, nil
}

// GetAllPasswords - function for getting all bank card credentials for user.
func (pg *PostgreSQLConnection) GetAllCardsCredentials(ctx context.Context, bankCards *[]*pb.BankCardMessage) (map[string][]byte, error) {
	cvcVector := make(map[string][]byte, 100)

	rows, err := pg.dbConn.QueryContext(ctx, "SELECT cardNumber, cvcCode, date, bank, metadata, initVector FROM BankCards WHERE userID=$1", ctx.Value(ut.IDKey))
	if err != nil {
		return nil, fmt.Errorf("error while getting bank credentials for user with id %s: %w", ctx.Value(ut.IDKey), err)
	}

	defer rows.Close()
	for rows.Next() {
		var initVector []byte
		var bankCardInfo pb.BankCardMessage
		var date string
		err = rows.Scan(&bankCardInfo.CardNumber, &bankCardInfo.CvcCode, &date, &bankCardInfo.Bank, &bankCardInfo.Metadata, &initVector)
		if err != nil {
			return nil, fmt.Errorf("error while scanning card credentials data for user with id %s: %w", ctx.Value(ut.IDKey), err)
		}
		t, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return nil, fmt.Errorf("error while parsing date to format MM/YY: %w", err)
		}

		bankCardInfo.Data = t.Format("01/06")

		*bankCards = append(*bankCards, &bankCardInfo)
		cvcVector[bankCardInfo.CvcCode] = initVector
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while saving all card credentials for user id %s: %w", ctx.Value(ut.IDKey), err)
	}

	return cvcVector, nil

}

// GetAllFilesInfo - function for getting info about all files for user.
func (pg *PostgreSQLConnection) GetAllFilesInfo(ctx context.Context) (map[string]string, error) {
	fileInfo := make(map[string]string, 100)

	rows, err := pg.dbConn.QueryContext(ctx, "SELECT fileName, metaData FROM UserFiles WHERE userID=$1", ctx.Value(ut.IDKey))
	if err != nil {
		return nil, fmt.Errorf("error while getting data about files for user with id %s: %w", ctx.Value(ut.IDKey), err)
	}

	defer rows.Close()
	for rows.Next() {
		var fileName, metaData string

		err = rows.Scan(&fileName, &metaData)
		if err != nil {
			return nil, fmt.Errorf("error while scanning file info for user with id %s: %w", ctx.Value(ut.IDKey), err)
		}

		fileInfo[fileName] = metaData
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error while saving all info about files for user id %s: %w", ctx.Value(ut.IDKey), err)
	}

	return fileInfo, nil
}
