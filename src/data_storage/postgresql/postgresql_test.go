package postgresql

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresTestSuite struct {
	suite.Suite
	QueryTimeout time.Duration
	tc           *tcpostgres.PostgresContainer
	cfg          *PostgreSQLConnection
}

func (ts *PostgresTestSuite) SetupSuite() {

	cfg := &PostgreSQLConnection{
		UserName: "postgres",
		Password: "postgres",
		DBName:   "postgres",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// create container with docker image for postgresql with database
	pgc, err := tcpostgres.Run(ctx,
		"docker.io/postgres:latest",
		tcpostgres.WithDatabase(cfg.DBName),
		tcpostgres.WithUsername(cfg.UserName),
		tcpostgres.WithPassword(cfg.Password),
		// wait for the condition
		// in the case wait for the string from log
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5*time.Second)),
	)

	// check if no error arrives
	require.NoError(ts.T(), err)

	cfg.Host, err = pgc.Host(ctx)
	require.NoError(ts.T(), err)

	//get mapped port from container to external
	require.NoError(ts.T(), err)

	ts.tc = pgc
	ts.cfg = cfg
	ts.QueryTimeout = 5 * time.Second
	require.NoError(ts.T(), cfg.Connect())

	ts.T().Logf("started postgres at %s", ts.cfg.Host)
}

// remove container with postgres
func (ts *PostgresTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(ts.T(), ts.tc.Terminate(ctx))
}

// clean all tables in database for running separate tests
func (ts *PostgresTestSuite) clean(ctx context.Context) error {
	newctx, cancel := context.WithTimeout(ctx, ts.QueryTimeout)
	defer cancel()

	_, err := ts.cfg.dbConn.ExecContext(newctx, "DELETE FROM metrics")
	return err
}

// function, that is running before every test case
func (ts *PostgresTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

// function, that is applied after test case for cleaning up test environment
func (ts *PostgresTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *PostgresTestSuite) TestRepositoryAddPassword() {

}

func (ts *PostgresTestSuite) TestGetPassword() {

}

func (ts *PostgresTestSuite) TestGetAllPasswords() {

}

func (ts *PostgresTestSuite) TestUpdatePassword() {

}

func (ts *PostgresTestSuite) TestDeletePassword() {

}

func (ts *PostgresTestSuite) TestRepositoryAdd() {

}

func (ts *PostgresTestSuite) TestGetPasswordCardCredentials() {

}

func (ts *PostgresTestSuite) TestGetAllCredentials() {

}

func (ts *PostgresTestSuite) TestUpdateCardCredentials() {

}

func (ts *PostgresTestSuite) TestDeleteCardCredentials() {

}
