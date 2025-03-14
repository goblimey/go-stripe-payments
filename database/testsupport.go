package database

import (
	"fmt"
)

const TestUserIDPostgres = 9
const TestAssociateIDPostgres = 1545
const TestUserEmail = "foo@bar.com"
const TestUserFirstName = "Simon"
const TestUserLastName = "Ritchie"
const TestAssociateEmail = "goo@bar.com"
const TestAssociateFirstName = "Luigi"
const TestAssociateLastName = "Schmidt"

var TestUserIDSQLite int
var TestAssociateIDSQLite int
var dbConfigForTestingWithPostgres DBConfig
var dbConfigForTestingWithSQLite DBConfig

func init() {
	dbConfigForTestingWithPostgres = DBConfig{
		Type: "postgres",
		Host: "localhost",
		Port: "5432",
		Name: "dev",
		User: "postgres",
	}
	dbConfigForTestingWithSQLite = DBConfig{
		Type: "sqlite",
		Name: ":memory:",
	}
}

// SetupDBForTesting is a helper function that sets up a database of the
// given type and loads any reference data.
func SetupDBForTesting(dbType string) (*Database, error) {

	var db *Database
	var connError error
	if dbType == "sqlite" {
		db = New(dbConfigForTestingWithSQLite)
		connError = db.Connect()
		if connError != nil {
			return nil, connError
		}

		prepError := prepareSQLiteTables(db)
		if prepError != nil {
			return nil, prepError
		}
	} else {
		db = New(dbConfigForTestingWithPostgres)
		connError = db.Connect()
		if connError != nil {
			return nil, connError
		}
	}

	return db, nil
}

// prepareSQLIteTables is a helper function that creates tables and loads
// loads test data into them.  It assumes that the database is SQLite.
func prepareSQLiteTables(db *Database) error {

	db.BeginTx()
	defer db.Rollback()

	createRolesTableSQL := `
		CREATE TABLE adm_roles (
		rol_id INTEGER PRIMARY KEY NOT NULL,
		rol_name varchar(20)		
	  );`

	sRoles, rolesError := db.Transaction.Prepare(createRolesTableSQL)
	if rolesError != nil {
		return rolesError
	}
	_, createRolesError := sRoles.Exec() // Create Roles table.
	if createRolesError != nil {
		return createRolesError
	}

	const fillRolesSQL = `
		insert into adm_roles(rol_id, rol_name) values(1, "Admin");
		insert into adm_roles(rol_id, rol_name) values(2, "Member");
	`

	_, fillRolesError := db.Transaction.Exec(fillRolesSQL)
	if fillRolesError != nil {
		return fillRolesError
	}

	const createADMUsersSQL = `
		CREATE TABLE adm_users (
			usr_id INTEGER PRIMARY KEY NOT NULL,
			usr_uuid character varying(36) NOT NULL,
			usr_login_name character varying(254),
			usr_password character varying(255),
			usr_photo byte,
			usr_text text,
			usr_pw_reset_id character varying(50),
			usr_pw_reset_timestamp varchar(30),
			usr_last_login varchar(30),
			usr_actual_login varchar(30),
			usr_number_login integer DEFAULT 0 NOT NULL,
			usr_date_invalid varchar(30),
			usr_number_invalid smallint DEFAULT 0 NOT NULL,
			usr_usr_id_create integer,
			usr_timestamp_create varchar(30),
			usr_usr_id_change integer,
			usr_timestamp_change varchar(30),
			usr_valid boolean DEFAULT false NOT NULL
		);
	`

	usersError := createTableForTesting(db, createADMUsersSQL)
	if usersError != nil {
		return usersError
	}

	insertUser := "insert into adm_users(usr_id, usr_uuid, usr_login_name) values(%d, '%s','%s');"
	nextSQLiteID++
	TestUserIDSQLite = nextSQLiteID
	createUser1SQL :=
		fmt.Sprintf(insertUser, TestUserIDSQLite, "a", TestUserEmail)

	_, createUser1Error := db.Transaction.Exec(createUser1SQL)
	if createUser1Error != nil {
		return createUser1Error
	}

	nextSQLiteID++
	TestAssociateIDSQLite = nextSQLiteID
	createUser2SQL :=
		fmt.Sprintf(insertUser, TestAssociateIDSQLite, "b", TestAssociateEmail)

	_, createUser2Error := db.Transaction.Exec(createUser2SQL)
	if createUser2Error != nil {
		return createUser2Error
	}

	// mem_id | mem_rol_id | mem_usr_id | mem_uuid | mem_begin | mem_end |
	// mem_leader | mem_usr_id_create | mem_timestamp_create |
	// mem_usr_id_change | mem_timestamp_change | mem_approved |
	// mem_comment | mem_count_guests
	createMembersTableSQL := `
		CREATE TABLE adm_members (
			mem_id INTEGER PRIMARY KEY NOT NULL,
			mem_rol_id integer NOT NULL,
			mem_usr_id integer NOT NULL,
			mem_uuid character varying(36) NOT NULL,
			mem_begin varchar(30) NOT NULL,
			mem_end varchar(30) NOT NULL,
			mem_leader boolean DEFAULT false NOT NULL,
			mem_usr_id_create integer,
			mem_timestamp_create varchar(30),
			mem_usr_id_change integer,
			mem_timestamp_change varchar(30),
			mem_approved integer,
			mem_comment character varying(4000),
			mem_count_guests integer DEFAULT 0 NOT NULL
		);
	`

	sMembers, membersError := db.Transaction.Prepare(createMembersTableSQL)
	if membersError != nil {
		return membersError
	}
	_, createMembersError := sMembers.Exec() // Create table.
	if createMembersError != nil {
		return createMembersError
	}

	createMember := `
		insert into adm_members(mem_id, mem_usr_id, mem_uuid, mem_rol_id, mem_begin, mem_end)
		values(%d, %d, '%s', 2, '%s', '%s');`

	nextSQLiteID++
	createMember1SQL := fmt.Sprintf(
		createMember, nextSQLiteID, TestUserIDSQLite, "a", "1970-01-01", "1970-01-01")

	_, createMember1Error := db.Transaction.Exec(createMember1SQL)
	if createMember1Error != nil {
		return createMember1Error
	}

	nextSQLiteID++
	createMember2SQL := fmt.Sprintf(
		createMember, nextSQLiteID, TestAssociateIDSQLite, "b", "1970-01-01", "1970-01-01")

	_, createMember2Error := db.Transaction.Exec(createMember2SQL)
	if createMember2Error != nil {
		return createMember2Error
	}

	createUserFieldsTableSQL := `
		CREATE TABLE adm_user_fields (
		usf_id INTEGER PRIMARY KEY NOT NULL,
		usf_name_intern varchar(30) NOT NULL		
	  );`

	sUserFields, err := db.Transaction.Prepare(createUserFieldsTableSQL)
	if err != nil {
		return err
	}
	_, createUserFieldsError := sUserFields.Exec() // Create Users table.
	if createUserFieldsError != nil {
		return createUserFieldsError
	}

	const fillUserFieldsSQL = `
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(1, "FIRST_NAME");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(2, "LAST_NAME");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(3, "EMAIL");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(4, "GIFT_AID");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(5, "MEMBERS_AT_ADDRESS");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(6, "VALUE_OF_LAST_PAYMENT");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(7, "FRIEND_OF_THE_MUSEUM");
		insert into adm_user_fields(usf_id, usf_name_intern)
			values(8, "DATE_LAST_PAID");	
	`

	_, fillUserFieldsError := db.Transaction.Exec(fillUserFieldsSQL)
	if fillUserFieldsError != nil {
		return fillUserFieldsError
	}

	// usd_id | usd_usr_id | usd_usf_id | usd_value
	createUserDataTableSQL := `
		CREATE TABLE adm_user_data (
		usd_id INTEGER PRIMARY KEY NOT NULL,
		usd_usr_id     integer NOT NULL,
		usd_usf_id     integer NOT NULL,
		usd_value      varchar(30) NOT NULL		
	  );
	`

	sUserData, err := db.Transaction.Prepare(createUserDataTableSQL)
	if err != nil {
		return err
	}
	_, createUserDataError := sUserData.Exec() // Create Users table.
	if createUserDataError != nil {
		return createUserDataError
	}

	const fillUserDataSQLTemplate = `
		insert into adm_user_data(usd_id,usd_usr_id,usd_usf_id,usd_value)
		values(1, %d, 1, '%s');
		insert into adm_user_data(usd_id,usd_usr_id,usd_usf_id,usd_value)
		values(2, %d, 2, '%s');
		insert into adm_user_data(usd_id,usd_usr_id,usd_usf_id,usd_value)
		values(3, %d, 3, '%s');
	`
	fillUserDataSQL := fmt.Sprintf(fillUserDataSQLTemplate,
		TestUserIDSQLite, TestAssociateFirstName,
		TestUserIDSQLite, TestAssociateLastName,
		TestUserIDSQLite, TestAssociateEmail)

	_, fillUserDataError := db.Transaction.Exec(fillUserDataSQL)
	if fillUserDataError != nil {
		return fillUserDataError
	}

	const createMembershipSalesSQL = `
		CREATE TABLE membership_sales
		(
			ms_id INTEGER PRIMARY KEY NOT NULL,
			ms_payment_service CHARACTER VARYING(36) NOT NULL,
			ms_payment_status CHARACTER VARYING(20) NOT NULL,
			ms_payment_id CHARACTER VARYING(200),
			ms_transaction_type varchar(30) NOT NULL DEFAULT 'membership renewal',
			ms_membership_year integer NOT NULL,
			ms_usr1_id integer NOT NULL,
			ms_usr1_fee REAL NOT NULL,
			ms_usr1_friend boolean NOT NULL DEFAULT false,
			-- 0.0 if not a friend.
			ms_usr1_friend_fee REAL NOT NULL default 0.0,
			ms_usr1_first_name varchar (50),
			ms_usr1_last_name varchar (50),
			ms_user1_email varchar (50),
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
			timestamp_create varchar(30) NOT NULL
		);
	`

	membersCreateError := createTableForTesting(db, createMembershipSalesSQL)
	if membersCreateError != nil {
		return membersCreateError
	}

	db.Commit()

	return nil
}

func createTableForTesting(db *Database, sql string) error {

	// Prepare the statement.
	stmt, prepareError := db.Transaction.Prepare(sql)
	if prepareError != nil {
		return prepareError
	}

	// Create the table.
	_, createError := stmt.Exec()
	if createError != nil {
		return createError
	}

	return nil
}
