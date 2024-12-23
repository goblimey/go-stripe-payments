package database

import (
	"fmt"
)

const TestUserIDPostgres = 9
const TestUserIDSQLite = 1
const TestEmail = "foo@bar.com"
const TestFirstName = "Luigi"
const TestLastName = "Schmidt"

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
		"usr_id" integer NOT NULL PRIMARY KEY,
		"usr_login_name" varchar(30) NOT NULL		
	  );`

	sUsers, err := db.Connection.Prepare(createUsersTableSQL)
	if err != nil {
		return err
	}
	_, createUsersError := sUsers.Exec() // Create Users table.
	if createUsersError != nil {
		return createUsersError
	}

	fillUsersSQL :=
		fmt.Sprintf("insert into adm_users(usr_id, usr_login_name) values(%d, '%s');",
			TestUserIDSQLite, TestEmail)

	_, fillUsersError := db.Connection.Exec(fillUsersSQL)
	if fillUsersError != nil {
		return fillUsersError
	}

	createRolesTableSQL := `CREATE TABLE adm_roles (
		"rol_id" integer NOT NULL PRIMARY KEY,
		"rol_name" varchar(20)		
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
		insert into adm_roles(rol_id, rol_name)
		values(2, "Member")
	`
	_, fillRolesError := db.Connection.Exec(fillRolesSQL)
	if fillRolesError != nil {
		return fillRolesError
	}

	// mem_id | mem_rol_id | mem_usr_id | mem_uuid | mem_begin | mem_end |
	// mem_leader | mem_usr_id_create | mem_timestamp_create |
	// mem_usr_id_change | mem_timestamp_change | mem_approved |
	// mem_comment | mem_count_guests
	createMembersTableSQL := `CREATE TABLE adm_members (
		"mem_id" integer NOT NULL PRIMARY KEY,
		"mem_usr_id" integer NOT NULL,
		"mem_rol_id" integer NOT NULL,
		"mem_begin" varchar(20),
		"mem_end" varchar(20)	
	  );`

	sMembers, membersError := db.Connection.Prepare(createMembersTableSQL)
	if membersError != nil {
		return membersError
	}
	_, createMembersError := sMembers.Exec() // Create table.
	if createMembersError != nil {
		return createMembersError
	}

	fillMembersSQL := fmt.Sprintf(`
		insert into adm_members(mem_id, mem_usr_id, mem_rol_id)
		values(1, %d, 2);`, TestUserIDSQLite)

	_, fillMembersError := db.Connection.Exec(fillMembersSQL)
	if fillMembersError != nil {
		return fillMembersError
	}

	createUserFieldsTableSQL := `CREATE TABLE adm_user_fields (
		"usf_id" integer NOT NULL PRIMARY KEY,
		"usf_name_intern" varchar(30) NOT NULL		
	  );`

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
	`

	_, fillUserFieldsError := db.Connection.Exec(fillUserFieldsSQL)
	if fillUserFieldsError != nil {
		return fillUserFieldsError
	}

	// usd_id | usd_usr_id | usd_usf_id | usd_value
	createUserDataTableSQL := `CREATE TABLE adm_user_data (
		"usd_id" integer NOT NULL PRIMARY KEY,
		"usd_usr_id" integer NOT NULL,
		"usd_usf_id" integer NOT NULL,
		"usd_value" varchar(30) NOT NULL		
	  );`

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
		TestUserIDSQLite, TestFirstName,
		TestUserIDSQLite, TestLastName,
		TestUserIDSQLite, TestEmail)

	_, fillUserDataError := db.Connection.Exec(fillUserDataSQL)
	if fillUserDataError != nil {
		return fillUserDataError
	}

	return nil
}
