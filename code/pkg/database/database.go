// database provides a set of methods to access a variety of databases,
// postgres, sqlite and potentially others.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	// Database drivers.
	_ "github.com/lib/pq"

	_ "modernc.org/sqlite"
	// _ "github.com/mattn/go-sqlite3"

	"github.com/goblimey/go-tools/testsupport"

	"github.com/goblimey/go-stripe-payments/code/pkg/config"
)

const EmailPermNameIntern = "PERMISSION_TO_SEND_EMAILS"
const DataStoragePermNameIntern = "DATA_PROTECTION_PERMISSION"

type DBConfig struct {
	Type   string       // The type of database, for example "postgres" or "sqlite".
	User   string       // The user connecting to the database.
	Pass   string       // the password of the iser connecting.
	Host   string       // The host machine running the database.
	Port   string       // The port on the host machine that the database uses.
	Name   string       // the name of the database (AKA "schema")
	Logger *slog.Logger // The structured logger for trace and error messages.
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

type Organisation struct {
	ID        int64
	UUID      string
	Shortname string
	Longname  string
	HomePage  string
}

// User holds the contents of a row from the adm_users table.
type User struct {
	ID                    int64     `json:"usr_id"`
	UUID                  string    `json:"usr_uuid"`
	LoginName             string    `json:"usr_login_name"`
	Password              string    `json:"usr_password"`
	Photo                 []byte    `json:"usr_photo"`
	Text                  string    `json:"usr_text"`
	PasswordResetID       string    `json:"usr_pw_reset_id"`
	PasswordRestTimestamp time.Time `json:"usr_pw_reset_timestamp"`
	LastLogin             time.Time `json:"usr_last_login"`
	ActualLogin           time.Time `json:"usr_actual_login"`
	NumberLogin           int       `json:"usr_number_login"`
	DateInvalid           time.Time `json:"usr_date_invalid"`
	NumberInvalid         int       `json:"usr_number_invalid"`
	IDCreate              int       `json:"usr_usr_id_create"`
	TimeStampCreate       time.Time `json:"usr_timestamp_create"`
	IDChange              int       `json:"usr_usr_id_change"`
	TimestampChange       time.Time `json:"usr_timestamp_change"`
	Valid                 bool      `json:"usr_valid"`
}

func NewUser(loginName string) *User {
	u := User{
		LoginName: loginName,
	}

	return &u
}

type Category struct {
	ID         int64
	UUID       string
	Org        *Organisation // The organisation that the category belongs to.
	Type       string        //  The type, for example "ROL" to supprt a role.
	NameIntern string        // The internal name, eg "COMMON"
	Name       string        // The name, eg "SYS_COMMON"
	System     bool          // System category or not.
	Default    bool          // Default category or not.
	Sequence   int           // The order in which the categories appear on the user profile page.
	CreateUser *User         // The user that created the category.
}

func NewCategory(org *Organisation, categoryType, nameIntern, name string, system, defaultStatus bool, sequence int, createUser *User) *Category {
	cat := Category{
		Org:        org,
		Type:       categoryType,
		NameIntern: nameIntern,
		Name:       name,
		System:     system,
		Default:    defaultStatus,
		Sequence:   sequence,
		CreateUser: createUser,
	}
	return &cat
}

// Role holds the data about a role in the adm_roles table.  The roles
// are already created.  The adm_roles table has a lot of fields but we
// can ignore most of them.  We only need the rol_id and the rol_name
// fields.
type Role struct {
	ID            int64  `json:"rol_id"`
	UUID          string `json:"rol_uuid"`
	Name          string `json:"rol_name"`
	RoleCategory  *Category
	CreateUser    *User
	Administrator bool
	Valid         bool
}

const RoleNameAdmin = "Administrator"
const RoleNameMember = "Member"

func NewRole(name string, category *Category, user *User, admin, valid bool) *Role {
	role := Role{
		Name:          name,
		RoleCategory:  category,
		CreateUser:    user,
		Administrator: admin,
		Valid:         valid,
	}
	return &role
}

// Member holds data from an adm_members record.  There are
// many members for each user, one per role (Member, Admin etc)
type Member struct {
	ID        int64  `json:"mem_id"`
	UserID    int64  `json:"mem_usr_id"`
	RoleID    int64  `json:"mem_rol_id"`
	UUID      string `json:"mem_uuid"`
	StartDate string `json:"mem_begin"`
	EndDate   string `json:"mem_end"`
	Approved  int    `json:"mem_approved"`
}

// MembershipSale represents the payment of a membership sale - the annual
// membership fee.
type MembershipSale struct {
	ID                    int64
	PaymentService        string  // The payment processor eg "Stripe".
	PaymentStatus         string  // "pending", "complete" or "cancelled"
	PaymentID             string  // The transaction Id from the payment processor.
	TransactionType       string  // The transaction type, eg 'membership renewal'
	MembershipYear        int     // The membership year paid for.
	UserID                int64   // The user ID of the member
	OrdinaryMemberFeePaid float64 // The fee paid for ordinary membership.
	Friend                bool    // True if the ordinary member is a friend of the museum.
	FriendFeePaid         float64 // The fee paid for the ordinary member to be a friend.
	FirstName             string  // First name (for new members)
	LastName              string  // Last name (for new members)
	Email                 string  // Email address (for new members)
	DonationToSociety     float64 // donation to the society.
	DonationToMuseum      float64 // donation to the museum.
	Giftaid               bool    // True if the member consents to Giftaid.
	AssocUserID           int64   // The user ID of the associate member.
	AssocFeePaid          float64 // the fee paid for associate membership.
	AssocFriend           bool    // True if the associate member is a friend of the museum.
	AssocFriendFeePaid    float64 // The fee paid for associate member to be a fiend.
	AssocFirstName        string  // First name (for new members)
	AssocLastName         string  // Last name (for new members)
	AssocEmail            string  // Email address (for new members)

	// Some HTML views are passed a sale object when the template is executed.  These
	// fields are used only by those views.  They are not stored in the database, but
	// should be set from the config when necessary before executing the template.
	EnableOtherMemberTypes   bool   // Enable associate members, friends etc.
	EnableGiftaid            bool   // Enable Giftaid.
	OrganisationName         string // Name of the organisation charging (quoted in various pages)
	EmailAddressForQuestions string // Emai address for questions (quoted in various pages)
	EmailAddressForFailures  string // Email addess for Failures after (quoted in various pages)

}

// NewMembershipSale creates a MembershipSale object.
func NewMembershipSale(c *config.Config) *MembershipSale {

	sale := MembershipSale{
		OrganisationName:         c.OrganisationName,
		OrdinaryMemberFeePaid:    c.OrdinaryMemberFee,
		EnableOtherMemberTypes:   c.EnableOtherMemberTypes,
		EnableGiftaid:            c.EnableGiftaid,
		EmailAddressForQuestions: c.EmailAddressForQuestions,
		EmailAddressForFailures:  c.EmailAddressForFailures,
	}

	return &sale
}

// Total calculates and returns the total cost of the purchase.  It's used in HTML
// templates so is parameterless and single-valued.  To gaurd against an attack that
// injects dangerous data into the form such as negative numbers, if any values are
// obviously illegal, the result is zero, which never happens with real data.  The
// back end should watch out for this and stop processing rather than displaying
// that value.
func (ms *MembershipSale) Total() float64 {
	switch {
	case ms.FriendFeePaid < 0:
		return 0.0
	case ms.AssocFeePaid < 0:
		return 0.0
	case ms.AssocFriendFeePaid < 0:
		return 0.0
	case ms.DonationToSociety < 0:
		return 0.0
	case ms.DonationToMuseum < 0:
		return 0.0
	}

	total := ms.OrdinaryMemberFeePaid +
		ms.FriendFeePaid +
		ms.DonationToSociety +
		ms.DonationToMuseum +
		ms.AssocFeePaid +
		ms.AssocFriendFeePaid

	return total
}

func (ms *MembershipSale) TotalForDisplay() string {

	total := ms.Total()

	if total == 0.0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", total)
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (ms *MembershipSale) OrdinaryMemberFeeForDisplay() string {
	if ms.OrdinaryMemberFeePaid == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", ms.OrdinaryMemberFeePaid)
}

// OrdinaryMemberFriendFeeForDisplay gets the ordinary member's
// museum friend fee for display - a number to two decimal places.
// If the member is not a friend, it returns "0.0".
func (ms *MembershipSale) FriendFeeForDisplay() string {

	if !ms.Friend {
		return ""
	}

	if ms.FriendFeePaid == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", ms.FriendFeePaid)
}

// DonationToSocietyForDisplay gets the donation to the society
// for a display - a number to two decimal places.
func (ms *MembershipSale) DonationToSocietyForDisplay() string {
	if ms.DonationToSociety == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", ms.DonationToSociety)
}

// DonationToMuseumForDisplay gets the donation to museum
// for a display - a number to two decimal places.
func (ms *MembershipSale) DonationToMuseumForDisplay() string {
	if ms.DonationToMuseum == 0 {
		return ""
	}
	return fmt.Sprintf("£%.2f", ms.DonationToMuseum)
}

// AssociateMemberFeeForDisplay gets the associate membership fee
// for display - a number to two decimal places.  If the value is
// zero or there is no associate, it returns "".
func (ms *MembershipSale) AssocFeeForDisplay() string {

	if ms.AssocFeePaid == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", ms.AssocFeePaid)
}

// AssociateMemberFriendFeeForDisplay gets the associate member's
// museum friend fee for display - a number to two decimal places.
// If there is no associate or the associate is not a friend, it
// returns "0.0".
func (ms *MembershipSale) AssocFriendFeeForDisplay() string {

	if !ms.AssocFriend {
		return ""
	}

	if ms.AssocFeePaid == 0 {
		return ""
	}

	return fmt.Sprintf("£%.2f", ms.AssocFriendFeePaid)
}

// FieldData holds the IDs of the fields in adm_user_fields.
type FieldData struct {
	ID         int64 // The usf_id.
	UUID       string
	Name       string // The usf_name.
	NameIntern string // The usf_name_intern.
	Type       string // The usf_type eg "EMAIL", "TEXT".
	Sequence   int    // The position on the profile page.
	CreateUser *User
	Cat        *Category
}

type Database struct {
	Config        *DBConfig             // The database config.
	Connection    *sql.DB               // The database connection.
	Transaction   *sql.Tx               // The transaction.
	SQLiteTempDir string                // The directory in /tmp used to store the SQLite DB.
	UserField     map[string]*FieldData // A cache of adm_user_field entries.
	Logger        *slog.Logger          // The structured daily logger
}

// New creates a database object using the given configuration.
func New(config *DBConfig) *Database {

	db := Database{
		Config: config,
	}
	return &db
}

// Connect connects to the given database and sets the connection in the object.
func (db *Database) Connect() error {

	switch db.Config.Type {
	case "postgres":
		var err error
		db.Connection, err = ConnectToPostgres(db.Config)
		if err != nil {
			db.Config.Logger.Error("Connect: " + err.Error())
			return err
		}

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

	default:
		return errors.New("no database config")
	}

	db.UserField = make(map[string]*FieldData)

	return nil
}

// Close closes the database connection.
func (db *Database) Close() error {

	closeError := db.Connection.Close()

	if db.Config.Type == "sqlite" {
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

	if db.Config.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row, err := db.Transaction.Query(query, args...)
	if db.Config.Type == "sqlite" {
		tables := db.ListSQLiteTables()
		_ = tables
	}
	return row, err
}

// QueryRow executes a query that is expected to return at most one row.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.QueryRow to do the work.
func (db *Database) QueryRow(query string, args ...any) *sql.Row {

	if db.Config.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	row := db.Transaction.QueryRow(query, args...)

	return row
}

// Exec executes an SQL statement such as an insert.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.Exec to do the work.
func (db *Database) Exec(query string, args ...any) (sql.Result, error) {

	if db.Config.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	result, err := db.Transaction.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateRow executes the given query and returns the id of the row.  It assumes that
// the query is an insert.  It massages the query parameter placeholders into the
// correct form for the database and uses db.sql.Query to do the work.
func (db *Database) CreateRow(query string, args ...any) (int64, error) {

	var id int64

	switch db.Config.Type {
	case "postgres":
		// Postgress doesn't support LastInsertID so the query for postgress should contain a
		// RETURNING clause that produces the ID for the scan.
		err := db.Transaction.QueryRow(query, args...).Scan(&id)
		if err != nil {
			return 0, err
		}

	default:
		// Databases such as SQLite supply the ID via LastInsertID.
		query = postgresParamsToSQLiteParams(query)
		res, err := db.Transaction.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		var err2 error
		id, err2 = res.LastInsertId()
		if err2 != nil {
			return 0, err2
		}
	}

	return id, nil
}

// UpdateRow executes the given query and returns the number of rows affected.  It assumes
// that the query is an update.  It massages the query parameter placeholders into the
// correct form for the database and uses db.sql.Query to do the work.
func (db *Database) UpdateRow(query string, args ...any) (int64, error) {

	res, err := db.Transaction.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rows, rError := res.RowsAffected()
	if rError != nil {
		return 0, rError
	}
	return rows, nil

}

// DeleteRow executes a query which should be a delete.
// It massages the query parameter placeholders into the correct form for
// the database and uses db.sql.QueryRow to do the work.
func (db *Database) DeleteRow(query string, args ...any) (int64, error) {

	if db.Config.Type == "sqlite" {
		query = postgresParamsToSQLiteParams(query)
	}

	res, err1 := db.Transaction.Exec(query, args...)
	if err1 != nil {
		return 0, err1
	}

	numRows, err2 := res.RowsAffected()
	if err2 != nil {
		return 0, err2
	}

	return numRows, nil
}

// ListSQLiteTables returns a list of the SQLite tables.
// (Used for debugging.)
func (db *Database) ListSQLiteTables() []string {

	result := make([]string, 1)
	if db.Config.Type != "sqlite" {
		result := append(result, "not implemented for DB "+db.Config.Type)
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

	slog.Debug("ConnectToSQLite: " + connectionDetails)

	// If you use the "modernc.org/sqlite" driver, open as follows:
	// conn, err := sql.Open("sqlite", connectionDetails)
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
