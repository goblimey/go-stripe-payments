// database provides a set of methods to access a variety of databases,
// postgres, sqlite and potentially others.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	// Database drivers.
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
	// "github.com/mattn/go-sqlite3"
)

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
	Type       string   // The type of database - "postgres", "sqlite" etc.
	Config     DBConfig // The database config.
	Connection *sql.DB  // The database connection.
	// ID of the adm_user_fields table, created once and reused.
	firstNameID         int // ID of the First Name field
	lastNameID          int // ID of the Last Name field
	emailID             int // ID of the email address field
	friendID            int // ID of the friend field
	giftaidID           int // ID of the giftaid field
	lastPaymentID       int // ID of the last payment field
	donationToSocietyID int // ID of the donation to society field
	donationToMuseumID  int // ID of the donation to museum field
	membersAtAddressID  int // ID of the members at this address field
	friendsAtAddressID  int // ID of the number of friends at this address field
	dateLastPaidID      int // ID of the number of friends at this address field
}

func New(dbConfig DBConfig) *Database {
	db := Database{Config: dbConfig}
	return &db
}

func (db *Database) String() string {
	s := fmt.Sprintf("Database{Type: %s,Connection %v}", db.Type, db.Connection)
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
		var err error
		db.Connection, err = ConnectToSQLite()
		if err != nil {
			return err
		}
		db.Type = db.Config.Type
		return nil
	default:
		return errors.New("no database config")
	}
}

// Close closes the database connection.
func (db *Database) Close() error {
	err := db.Connection.Close()
	return err
}

// QueryRow executes a query that is expected to return at most one row.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.QueryRow to do the work.
func (db *Database) Query(query string, args ...any) (*sql.Rows, error) {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row, err := db.Connection.Query(query, args...)

	return row, err
}

// QueryRow executes a query that is expected to return at most one row.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.QueryRow to do the work.
func (db *Database) QueryRow(query string, args ...any) *sql.Row {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row := db.Connection.QueryRow(query, args...)

	return row
}

// Exec executes an SQL statement such as an insert.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.Exec to do the work.
func (db *Database) Exec(query string, args ...any) (sql.Result, error) {

	if db.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	result, err := db.Connection.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
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
func ConnectToSQLite() (*sql.DB, error) {

	// At present we only use in-memory sqlite databases.  If we
	// ever use a database on disk, this part may need reworking.

	connectionDetails := ":memory:"

	conn, err := sql.Open("sqlite", connectionDetails)
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
