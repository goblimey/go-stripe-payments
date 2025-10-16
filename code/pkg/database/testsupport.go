package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/goblimey/go-tools/dailylogger"

	"github.com/goblimey/go-tools/testsupport"
)

var DBConfigForTestingWithPostgres DBConfig
var DBConfigForTestingWithSQLite DBConfig
var ourOrganisation *Organisation
var systemUser *User
var catBasic *Category
var catCommon *Category

func init() {

	logger := createLoggerForTesting()

	DBConfigForTestingWithPostgres = DBConfig{
		Type:   "postgres",
		Host:   "localhost",
		Port:   "5432",
		User:   "postgres",
		Name:   "testdb",
		Pass:   "secret",
		Logger: logger,
	}
	DBConfigForTestingWithSQLite = DBConfig{
		Type:   "sqlite",
		Logger: logger,
	}
}

// ConnectForTesting connects to a local test database.
// creates an SQLite database in a temporary directory
// and sets the directory name in the Database object.
func (db *Database) ConnectForTestingWithSQLite() error {

	// Attempts to use an in-memory database produced random failures
	// due to the database being closed and cleared down prematurely
	// after various queries had run.  Instead we create a temporary
	// directory and use a file database in that.  The caller is
	// expected to remove the temporary directory once it's done with it.

	tempDir, tempError := testsupport.CreateWorkingDirectory()
	if tempError != nil {
		return tempError
	}

	connectionDetails := "file:" + tempDir + "/sqlite.db"

	var connErr error
	db.Connection, connErr = ConnectToSQLite(connectionDetails)
	if connErr != nil {
		return connErr
	}
	db.SQLiteTempDir = tempDir

	connError := db.Connect()
	if connError != nil {
		slog.Error(connError.Error())
		os.Exit(-1)
	}

	return nil
}

// CloseAndDelete closes the database connection and deletes the temporary
// directory where the database is stored.
func (db *Database) CloseAndDelete() error {

	// We only use temporary SQLite databases so this should never
	// be called with any other DB type.  If it is, just close the
	// database.
	switch db.Config.Type {

	case "sqlite":
		tempDir := db.SQLiteTempDir
		ConnErr := db.Connection.Close() //  Ignore the error
		_ = ConnErr
		// This is not thread safe!
		if len(db.SQLiteTempDir) == 0 {
			// The directory was never created or has already been removed.
			return nil
		}
		err := testsupport.RemoveWorkingDirectory(tempDir)
		if err != nil {
			return err
		}
		db.SQLiteTempDir = ""
		return nil

	default:
		return db.Connection.Close()
	}
}

func OpenDBForTesting(dbType string) (*Database, error) {

	var db *Database
	if dbType == "sqlite" {
		db = New(&DBConfigForTestingWithSQLite)
		db.UserField = make(map[string]*FieldData)
		db.Config.Logger = createLoggerForTesting()
		connError := db.ConnectForTestingWithSQLite()
		if connError != nil {
			return nil, connError
		}
	} else {
		db = New(&DBConfigForTestingWithPostgres)
		db.UserField = make(map[string]*FieldData)
		db.Config.Logger = createLoggerForTesting()
		connError := db.Connect()
		if connError != nil {
			return nil, connError
		}
	}

	return db, nil
}

// CreateTablesForTesting is a helper function that creates the tables
// needed for testing.  The Postgres tables are already created so only
// SQLite needs this.
func CreateTablesForTesting(db *Database) error {

	// The test postgres DB is permanent and the tables are created just once
	// once manually using admidio.4.3.14.schema.sql.
	// The sqlite DB is temporary, created at the start of each test and
	// deleted at the end, so the tables are created over and over using this
	// function.

	if db.Config.Type == "sqlite" {

		// There must be one user to support the category table.

		const createADMUsersSQL = `
		CREATE TABLE IF NOT EXISTS adm_users (
			usr_id INTEGER PRIMARY KEY,
			usr_uuid character varying(36) NOT NULL,
			usr_login_name character varying(254),
			usr_password character varying(255),
			usr_photo bytea,
			usr_text text,
			usr_pw_reset_id character varying(50),
			usr_pw_reset_timestamp timestamp without time zone,
			usr_last_login timestamp without time zone,
			usr_actual_login timestamp without time zone,
			usr_number_login integer DEFAULT 0 NOT NULL,
			usr_date_invalid timestamp without time zone,
			usr_number_invalid smallint DEFAULT 0 NOT NULL,
			usr_usr_id_create integer,
			usr_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
			usr_usr_id_change integer,
			usr_timestamp_change timestamp without time zone,
			usr_valid boolean DEFAULT false NOT NULL
		);
	`

		usersError := createTableForTesting(db, createADMUsersSQL)
		if usersError != nil {
			return usersError
		}

		const createOrgSQL = `
		CREATE TABLE IF NOT EXISTS adm_organizations (
			org_id integer PRIMARY KEY NOT NULL,
			org_uuid character varying(36) NOT NULL,
			org_shortname character varying(10) NOT NULL,
			org_longname character varying(60) NOT NULL,
			org_org_id_parent integer,
			org_homepage character varying(60) NOT NULL
		);
	`

		orgError := createTableForTesting(db, createOrgSQL)
		if orgError != nil {
			return orgError
		}

		const createCatSQL = `
	CREATE TABLE IF NOT EXISTS adm_categories (
		cat_id integer PRIMARY KEY NOT NULL,
		cat_org_id integer,
		cat_uuid character varying(36) NOT NULL,
		cat_type character varying(10) NOT NULL,
		cat_name_intern character varying(110) NOT NULL,
		cat_name character varying(100) NOT NULL,
		cat_system boolean DEFAULT false NOT NULL,
		cat_default boolean DEFAULT false NOT NULL,
		cat_sequence smallint NOT NULL,
		cat_usr_id_create integer,
		cat_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		cat_usr_id_change integer,
		cat_timestamp_change timestamp without time zone
	);
`

		catError := createTableForTesting(db, createCatSQL)
		if catError != nil {
			return catError
		}

		const createRolesTableSQL = `
		CREATE TABLE IF NOT EXISTS adm_roles (
			rol_id INTEGER PRIMARY KEY,
			rol_cat_id integer NOT NULL,
			rol_uuid varchar(50),
			rol_name varchar(20),
			rol_usr_id_create integer,
				rol_valid boolean DEFAULT true NOT NULL,
			rol_system boolean DEFAULT false NOT NULL,
			rol_administrator boolean DEFAULT false NOT NULL
			);`

		rolesError := createTableForTesting(db, createRolesTableSQL)
		if rolesError != nil {
			return rolesError
		}

		const createMembersTableSQL = `
		CREATE TABLE IF NOT EXISTS adm_members (
			mem_id INTEGER PRIMARY KEY,
			mem_rol_id integer NOT NULL,
			mem_usr_id integer NOT NULL,
			mem_uuid character varying(36) NOT NULL,
			mem_begin varchar(30) NOT NULL,
			mem_end varchar(30) NOT NULL,
			mem_leader boolean DEFAULT false NOT NULL,
			mem_usr_id_create integer,
			mem_timestamp_create varchar(30) NOT NULL DEFAULT CURRENT_TIMESTAMP,
			mem_usr_id_change integer,
			mem_timestamp_change varchar(30),
			mem_approved integer,
			mem_comment character varying(4000),
			mem_count_guests integer DEFAULT 0 NOT NULL
		);
	`

		createMembersError := createTableForTesting(db, createMembersTableSQL)
		if createMembersError != nil {
			return createMembersError
		}

		const createUserFieldsTableSQL = `
		CREATE TABLE IF NOT EXISTS adm_user_fields (
			usf_id integer PRIMARY KEY,
			usf_uuid character varying(36) NOT NULL,
			usf_type character varying(30) NOT NULL,
			usf_cat_id integer NOT NULL,
			usf_name_intern character varying(110) NOT NULL,
			usf_name character varying(100) NOT NULL,
			usf_sequence smallint NOT NULL,
			usf_usr_id_create integer
		);
	`

		createFieldsError := createTableForTesting(db, createUserFieldsTableSQL)
		if createFieldsError != nil {
			return createFieldsError
		}

		const createUserDataTableSQL = `
		CREATE TABLE IF NOT EXISTS adm_user_data (
		usd_id INTEGER PRIMARY KEY,
		usd_usr_id     integer NOT NULL,
		usd_usf_id     integer NOT NULL,
		usd_value      varchar(30) NOT NULL		
		);
	`

		createUserDataError := createTableForTesting(db, createUserDataTableSQL)
		if createUserDataError != nil {
			return createUserDataError
		}

		const createMembershipSalesSQL = `
	CREATE TABLE IF NOT EXISTS membership_sales (
		ms_id INTEGER PRIMARY KEY,
		ms_payment_service CHARACTER VARYING(36) NOT NULL,
		ms_payment_status CHARACTER VARYING(20) NOT NULL,
		ms_payment_id CHARACTER VARYING(200),
		ms_transaction_type varchar(30) NOT NULL DEFAULT 'membership renewal',
		ms_membership_year integer NOT NULL,
		ms_usr1_id integer DEFAULT NULL,
		ms_usr1_fee REAL NOT NULL,
		ms_usr1_friend boolean NOT NULL DEFAULT false,
		-- 0.0 if not a friend.
		ms_usr1_friend_fee REAL NOT NULL default 0.0,
		ms_usr1_first_name varchar (50),
		ms_usr1_last_name varchar (50),
		ms_usr1_email varchar (50),
		-- 0 if no associate
		ms_usr2_id integer DEFAULT NULL,
		-- 0.0 if no associate
		ms_usr2_fee REAL NOT NULL default 0.0,
		-- false if no associate.
		ms_usr2_friend boolean NOT NULL DEFAULT false,
		-- 0.0 if no associate.
		ms_usr2_friend_fee REAL NOT NULL DEFAULT 0.0,
		ms_usr2_first_name varchar (50),
		ms_usr2_last_name varchar(50),
		ms_usr2_email varchar (50),
		-- 0.0 if no donation.
		ms_donation REAL NOT NULL DEFAULT 0.0,
		-- 0.0 if no donation to museum.
		ms_donation_museum REAL NOT NULL DEFAULT 0.0,
		ms_giftaid boolean NOT NULL DEFAULT false,
		ms_timestamp_create varchar(30) NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ms_organisation_name varchar(30), 
		ms_email_address_for_questions varchare(30),
		ms_email_address_for_failures varchar(30)
	);
`

		membersCreateError := createTableForTesting(db, createMembershipSalesSQL)
		if membersCreateError != nil {
			return membersCreateError
		}
	}

	return nil
}

// PopulateTestTables is a helper function that loads the reference
// data into the tables.  The Postgres test database is fixed so the
// data is only loaded once.  After that, the function just checks
// that it's loaded.  If the database is SQLite, the data are created for
// every test.
func PopulateTestTables(db *Database) error {

	// Ensure the system user is set up.
	u, fetchUserError := db.GetUsersByLoginName("System")
	if fetchUserError != nil {
		return fetchUserError
	}

	if len(u) == 0 {
		systemUser = NewUser("System")
		err := db.CreateUserWithNullPassword(systemUser)
		if err != nil {
			return err
		}
	} else {
		systemUser = &u[0]
	}

	// Ensure that the organisation is set up.
	o, fetchOrgError := db.GetOrganisationsByShortName("org")
	if fetchOrgError != nil {
		return fetchOrgError
	}

	if len(o) > 0 {
		// The organisation exists.  Set the global pointer.
		ourOrganisation = &o[0]
	} else {
		// The organisation does not exist.  Create it.
		org := Organisation{
			Shortname: "org",
			Longname:  "the org",
			HomePage:  "https://example.com/",
		}

		// Set the global pointer.
		ourOrganisation = &org

		createOrgError := db.CreateOrganisation(ourOrganisation)
		if createOrgError != nil {
			return createOrgError
		}
	}

	// Ensure that the Categories are set up - "COMMON" and "BASIC_DATA".

	var fetchC1Error error
	catCommon, fetchC1Error = db.GetCategoryByNameIntern("COMMON")
	if fetchC1Error != nil && fetchC1Error != sql.ErrNoRows {
		return fetchC1Error
	}
	if fetchC1Error == sql.ErrNoRows {
		// This category is not set up.
		catCommon = NewCategory(ourOrganisation, "ROL", "COMMON", "SYS_COMMON", false, true, 1, systemUser)

		err := db.CreateCategory(catCommon)
		if err != nil {
			return err
		}
	}

	var fetchC2Error error
	catBasic, fetchC2Error = db.GetCategoryByNameIntern("BASIC_DATA")
	if fetchC2Error != nil && fetchC1Error != sql.ErrNoRows {
		return fetchC2Error
	}

	if fetchC2Error == sql.ErrNoRows {
		catBasic = NewCategory(ourOrganisation, "USF", "BASIC_DATA", "SYS_BASIC_DATA", false, true, 1, systemUser)

		err := db.CreateCategory(catBasic)
		if err != nil {
			return err
		}
	}

	// Ensure that the roles Admin and Member are set up.

	roleName := []string{RoleNameAdmin, RoleNameMember}

	for _, name := range roleName {

		_, fetchError := db.GetRole(name)

		switch {
		case fetchError != nil && fetchError != sql.ErrNoRows:
			// Fatal error.
			return fetchError

		case fetchError == sql.ErrNoRows:
			// Expected error - the role doesn't exist yet.  Create it.
			var isAdmin bool
			if name == RoleNameAdmin {
				isAdmin = true
			}
			// Roles we create are always valid.
			role := NewRole(name, catCommon, systemUser, isAdmin, true)
			createError := db.CreateRole(role)
			if createError != nil {
				return createError
			}
		}
	}

	// UserFieldsForTesting is the list of adm_user_fields rows to be set up for testing.
	var UserFieldsForTesting = []*FieldData{
		{0, "", "SYS_EMAIL", "EMAIL", "EMAIL", 6, systemUser, catBasic},
		{0, "", "Salutation", "SALUTATION", "TEXT", 2, systemUser, catBasic},
		{0, "", "Initials", "INITIALS", "TEXT", 4, systemUser, catBasic},
		{0, "", "SYS_FIRSTNAME", "FIRST_NAME", "TEXT", 3, systemUser, catBasic},
		{0, "", "SYS_LASTNAME", "LAST_NAME", "TEXT", 5, systemUser, catBasic},
		{0, "", "Address line 1", "STREET", "TEXT", 7, systemUser, catBasic},
		{0, "", "Address line 2", "ADDRESS_LINE_2", "TEXT", 8, systemUser, catBasic},
		{0, "", "address line 3", "ADDRESS_LINE_3", "TEXT", 9, systemUser, catBasic},
		{0, "", "County", "COUNTY", "TEXT", 11, systemUser, catBasic},
		{0, "", "SYS_POSTCODE", "POSTCODE", "TEXT", 13, systemUser, catBasic},
		{0, "", "SYS_COUNTRY", "COUNTRY", "TEXT", 12, systemUser, catBasic},
		{0, "", "City", "CITY", "TEXT", 10, systemUser, catBasic},
		{0, "", "SYS_PHONE", "PHONE", "PHONE", 15, systemUser, catBasic},
		{0, "", "SYS_MOBILE", "MOBILE", "PHONE", 16, systemUser, catBasic},
		{0, "", "gift aid", "GIFT_AID", "CHECKBOX", 30, systemUser, catBasic},
		{0, "", "Total value of last payment", "VALUE_OF_LAST_PAYMENT", "DECIMAL", 22, systemUser, catBasic},
		{0, "", "Friend of the Museum", "FRIEND_OF_THE_MUSEUM", "CHECKBOX", 23, systemUser, catBasic},
		{0, "", "date last paid", "DATE_LAST_PAID", "DATE", 19, systemUser, catBasic},
		{0, "", "Notices by email", "NOTICES_BY_EMAIL", "CHECKBOX", 31, systemUser, catBasic},
		{0, "", "Number of members of LDLHS at address", "MEMBERS_AT_ADDRESS", "NUMBER", 17, systemUser, catBasic},
		{0, "", "Number of Friends of the Museum at this address", "NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS", "NUMBER", 18, systemUser, catBasic},
		{0, "", "Notices by post", "NOTICES_BY_POST", "CHECKBOX", 32, systemUser, catBasic},
		{0, "", "Newsletter by Email", "NEWSLETTER_BY_EMAIL", "CHECKBOX", 33, systemUser, catBasic},
		{0, "", "Permission to send emails", EmailPermNameIntern, "CHECKBOX", 33, systemUser, catBasic},
		{0, "", "donation to the society", "DONATION_TO_SOCIETY", "DECIMAL", 34, systemUser, catBasic},
		{0, "", "Donation to the museum.", "DONATION_TO_MUSEUM", "DECIMAL", 35, systemUser, catBasic},
		{0, "", "Total value of last payment", "VALUE_OF_LAST_PAYMENT", "DECIMAL", 36, systemUser, catBasic},
		{0, "", "data protection permission", "DATA_PROTECTION_PERMISSION", "checkbox", 37, systemUser, catBasic},
	}

	// Create the field names in adm_user_fields.  The names of the fields are given
	// by the FieldNamesForTesting list.  In a postgres database the fields should
	// be set up in the database the first time this is run, and then be committed.
	// In an SQLite database, we need to create the rows for evey test.
	for _, f := range UserFieldsForTesting {

		_, fetchError := db.GetUserDataFieldByNameIntern(f.NameIntern)
		if fetchError != nil && fetchError != sql.ErrNoRows {
			return fetchError
		}

		if fetchError == sql.ErrNoRows {
			// The field is not in adm_user_fields.  Add it.
			fd := NewUserField(f.Name, f.NameIntern, f.Type, systemUser, catBasic)
			err := db.CreateUserField(fd)
			if err != nil {
				em := fmt.Sprintf("error creating field name %s - %v", f.NameIntern, err)
				return errors.New(em)
			}
		}
	}

	// If the DB is postgres, all the records should be created just once
	// and committed.  Tests in the future will find them set up already.
	// If it's SQLite, the database is set up in a temporary directory.
	// The data are set up at the start of each test and all changes rolled
	// back at the end.  Then the temporary directory is removed.

	if db.Config.Type == "postgres" {
		db.Commit()
		db.BeginTx() // The code that follows expects a transaction.
	}

	return nil
}

func PrepareTestTables(db *Database) error {
	// This only creates tables when the DB is sqlite.  The postgress
	// test database is set up permanently.  Some postgres objects need to be
	// set up, but only once, the first time this function is run on an empty
	// database.

	if db.Config.Type == "postgres" {
		return nil
	}

	e := CreateTablesForTesting(db)
	if e != nil {
		return e
	}

	e2 := PopulateTestTables(db)
	if e2 != nil {
		return e2
	}

	return nil
}

// createTablesForTesting ensures that the test tables are set up and
// populated with the reference data.
func createTableForTesting(db *Database, sql string) error {

	// Create the table.
	_, createError := db.Exec(sql)
	if createError != nil {
		return createError
	}

	return nil
}

func createLoggerForTesting() *slog.Logger {
	dailyLogWriter := dailylogger.New(".", "test.", ".log")

	// Create a structured logger that writes to the dailyLogWriter.
	logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))

	return logger
}
