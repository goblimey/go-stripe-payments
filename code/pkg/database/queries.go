package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// values for the ms_transaction_type field of the membershipsale database
// table.
const TransactionTypeNewMember = "new member"
const TransactionTypeRenewal = "membership renewal"
const PaymentStatusPending = "pending"
const PaymentStatusComplete = "complete"

var regExpForPostgresParamsToSQLiteParams *regexp.Regexp

// init should always work but if any of the calls in it fail, it will
// crash the application.
func init() {
	// Set up the regular expression or die.
	regExpForPostgresParamsToSQLiteParams = regexp.MustCompile(`\$[0-9]+`)
}

// CreateRole creates a role with the given name.
// It's assumed that a transaction is set up in the db object.
func (db *Database) CreateRole(role *Role) error {

	var uError error
	role.UUID, uError = CreateUuid(db.Transaction, "rol_uuid", "adm_roles")
	if uError != nil {
		return uError
	}

	// rol_adminstrator contains "t" or "f".
	adminFlag := "f"
	if role.Administrator {
		adminFlag = "t"
	}

	const sqlPostgres = `
	insert into adm_roles(
		rol_uuid, rol_name, rol_cat_id, rol_usr_id_create, 
		rol_administrator, rol_valid) 
		values($1, $2, $3, $4, $5,'t') 
		RETURNING rol_id;`

	const sqlSQLite = `
	insert into adm_roles(
		rol_uuid, rol_name, rol_cat_id, rol_usr_id_create, 
		rol_administrator, rol_valid) 
		values(?, ?, ?, ?, ?, 't');`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = sqlPostgres
	default:
		q = sqlSQLite
	}

	id, err := db.CreateRow(q, role.UUID, role.Name, role.RoleCategory.ID, role.RoleCategory.CreateUser.ID, adminFlag)
	if err != nil {
		return err
	}

	role.ID = id

	return nil
}

// GetRole gets the role with the given name.
func (db *Database) GetRole(name string) (*Role, error) {

	role := Role{}
	q := `select rol_id, rol_uuid, rol_name, rol_cat_id, rol_usr_id_create, rol_administrator, rol_valid from adm_roles where rol_name=$1;`
	row := db.QueryRow(q, name)
	if row.Err() != nil {
		return nil, row.Err()

	}
	var categoryID, createUserID int64
	var admin, valid string
	err := row.Scan(&role.ID, &role.UUID, &role.Name, &categoryID, &createUserID, &admin, &valid)
	if err != nil {
		return nil, err
	}

	if admin == "t" {
		role.Administrator = true
	}
	if valid == "t" {
		role.Valid = true
	}

	var catError error
	role.RoleCategory, catError = db.GetCategory(categoryID)
	if catError != nil {
		return nil, catError
	}

	var userError error
	role.CreateUser, userError = db.GetUser(createUserID)
	if userError != nil {
		return nil, userError
	}

	return &role, nil
}

// CreateOrganisation creates an organisation (adm_organization) with
// the given names.  It's assumed that a transaction is set up in the db object.
func (db *Database) CreateOrganisation(org *Organisation) error {

	var uError error
	org.UUID, uError = CreateUuid(db.Transaction, "org_uuid", "adm_organizations")
	if uError != nil {
		return uError
	}

	const sqlPostgres = `insert into adm_organizations(org_uuid, org_shortname, org_longname, org_homepage) 
		values($1, $2, $3, $4) RETURNING org_id;`
	const sqlSQLite = `insert into adm_organizations(org_uuid, org_shortname, org_longname, org_homepage) 
		values(?,?,?,?);`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = sqlPostgres
	default:
		q = sqlSQLite
	}

	var err error
	org.ID, err = db.CreateRow(q, org.UUID, org.Shortname, org.Longname, org.HomePage)
	if err != nil {
		return err
	}

	return nil
}

// GetOrganisationById gets the oganisation with the given ID.  If there is no such organisation
// it returns sql.ErrNoRow.
func (db *Database) GetOrganisationById(id int64) (*Organisation, error) {

	org := Organisation{}
	q := `select org_id, org_uuid, org_shortname, org_longname, org_homepage from adm_organizations where org_id=$1;`
	row := db.QueryRow(q, id)
	if row.Err() != nil {
		return nil, row.Err()
	}

	err := row.Scan(&org.ID, &org.UUID, &org.Shortname, &org.Longname, &org.HomePage)
	if err != nil {
		em := fmt.Sprintf("GetOrganisationById: %d not found - %v", id, err)
		return nil, errors.New(em)
	}

	// Return the zero-valued organisation.
	return &org, nil
}

// GetOrganisationsByShortName gets the oragniasation(s) with the given  shortname.  That should
// return a list containing just one organisation but in theory there could be more.
func (db *Database) GetOrganisationsByShortName(name string) ([]Organisation, error) {
	const q = `
		SELECT org_id, org_uuid, org_shortname, org_longname, org_homepage
		FROM adm_organizations
		WHERE org_shortname = $1;
	`

	orgs := make([]Organisation, 0)

	rows, err := db.Query(q, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return orgs, nil
		}
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}
		org := Organisation{}
		err := rows.Scan(&org.ID, &org.UUID, &org.Shortname, &org.Longname, &org.HomePage)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// CreateCategory creates an category (adm_categories) using the given category
// data.  It's assumed that a transaction is set up in the db object.
func (db *Database) CreateCategory(cat *Category) error {

	var uError error
	cat.UUID, uError = CreateUuid(db.Transaction, "cat_uuid", "adm_categories")
	if uError != nil {
		return uError
	}

	sys := "f"
	if cat.System {
		sys = "t"
	}
	df := "f"
	if cat.Default {
		df = "t"
	}

	// For postgress, the SQL is something like:
	// insert into adm_categories(cat_uuid, cat_type, cat_name_intern, cat_name,
	// cat_system, cat_default, cat_sequence, cat_org_id, cat_usr_id_create)
	// values($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING cat_id;
	//
	// However, cat.Org and/or cat.CreateUser may be nill, in which case we need an explicit NULL
	// value in the SQL.
	const sqlLeader = `
		insert into adm_categories(cat_uuid, cat_type, cat_name_intern, cat_name, 
		cat_system, cat_default, cat_sequence, cat_org_id, cat_usr_id_create)

	`
	var q string
	var err error
	switch {
	case cat.Org == nil && cat.CreateUser == nil:
		switch db.Config.Type {
		case "postgres":
			q = sqlLeader + `values($1, $2, $3, $4, $5, $6, $7, NULL, NULL) RETURNING cat_id;`
		default:
			q = sqlLeader + `values(?, ?, ?, ?, ?, ?, ?, NULL, NULL);`
		}
		cat.ID, err = db.CreateRow(
			q, cat.UUID, cat.Type, cat.NameIntern, cat.Name,
			sys, df, cat.Sequence,
		)

		if err != nil {
			return err
		}

	case cat.Org == nil:
		switch db.Config.Type {
		case "postgres":
			q = sqlLeader + `values($1, $2, $3, $4, $5, $6, $7, NULL, $8) RETURNING cat_id;`
		default:
			q = sqlLeader + `values(?, ?, ?, ?, ?, ?, ?, NULL, ?);`
		}

		cat.ID, err = db.CreateRow(
			q, cat.UUID, cat.Type, cat.NameIntern, cat.Name,
			sys, df, cat.Sequence, cat.CreateUser.ID,
		)

		if err != nil {
			return err
		}

	case cat.CreateUser == nil:
		switch db.Config.Type {
		case "postgres":
			q = sqlLeader + `values($1, $2, $3, $4, $5, $6, $7, $8, NULL) RETURNING cat_id;`
		default:
			q = sqlLeader + `values(?, ?, ?, ?, ?, ?, ?, ?, NULL);`
		}

		cat.ID, err = db.CreateRow(
			q, cat.UUID, cat.Type, cat.NameIntern, cat.Name,
			sys, df, cat.Sequence, cat.Org.ID,
		)

		if err != nil {
			return err
		}

	default:
		switch db.Config.Type {
		case "postgres":
			q = sqlLeader + `values($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING cat_id;`
		default:
			q = sqlLeader + `values(?, ?, ?, ?, ?, ?, ?, ?, ?);`

		}

		cat.ID, err = db.CreateRow(
			q, cat.UUID, cat.Type, cat.NameIntern, cat.Name,
			sys, df, cat.Sequence, cat.Org.ID, cat.CreateUser.ID,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// GetCategory gets the category with the given ID.  If there is no such category
// it returns sql.ErrNoRows.
func (db *Database) GetCategory(id int64) (*Category, error) {

	cat := Category{}
	// cat_org_id and cat_usr_id_create can be null.  If so, set to zero.
	qTemplate := `select cat_id, cat_uuid, %s(cat_org_id, 0), cat_type, cat_name_intern, cat_name, 
		cat_system, cat_default, cat_sequence, %s(cat_usr_id_create, 0) 
		FROM adm_categories
		WHERE cat_id=$1;`
	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(qTemplate, "COALESCE", "COALESCE")
	default:
		q = fmt.Sprintf(qTemplate, "IFNULL", "IFNULL")
	}
	row := db.QueryRow(q, id)

	var orgID, createUserID int64
	var system, def string

	err := row.Scan(&cat.ID, &cat.UUID, &orgID, &cat.Type, &cat.NameIntern, &cat.Name, &system, &def, &cat.Sequence, &createUserID)
	if err != nil {
		em := fmt.Sprintf("GetCategory: %d not found - %v", id, err)
		return nil, errors.New(em)
	}

	completeError := db.completeCategory(&cat, system, def, orgID, createUserID)
	if completeError != nil {
		em := fmt.Sprintf("cannot complete category %d - %v", cat.ID, completeError)
		slog.Error(em)
		return nil, completeError
	}

	// Success.  Return the organisation
	return &cat, nil
}

// GetCategoriessByName gets the category with the given name.
func (db *Database) GetCategoryByNameIntern(name string) (*Category, error) {
	// cat_org_id and cat_usr_id_create can be null.  If so, set to zero.
	const qTemplate = `select cat_id, cat_uuid, %s(cat_org_id, 0), cat_type, cat_name_intern, cat_name, 
		cat_system, cat_default, cat_sequence, %s(cat_usr_id_create, 0) 
		FROM adm_categories
		WHERE cat_name_intern=$1;`
	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(qTemplate, "COALESCE", "COALESCE")
	default:
		q = fmt.Sprintf(qTemplate, "IFNULL", "IFNULL")
	}

	var orgID, createUserID int64
	var system, def string
	cat := Category{}
	rows := db.QueryRow(q, name)
	err := rows.Scan(&cat.ID, &cat.UUID, &orgID, &cat.Type, &cat.NameIntern, &cat.Name,
		&system, &def, &cat.Sequence, &createUserID)
	if err != nil {
		return nil, err
	}

	completeError := db.completeCategory(&cat, system, def, orgID, createUserID)
	if completeError != nil {
		em := fmt.Sprintf("completeCategory: %d - %v", cat.ID, completeError)
		slog.Error(em)
		return &cat, completeError
	}

	return &cat, nil
}

// completeCategory is a helper function.  It fills in the fields of a category that are not easy
// to fetch directly from the database.
func (db *Database) completeCategory(category *Category, system, def string, organisationID, createUserID int64) error {

	if category == nil {
		return errors.New("cannot complete category - no category")
	}

	// Postgres return "true" and "false", SQLite returns "t"and "f".
	if system == "true" || system == "t" {
		category.System = true
	}
	if def == "true" || def == "t" {
		category.Default = true
	}

	if organisationID != 0 {
		// Get the embedded organisation.
		var err error
		category.Org, err = db.GetOrganisationById(organisationID)
		if err != nil {
			return errors.New("completeCategory:" + err.Error())
		}
	}
	if createUserID != 0 {
		// Get the embedded user.
		var err error
		category.CreateUser, err = db.GetUser(createUserID)
		if err != nil {
			return errors.New("completeCategory:" + err.Error())
		}
	}

	return nil
}

// CreateUser creates a user with the valid flag set.  The password is
// locked so they need to use the password change mechanism to log in.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) CreateUser(user *User) error {

	var uuidError error
	user.UUID, uuidError = CreateUuid(db.Transaction, "usr_uuid", "adm_users")
	if uuidError != nil {
		return uuidError
	}

	const postgresSQL = `
		insert into adm_users
		(usr_uuid, usr_login_name, usr_password, usr_valid)
		values($1, $2, '*LK*', 't')
		RETURNING usr_id;
	`
	const sqliteSQL = `
		insert into adm_users
		(usr_uuid, usr_login_name, usr_password, usr_valid)
		values(?, ?, '*LK*', 't');
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = postgresSQL
	default:
		q = sqliteSQL
	}

	id, createError := db.CreateRow(q, user.UUID, user.LoginName)

	if createError != nil {
		// This error will mess up the whole process, so log it.
		em := fmt.Sprintf("CreateUser: %s - %v", user.LoginName, createError)
		slog.Error(em)
		return createError
	}

	user.ID = id
	user.Valid = true

	return nil
}

// CreateUserWithNullPassword creates a user with a NULL password.  Such a user cannot log in
// to the Admidio website.  The user "System" is set up that way and we need to set up that
// user in the test environment.  The function assumes that a transaction is already set up
// in the db object.
func (db *Database) CreateUserWithNullPassword(user *User) error {

	var uuidError error
	user.UUID, uuidError = CreateUuid(db.Transaction, "usr_uuid", "adm_users")
	if uuidError != nil {
		return uuidError
	}

	const q = `
		insert into adm_users
		(usr_uuid, usr_login_name, usr_valid)
		values($1, $2, 't')
		RETURNING usr_id;
	`

	id, createError := db.CreateRow(q, user.UUID, user.LoginName)

	if createError != nil {
		return createError
	}

	user.ID = id
	user.Valid = true

	return nil
}

// UpdateUser saves a (presumably) updated User.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) UpdateUser(user *User) error {

	const sql = `
		UPDATE adm_users 
		SET usr_uuid=$1, usr_login_name=$2, usr_password=$3, usr_valid=$4
		WHERE usr_id=$5;
	`

	var valid string = "f"
	if user.Valid {
		valid = "t"
	}

	_, updateError := db.Exec(sql, user.UUID, user.LoginName, user.Password, valid, user.ID)

	if updateError != nil {
		return updateError
	}

	return nil
}

// GetUser gets the user with the given ID.  It's assumed that a transaction is already
// set up in the db object.
func (db *Database) GetUser(id int64) (*User, error) {

	// The password may be null (which prevents login).
	// Postgres uses COALESCE to convert NULL to a readable value, SQLite uses IFNULL
	const queryTemplate = `
		SELECT usr_id, usr_uuid, usr_login_name, %s(usr_password, ''), usr_valid
		FROM adm_users
		WHERE usr_id = $1;
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(queryTemplate, "COALESCE")
	default:
		q = fmt.Sprintf(queryTemplate, "IFNULL")
	}

	u := NewUser("name to be overwritten")

	err := db.QueryRow(q, id).Scan(&u.ID, &u.UUID, &u.LoginName, &u.Password, &u.Valid)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetUsers gets all of the users.  It's assumed that a transaction is already set up
// in the db object.
func (db *Database) GetUsers() ([]User, error) {

	// The password may be null (which prevents login).
	// Postgres uses COALESCE to convert NULL to a readable value, SQLite uses IFNULL
	const queryTemplate = `
		SELECT usr_id, usr_uuid, usr_login_name, %s(usr_password, ''), usr_valid
		FROM adm_users
		ORDER BY usr_login_name;
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(queryTemplate, "COALESCE")
	default:
		q = fmt.Sprintf(queryTemplate, "IFNULL")
	}

	users := make([]User, 0)

	rows, err := db.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}
		user := User{LoginName: "junk name to be replaced"}
		err := rows.Scan(&user.ID, &user.UUID, &user.LoginName, &user.Password, &user.Valid)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUsersByLoginName gets the users with the given login name.  The search is case-insensitive.
// The result should be a list containing just one user but in theory there could be more.  It's
// assumed that a transaction is already set up in the db object.
func (db *Database) GetUsersByLoginName(name string) ([]User, error) {

	// The password may be null (which prevents login).
	// Postgres uses COALESCE to convert NULL to a readable value, SQLite uses IFNULL.
	// The name comarison is case-insensitive because it may be supplied by the user
	// and they may type it differently in different years.
	const queryTemplate = `
		SELECT usr_id, usr_uuid, %s(usr_password, ''), usr_valid
		FROM adm_users
		WHERE lower(usr_login_name) = lower($1);
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(queryTemplate, "COALESCE")
	default:
		q = fmt.Sprintf(queryTemplate, "IFNULL")
	}

	users := make([]User, 0)

	rows, err := db.Query(q, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}
		user := User{LoginName: name}
		err := rows.Scan(&user.ID, &user.UUID, &user.Password, &user.Valid)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByLoginName gets the user with the given login name (case-insensitive).
// That field has a unique constraint so there will be at most one result.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetUserByLoginName(name string) (*User, error) {

	// The password may be null (which prevents login).
	const queryTemplate = `
		SELECT usr_id, usr_uuid, %s(lower(usr_password), ''), usr_valid
		FROM adm_users
		WHERE lower(usr_login_name) = lower($1);
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = fmt.Sprintf(queryTemplate, "COALESCE")
	default:
		q = fmt.Sprintf(queryTemplate, "IFNULL")
	}

	user := NewUser(name)
	err := db.QueryRow(q, name).Scan(&user.ID, &user.UUID, &user.Password, &user.Valid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes the given user from the database and sets the ID in the object
// to zero.  It's assumed that a transaction is already set up in the db object.
func (db *Database) DeleteUser(u *User) error {
	const q = `
		DELETE FROM adm_users
		WHERE usr_id = $1;
	`

	numRows, err := db.DeleteRow(q, u.ID)

	if err != nil {
		msg := fmt.Sprintf("DeleteUser: error deleting user with ID %d", u.ID)
		return errors.New(msg)
	}
	if numRows != 1 {
		msg := fmt.Sprintf("DeleteUser: returned ID %d not 1", numRows)
		return errors.New(msg)
	}

	// Success!
	u.ID = 0
	return nil
}

// NewUserField creates a user field using the given data.
func NewUserField(name, nameIntern, fieldType string, user *User, cat *Category) *FieldData {
	uf := FieldData{
		NameIntern: nameIntern,
		Name:       name,
		Type:       fieldType,
		CreateUser: user,
		Cat:        cat,
	}
	return &uf
}

// CreateUserField creates a user field in the database using the give field data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) CreateUserField(uf *FieldData) error {

	var uError error
	uf.UUID, uError = CreateUuid(db.Transaction, "usf_uuid", "adm_user_fields")
	if uError != nil {
		return uError
	}

	const postgresSQL = `
		insert into adm_user_fields
		(usf_uuid, usf_name_intern, usf_name, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
		values($1, $2, $3, $4, $5, $6, $7)
		RETURNING usf_id;
	`
	const sqliteSQL = `
		insert into adm_user_fields
		(usf_uuid, usf_name_intern, usf_name, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
		values(?,?,?,?,?,?,?);
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = postgresSQL
	default:
		q = sqliteSQL
	}

	id, createError := db.CreateRow(
		q, uf.UUID, uf.NameIntern, uf.Name, uf.Type, uf.Sequence, uf.Cat.ID, uf.CreateUser.ID)

	if createError != nil {
		return createError
	}

	uf.ID = int64(id)

	return nil
}

// GetUserDataFieldByNameIntern gets the user data field with the given internal name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetUserDataFieldByNameIntern(name string) (*FieldData, error) {

	const q = `
		SELECT usf_id, usf_uuid, usf_name_intern, usf_name, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create
		FROM adm_user_fields
		WHERE usf_name_intern = $1;
	`

	var catID, userID int64
	uf := NewUserField("name to be overwritten", "", "", nil, nil)

	err := db.QueryRow(q, name).Scan(
		&uf.ID, &uf.UUID, &uf.NameIntern, &uf.Name, &uf.Type, &uf.Sequence, &catID, &userID)
	if err != nil {
		return nil, err
	}

	// Fetch the embedded category.
	var fetchCatError error
	uf.Cat, fetchCatError = db.GetCategory(catID)
	if fetchCatError != nil {
		return nil, fetchCatError
	}

	// Fetch the embedded user.
	var ue error
	uf.CreateUser, ue = db.GetUser(userID)
	if ue != nil {
		return nil, ue
	}

	return uf, nil
}

// Create creates a MembershipSale record in the database.
// It's assumed that a transaction is already set up in the db object.
func (ms *MembershipSale) Create(db *Database) (int64, error) {

	// Postgres doesn't support LastInsertId.  We use the form recommended here:
	// https://github.com/jackc/pgx/issues/1483, which is
	// err := db.QueryRow("INSERT INTO user (name) VALUES ('John') RETURNING id").Scan(&id)
	//
	// Initially the payment_id may be an empty string.   It will be set to the
	// Stripe transaction ID later.
	//
	// Postgres uses a sequence to give  the next ID, SQLite uses autoincrement.
	// Postgres: Insert into membership_sales(ms_id, ms_payment_service, ...)
	//           VALUES(nextval('membership_sales_ms_id_seq'), $1, ...) RETURNING ms_id;
	// SQLite:   Insert into membership_sales(ms_payment_service, $1, $2...)
	//           VALUES(?, ?, ...);
	//
	// If the ID of the ordinary user or the associate user is not given, that column
	// is set to NULL in the table.  That satisfies the relational constraint that
	// the value must be an ID in the adm_users table.

	var id int64
	var createError error
	switch {
	case ms.UserID <= 0 && ms.AssocUserID <= 0:
		// Both foreign keys should be NULL.
		const sqlTemplate = `
				INSERT INTO membership_sales (
					%s
					ms_usr1_id,
					ms_usr2_id,

					ms_payment_service,
					ms_payment_status,
					ms_payment_id,
					ms_transaction_type,
					ms_membership_year,
					ms_usr1_fee,
					ms_usr1_friend,
					ms_usr1_friend_fee,
					ms_usr1_title,
					ms_usr1_first_name,

					ms_usr1_last_name,
					ms_usr1_email,
					ms_usr2_fee,
					ms_usr2_friend,
					ms_usr2_friend_fee,
					ms_usr2_title,
					ms_usr2_first_name,
					ms_usr2_last_name,
					ms_usr2_email,
					ms_donation,

					ms_donation_museum,
					ms_giftaid
				)
				VALUES
				(
					%s
					NULL, NULL,
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
					$16, $17, $18, $19, $20, $21, $22
				)
				%s;
			`

		sql := ms.fixSaleFormCreateStatement(db, sqlTemplate)

		id, createError = db.CreateRow(
			sql,
			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.OrdinaryMemberFeePaid,
			ms.Friend,
			ms.FriendFeePaid,
			ms.Title,
			ms.FirstName,

			ms.LastName,
			ms.Email,
			ms.AssocFeePaid,
			ms.AssocFriend,
			ms.AssocFriendFeePaid,
			ms.AssocTitle,
			ms.AssocFirstName,
			ms.AssocLastName,
			ms.AssocEmail,
			ms.DonationToSociety,

			ms.DonationToMuseum,
			ms.Giftaid,
		)

	case ms.AssocUserID <= 0:
		// The user's foreign key is set but the associate foreign key should be null.
		// This is an ordinary user with no associate.
		const sqlTemplate = `
				INSERT INTO membership_sales (
					%s
					ms_usr2_id,

					ms_payment_service,
					ms_payment_status,
					ms_payment_id,
					ms_transaction_type,
					ms_membership_year,
					ms_usr1_id,
					ms_usr1_fee,
					ms_usr1_friend,
					ms_usr1_friend_fee,
					ms_usr1_title,

					ms_usr1_first_name,
					ms_usr1_last_name,
					ms_usr1_email,
					ms_usr2_fee,
					ms_usr2_friend,
					ms_usr2_friend_fee,
					ms_usr2_title,
					ms_usr2_first_name,
					ms_usr2_last_name,
					ms_usr2_email,

					ms_donation,
					ms_donation_museum,
					ms_giftaid
				)
				VALUES
				(
					%s
					NULL,
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
					$16, $17, $18, $19, $20, $21, $22, $23
				)
				%s;
			`

		sql := ms.fixSaleFormCreateStatement(db, sqlTemplate)

		id, createError = db.CreateRow(
			sql,

			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.UserID,
			ms.OrdinaryMemberFeePaid,
			ms.Friend,
			ms.FriendFeePaid,
			ms.Title,

			ms.FirstName,
			ms.LastName,
			ms.Email,
			ms.AssocFeePaid,
			ms.AssocFriend,
			ms.AssocFriendFeePaid,
			ms.AssocTitle,
			ms.AssocFirstName,
			ms.AssocLastName,
			ms.AssocEmail,

			ms.DonationToSociety,
			ms.DonationToMuseum,
			ms.Giftaid,
		)

	case ms.UserID <= 0:
		// Edge case.  The user's foreign key should be null but the associate foreign
		// key is set.  (This should not happen in this app.)
		const sqlTemplate = `
				INSERT INTO membership_sales (
					%s
					ms_usr1_id,

					ms_payment_service,
					ms_payment_status,
					ms_payment_id,
					ms_transaction_type,
					ms_membership_year,
					ms_usr1_fee,
					ms_usr1_friend,
					ms_usr1_friend_fee,
					ms_usr1_title,
					ms_usr1_first_name,

					ms_usr1_last_name,
					ms_usr1_email,
					ms_usr2_id,
					ms_usr2_fee,
					ms_usr2_friend,
					ms_usr2_friend_fee,
					ms_usr2_title,
					ms_usr2_first_name,
					ms_usr2_last_name,
					ms_usr2_email,

					ms_donation,
					ms_donation_museum,
					ms_giftaid
				)
				VALUES
				(
					%s
					NULL,
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
					$16, $17, $18, $19, $20, $21, $22, $23
				)
				%s;
			`

		sql := ms.fixSaleFormCreateStatement(db, sqlTemplate)

		id, createError = db.CreateRow(
			sql,

			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.OrdinaryMemberFeePaid,
			ms.Friend,
			ms.FriendFeePaid,
			ms.Title,
			ms.FirstName,

			ms.LastName,
			ms.Email,
			ms.AssocUserID,
			ms.AssocFeePaid,
			ms.AssocFriend,
			ms.AssocFriendFeePaid,
			ms.AssocTitle,
			ms.AssocFirstName,
			ms.AssocLastName,
			ms.AssocEmail,

			ms.DonationToSociety,
			ms.DonationToMuseum,
			ms.Giftaid,
		)

	default:
		const sqlTemplate = `
			INSERT INTO membership_sales (
				%s 

				ms_payment_service,
				ms_payment_status,
				ms_payment_id,
				ms_transaction_type,
				ms_membership_year,
				ms_usr1_id,
				ms_usr1_fee,
				ms_usr1_friend,
				ms_usr1_friend_fee,
				ms_usr1_title,

				ms_usr1_first_name,
				ms_usr1_last_name,
				ms_usr1_email,
				ms_usr2_id,
				ms_usr2_fee,
				ms_usr2_friend,
				ms_usr2_friend_fee,
				ms_usr2_title,
				ms_usr2_first_name,
				ms_usr2_last_name,

				ms_usr2_email,
				ms_donation,
				ms_donation_museum,
				ms_giftaid       
			) 
			VALUES
			(
				%s 
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
				$15, $16, $17, $18, $19, $20, $21, $22, $23, $24
			)
			%s;
		`

		sql := ms.fixSaleFormCreateStatement(db, sqlTemplate)

		id, createError = db.CreateRow(
			sql,

			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.UserID,
			ms.OrdinaryMemberFeePaid,
			ms.Friend,
			ms.FriendFeePaid,
			ms.Title,

			ms.FirstName,
			ms.LastName,
			ms.Email,
			ms.AssocUserID,
			ms.AssocFeePaid,
			ms.AssocFriend,
			ms.AssocFriendFeePaid,
			ms.AssocTitle,
			ms.AssocFirstName,
			ms.AssocLastName,

			ms.AssocEmail,
			ms.DonationToSociety,
			ms.DonationToMuseum,
			ms.Giftaid,
		)
	}

	if createError == nil {
		// Set the id in the membership sales object.
		ms.ID = id
	}

	return id, createError

}

// fixMembershipSaleCreateStatement takes an Sprintf template with three string placeholders
// and expands and returns it.
func (ms *MembershipSale) fixSaleFormCreateStatement(db *Database, createStatementPattern string) string {
	// Postgres uses a sequence to give  the next ID, SQLite uses autoincrement.
	// Postgres: Insert into membership_sales(ms_id, ms_payment_service, ...)
	//           VALUES(nextval('membership_sales_ms_id_seq'), $1, ...) RETURNING ms_id;
	// SQLite:   Insert into membership_sales(ms_payment_service, $1, $2...)
	//           VALUES(?, ?, ...);
	//
	// For SQLite, $1, $2 ... are converted to ?, ? ... later.
	var result string
	switch db.Config.Type {
	case "postgres":
		result =
			fmt.Sprintf(createStatementPattern, "ms_id,", "nextval('membership_sales_ms_id_seq'),", "RETURNING ms_id")
	default:
		result = fmt.Sprintf(createStatementPattern, "", "", "")
	}
	return result
}

// GetMembershipSale gets the membership_sale record for the user with
// the given ID.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMembershipSale(id int64) (*MembershipSale, error) {

	// Postgres uses COALESCE to convert NULL to a readable value, SQLite uses IFNULL
	const queryTemplate = `
	SELECT 
		ms_id,
		ms_payment_service,
		ms_payment_status,
		ms_payment_id,
		ms_membership_year,
		ms_transaction_type,
		%s(ms_usr1_id, 0),
		%s(ms_usr1_title, ''),
		ms_usr1_first_name,
		ms_usr1_last_name,

		ms_usr1_email,
		ms_usr1_fee,
		ms_usr1_friend,
		ms_usr1_friend_fee,
		ms_donation,
		ms_donation_museum,
		ms_giftaid,
		%s(ms_usr2_id, 0),
		%s(ms_usr2_title, ''),
		%s(ms_usr2_first_name, ''),
		%s(ms_usr2_last_name, ''),
		%s(ms_usr2_email, ''),
		ms_usr2_fee,
		ms_usr2_friend,
		ms_usr2_friend_fee
		
	FROM membership_sales
	WHERE ms_id = $1;
`

	var query string
	switch db.Config.Type {
	case "postgres":
		query = fmt.Sprintf(queryTemplate, "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE")
		// query = fmt.Sprintf(queryTemplate, "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE")
		// query = fmt.Sprintf(queryTemplate, "COALESCE", "COALESCE", "COALESCE", "COALESCE", "COALESCE")
	default:
		query = fmt.Sprintf(queryTemplate, "IFNULL", "IFNULL", "IFNULL", "IFNULL", "IFNULL", "IFNULL", "IFNULL")
		// query = fmt.Sprintf(queryTemplate, "IFNULL", "IFNULL", "IFNULL", "IFNULL", "IFNULL")
	}

	row, searchErr := db.Query(query, id)
	if searchErr != nil {
		return nil, searchErr
	}
	defer row.Close()

	if !row.Next() {
		return nil, errors.New("GetMembershipSale: no matching record")
	}

	var ms MembershipSale

	err := row.Scan(
		&ms.ID,
		&ms.PaymentService,
		&ms.PaymentStatus,
		&ms.PaymentID,
		&ms.MembershipYear,
		&ms.TransactionType,
		&ms.UserID,
		&ms.Title,
		&ms.FirstName,
		&ms.LastName,

		&ms.Email,
		&ms.OrdinaryMemberFeePaid,
		&ms.Friend,
		&ms.FriendFeePaid,
		&ms.DonationToSociety,
		&ms.DonationToMuseum,
		&ms.Giftaid,
		&ms.AssocUserID,
		&ms.AssocTitle,
		&ms.AssocFirstName,

		&ms.AssocLastName,
		&ms.AssocEmail,
		&ms.AssocFeePaid,
		&ms.AssocFriend,
		&ms.AssocFriendFeePaid,
	)
	if err != nil {
		return nil, err
	}

	return &ms, nil

}

// fixMembershipSaleCreateStatement takes an Sprintf template with three string placeholders
// and expands and returns it.
func (ms *MembershipSale) fixMembershipSaleCreateStatement(db *Database, createStatementPattern string) string {
	// Postgres uses a sequence to give  the next ID, SQLite uses autoincrement.
	// Postgres: Insert into membership_sales(ms_id, ms_payment_service, ...)
	//           VALUES(nextval('membership_sales_ms_id_seq'), $1, ...) RETURNING ms_id;
	// SQLite:   Insert into membership_sales(ms_payment_service, $1, $2...)
	//           VALUES(?, ?, ...);
	//
	// For SQLite, $1, $2 ... are converted to ?, ? ... later.
	var result string
	switch db.Config.Type {
	case "postgres":
		result =
			fmt.Sprintf(createStatementPattern, "ms_id,", "nextval('membership_sales_ms_id_seq'),", "RETURNING ms_id")
	default:
		result = fmt.Sprintf(createStatementPattern, "", "", "")
	}
	return result
}

// Update updates a MembershipSale record in the database from the data supplied
// in the Membeship.
// It's assumed that a transaction is already set up in the db object.
func (ms *MembershipSale) Update(db *Database) error {

	// Booleans are stored in the database as "t" or "f".
	friend := "f"
	if ms.Friend {
		friend = "t"
	}

	giftaid := "f"
	if ms.Giftaid {
		giftaid = "t"
	}

	var rowsAffected int64
	var createError error

	// There is an ordinary member and an associate member.
	const sql = `
			UPDATE membership_sales SET
				ms_payment_service = $1,
				ms_payment_status = $2,
				ms_payment_id = $3,
				ms_transaction_type = $4,
				ms_membership_year = $5,
				ms_usr1_id = NULLIF($6, 0),
				ms_usr1_fee = $7,
				ms_usr1_friend = $8,
				ms_usr1_friend_fee = $9,
				ms_usr1_title = $10,

				ms_usr1_first_name = $11,
				ms_usr1_last_name = $12,
				ms_usr1_email = $13,
				ms_usr2_id = NULLIF($14, 0),
				ms_usr2_fee = $15,
				ms_usr2_friend = $16,
				ms_usr2_friend_fee = $17,
				ms_usr2_title = $18,
				ms_usr2_first_name = $19,
				ms_usr2_last_name = $20,

				ms_usr2_email = $21,
				ms_donation=  $22,
				ms_donation_museum = $23,
				ms_giftaid = $24

			WHERE ms_id=$25;
		`

	rowsAffected, createError = db.UpdateRow(
		sql,
		ms.PaymentService,
		ms.PaymentStatus,
		ms.PaymentID,
		ms.TransactionType,
		ms.MembershipYear,
		ms.UserID,
		ms.OrdinaryMemberFeePaid,
		friend,
		ms.FriendFeePaid,
		ms.Title,

		ms.FirstName,
		ms.LastName,
		ms.Email,
		ms.AssocUserID,
		ms.AssocFeePaid,
		ms.AssocFriend,
		ms.AssocFriendFeePaid,
		ms.AssocTitle,
		ms.AssocFirstName,
		ms.AssocLastName,

		ms.AssocEmail,
		ms.DonationToSociety,
		ms.DonationToMuseum,
		giftaid,

		ms.ID, // for the WHERE clause.
	)

	if createError != nil {
		return createError
	}

	if rowsAffected != 1 {
		// Too many rows affected - should be only one.
		return fmt.Errorf("update affected %d rows - expected just 1", rowsAffected)
	}

	// Success.
	return nil
}

// Delete deletes a MembershipSale record in the database.
// It's assumed that a transaction is already set up in the db object.
func (ms *MembershipSale) Delete(db *Database) error {

	q := `
		DELETE FROM MEMBERSHIP_SALES 
		WHERE ms_id=$1;
	`

	id := ms.ID
	numRows, execAndScanError := db.DeleteRow(q, ms.ID)
	if execAndScanError != nil {
		return execAndScanError
	}

	if numRows != 1 {
		em := fmt.Sprintf(
			"MembershipSale.Delete: deleting ID %d, want 1 row deleted got %d", id, numRows)
		return errors.New(em)
	}

	// Set the ID in the struct to zero to mark it as deleted.
	ms.ID = 0

	// Success.
	return nil
}

// NewMember creates a Member object from the given data.
func NewMember(user *User, role *Role, startTime, endTime time.Time) *Member {

	m := Member{
		UserID:    user.ID,
		RoleID:    role.ID,
		StartDate: startTime.Format("2006-01-02"),
		EndDate:   endTime.Format("2006-01-02"),
	}

	return &m
}

// CreateMember creates a member record from the given member data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) CreateMember(member *Member) error {

	var uError error
	member.UUID, uError = CreateUuid(db.Transaction, "mem_uuid", "adm_members")
	if uError != nil {
		return uError
	}

	// Create a member record with the approved flag set.
	const q = `
		insert into adm_members(mem_uuid, mem_usr_id, mem_rol_id, mem_begin, mem_end, mem_approved) 
		values($1, $2, $3, $4, $5, $6) RETURNING mem_id;
	`

	id, err := db.CreateRow(
		q,
		member.UUID,
		member.UserID,
		member.RoleID,
		member.StartDate,
		member.EndDate,
		1, // Approve flag.
	)

	if err != nil {
		return err
	}

	member.ID = id

	return nil
}

// DeleteMember deletes the given member from the database and sets the ID in the
// object to zero.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) DeleteMember(m *Member) error {
	q := `
		DELETE FROM adm_members
		WHERE mem_id = $1;
	`

	got, err := db.DeleteRow(q, m.ID)

	if err != nil {
		msg := fmt.Sprintf("DeleteMember: error deleting member with ID %d", m.ID)
		return errors.New(msg)
	}
	if got != 1 {
		msg := fmt.Sprintf("DeleteUser: want num rows deleted 1 got %d", got)
		return errors.New(msg)
	}

	// Success!
	m.ID = 0
	return nil
}

// GetMember gets the member with the given ID.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMember(id int64) (*Member, error) {
	const q = `
		SELECT mem_id, mem_uuid, mem_rol_id, mem_usr_id, mem_begin, mem_end, mem_approved
		FROM adm_members
		WHERE mem_id = $1;
	`

	// Create a new member object.  Any values passed will be overwritten by the scan.
	m := NewMember(nil, nil, time.Now(), time.Now())

	err := db.QueryRow(q, id).Scan(&m.ID, &m.UUID, &m.RoleID, &m.UserID, &m.StartDate, &m.EndDate)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// GetInterests get the list of interests from adm_members_interest.
func (db *Database) GetInterests() ([]Interest, error) {

	mis := make([]Interest, 0)

	q := `
		select ntrst_id, ntrst_name 
		from adm_interests
		order by ntrst_id;
	`

	rows, err := db.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			// The query worked but the table is empty.
			return mis, nil
		}
		// the query failed.
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}

		var mi Interest
		err := rows.Scan(&mi.ID, &mi.Name)
		if err != nil {
			return nil, err
		}
		mis = append(mis, mi)
	}

	return mis, nil
}

// CreateMembersInterest links a user to an interest if a linkage does not
// already exist.  If the linkage already exists, the ID will remain set to
// zero. It's assumed that a transaction is set up in the db object.
func (db *Database) CreateMembersInterest(mi *MembersInterest) error {

	// Insert the record if not already done (upsert).

	const qPostgres = `
	insert into adm_members_interests(mi_usr_id, mi_interest_id) 
		values($1, $2)
		on conflict(mi_usr_id, mi_interest_id) do nothing
		RETURNING mi_id;`

	const qSQLite = `
	insert into adm_members_interests(mi_usr_id, mi_interest_id)
		values(?, ?)
		on conflict(mi_usr_id, mi_interest_id) do nothing;`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = qPostgres
	default:
		q = qSQLite
	}

	// If the database is postgres and the unique constraint prevents the insert,
	// the attempt to get the mi_id will fail and produce a "no rows" error.  That's
	// expected behaviour.  Leave mi.ID set to zero.
	id, err := db.CreateRow(q, mi.UserID, mi.InterestID)
	if err != nil {
		if db.Config.Type == "postgres" && err.Error() == "sql: no rows in result set" {
			return nil
		}
		return err
	}

	mi.ID = id

	return nil
}

// GetMembersInterestsForUser gets the adm_member_interest rows for the given user.
// It's assumed that a transaction is already set up.
func (db *Database) GetMembersInterests(userID int64) ([]MembersInterest, error) {

	mis := make([]MembersInterest, 0)
	// Ordering the result helps testing.
	q := `
		select mi_id, mi_usr_id, mi_interest_id 
		from adm_members_interests 
		where mi_usr_id=$1
		order by mi_id;
	`

	rows, err := db.Query(q, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return mis, nil
		}
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}

		var mi MembersInterest
		err := rows.Scan(&mi.ID, &mi.UserID, &mi.InterestID)
		if err != nil {
			return nil, err
		}
		mis = append(mis, mi)
	}

	return mis, nil
}

// GetMembersOtherInterestsForUser gets the members other interests for the user with the
// given ID.  It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMembersOtherInterests(userID int64) (*MembersOtherInterests, error) {

	const query = `
		SELECT moi_id, moi_usr_id, moi_interests
		FROM adm_members_other_interests
		WHERE moi_usr_id = $1;
	`

	var moi MembersOtherInterests

	err := db.QueryRow(query, userID).Scan(&moi.ID, &moi.UserID, &moi.Interests)
	if err != nil {
		return nil, err
	}

	return &moi, nil
}

// CreateMembersOtherInterests creates the given user's other interests record.
//
//	The function assumes that a transaction is already set up in the db object.
func (db *Database) CreateMembersOtherInterests(moi *MembersOtherInterests) error {

	// Insert the interests if not already done (upsert).
	const qPostgres = `
		insert into adm_members_other_interests
		(moi_usr_id,moi_interests)
		values($1, $2)
		RETURNING moi_id;
	`

	const qSQLite = `
		insert into adm_members_other_interests
		(moi_usr_id,moi_interests)
		values(?, ?);
	`

	var q string
	switch db.Config.Type {
	case "postgres":
		q = qPostgres
	default:
		q = qSQLite
	}

	id, createError := db.CreateRow(q, moi.UserID, moi.Interests)

	if createError != nil {
		return createError
	}

	moi.ID = id

	return nil
}

// UpdateMembersOtherInterests updates the given user's other interests.
func (db *Database) UpdateMembersOtherInterests(moi *MembersOtherInterests) error {
	const sql = `
		UPDATE adm_members_other_interests 
		SET moi_interests=$1
		WHERE moi_usr_id=$2;
	`

	_, updateError := db.Exec(sql, moi.Interests, moi.UserID)

	if updateError != nil {
		return updateError
	}

	return nil
}

// Update the users other interests or, if a record already exists, update it.
// The function assumes that a transaction has been set up.
func (db *Database) UpsertMembersOtherInterests(moi *MembersOtherInterests) error {
	moiFetched, fetchError := db.GetMembersOtherInterests(moi.UserID)
	if fetchError == nil {
		// There is already a record.  Update it.
		moiFetched.Interests = moi.Interests
		err := db.UpdateMembersOtherInterests(moiFetched)
		if err != nil {
			return err
		}
	} else {
		// Thee is no record yet.  Create one.
		err := db.CreateMembersOtherInterests(moi)
		if err != nil {
			return err
		}
	}

	// Success.
	return nil
}

// CreateUserAndMember creates a user and associated records (member, first name field etc).
// It's assumed that a transaction is already set up in the db object.
func (db *Database) CreateUserAndMember(loginName, title, firstName, lastName string, role *Role, startTime, endTime time.Time) (*User, *Member, error) {

	user := NewUser(loginName)
	uError := db.CreateUser(user)
	if uError != nil {
		return nil, nil, uError
	}

	em, eme := db.GetUserDataFieldIDByNameIntern("EMAIL")
	if eme != nil {
		return nil, nil, eme
	}
	emError := SetUserDataField(db, em, user.ID, loginName)
	if emError != nil {
		return nil, nil, emError
	}

	// Set Title (Mr, Ms, Dr etc) - Salutation.
	it, tne := db.GetUserDataFieldIDByNameIntern("SALUTATION")
	if tne != nil {
		return nil, nil, tne
	}
	tError := SetUserDataField(db, it, user.ID, title)
	if tError != nil {
		return nil, nil, tError
	}

	ifn, fne := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
	if fne != nil {
		return nil, nil, fne
	}
	fnError := SetUserDataField(db, ifn, user.ID, firstName)
	if fnError != nil {
		return nil, nil, fnError
	}

	member := NewMember(user, role, startTime, endTime)
	mError := db.CreateMember(member)
	if mError != nil {
		return nil, nil, mError
	}

	iln, lne := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
	if lne != nil {
		return nil, nil, lne
	}
	snError := SetUserDataField(db, iln, user.ID, lastName)
	if snError != nil {
		return nil, nil, snError
	}

	return user, member, nil
}

// GetMemberOfUser gets the adm_members record with role 'Member' associated
// with the user with the given ID.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMemberOfUser(user *User) (*Member, error) {

	// Create a new member object.  Any values passed will be overwritten by the scan.
	u := NewUser("to be removed")
	var role Role
	m := NewMember(u, &role, time.Now(), time.Now())

	const q = `
		SELECT m.mem_id, m.mem_uuid, m.mem_rol_id, m.mem_usr_id, 
			m.mem_begin, m.mem_end, m.mem_approved
		FROM adm_members as m
		LEFT JOIN adm_roles as r
			ON r.rol_id = m.mem_rol_id
			AND r.rol_name = 'Member'
		WHERE m.mem_usr_id = $1;
	`

	err := db.QueryRow(q, user.ID).Scan(&m.ID, &m.UUID, &m.RoleID, &m.UserID, &m.StartDate, &m.EndDate, &m.Approved)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// GetMembers gets all of the members.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMembers() ([]Member, error) {

	role, roleError := db.GetRole("Member")
	if roleError != nil {
		return nil, roleError
	}

	const q = `
		SELECT mem_id, mem_uuid, mem_usr_id, mem_rol_id,
			mem_begin, mem_end, mem_approved
		FROM adm_members
		ORDER BY mem_id;
	`

	members := make([]Member, 0)

	rows, err := db.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			return members, nil
		}
		return nil, err
	}
	defer rows.Close()

	for {
		if !rows.Next() {
			break
		}
		user := NewUser("name will be overwritten")
		member := NewMember(user, role, time.Now(), time.Now())
		err := rows.Scan(&member.ID, &member.UUID, &member.UserID, &member.RoleID,
			&member.StartDate, &member.EndDate, &member.Approved)
		if err != nil {
			return nil, err
		}
		members = append(members, *member)
	}

	return members, nil
}

// CreateAccounts creates accounts for an ordinary member and, if given, for an
// associate member.  Each account is represented by a record in the adm_users table,
// a linked record in adm_members with role Member and, if an email address is
// supplied, a linked record in adm_user_data giving the user's email address (which
// is required to change their password).  The given membership sale record supplies
// the data.  It's assumed that the db object contains a transaction.  The ID of the
// ordinary user is returned.
//
// If two members live at the same address it's quite common for them to both use the
// same email address or for the associate not to supply an email address but the system
// only works properly if each email address in the adm_users records is unique - if not,
// the password change process doesn't work.  So both members give the same email address,
// it can only be used in one record.  If the email address is not given or it's already
// used, the login name is formed from their first and last name ("first.last").
// HOWEVER without an email address they can't get control of their account by setting
// their password.  Their record just marks that they are a paid-up member.
//
// It's assumed that a transaction is already set up in the db object.
func (db *Database) CreateAccounts(sale *MembershipSale, now time.Time, endTime time.Time) (int64, int64, error) {

	// Get the login name(s) from the sale.
	name, namesError := getLoginNames(sale)
	if namesError != nil {
		return 0, 0, namesError
	}

	// Create the user for the ordinary account.
	ordinaryUser := NewUser(name[0])
	createOrdinaryUserError := db.CreateUser(ordinaryUser)
	if createOrdinaryUserError != nil {
		return 0, 0, createOrdinaryUserError
	}

	// Update the sale.
	sale.UserID = ordinaryUser.ID

	roleMember, roleError := db.GetRole("Member")
	if roleError != nil {
		return 0, 0, roleError
	}
	// Create the member record for the ordinary account.
	member := NewMember(ordinaryUser, roleMember, now, endTime)
	createMemberError := db.CreateMember(member)
	if createMemberError != nil {
		return 0, 0, createMemberError
	}

	if len(name) == 1 {
		// There is no associate user.
		return ordinaryUser.ID, 0, nil
	}

	// The sale includes payment for an associate member.  Set up records
	// for them too.
	assocUser := NewUser(name[1])
	createAssocUserError := db.CreateUser(assocUser)
	if createAssocUserError != nil {
		return 0, 0, createAssocUserError
	}

	// Update the sale.
	sale.AssocUserID = assocUser.ID

	// Create the member record for the associate account.
	assocMember := NewMember(assocUser, roleMember, now, endTime)
	createAssocMemberError := db.CreateMember(assocMember)
	if createAssocMemberError != nil {
		return 0, 0, createAssocMemberError
	}

	return ordinaryUser.ID, assocUser.ID, nil
}

// getLoginNames creates an returns the ordinary member's login name (their email
// address) and, if there is an associate, their login name (email if given,
// otherwise their name in the form  "first.last").
// It's assumed that a transaction is already set up in the db object.
func getLoginNames(sale *MembershipSale) ([]string, error) {

	result := make([]string, 0, 2)

	// The ordinary user must have an email address.  The incoming data has
	// been validated.  This check defends against it getting lost along the way.
	if len(sale.Email) == 0 {
		return result, errors.New("getLoginNames: no email address given")
	}

	result = append(result, sale.Email)

	if len(sale.AssocFirstName) > 0 {
		// The sale includes an associate member, who may or may not have an
		//email address.  If it's the same as the ordinary member, ignore it.
		// (If they share an email address they don't want two copies of our
		// emails.  In any case, two members with the same email address causes
		// problems when they try to change their Admidio password.)
		if len(sale.AssocEmail) <= 0 || sale.AssocEmail == sale.Email {
			// Use the name - "first.last" as the associate user's account name.
			loginName := sale.AssocFirstName + "." + sale.AssocLastName
			result = append(result, loginName)
		} else {
			// Use the eamil address as the associate user's account name.
			result = append(result, sale.AssocEmail)
		}
	}

	return result, nil
}

// GetMembershipYear gets the membership year that we are currently selling.
// It differs from organisation to organisation.
func GetMembershipYear(now time.Time) int {

	// The L&DLHS invites members to pay for year N+1 from the
	// 1st October in year N.  If a new member signs up during
	// year N after that date they get membership until the end
	// of year N+1.

	// Take the current date and figure out which year we are
	// selling.

	timezone := now.Location()
	currentYear := now.Year()
	startOfSellingYear := time.Date(currentYear, time.October, 1, 0, 0, 0, 0, timezone)
	var sellingYear int
	if now.Before(startOfSellingYear) {
		sellingYear = now.Year()
	} else {
		sellingYear = now.Year() + 1
	}

	return sellingYear
}

// GetUserIDofMember returns the userID of the user with a matching
// email address OR matching first name and last name.  The user must
// have the role Member.  The name and email search is case-insensitive.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetUserIDofMember(firstName, lastName, emailAddress string) (int64, error) {

	// The email address may be an empty string.  The first name and last name should always
	// be non-empty strings.
	users := make([]User, 0, 1)
	var userError error
	if len(emailAddress) > 0 {
		users, userError = db.GetUsersByLoginName(emailAddress)

		if userError != nil {
			return 0, userError
		}
	}

	if len(users) > 0 {
		// Success - found a user by the email address.
		return users[0].ID, nil
	}

	// We have to look up the user by the first and last names.
	lastNameID, lastNameIDErr := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
	if lastNameIDErr != nil {
		return 0, lastNameIDErr
	}

	firstNameID, firstNameIDErr := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
	if firstNameIDErr != nil {
		return 0, firstNameIDErr
	}

	// This queries searches for the member.  It uses lower()
	// which works in both Postgres and sqlite.
	const sql = `
		SELECT users.usr_id as id
        FROM adm_users as users
        LEFT JOIN adm_members as members
        ON users.usr_id = members.mem_usr_id
        LEFT JOIN adm_roles as roles
			ON roles.rol_id = members.mem_rol_id
			AND roles.rol_name = 'Member'
		LEFT JOIN adm_user_data AS firstName
			ON firstName.usd_usr_id = users.usr_id
			AND firstName.usd_usf_id = $1
		LEFT JOIN adm_user_data AS lastName
			ON lastName.usd_usr_id = users.usr_id
			AND lastName.usd_usf_id = $2
		WHERE (
				lower(firstName.usd_value) = lower($3) 
			AND lower(lastName.usd_value) = lower($4)
		);
		`

	rows, searchErr := db.Query(sql, firstNameID, lastNameID, firstName, lastName)
	if searchErr != nil {
		return 0, searchErr
	}
	defer rows.Close()

	if !rows.Next() {
		// No such member - a new member is registering.
		return 0, nil
	}

	var id int64
	err := rows.Scan(&id)
	if err != nil {
		return 0, err
	}
	rows.Close()
	return id, nil
}

// GetMembershipYearOfUser returns the user's membership year as a four-digit int,
// for example 2025.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMembershipYearOfUser(userID int64) (int, error) {
	var dateStr string
	var year int

	if db.Config.Type == "sqlite" {

		// SQLite stores dates as string, int or float.  We use strings
		// in the format "YYYY-MM-DD HH:MM:SS.SSS"
		const sqlForSQLite = `
			SELECT m.mem_end
			FROM adm_members AS m
			LEFT JOIN adm_roles as r
				ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			WHERE m.mem_usr_id = ?
		`

		getDateError := db.QueryRow(sqlForSQLite, userID).
			Scan(&dateStr)

		if getDateError != nil {
			return 0, getDateError
		}

		// We should have a string at least four characters long.
		// The first four characters are the year.

		if len(dateStr) == 0 {
			return 0, errors.New("date not set")
		}

		if len(dateStr) < 4 {
			return 0, errors.New("illegal date " + dateStr)
		}

		//  Get the year part of the string and convert to int.
		yearStr := dateStr[:4]

		var formatError error

		year, formatError = strconv.Atoi(yearStr)

		if formatError != nil {
			return 0, errors.New("illegal year " + yearStr)
		}

	} else {

		// The postgres to_char function does most of the work.
		const sqlForPostgres = `
			SELECT to_char(m.mem_end, 'YYYY')
			FROM adm_members AS m
			LEFT JOIN adm_roles as r
				ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			WHERE m.mem_usr_id = $1
		`

		getYearError := db.QueryRow(sqlForPostgres, userID).Scan(&year)

		if getYearError != nil {
			return 0, getYearError
		}
	}

	return year, nil
}

// SetMemberEndDate sets the end date of a member to the end of the current membership year.
// It's intended use is to allow an admin to revive a user account when the user renews their
// membership manually, eg using a paper form and a cheque.  There's a facility on the website
// that calls this function.
// The function returns an error if the user does not exist or has no member record with role
// 'Member'.  It's assumed that a transaction is already set up in the db object.
func (db *Database) SetMemberEndDate(userID int64, year int) error {

	// This query gets the member ID, start and end date of a member, given their user id.
	// A user with many roles has many adm_members records, one per role (admin, member etc).
	// We need the one with role "Member".

	const funcName = "Database.SetMemberEndDate"

	const getMemberIDSQL = `
		SELECT m.mem_id
		FROM adm_members AS m
		LEFT JOIN adm_users AS u
			ON m.mem_usr_id=u.usr_id
		LEFT JOIN adm_roles as r
			ON r.rol_id = m.mem_rol_id
		WHERE r.rol_name = 'Member'
		AND u.usr_id = $1;
	`

	// If everything is working properly we should only get exactly
	// one result, the ID of the user's member record with role Member.
	rows, getMemberIDError := db.Query(getMemberIDSQL, userID)
	if getMemberIDError != nil {
		if getMemberIDError == sql.ErrNoRows {
			em := fmt.Sprintf("%s: no member for user %d",
				funcName, userID)
			return errors.New(em)
		}
		em := fmt.Sprintf("%s %v", funcName, getMemberIDError)
		return errors.New(em)
	}
	defer rows.Close()

	// The user has a member record with 'Member' role.  Set the
	// end date, for example "2024-12-31 23:59:59 999999 +00".
	// That's the last microsecond of the last second of the year
	// in UTC.  It's safe to use this form for dates when we are
	// in GMT, but not for dates during BST.  We are setting dates
	// in the winter so we are OK.
	//
	// If we get more than one record (which shouldn't happen) set
	// the end date in all of them

	ids := make([]int, 0)
	for {
		if !rows.Next() {
			break
		}
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return errors.New(funcName + err.Error())
		}
		ids = append(ids, id)
	}

	// We must close the rows before we run another query.
	rows.Close()

	endDate := fmt.Sprintf("%04d-12-31 23:59:59 999999 +00", year)

	for _, id := range ids {

		var updateSQL string

		if db.Config.Type == "sqlite" {

			// SQLite has no special date or timestamp format.  We store
			// timestamps as strings in the format "YYYY-MM-DD HH:MM:SS.SSS".
			// It supports rowsAffected.
			updateSQL = `
				UPDATE adm_members
				SET mem_end = ?
				WHERE mem_id =?;
			`

		} else {

			// Postgres has a format for timestamps and a converter function
			// to turn a string into a timestamp.  It doesn't support
			// rowsAffected, so we use RETURNING.
			updateSQL = `
				UPDATE adm_members
				SET mem_end = to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS US TZH')
				WHERE mem_id =$2
				RETURNING mem_id;
			`
		}

		returnedID, setDateError := db.CreateRow(updateSQL, endDate, id)

		if setDateError != nil {
			em := fmt.Sprintf("%s: %v", funcName, setDateError)
			return errors.New(em)
		}

		if returnedID == 0 {
			em := fmt.Sprintf("%s: ID zero returned updating ID %d", funcName, id)
			return errors.New(em)
		}
	}

	// Success!
	return nil
}

// GetUserDataFieldIDByNameIntern gets the row from adm_user_fields with
// the given internal name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetUserDataFieldIDByNameIntern(nameIntern string) (int64, error) {

	f := "GetUserDataFieldIDByNameIntern: "

	const q = `
		SELECT usf_id
		FROM adm_user_fields
		WHERE usf_name_intern = $1
	`
	var fieldID int64
	err := db.QueryRow(q, nameIntern).Scan(&fieldID)
	if err != nil {
		em := fmt.Sprintf("%s: %v - %s", f, err, nameIntern)
		return 0, errors.New(em)
	}

	return fieldID, nil
}

// GetUserDataField gets the value of type T from the adm_user_data row with the given
// field ID and belonging to the given user.  If there is no row found it returns the
// zero value of the return type (an empty string, a zero or false).  (To detect no row,
// use GetUserDataFieldErrorOnNotFound().)
// It's assumed that a transaction is already set up in the db object.
func GetUserDataField[T int | float64 | bool | string](db *Database, fieldID, userID int64) (T, error) {

	var result T
	var zero T // The zero value of the type T.
	var err error
	result, err = GetUserDataFieldErrorOnNotFound[T](db, fieldID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// The query worked but found no rows.  This is correct behaviour.
			// Return the zero value of the type.
			return zero, nil
		}
		em := fmt.Sprintf("GetUserDataField: %v - %d %d", err, fieldID, userID)
		return zero, errors.New(em)
	}

	// Found a record.
	return result, nil
}

// GetUserDataFieldErrorOnNotFound gets the value of type T from the adm_user_data row with
//
//	given field ID and belonging to the given user.  If there is no row, it returns an
//
// sql.ErrNoRows error.
// It's assumed that a transaction is already set up in the db object.
func GetUserDataFieldErrorOnNotFound[T int | float64 | bool | string](db *Database, fieldID, userID int64) (T, error) {

	const q = `
		SELECT usd_value
		FROM adm_user_data
		WHERE usd_usr_id = $1
		AND usd_usf_id = $2;
	`
	var zero T
	var result T
	queryError := db.QueryRow(q, userID, fieldID).Scan(&result)
	if queryError != nil {
		// Return the zero value and the error.
		return zero, queryError
	}

	// Found a record.
	return result, nil
}

// SetDataProtectionField sets the DATA_PROTECTION_PERMISSION field for the user in
// adm_user_data.  In the DB, tick box fields are set to 0 or 1.
func (db *Database) SetDataProtectionField(userID int64, ticked bool) error {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern(DataStoragePermNameIntern)
	if fieldError != nil {
		return fieldError
	}
	// If the member consents to giftaid, fill in the box.  In case
	// it's already set from last year but not this year, ensure that
	// the value in the DB record is reset.
	if ticked {
		return SetUserDataField(db, fieldID, userID, 1)
	} else {
		return SetUserDataField(db, fieldID, userID, 0)
	}
}

// GetDataProtectionField gets the DATA_PROTECTION_PERMISSION field for the user from
// adm_user_data.  In the DB, Tick box fields are set to 0 or 1.
func (db *Database) GetDataProtectionField(userID int64) (bool, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern(DataStoragePermNameIntern)
	if fieldError != nil {
		em := fmt.Sprintf("GetDataProtectionField: %v", fieldError)
		return false, errors.New(em)
	}

	var fetchedValue int
	fetchedValue, fetchError := GetUserDataField[int](db, fieldID, userID)
	if fetchError != nil {
		return false, fetchError
	}

	// 0 is false, any other value is true
	if fetchedValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// SetReceiveEmailField sets the notices by email field for the user in
// adm_user_data.  In the DB, tick box fields are set to 0 or 1.
func (db *Database) SetReceiveEmailField(userID int64, ticked bool) error {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("NOTICES_BY_EMAIL")
	if fieldError != nil {
		return fieldError
	}
	// If the member consents to giftaid, fill in the box.  In case
	// it's already set from last year but not this year, ensure that
	// the value in the DB record is reset.
	if ticked {
		return SetUserDataField(db, fieldID, userID, 1)
	} else {
		return SetUserDataField(db, fieldID, userID, 0)
	}
}

// GetNoticesByEmailField gets the notices by email field for the user from
// adm_user_data.  In the DB, Tick box fields are set to 0 or 1.
func (db *Database) GetReceiveEmailField(userID int64) (bool, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("NOTICES_BY_EMAIL")
	if fieldError != nil {
		em := fmt.Sprintf("GetReceiveEmailField: %v", fieldError)
		return false, errors.New(em)
	}

	var fetchedValue int
	fetchedValue, fetchError := GetUserDataField[int](db, fieldID, userID)
	if fetchError != nil {
		return false, fetchError
	}

	// 0 is false, any other value is true
	if fetchedValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// SetTitle sets the user's title.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetTitle(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("SALUTATION")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetTitle gets the user's title.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetTitle(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("SALUTATION")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetFirstName sets the user's last name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetFirstName(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetFirstName gets the user's last name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetFirstName(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetLastName sets the user's first name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetLastName(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetLastName gets the user's last name.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetLastName(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetEmail sets the user's email address in adm_user_data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetEmail(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("EMAIL")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetEmail gets the user's Email address from adm_user_data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetEmail(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("EMAIL")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetAddressline1 sets the first line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetAddressLine1(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("STREET")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetAddressline1 gets the first line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetAddressLine1(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("STREET")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetAddressLine2 sets the second line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetAddressLine2(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_2")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetAddressline2 gets the second line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetAddressLine2(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_2")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetAddressLine3 sets the third line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetAddressLine3(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_3")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetAddressline3 gets the third line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetAddressLine3(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_3")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetTown sets the town in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetTown(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("CITY")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetTown gets the second line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetTown(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("CITY")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetCounty sets the county in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetCounty(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTY")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetCounty gets the second line of the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetCounty(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTY")
	if fieldError != nil {
		return "", fieldError
	}

	return GetUserDataField[string](db, fieldID, userID)
}

// SetPostcode sets the postcode in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetPostcode(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("POSTCODE")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetPostcode gets the postcode in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetPostcode(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("POSTCODE")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetCountryCode sets the country in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetCountryCode(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTRY")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetCountryCode gets the country in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetCountryCode(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTRY")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetCountry sets the country in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetCountry(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTRY")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetCountry gets the country in the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetCountry(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("COUNTRY")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetPhone sets the user's landline phone number in their profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetPhone(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("PHONE")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetPhone gets the landline phone number from the user's recvords.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetPhone(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("PHONE")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetMobile sets the user's mobile phone number in their profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetMobile(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("MOBILE")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetMobile gets the mobile phone number from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMobile(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("MOBILE")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetLocationOfInterest sets the user's location of interest (parish) in their profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetLocationOfInterest(userID int64, val string) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("LOCATION_OF_INTEREST")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, val)
}

// GetLocationOfInterest gets the mobile phone number from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetLocationOfInterest(userID int64) (string, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("LOCATION_OF_INTEREST")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetLastPayment sets the date of last payment field in adm_user_data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetLastPayment(userID int64, payment float64) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_LAST_PAYMENT")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, payment)
}

// SetDonationToSociety sets the donation to society field in adm_user_data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetDonationToSociety(userID int64, payment float64) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_LDLHS")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, payment)
}

// GetDonationToSociety gets the donation to society from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetDonationToSociety(userID int64) (float64, error) {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_LDLHS")
	if fieldError != nil {
		return 0.0, fieldError
	}

	v, e := GetUserDataField[float64](db, fieldID, userID)
	if e != nil {
		return 0.0, e
	}

	return v, nil
}

// SetDonationToMuseum sets the donation to the museum.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetDonationToMuseum(userID int64, payment float64) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_THE_MUSEUM")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, payment)
}

// GetDonationToMuseum gets the donation to museum from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetDonationToMuseum(userID int64) (float64, error) {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_THE_MUSEUM")
	if fieldError != nil {
		return 0.0, fieldError
	}

	v, e := GetUserDataField[float64](db, fieldID, userID)
	if e != nil {
		return 0.0, e
	}

	return v, nil
}

// SetDateLastPaid sets the date last paid field in adm_user_data.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) SetDateLastPaid(userID int64, d time.Time) error {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("DATE_LAST_PAID")
	if fieldError != nil {
		return fieldError
	}

	return db.SetDateFieldInUserData(fieldID, userID, d)
}

// GetDateLastPaid gets the date last paid from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetDateLastPaid(userID int64) (string, error) {
	// This is only used for testing.  The date string is not converted
	// to a date object.  If you want to do that, note that Postgres
	// and SQLite store them differently:
	// postgres: "2025-12-31T00:00:00Z"
	// SQLite: "2025-12-31 23:59:59 999999 +00"

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("DATE_LAST_PAID")
	if fieldError != nil {
		return "", fieldError
	}

	v, e := GetUserDataField[string](db, fieldID, userID)
	if e != nil {
		return "", e
	}

	return v, nil
}

// SetFriendField sets the friend of the museum field for the user in
// adm_user_data.  Tick box fields are set to 0 or 1.
func (db *Database) SetFriendField(userID int64, ticked bool) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("FRIEND_OF_THE_MUSEUM")
	if fieldError != nil {
		return fieldError
	}

	if ticked {
		return SetUserDataField(db, fieldID, userID, 1)
	} else {
		return SetUserDataField(db, fieldID, userID, 0)
	}

}

// SetGiftaid sets the giftaid field for the user in
// adm_user_data.  In the DB, tick box fields are set to 0 or 1.
func (db *Database) SetGiftaid(userID int64, ticked bool) error {

	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("GIFT_AID")
	if fieldError != nil {
		return fieldError
	}
	// If the member consents to giftaid, fill in the box.  In case
	// it's already set from last year but not this year, ensure that
	// the value in the DB record is reset.
	if ticked {
		return SetUserDataField(db, fieldID, userID, 1)
	} else {
		return SetUserDataField(db, fieldID, userID, 0)
	}
}

// GetGiftaid gets the giftaid field for the user from
// adm_user_data.  In the DB, Tick box fields are set to 0 or 1.
func (db *Database) GetGiftaid(userID int64) (bool, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("GIFT_AID")
	if fieldError != nil {
		em := fmt.Sprintf("GetGiftaid: %v", fieldError)
		return false, errors.New(em)
	}

	var fetchedValue int
	fetchedValue, fetchError := GetUserDataField[int](db, fieldID, userID)
	if fetchError != nil {
		return false, fetchError
	}

	// 0 is false, any other value is true
	if fetchedValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// SetMembersAtAddress sets the number of members at the user's address in
// adm_user_data.
func (db *Database) SetMembersAtAddress(userID int64, members int) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("MEMBERS_AT_ADDRESS")
	if fieldError != nil {
		return fieldError
	}
	return SetUserDataField(db, fieldID, userID, members)
}

// GetMembersAtAddress gets the number of members at address from the user's profile.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetMembersAtAddress(userID int64) (int, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("MEMBERS_AT_ADDRESS")
	if fieldError != nil {
		return 0, fieldError
	}

	v, e := GetUserDataField[int](db, fieldID, userID)
	if e != nil {
		return 0, e
	}

	return v, nil
}

// SetFriendsAtAddress sets the number of friends of the museum at the
// user's address in adm_user_data.
func (db *Database) SetFriendsAtAddress(userID int64, members int) error {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS")
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField(db, fieldID, userID, members)
}

// GetFriendsAtAddress gets the number of friends at the user's address.
// It's assumed that a transaction is already set up in the db object.
func (db *Database) GetFriendsAtAddress(userID int64) (int, error) {
	fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS")
	if fieldError != nil {
		return 0, fieldError
	}

	v, e := GetUserDataField[int](db, fieldID, userID)
	if e != nil {
		return 0, e
	}

	return v, nil
}

// SetUserDataField sets the field with ID fieldID in adm_user_data
// for the given user to the given value.  If a record for the field is
// missing, one is created.
func SetUserDataField[T int | float64 | bool | string](db *Database, fieldID, userID int64, val T) error {

	fn := "SetUserDataField"

	var query string
	var err error
	var returnedID int64

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update the value.
		const postgresSQL = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_id;
		`
		const sqliteSQL = `
			UPDATE adm_user_data
			SET usd_value = ?
			WHERE usd_usr_id = ?
			AND usd_usf_id = ?;
		`

		switch db.Config.Type {
		case "postgres":
			query = postgresSQL
		default:
			query = sqliteSQL
		}

		returnedID, err = db.CreateRow(query, val, userID, fieldID)

	} else {

		// There is no record for that field.  Create one.

		const postgresSQL = `
			INSERT INTO adm_user_data(usd_usr_id, usd_usf_id, usd_value)
			VALUES ($1, $2, $3)
			RETURNING usd_id;
		`
		const sqliteSQL = `
			INSERT INTO adm_user_data(usd_usr_id, usd_usf_id, usd_value)
			VALUES (?, ?, ?);
		`

		switch db.Config.Type {
		case "postgres":
			query = postgresSQL
		default:
			query = sqliteSQL
		}

		returnedID, err = db.CreateRow(query, userID, fieldID, val)
	}

	if err != nil {
		return err
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: zero return updating ID %d", fn, returnedID)
		return errors.New(em)
	}

	return nil
}

// SetDateFieldInUserData sets the field with ID fieldID in adm_user_data to an
// date value, eg '2025-10-30'.
func (db *Database) SetDateFieldInUserData(fieldID, userID int64, t time.Time) error {

	f := "SetDateFieldInUserData"

	dateStr := t.Format("2006-01-02")

	// Neither Postgres nor SQLite support rowsAffected.  Use RETURNING.
	var q string

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update it.
		const postgresSQL = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_usr_id;
		`
		const sqliteSQL = `
			UPDATE adm_user_data
			SET usd_value = ?
			WHERE usd_usr_id = ?
			AND usd_usf_id = ?;
		`
		switch db.Config.Type {
		case "postgres":
			q = postgresSQL
		default:
			q = sqliteSQL
		}

	} else {

		// There is no record for this field.  Create and set it.
		const postgresSQL = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES ($1, $2, $3)
			RETURNING usd_usr_id;
		`
		const sqliteSQL = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES (?, ?, ?);
		`
		switch db.Config.Type {
		case "postgres":
			q = postgresSQL
		default:
			q = sqliteSQL
		}
	}

	returnedID, execAndScanError :=
		db.CreateRow(q, dateStr, userID, fieldID)

	if execAndScanError != nil {
		em := fmt.Sprintf("%s: %v", f, execAndScanError)
		return errors.New(em)
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: ID zero returned updating ID %d", f, userID)
		return errors.New(em)
	}

	return nil
}

// SetTimeFieldInUserData sets the field with ID fieldID in adm_user_data to an
// time value.
func (db *Database) SetTimeFieldInUserData(fieldID, userID int64, t time.Time) error {

	f := "SetTimeFieldInUserData"

	var q string

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update it.

		// Postgres uses to_timestamp() to set the timestamp from a string.
		const postgresSQL = `
		UPDATE adm_user_data
		SET usd_value = to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS')
		WHERE usd_usr_id = $2
		AND usd_usf_id = $3
		RETURNING usd_usr_id;
	`
		const sqliteSQL = `
		UPDATE adm_user_data
		SET usd_value = ?
		WHERE usd_usr_id = ?
		AND usd_usf_id = ?;
	`
		switch db.Config.Type {
		case "postgres":
			q = postgresSQL
		default:
			q = sqliteSQL
		}

	} else {

		// There is no record for this field.  Create and set it.
		const postgresSQL = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES (to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS+HH'), $2, $3)
			RETURNING usd_usr_id;
		`
		const sqliteSQL = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES (?, ?, ?)
			RETURNING usd_usr_id;
		`
		switch db.Config.Type {
		case "postgres":
			q = postgresSQL
		default:
			q = sqliteSQL
		}
	}

	timeStr := t.Format("2006-01-02 15:04:05")

	returnedID, err := db.CreateRow(q, timeStr, userID, fieldID)
	if err != nil {
		em := fmt.Sprintf("%s: %v", f, err)
		return errors.New(em)
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: ID zero returned updating ID %d", f, userID)
		return errors.New(em)
	}

	return nil
}

// FieldSet checks whether the given field is set in adm_user_data
// for the given user.
func (db *Database) FieldSet(fieldID, userID int64) bool {
	const sqlSelect = `
		SELECT usd_id FROM adm_user_data
		WHERE usd_usr_id = $1
		AND usd_usf_id = $2;
	`

	// Check if the user already has the field set.
	rows, selectError := db.Query(sqlSelect, userID, fieldID)
	if selectError != nil {
		return false
	}
	defer rows.Close()

	if rows.Next() {
		// The field is set.
		return true
	} else {
		// The field is not set
		return false
	}
}

// MemberExists returns true if there is already a member with the given
// username or email address.
func (db *Database) MemberExists(username, emailAddress string) (bool, error) {

	// Check for a user with the given username.
	users, userError := db.GetUsersByLoginName(username)
	if userError != nil {
		return false, nil
	}

	// There should be exactly one user (user exists) or none (no such user)
	// but in theory there could be more than one user.
	if len(users) > 0 {
		return true, nil
	}

	// Check if any user is already set up with this email address regardless of their
	// user name.

	// Get the ID just once.
	const emailFieldName = "EMAIL"
	emailID, fieldError := db.GetUserDataFieldIDByNameIntern(emailFieldName)
	if fieldError != nil && fieldError != sql.ErrNoRows {
		return false, errors.New("MemberExists: checking email field ID - " + fieldError.Error())
	}

	const q = `
	SELECT email.usd_id as id
        FROM adm_user_data AS email
		WHERE email.usd_usf_id = $1
			AND (lower(email.usd_value) = lower($2));
	`

	rows, emailError := db.Query(
		q,
		emailID,
		emailAddress,
	)
	if emailError != nil {
		slog.Error(emailError.Error())
		return false, emailError
	}
	defer rows.Close()

	// Get the results - a list of email addresses.  If there are any,
	// we have a match, so return true.  If not, return false.
	if !rows.Next() {
		// No users found with this email address.
		return false, nil
	}

	var result int
	scanError := rows.Scan(&result)
	if scanError != nil {
		slog.Error(scanError.Error())
	}

	// At least one user found with this email address.  Enough.
	return true, nil
}

// GetExtraDetails gets the given user's extra details (address, phone number etc)
// and fills in the appropriate fields of the given membership sale record.
// The function assumes that a transaction has been set up.
func (db *Database) GetExtraDetails(ms *MembershipSale) error {

	if ms.UserID <= 0 {
		return errors.New("GetExtraDetails.Save: no userID")
	}

	a1, a1e := db.GetAddressLine1(ms.UserID)
	if a1e != nil {
		return a1e
	}

	a2, a2e := db.GetAddressLine2(ms.UserID)
	if a2e != nil {
		return a2e
	}

	a3, a3e := db.GetAddressLine3(ms.UserID)
	if a3e != nil {
		return a3e
	}

	t, te := db.GetTown(ms.UserID)
	if te != nil {
		return te
	}

	c, ce := db.GetCounty(ms.UserID)
	if ce != nil {
		return ce
	}

	pc, pce := db.GetPostcode(ms.UserID)
	if pce != nil {
		return pce
	}

	ct, cte := db.GetCountry(ms.UserID)
	if cte != nil {
		return cte
	}

	p, pe := db.GetPhone(ms.UserID)
	if pe != nil {
		return pe
	}

	mob, me := db.GetMobile(ms.UserID)
	if me != nil {
		return me
	}

	loi, loie := db.GetLocationOfInterest(ms.UserID)
	if loie != nil {
		return loie
	}

	mi, mie := db.GetMembersInterests(ms.UserID)
	if mie != nil {
		return mie
	}

	moi, moie := db.GetMembersOtherInterests(ms.UserID)
	if moie != nil {
		return moie
	}
	ms.AddressLine1 = a1
	ms.AddressLine2 = a2
	ms.AddressLine3 = a3
	ms.Town = t
	ms.County = c
	ms.Postcode = pc
	ms.Country = ct
	ms.Phone = p
	ms.Mobile = mob
	ms.LocationOfInterest = loi
	if ms.TopicsOfInterest == nil {
		ms.TopicsOfInterest = make(map[int64]interface{})
	}
	for _, nt := range mi {
		ms.TopicsOfInterest[nt.InterestID] = nil
	}
	ms.OtherTopicsOfInterest = moi.Interests

	return nil
}

// SaveExtraDetails saves the given user's extra details - address, phone number etc.
// The function assumes that a transaction has been set up.
func (db *Database) SaveExtraDetails(ms *MembershipSale) error {

	if ms.UserID <= 0 {
		return errors.New("saveExtraDetails.Save: no userID")
	}

	if len(ms.AddressLine1) > 0 {
		err := db.SetAddressLine1(ms.UserID, ms.AddressLine1)
		if err != nil {
			return err
		}
	}

	if len(ms.AddressLine2) > 0 {
		err := db.SetAddressLine2(ms.UserID, ms.AddressLine2)
		if err != nil {
			return err
		}
	}

	if len(ms.AddressLine3) > 0 {
		err := db.SetAddressLine3(ms.UserID, ms.AddressLine3)
		if err != nil {
			return err
		}
	}

	if len(ms.Town) > 0 {
		err := db.SetTown(ms.UserID, ms.Town)
		if err != nil {
			return err
		}
	}

	if len(ms.County) > 0 {
		err := db.SetCounty(ms.UserID, ms.County)
		if err != nil {
			return err
		}
	}

	if len(ms.Postcode) > 0 {
		err := db.SetPostcode(ms.UserID, ms.Postcode)
		if err != nil {
			return err
		}
	}

	if len(ms.CountryCode) > 0 {
		err := db.SetCountryCode(ms.UserID, ms.CountryCode)
		if err != nil {
			return err
		}
	}

	if len(ms.Phone) > 0 {
		err := db.SetPhone(ms.UserID, ms.Phone)
		if err != nil {
			return err
		}
	}

	if len(ms.Mobile) > 0 {
		err := db.SetMobile(ms.UserID, ms.Mobile)
		if err != nil {
			return err
		}
	}

	if len(ms.LocationOfInterest) > 0 {
		err := db.SetLocationOfInterest(ms.UserID, ms.LocationOfInterest)
		if err != nil {
			return err
		}
	}

	if len(ms.TopicsOfInterest) > 0 {
		for interestID := range ms.TopicsOfInterest {
			mi := NewMembersInterest(ms.UserID, interestID)
			err := db.CreateMembersInterest(mi)
			if err != nil {
				return err
			}
		}
	}

	if len(ms.OtherTopicsOfInterest) > 0 {
		moi := NewMembersOtherInterests(ms.UserID, ms.OtherTopicsOfInterest)
		err := db.CreateMembersOtherInterests(moi)
		if err != nil {
			return err
		}
	}

	if ms.AssocUserID > 0 {
		// There is an assocate user.  Set the same address and landline number.
		if len(ms.AddressLine1) > 0 {
			err := db.SetAddressLine1(ms.AssocUserID, ms.AddressLine1)
			if err != nil {
				return err
			}
		}

		if len(ms.AddressLine2) > 0 {
			err := db.SetAddressLine2(ms.AssocUserID, ms.AddressLine2)
			if err != nil {
				return err
			}
		}

		if len(ms.AddressLine3) > 0 {
			err := db.SetAddressLine3(ms.AssocUserID, ms.AddressLine3)
			if err != nil {
				return err
			}
		}

		if len(ms.Town) > 0 {
			err := db.SetTown(ms.AssocUserID, ms.Town)
			if err != nil {
				return err
			}
		}

		if len(ms.County) > 0 {
			err := db.SetCounty(ms.AssocUserID, ms.County)
			if err != nil {
				return err
			}
		}

		if len(ms.Postcode) > 0 {
			err := db.SetPostcode(ms.AssocUserID, ms.Postcode)
			if err != nil {
				return err
			}
		}

		if len(ms.CountryCode) > 0 {
			err := db.SetCountryCode(ms.AssocUserID, ms.CountryCode)
			if err != nil {
				return err
			}
		}

		if len(ms.Phone) > 0 {
			err := db.SetPhone(ms.AssocUserID, ms.Phone)
			if err != nil {
				return err
			}
		}

		// The extra details form may have specified the associate user's mobile number.
		if len(ms.AssocMobile) > 0 {
			err := db.SetMobile(ms.AssocUserID, ms.AssocMobile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

const NoUserNameError = "getUserName: no text supplied"

// GetUserName constructs the user name.  Any upper case is converted to lower
// case.  If the email address is present, use that, otherwise construct one
// from the first and last name but beware of white space.  Replace any with a
// single full stop, so first name "Herbert George" last name "Wells" produces
// "herbert.george.wells" and first name "\tH   G", last name "Wells" becomes
// "h.g.wells".
func GetUserName(email, firstname, lastname string) (string, error) {

	// Trim space at the start and end of each string.
	email = strings.TrimSpace(email)
	firstname = strings.TrimSpace(firstname)
	lastname = strings.TrimSpace(lastname)

	// If there is nothing left, that's an error.
	if len(email) == 0 && len(firstname) == 0 && len(lastname) == 0 {
		return "", errors.New(NoUserNameError)
	}

	email = strings.ToLower(email)
	firstname = strings.ToLower(firstname)
	lastname = strings.ToLower(lastname)

	// Replace all multiple white space within the strings with a single dot.
	// "a\t B" -> "a.B".
	re := regexp.MustCompile("[ \t]+")
	email = re.ReplaceAllString(email, ".")
	firstname = re.ReplaceAllString(firstname, ".")
	lastname = re.ReplaceAllString(lastname, ".")

	// At least one of the fields contains some text.
	if len(email) > 0 {
		return email, nil
	}

	// The first or second name field contains text but maybe not both.
	switch {
	case len(firstname) > 0 && len(lastname) > 0:
		return firstname + "." + lastname, nil
	case len(firstname) > 0:
		return firstname, nil
	default:
		// If we get to here, lastname must contain some text.
		return lastname, nil
	}
}

// CreateUuid creates and returns a UUID which is unique in the given row of the given table.
// The technque uses a random source which makes a duplicate extremely unlikely. although
// possible.
func CreateUuid(tx *sql.Tx, field, table string) (string, error) {

	// Do this up to ten times until you get a UUID that's not already
	// used.  Each attempt is very unlikely to fail.
	for i := 0; i < 10; i++ {

		// Create a UUID.
		uid, randError := uuid.NewRandom()
		if randError != nil {
			return "", randError
		}

		// Check that the UUID is not already in the table.
		// (This is theoretically possible but unlikely.)
		q := fmt.Sprintf("select %s from %s where %s = $1;",
			field, table, field)

		resultSet, err := tx.Query(q, uid.String())
		if err != nil {
			// If there is no match under Postgres, this may return the error
			// "no rows in result set".
			if err == sql.ErrNoRows {
				// Success!
				return uid.String(), nil
			}
			// If this happens under SQLite, it's a genuine error.
			em := fmt.Sprintf("createUuid: %s %s %s", field, table, err.Error())
			return "", errors.New(em)
		}

		// Under Postgres the ResultSet may be nil if there are no matching entries.
		if resultSet == nil {
			// Success!
			return uid.String(), nil
		}

		// We are only interested in any error.  Leaving the result set open can cause
		// problems so close it now.
		closeError := resultSet.Close()
		if closeError != nil {
			return "", closeError
		}

		// If there is no match under SQLite we get to here and
		// the list of uuids is empty.

		// Get the results - a list of uuids.  Should be no items
		// or one item.  If there are no items, u is unique so
		// return it.  If we find any items, the uuid is already
		// in the table, it's not unique.  We try again until we
		// find one that is unique.

		if !resultSet.Next() {
			// Success!
			return uid.String(), nil
		}

		var fetchedUUID string
		scanError := resultSet.Scan(&fetchedUUID)
		if scanError != nil {
			return "", scanError
		}
	}

	// All attempts have failed.  This is very very unlikely but
	// possible.
	return "", errors.New("CreateUUID: clash creating ID for table " + table)
}
