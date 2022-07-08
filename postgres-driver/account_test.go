package postgresdriver

import (
	"errors"
	"math/big"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	indexer "github.com/pokt-foundation/pocket-indexer-lib"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_WriteAccount(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	mock.ExpectExec("INSERT into accounts").WithArgs("00353abd21ef72725b295ba5a9a5eb6082548e21",
		21, "node", "212121", "upokt").
		WillReturnResult(sqlmock.NewResult(1, 1))

	driver := NewPostgresDriverFromSQLDBInstance(db)

	err = driver.WriteAccount(&indexer.Account{
		Address:             "00353abd21ef72725b295ba5a9a5eb6082548e21",
		Height:              21,
		AccountType:         indexer.AccountTypeNode,
		Balance:             big.NewInt(212121),
		BalanceDenomination: "upokt",
	})
	c.NoError(err)

	mock.ExpectExec("INSERT into accounts").WithArgs("00353abd21ef72725b295ba5a9a5eb6082548e21",
		21, "node", "212121", "upokt").
		WillReturnError(errors.New("dummy error"))

	err = driver.WriteAccount(&indexer.Account{
		Address:             "00353abd21ef72725b295ba5a9a5eb6082548e21",
		Height:              21,
		AccountType:         indexer.AccountTypeNode,
		Balance:             big.NewInt(212121),
		BalanceDenomination: "upokt",
	})
	c.EqualError(err, "dummy error")
}

func TestPostgresDriver_ReadAccountByAddress(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "address", "height", "account_type", "balance", "balance_denomination"}).
		AddRow(1, "00353abd21ef72725b295ba5a9a5eb6082548e21", 21, "node", "212121", "upokt")

	mock.ExpectQuery("^SELECT (.+) FROM accounts (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	account, err := driver.ReadAccountByAddress("00353abd21ef72725b295ba5a9a5eb6082548e2", &ReadAccountByAddressOptions{Height: 21})
	c.Equal(ErrInvalidAddress, err)
	c.Empty(account)

	account, err = driver.ReadAccountByAddress("00353abd21ef72725b295ba5a9a5eb6082548e21", &ReadAccountByAddressOptions{Height: 21})
	c.NoError(err)
	c.NotEmpty(account)

	rows = sqlmock.NewRows([]string{"id", "address", "height", "account_type", "balance", "balance_denomination"}).
		AddRow(1, "00353abd21ef72725b295ba5a9a5eb6082548e21", 21, "node", "212121", "upokt")

	mock.ExpectQuery("^SELECT (.+) FROM accounts (.+)").WillReturnRows(rows)

	account, err = driver.ReadAccountByAddress("00353abd21ef72725b295ba5a9a5eb6082548e21", nil)
	c.NoError(err)
	c.NotEmpty(account)

	mock.ExpectQuery("^SELECT (.+) FROM accounts (.+)").WillReturnError(errors.New("dummy error"))

	account, err = driver.ReadAccountByAddress("00353abd21ef72725b295ba5a9a5eb6082548e21", &ReadAccountByAddressOptions{Height: 21})
	c.EqualError(err, "dummy error")
	c.Empty(account)

	mock.ExpectQuery("^SELECT (.+) FROM accounts (.+)").WillReturnError(errors.New("dummy error"))

	account, err = driver.ReadAccountByAddress("00353abd21ef72725b295ba5a9a5eb6082548e21", nil)
	c.EqualError(err, "dummy error")
	c.Empty(account)
}

func TestPostgresDriver_ReadAccounts(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "address", "height", "account_type", "balance", "balance_denomination"}).
		AddRow(1, "00353abd21ef72725b295ba5a9a5eb6082548e21", 21, "node", "212121", "upokt").
		AddRow(2, "00353abd21ef72725b295ba5a9a5eb6082548e22", 21, "node", "212121", "upokt")

	mock.ExpectBegin()
	mock.ExpectQuery(".*").WillReturnRows(rows)
	mock.ExpectCommit()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	accounts, err := driver.ReadAccounts(&ReadAccountsOptions{Page: 21, PerPage: 7, Height: 21})
	c.NoError(err)
	c.Len(accounts, 2)

	mock.ExpectBegin()
	mock.ExpectQuery(".*").WillReturnError(errors.New("dummy error"))
	mock.ExpectCommit()

	accounts, err = driver.ReadAccounts(&ReadAccountsOptions{})
	c.EqualError(err, "dummy error")
	c.Empty(accounts)
}

func TestPostgresDriver_GetAccountsQuantity(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(100)

	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	maxHeight, err := driver.GetAccountsQuantity(&GetAccountsQuantityOptions{Height: 21})
	c.NoError(err)
	c.Equal(int64(100), maxHeight)

	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnError(errors.New("dummy error"))

	maxHeight, err = driver.GetAccountsQuantity(nil)
	c.EqualError(err, "dummy error")
	c.Empty(maxHeight)
}
