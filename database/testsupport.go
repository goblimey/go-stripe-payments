package database

import (
	"fmt"
)

const TestUserIDPostgres = 9
const TestAssociateIDPostgres = 1545
const TestUserIDSQLite = 1
const TestAssociateIDSQLite = 2
const TestUserEmail = "foo@bar.com"
const TestUserFirstName = "Simon"
const TestUserLastName = "Ritchie"
const TestAssociateEmail = "goo@bar.com"
const TestAssociateFirstName = "Luigi"
const TestAssociateLastName = "Schmidt"

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

	createUsersTableSQL := `CREATE TABLE adm_users (
		usr_id         integer NOT NULL PRIMARY KEY,
		usr_login_name varchar(30) NOT NULL		
	  );`

	// use db.Connection to avoid tweaking the bespoke query.
	sUsers, err := db.Connection.Prepare(createUsersTableSQL)
	if err != nil {
		return err
	}
	_, createUsersError := sUsers.Exec() // Create Users table.
	if createUsersError != nil {
		return createUsersError
	}

	insertUser := "insert into adm_users(usr_id, usr_login_name) values(%d, '%s');"
	createUser1SQL :=
		fmt.Sprintf(insertUser, TestUserIDSQLite, TestUserEmail)

	// use db.Connection to avoid tweaking the bespoke query.
	_, createUser1Error := db.Connection.Exec(createUser1SQL)
	if createUser1Error != nil {
		return createUser1Error
	}

	createUser2SQL :=
		fmt.Sprintf(insertUser, TestAssociateIDSQLite, TestAssociateEmail)
	// use db.Connection to avoid tweaking the bespoke query.
	_, createUser2Error := db.Connection.Exec(createUser2SQL)
	if createUser2Error != nil {
		return createUser2Error
	}

	createRolesTableSQL := `CREATE TABLE adm_roles (
		rol_id integer NOT NULL PRIMARY KEY,
		rol_name varchar(20)		
	  );`

	sRoles, rolesError := db.Connection.Prepare(createRolesTableSQL)
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
	// use db.Connection to avoid tweaking the bespoke query.
	_, fillRolesError := db.Connection.Exec(fillRolesSQL)
	if fillRolesError != nil {
		return fillRolesError
	}

	// mem_id | mem_rol_id | mem_usr_id | mem_uuid | mem_begin | mem_end |
	// mem_leader | mem_usr_id_create | mem_timestamp_create |
	// mem_usr_id_change | mem_timestamp_change | mem_approved |
	// mem_comment | mem_count_guests
	createMembersTableSQL := `CREATE TABLE adm_members (
		mem_id     integer NOT NULL PRIMARY KEY,
		mem_usr_id integer NOT NULL,
		mem_rol_id integer NOT NULL,
		mem_begin  varchar(20),
		mem_end    varchar(20)	
	  );`

	// use db.Connection to avoid tweaking the bespoke query.
	sMembers, membersError := db.Connection.Prepare(createMembersTableSQL)
	if membersError != nil {
		return membersError
	}
	_, createMembersError := sMembers.Exec() // Create table.
	if createMembersError != nil {
		return createMembersError
	}

	createMember1SQL := fmt.Sprintf(`
		insert into adm_members(mem_id, mem_usr_id, mem_rol_id)
		values(1, %d, 2);`, TestUserIDSQLite)

	// use db.Connection to avoid tweaking the bespoke query.
	_, createMember1Error := db.Connection.Exec(createMember1SQL)
	if createMember1Error != nil {
		return createMember1Error
	}

	createMember2SQL := fmt.Sprintf(`
		insert into adm_members(mem_id, mem_usr_id, mem_rol_id)
		values(2, %d, 2);`, TestAssociateIDSQLite)

	// use db.Connection to avoid tweaking the bespoke query.
	_, createMember2Error := db.Connection.Exec(createMember2SQL)
	if createMember2Error != nil {
		return createMember2Error
	}

	createUserFieldsTableSQL := `CREATE TABLE adm_user_fields (
		usf_id integer  NOT NULL PRIMARY KEY,
		usf_name_intern varchar(30) NOT NULL		
	  );`

	// use db.Connection to avoid tweaking the bespoke query.
	sUserFields, err := db.Connection.Prepare(createUserFieldsTableSQL)
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

	// use db.Connection to avoid tweaking the bespoke query.
	_, fillUserFieldsError := db.Connection.Exec(fillUserFieldsSQL)
	if fillUserFieldsError != nil {
		return fillUserFieldsError
	}

	// usd_id | usd_usr_id | usd_usf_id | usd_value
	createUserDataTableSQL := `CREATE TABLE adm_user_data (
		usd_id integer NOT NULL PRIMARY KEY,
		usd_usr_id     integer NOT NULL,
		usd_usf_id     integer NOT NULL,
		usd_value      varchar(30) NOT NULL		
	  );`

	// use db.Connection to avoid tweaking the bespoke query.
	sUserData, err := db.Connection.Prepare(createUserDataTableSQL)
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

	// use db.Connection to avoid tweaking the bespoke query.
	_, fillUserDataError := db.Connection.Exec(fillUserDataSQL)
	if fillUserDataError != nil {
		return fillUserDataError
	}

	const createMembershipSalesSQL = `
		CREATE TABLE membership_sales
(
    ms_id integer NOT NULL,
    ms_payment_service CHARACTER VARYING(36) NOT NULL,
    ms_payment_status CHARACTER VARYING(20) NOT NULL,
    ms_payment_id CHARACTER VARYING(200),
    ms_membership_year integer NOT NULL,
    ms_usr1_id integer NOT NULL,
    ms_usr1_fee REAL NOT NULL,
    ms_usr1_friend boolean NOT NULL DEFAULT false,
    -- 0.0 if not a friend.
    ms_usr1_friend_fee REAL NOT NULL default 0.0,
    -- 0 if no associate
    ms_usr2_id integer DEFAULT 0,
    -- 0.0 if no associate
    ms_usr2_fee REAL NOT NULL default 0.0,
    -- false if no associate.
    ms_usr2_friend boolean NOT NULL DEFAULT false,
    -- 0.0 if no associate.
    ms_usr2_friend_fee REAL NOT NULL DEFAULT 0.0,
    -- 0.0 if no donation.
    ms_donation REAL NOT NULL DEFAULT 0.0,
    -- 0.0 if no donation to museum.
    ms_donation_museum REAL NOT NULL DEFAULT 0.0,
    ms_giftaid boolean NOT NULL DEFAULT false,
    timestamp_create varchar(20)
);
`
	// use db.Connection to avoid tweaking the bespoke query.
	membersCreate, membersPrepareError := db.Connection.Prepare(createMembershipSalesSQL)
	if membersPrepareError != nil {
		return membersPrepareError
	}
	_, membersCreateError := membersCreate.Exec() // Create Users table.
	if membersCreateError != nil {
		return membersCreateError
	}

	return nil
}
