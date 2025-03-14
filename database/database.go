// model provides a set of methods to access a variety of databases,
// postgres, sqlite and potentially others.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	// Database drivers.
	_ "github.com/lib/pq"

	// _ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"

	"github.com/goblimey/go-tools/testsupport"
)

// The next value for the ID when a record is created in an SQLite table.
var nextSQLiteID int

type DBConfig struct {
	Type string
	User string
	Pass string
	Host string
	Port string
	Name string
}

func GetDBConfigFromTheEnvironment() DBConfig {
	config := DBConfig{
		Type: os.Getenv("DBType"),
		User: os.Getenv("DBUser"),
		Pass: os.Getenv("DBPassword"),
		Host: os.Getenv("DBHost"),
		Port: os.Getenv("DBPort"),
		Name: os.Getenv("DBDatabase"),
	}

	return config
}

func (dbc *DBConfig) String() string {
	return fmt.Sprintf(
		"Type: %s,User: %s, Host: %s, Port: %s, Name: %s, Pass: %s",
		dbc.Type, dbc.User, dbc.Host, dbc.Port, dbc.Name, dbc.Pass)
}

type Database struct {
	Type          string   // The type of database - "postgres", "sqlite" etc.
	Config        DBConfig // The database config.
	Connection    *sql.DB  // The database connection.
	Transaction   *sql.Tx  // The transaction.
	SQLiteTempDir string   // The directory in /tmp used to store the SQLite DB.
}

func New(dbConfig DBConfig) *Database {
	db := Database{Config: dbConfig}
	return &db
}

func (db *Database) String() string {
	s := fmt.Sprintf("Database{Type: %s, Connection %v, Transaction %v}",
		db.Type, db.Connection, db.Transaction)
	return s
}

// Connect connects to the given database and sets the connection in the object.
func (db *Database) Connect() error {

	switch db.Config.Type {
	case "postgres":
		var err error
		db.Connection, err = ConnectToPostgres(&db.Config)
		if err != nil {
			fmt.Println("Connect: ", err.Error())
			return err
		}
		db.Type = db.Config.Type
		return nil

	case "sqlite":

		// Create a working directory.  Close() removes it.
		var wdErr error
		db.SQLiteTempDir, wdErr = testsupport.CreateWorkingDirectory()
		if wdErr != nil {
			return wdErr
		}

		// Attempts to use an in-memory database produced random failures
		// due to the database being closed and cleared down prematurely
		// after various queries had run.  Instead we use file databases
		// in /tmp.

		connectionDetails := "file:" + db.SQLiteTempDir + "/sqlite3.db"

		var connErr error
		db.Connection, connErr = ConnectToSQLite(connectionDetails)
		if connErr != nil {
			return connErr
		}
		db.Type = db.Config.Type
		return nil
	default:
		return errors.New("no database config")
	}
}

// Close closes the database connection.
func (db *Database) Close() error {

	closeError := db.Connection.Close()

	if db.Type == "sqlite" {
		// Whether the close worked or not, we must remove the DB file.
		testsupport.RemoveWorkingDirectory(db.SQLiteTempDir)
	}

	return closeError
}

// BeginTx starts a transaction using the background context and
// the default isolation options.  The transaction is stored in
// the Database object.  (The approach of storing a single
// transaction makes sense in a web solution that opens a
// transaction at the start of each HTTP request and closes
// it at the end.)
func (db *Database) BeginTx() error {
	var err error
	db.Transaction, err = db.Connection.BeginTx(context.Background(), nil)
	return err
}

// Commit commits the stored transaction.
func (db *Database) Commit() error {
	return db.Transaction.Commit()
}

// Rollback roles back the stored transaction.
func (db *Database) Rollback() error {
	return db.Transaction.Rollback()
}

// Query executes the given query and returns the rows.  It massages
// the query parameter placeholders into the correct form for the
// database and uses db.sql.Query to do the work.
func (db *Database) Query(query string, args ...any) (*sql.Rows, error) {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row, err := db.Transaction.Query(query, args...)
	if db.Type == "sqlite" {
		tables := db.ListSQLiteTables()
		_ = tables
	}
	return row, err
}

// QueryRow executes a query that is expected to return at most one row.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.QueryRow to do the work.
func (db *Database) QueryRow(query string, args ...any) *sql.Row {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row := db.Transaction.QueryRow(query, args...)

	return row
}

// Exec executes an SQL statement such as an insert.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.Exec to do the work.
func (db *Database) Exec(query string, args ...any) (sql.Result, error) {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	result, err := db.Transaction.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListSQLiteTables returns a list of the SQLite tables.
// (Used for debugging.)
func (db *Database) ListSQLiteTables() []string {

	result := make([]string, 1)
	if db.Type != "sqlite" {
		result := append(result, "not implemented for DB "+db.Type)
		return result
	}

	// SQLite stores dates as string, int or float.  We use strings
	// in the format "YYYY-MM-DD HH:MM:SS.SSS"
	const sql = `SELECT name FROM sqlite_master WHERE type='table'`

	rows, getNamesError := db.Connection.Query(sql)

	if getNamesError != nil {
		result = append(result, getNamesError.Error())
		return result
	}

	for {
		if !rows.Next() {
			break
		}
		var table string
		err := rows.Scan(&table)
		if err != nil {
			result = append(result, err.Error())
			return result
		}

		result = append(result, table)
	}

	return result
}

// ConnectToPostgres connects to the database specified in the Database object.
func ConnectToPostgres(dbConfig *DBConfig) (*sql.DB, error) {

	var connectionStr string
	if len(dbConfig.Pass) == 0 {
		// If the password is empty, don't supply "password=".
		connectionStr = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=disable",
			dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Name)
	} else {
		connectionStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Pass, dbConfig.Name)
	}

	// This checks the connection details, but doesn't open a connection!
	const driverName = "postgres"
	conn, errConn := sql.Open(driverName, connectionStr)
	if errConn != nil {
		return nil, errConn
	}

	// Ping actually opens the database connection.
	errPing := conn.Ping()
	if errPing != nil {
		return nil, errPing
	}

	return conn, nil
}

// Connect connects to the database.
func ConnectToSQLite(connectionDetails string) (*sql.DB, error) {

	// If you use the "modernc.org/sqlite" driver, open as follows:
	// conn, err := sql.Open("sqlite", connectionDetails)

	conn, err := sql.Open("sqlite3", connectionDetails)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// postgresParamsToSQLiteParams takes a query string and converts any
// Postgres-style parameter placeholders ('$1', '$2' etc) to sqlite-style
// placeholders ('?').  It uses a simple regular expression replacement
// so it can be defeated, for example by what looks like a placeholder
// within an SQL string - "select '$1' from foo where bar=$1".
func postgresParamsToSQLiteParams(query string) string {
	resultBytes := regExpForPostgresParamsToSQLiteParams.
		ReplaceAll([]byte(query), []byte("?"))
	result := string(resultBytes)
	return result
}
