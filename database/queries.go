package database

import (
	// database/sql is imported by database.go.
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var regExpForPostgresParamsToSQLiteParams *regexp.Regexp

// init should always work but if any of the calls in it fail, it will
// crash the application.
func init() {
	// Set up the regular expression or die.
	regExpForPostgresParamsToSQLiteParams = regexp.MustCompile(`\$[0-9]+`)
}

// MembershipSale represents the payment of a membership sale - the annual
// membership fee.
type MembershipSale struct {
	ID                       int
	PaymentService           string  // The payment processor eg "Stripe".
	PaymentStatus            string  // "pending", "complete" or "cancelled"
	PaymentID                string  // The transaction Id from the payment processor.
	MembershipYear           int     // The membership year paid for.
	FullMemberID             int     // The user ID of the member
	FullMemberFee            float64 // The fee paid for full membership.
	FullMemberIsFriend       bool    // True if the full member is a friend of the museum.
	FullMemberFriendFee      float64 // The fee paid for the full member to be a friend.
	AssociateMemberID        int     // The user ID of the associate member.
	AssociateMemberFee       float64 // the fee paid for associate membership.
	AssocMemberIsFriend      bool    // True if the associate member is a friend of the museum.
	AssociateMemberFriendFee float64 // the fee paid for associate member to be a fiend.
	DonationToSociety        float64
	DonationToMuseum         float64
}

// Create() creates a MembershipSale record in the database.
func (ms *MembershipSale) Create(db *Database) (int, error) {

	// Postgres doesn't support LastInsertId.  We use the form recommended here:
	// https://github.com/jackc/pgx/issues/1483, which is
	// err := db.QueryRow("INSERT INTO user (name) VALUES ('John') RETURNING id").Scan(&id)

	var createError error
	var id int

	if ms.AssociateMemberID > 0 {
		// createStatement is the template to create a record in MembershipSales
		// with a non-null ms_usr2_id (associate member).
		// Initially the payment_id is an empty string.   It will be set to the
		// Stripe transaction ID later.
		const createStatement = `
	INSERT INTO membership_sales (
		ms_id, 
		ms_payment_service,
		ms_payment_status,
		ms_payment_id,
		ms_membership_year,
		ms_usr1_id,
		ms_usr1_fee,
		ms_usr1_friend,
		ms_usr1_friend_fee,
		ms_usr2_id,
		ms_usr2_fee,
		ms_usr2_friend,
		ms_usr2_friend_fee,
		ms_donation,
		ms_donation_museum
	) Values(nextval('membership_sales_ms_id_seq'), 
	 	$1, $2, '', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	)
	RETURNING ms_id;
	`
		createError = db.QueryRow(
			createStatement,
			ms.PaymentService,
			ms.PaymentStatus,
			ms.MembershipYear,
			ms.FullMemberID,
			ms.FullMemberFee,
			ms.FullMemberIsFriend,
			ms.FullMemberFriendFee,
			ms.AssociateMemberID,
			ms.AssociateMemberFee,
			ms.AssocMemberIsFriend,
			ms.AssociateMemberFriendFee,
			ms.DonationToSociety,
			ms.DonationToMuseum,
		).Scan(&id)
	} else {
		// createStatement is the template to create a record in MembershipSales
		// with a null ms_usr2_id (no associate member).
		// Initially the payment_id is an empty string.   It will be set to the
		// Stripe transaction ID later.
		const createStatement = `
			INSERT INTO membership_sales (
				ms_id, 
				ms_payment_service,
				ms_payment_status,
				ms_payment_id,
				ms_membership_year,
				ms_usr1_id,
				ms_usr1_fee,
				ms_usr1_friend,
				ms_usr1_friend_fee,
				ms_usr2_id,
				ms_usr2_fee,
				ms_usr2_friend,
				ms_usr2_friend_fee,
				ms_donation,
				ms_donation_museum
				) Values(nextval('membership_sales_ms_id_seq'), 
					$1, $2, '', $3, $4, $5, $6, $7, NULL, 0.0, false, 0.0, $8, $9
				)
				RETURNING ms_id;
			`
		// The payment ID will be set to the Stripe transaction ID later.
		createError = db.QueryRow(
			createStatement,
			ms.PaymentService,
			ms.PaymentStatus,
			ms.MembershipYear,
			ms.FullMemberID,
			ms.FullMemberFee,
			ms.FullMemberIsFriend,
			ms.FullMemberFriendFee,
			ms.DonationToSociety,
			ms.DonationToMuseum,
		).Scan(&id)

	}

	if createError == nil {
		// Set the id in the membership sales object.
		ms.ID = id
	}

	return id, createError
}

// Update() updates a MembershipSale record in the database,
// setting the payment status (nt null) and the payment Id (nullable).
func (ms *MembershipSale) Update(db *Database, status, paymentID string) (int, error) {

	sql := `
		UPDATE MEMBERSHIP_SALES 
		SET ms_payment_status=$1, ms_payment_id=$2
		WHERE ms_id=$3
	`

	result, execError := db.Exec(sql, status, paymentID, ms.ID)
	if execError == nil {
		return 0, execError
	}

	rowsAffected, raError := result.RowsAffected()
	if raError != nil {
		return 0, raError
	}
	if rowsAffected != 1 {
		em := fmt.Sprintf("MembershipSale.Update: expected 1 row affected, got %d", rowsAffected)
		return 0, errors.New(em)
	}

	// Success.
	return ms.ID, nil
}

// Delete() deletes a MembershipSale record in the database,
func (ms *MembershipSale) Delete(db *Database) error {

	// We have to use this form - see the Create function.
	sql := `
		DELETE FROM MEMBERSHIP_SALES 
		WHERE ms_id=$1
		RETURNING ms_id;
	`

	var returnedID int
	execAndScanError := db.QueryRow(sql, ms.ID).Scan(&returnedID)
	if execAndScanError != nil {
		return execAndScanError
	}

	if returnedID == 0 {
		em := fmt.Sprintf("MembershipSale.Delete: zero return deleting ID %d", ms.ID)
		return errors.New(em)
	}

	if returnedID != ms.ID {
		em := fmt.Sprintf(
			"MembershipSale.Delete: deleting ID %d, got ID %d  back", ms.ID, returnedID)
		return errors.New(em)
	}

	// Set the ID in the struct to zero to mark it as deleted.
	ms.ID = 0

	// Success.
	return nil
}

// TotalPayment adds up the fees and returns the total.
func (ms *MembershipSale) TotalPayment() float64 {
	return ms.FullMemberFee +
		ms.AssociateMemberFee +
		ms.FullMemberFriendFee +
		ms.AssociateMemberFriendFee +
		ms.DonationToSociety +
		ms.DonationToMuseum
}

func (db *Database) GetMembershipSale(id int) (*MembershipSale, error) {
	const query = `
	SELECT 
		ms_id,
		ms_payment_service,
		ms_payment_status,
		ms_payment_id,
		ms_membership_year,
		ms_usr1_id,
		ms_usr1_friend,
		COALESCE(ms_usr2_id, 0),
		ms_usr2_friend,
		ms_usr1_fee,
		ms_usr2_fee,
		ms_usr1_friend_fee,
		ms_usr2_friend_fee,
		ms_donation,
		ms_donation_museum
	FROM membership_sales
	WHERE ms_id = $1;
`
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
		&ms.FullMemberID,
		&ms.FullMemberIsFriend,
		&ms.AssociateMemberID,
		&ms.AssocMemberIsFriend,
		&ms.FullMemberFee,
		&ms.AssociateMemberFee,
		&ms.FullMemberFriendFee,
		&ms.AssociateMemberFriendFee,
		&ms.DonationToSociety,
		&ms.DonationToMuseum,
	)
	if err != nil {
		return nil, err
	}

	return &ms, nil

}

// GetPaymentYear gets the membership year that we are
// currently selling.
func GetPaymentYear(now time.Time) int {

	// The L&DLHS invites members to pay for year N+1 from the
	// 1st October in year N.  If a new member signs up during
	// year N after that date they get membership until the end
	// of year N.

	// Take the current date and figure out which year we are
	// selling.

	timeZone := now.Local().Location()
	currentYear := now.Year()
	startOfSellingYear := time.Date(currentYear, time.October, 1, 0, 0, 0, 0, timeZone)
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
func (db *Database) GetUserIDofMember(firstName, lastName, emailAddress string) (int, error) {

	emailID, emailIDErr := db.getEmailID()
	if emailIDErr != nil {
		return 0, emailIDErr
	}

	lastNameID, lastNameIDErr := db.getLastNameID()
	if lastNameIDErr != nil {
		return 0, lastNameIDErr
	}

	firstNameID, firstNameIDErr := db.getFirstNameID()
	if firstNameIDErr != nil {
		return 0, firstNameIDErr
	}

	// This queries searches for the member.  It uses lower()
	// which works in both Postgres and sqlite.
	const getUserIDForMemberSQL = `
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
		LEFT JOIN adm_user_data AS email
			ON email.usd_usr_id = users.usr_id
			AND email.usd_usf_id = $3
		WHERE (lower(firstName.usd_value) = lower($4) 
			AND lower(lastName.usd_value) = lower($5))
			OR lower(email.usd_value) = lower($6);
		`

	row, searchErr := db.Query(getUserIDForMemberSQL,
		firstNameID, lastNameID, emailID,
		firstName, lastName, emailAddress)

	if searchErr != nil {
		return 0, searchErr
	}

	if !row.Next() {
		return 0, errors.New("no matching member")
	}

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	row.Close()
	return id, nil
}

func (db *Database) GetMembershipYear(userID int) (int, error) {
	var dateStr string
	var year int

	if db.Type == "sqlite" {
		// SQLite stores dates as string, int or float.  We use strings
		// in the format "YYYY-MM-DD HH:MM:SS.SSS"
		const sqlForSQLite = `
			SELECT m.mem_end
			FROM adm_members AS m
			LEFT JOIN adm_roles as r
				ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			WHERE m.mem_usr_id = $1
		`

		getDateError := db.Connection.QueryRow(sqlForSQLite, userID).
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

// SetMemberEndDate sets the end date of a member to the end
// of the current year.  It returns an error if the user does
// not exist or has no member record with role 'Member'.
func (db *Database) SetMemberEndDate(userID int, year int) error {

	// This query gets the member ID, start and end date of a
	// member, given their user id.  A user can have many members,
	// one per role (admin, member etc).  We need the one with
	// role "Member".

	const funcName = "SetMemberEndDate"

	// This works for Postgres and sqlite.
	const getMemberIDSQL = `
		SELECT m.mem_id
		FROM adm_members AS m
		LEFT JOIN adm_users AS u 
			ON m.mem_usr_id=u.usr_id
		LEFT JOIN adm_roles as r
			ON r.rol_id = m.mem_rol_id
			AND r.rol_name = 'Member'
		WHERE u.usr_id = $1
	`

	var memberID int

	// We should only get either no results or one result.
	getMemberIDError := db.Connection.QueryRow(getMemberIDSQL, userID).
		Scan(&memberID)

	if getMemberIDError != nil {
		if getMemberIDError == sql.ErrNoRows {
			em := fmt.Sprintf("%s: no member for user %d",
				funcName, userID)
			return errors.New(em)
		}
		em := fmt.Sprintf("%s %v", funcName, getMemberIDError)
		return errors.New(em)
	}

	// The user has a member record with 'Member' role.  Set the
	// end date, for example "2024-12-31 23:59:59 999999 +00".
	// That's the last microsecond of the last second of the year
	// in UTC.  It's safe to use this form for dates when we are
	// in GMT, but not for dates during BST.  We are setting dates
	// in the winter so we are OK.

	var result sql.Result
	var setDateError error

	if db.Type == "sqlite" {

		// Sqlite has no special date or timestamp format.  We store
		// timestamps as strings in the format "YYYY-MM-DD HH:MM:SS.SSS".
		dateStr := fmt.Sprintf("%04d-12-31 23:59:59.999", year)

		const setYearEndSQLForSQLite = `
		UPDATE adm_members
		SET mem_end = ?
	    WHERE mem_id =?`

		result, setDateError =
			db.Connection.Exec(setYearEndSQLForSQLite, dateStr, memberID)

	} else {

		// Postgres has a format for timestamps and converter functions
		// to turn a timestamp into a string and vice versa.
		endDate := fmt.Sprintf("%d-12-31 23:59:59 999999 +00", year)
		const setYearEndSQLForPostgres = `
			UPDATE adm_members
			SET mem_end = to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS US TZH')
			WHERE mem_id =$2`

		result, setDateError =
			db.Connection.Exec(setYearEndSQLForPostgres, endDate, memberID)
	}

	if setDateError != nil {
		em := fmt.Sprintf("%s %v", funcName, setDateError)
		return errors.New(em)
	}

	rowsAffected, rowsAffectedError := result.RowsAffected()
	if rowsAffectedError != nil {
		em := fmt.Sprintf("%s %v", funcName, rowsAffectedError)
		return errors.New(em)
	}

	if rowsAffected < 1 {
		em := fmt.Sprintf("%s: no rows updated for userID %d",
			funcName, userID)
		return errors.New(em)
	}

	// Success!
	return nil
}

// SetLastPayment sets the date of last payment field in adm_user_data.
func (db *Database) SetLastPayment(userID int, payment float64) error {
	fieldID, fieldError := db.getLastPaymentID()
	if fieldError != nil {
		return fieldError
	}

	return db.SetUserDataFloatField(fieldID, userID, payment)
}

// SetDonationToSociety sets the donation to society field in adm_user_data.
func (db *Database) SetDonationToSociety(userID int, payment float64) error {
	fieldID, fieldError := db.getDonationToSocietyID()
	if fieldError != nil {
		return fieldError
	}

	return db.SetUserDataFloatField(fieldID, userID, payment)
}

// SetDonationToMuseum sets the donation to museum field in adm_user_data.
func (db *Database) SetDonationToMuseum(userID int, payment float64) error {
	fieldID, fieldError := db.getDonationToMuseumID()
	if fieldError != nil {
		return fieldError
	}

	return db.SetUserDataFloatField(fieldID, userID, payment)
}

// SetDateLastPaid sets the dat last paid field in adm_user_data.
func (db *Database) SetDateLastPaid(userID int, d time.Time) error {

	fieldID, fieldError := db.getFieldIDOnce("DATE_LAST_PAID", &db.dateLastPaidID)
	if fieldError != nil {
		return fieldError
	}

	return db.SetTimeFieldInUserData(fieldID, userID, d)
}

// SetFriendTickBox sets the friend of the museum tick box for the user in
// adm_user_data.  Tick box fields are set to 0 or 1.
func (db *Database) SetFriendTickBox(userID int, ticked bool) error {
	fieldID, fieldError := db.getFriendID()
	if fieldError != nil {
		return fieldError
	}

	if ticked {
		return db.SetUserDataIntField(fieldID, userID, 1)
	} else {
		return db.SetUserDataIntField(fieldID, userID, 0)
	}

}

// SetMembersAtAddress sets the number of members at the user's address in
// adm_user_data.
func (db *Database) SetMembersAtAddress(userID int, members int) error {
	fieldID, fieldError := db.getMembersAtAddressID()
	if fieldError != nil {
		return fieldError
	}
	return db.SetUserDataIntField(fieldID, userID, members)
}

// SetFriendsAtAddress sets the number of friends of the museum at the
// user's address in adm_user_data.
func (db *Database) SetFriendsAtAddress(userID int, members int) error {
	fieldID, fieldError := db.getFriendsAtAddressID()
	if fieldError != nil {
		return fieldError
	}

	return db.SetUserDataIntField(fieldID, userID, members)
}

// SetUserDataFloatField sets the field with ID fieldID in adm_user_data to a
// float64 value.  If a record for the field is missing, one is created.
func (db *Database) SetUserDataFloatField(fieldID, userID int, floatValue float64) error {

	var result sql.Result

	if db.FieldSet(fieldID, userID) {

		// There is already a record for this field.  Update the value.
		const sqlUpdate = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3;
	`
		var execError error
		result, execError = db.Exec(sqlUpdate, floatValue, userID, fieldID)
		if execError == nil {
			return execError
		}

	} else {

		// There is no record for that field.  Create one.
		const sqlInsert = `
			INSERT INTO adm_user_data(usd_id, usd_usr_id, usd_usf_id, usd_value)
			VALUES (nextval('public.adm_user_data_usd_id_seq'), $1, $2, $3);
		`
		var execError error
		result, execError = db.Exec(sqlInsert, userID, fieldID, floatValue)
		if execError == nil {
			return execError
		}
	}

	rowsAffected, raError := result.RowsAffected()
	if raError != nil {
		return raError
	}
	if rowsAffected != 1 {
		em := fmt.Sprintf("SetLastPayment: insert expected 1 row affected, got %d", rowsAffected)
		return errors.New(em)
	}

	return nil
}

// SetUserDataIntField sets the field with ID fieldID in adm_user_data to an
// int value.  (This includes tick boxes, which are set to 0 or 1.)  If a record
// for the field is missing, one is created.
func (db *Database) SetUserDataIntField(fieldID, userID int, intValue int) error {

	var result sql.Result

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update the value.
		const sqlUpdate = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3;
		`
		var execError error
		result, execError = db.Exec(sqlUpdate, intValue, userID, fieldID)
		if execError == nil {
			return execError
		}
	} else {

		// There is no record for that field.  Create one.
		const sqlInsert = `
			INSERT INTO adm_user_data(usd_id, usd_usr_id, usd_usf_id, usd_value)
			VALUES (nextval('public.adm_user_data_usd_id_seq'), $1, $2, $3);
		`
		var execError error
		result, execError = db.Exec(sqlInsert, userID, fieldID, intValue)
		if execError == nil {
			return execError
		}
	}

	rowsAffected, raError := result.RowsAffected()
	if raError != nil {
		return raError
	}
	if rowsAffected != 1 {
		em := fmt.Sprintf("SetLastPayment: insert expected 1 row affected, got %d", rowsAffected)
		return errors.New(em)
	}

	return nil
}

// SetTimeFieldInUserData sets the field with ID fieldID in adm_user_data to an
// time value.
func (db *Database) SetTimeFieldInUserData(fieldID, userID int, t time.Time) error {

	// Postgres uses to_timestamp() to set the time from a string.
	const sqlUpdatePostgres = `
		UPDATE adm_user_data
		SET usd_value = to_timestamp($1, 'YYYY-MM-DD')
		WHERE usd_usr_id = $2
		AND usd_usf_id = $3;
	`
	const sqlInsertPostgres = `
		INSERT INTO adm_user_data(usd_usr_id, usd_usf_id, usd_value)
		VALUES ($1, $2, to_timestamp($3, 'YYYY-MM-DD'));
	`
	const sqlUpdateSQLite = `
		UPDATE adm_user_data
		SET usd_value = ?
		WHERE usd_usr_id = ?
		AND usd_usf_id = ?;
	`
	const sqlInsertSQLite = `
		INSERT INTO adm_user_data(usd_usr_id, usd_usf_id, usd_value)
		VALUES (?, ?, ?,);
	`

	var sqlUpdate string
	var sqlInsert string
	if db.Type == "postgres" {
		sqlUpdate = sqlUpdatePostgres
		sqlInsert = sqlInsertPostgres
	} else {
		sqlUpdate = sqlUpdateSQLite
		sqlInsert = sqlInsertSQLite
	}

	var sqlResult sql.Result
	timeStr := t.Format("2006-01-02")

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update the value.
		var execError error
		sqlResult, execError = db.Exec(sqlUpdate, timeStr, userID, fieldID)
		if execError != nil {
			return execError
		}
	} else {

		// There is no record for that field.  Create one.
		var execError error
		sqlResult, execError = db.Exec(sqlInsert, userID, fieldID, timeStr)
		if execError != nil {
			return execError
		}
	}

	rowsAffected, raError := sqlResult.RowsAffected()
	if raError != nil {
		return raError
	}
	if rowsAffected != 1 {
		em := fmt.Sprintf("SetTimeFieldInUserData: expected 1 row affected, got %d", rowsAffected)
		return errors.New(em)
	}

	return nil
}

func (db *Database) FieldSet(fieldID, userID int) bool {
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

	if rows.Next() {
		// The field is set.
		return true
	} else {
		// The field is not set
		return false
	}
}

// getFirstNameID returns the ID of the first name field.
func (db *Database) getFirstNameID() (int, error) {
	// Get the ID just once.
	const firstNameFieldName = "FIRST_NAME"
	if db.firstNameID != 0 {
		return db.firstNameID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(firstNameFieldName)
	if fetchError != nil {
		return 0, errors.New("getFirstNameID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.firstNameID = fieldID

	return fieldID, nil
}

// getLastNameID return the Id of the last name field.
func (db *Database) getLastNameID() (int, error) {
	const lastNameFieldName = "LAST_NAME"
	if db.lastNameID != 0 {
		return db.lastNameID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(lastNameFieldName)
	if fetchError != nil {
		return 0, errors.New("getLastNameID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.lastNameID = fieldID

	return fieldID, nil
}

// getEmailID gets the ID of the Email field.
func (db *Database) getEmailID() (int, error) {
	const emailFieldName = "EMAIL"
	if db.emailID != 0 {
		return db.emailID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(emailFieldName)
	if fetchError != nil {
		return 0, errors.New("getEmailID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.emailID = fieldID

	return fieldID, nil
}

// getLastPaymentID gets the ID of the last payment field.
func (db *Database) getLastPaymentID() (int, error) {

	if db.lastPaymentID != 0 {
		return db.lastPaymentID, nil
	}

	// This is the first call so we need to look up the ID.
	const fieldName = "VALUE_OF_LAST_PAYMENT"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getLastPaymentID: " + fetchError.Error())
	}

	// Set the ID so that we don't have to look up again.
	db.lastPaymentID = fieldID

	return fieldID, nil
}

// getDonationToSocietyID gets the ID of the donation to museum field.
func (db *Database) getDonationToSocietyID() (int, error) {
	const fieldName = "VALUE_OF_DONATION_TO_LDLHS"
	if db.donationToSocietyID != 0 {
		return db.donationToSocietyID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getDonationToSocietyID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.donationToSocietyID = fieldID

	return fieldID, nil
}

// getDonationMuseumID gets the ID of the donation to museum field.
func (db *Database) getDonationToMuseumID() (int, error) {
	const fieldName = "VALUE_OF_DONATION_TO_THE_MUSEUM"
	if db.donationToMuseumID != 0 {
		return db.donationToMuseumID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getDonationToMuseumID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.donationToMuseumID = fieldID

	return fieldID, nil
}

// getFriendID gets the ID of the donation to museum field.
func (db *Database) getFriendID() (int, error) {
	const fieldName = "FRIEND_OF_THE_MUSEUM"
	if db.friendID != 0 {
		return db.friendID, nil
	}

	// This is the first call so we need to look up the ID.
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getFriendID: " + fetchError.Error())
	}

	// Set the global ID so that we don't have to look up again.
	db.friendID = fieldID

	return fieldID, nil
}

// getMembersAtAddressID gets the ID of the members at this address field.
func (db *Database) getMembersAtAddressID() (int, error) {
	return db.getFieldIDOnce(
		"MEMBERS_AT_ADDRESS",
		&db.membersAtAddressID)
}

// getFriendsAtAddressID gets the ID of the friends at this address field.
func (db *Database) getFriendsAtAddressID() (int, error) {
	return db.getFieldIDOnce(
		"NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS",
		&db.friendsAtAddressID)
}

// getFieldIDOnce returns the ID of a field from adm_user_fields.
// It stores the value in the given cache and uses that in subsequent
// calls.
func (db *Database) getFieldIDOnce(fieldName string, cache *int) (int, error) {
	if *cache != 0 {
		return *cache, nil
	}

	// This is the first call so we need to look up the ID.
	var fetchError error
	*cache, fetchError = db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getFieldIDOnce: " + fetchError.Error())
	}

	return *cache, nil
}

// getFieldID gets the ID of the given field.
func (db *Database) getFieldID(name string) (int, error) {

	const sql = `SELECT usf_id from adm_user_fields where usf_name_intern = $1`

	result, queryError := db.Query(sql, name)
	if queryError != nil {
		return 0, queryError
	}

	defer result.Close()

	if !result.Next() {
		return 0, errors.New("getFieldID: " + name + " not found")
	}
	var id int
	scanError := result.Scan(&id)
	if scanError != nil {
		return 0, scanError
	}
	return id, nil
}

func createMembershipSalesRecord() {
	// Create a membership_sales row:

	// ms_id integer NOT NULL,
	// ms_membership_year integer NOT NULL,
	// ms_usr1_id integer NOT NULL,
	// ms_usr1_friend boolean NOT NULL,
	// ms_usr2_id integer,
	// ms_usr2_friend boolean,
	// ms_payment_service CHARACTER VARYING(36) NOT NULL,
	// ms_payment_id CHARACTER VARYING(200) NOT NULL,
	// ms_usr1_fee REAL NOT NULL,
	// ms_usr2_fee REAL,
	// ms_usr1_friend_fee REAL,
	// ms_usr2_friend_fee REAL,
	// ms_donation REAL,
	// ms_donation_museum REAL,
}

func a() {
	const sql = `SELECT users.usr_id as id
        FROM adm_users as users
        LEFT JOIN adm_members as members
        ON users.usr_id = members.mem_usr_id
        AND members.mem_begin <= NOW()
        AND members.mem_end >= NOW()
        INNER JOIN adm_roles as roles
			ON roles.rol_id = members.mem_rol_id
			AND roles.rol_name = 'Member'
		LEFT JOIN adm_user_data AS email
			ON email.usd_usr_id = users.usr_id
			AND email.usd_usf_id = $1
		LEFT JOIN adm_user_data AS permission
			ON permission.usd_usr_id = users.usr_id
			AND permission.usd_usf_id = $2
		WHERE email.usd_value IS NOT NULL
			AND permission.usd_value IS NULL
		ORDER BY users.usr_id
		`
}
