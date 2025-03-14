package database

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// values for the ms_transaction_type field of the membershipsale database
// table.
const TransactionTypeNewMember = "new member"
const TransactionTypeRenewal = "membership renewal"

var regExpForPostgresParamsToSQLiteParams *regexp.Regexp

// init should always work but if any of the calls in it fail, it will
// crash the application.
func init() {
	// Set up the regular expression or die.
	regExpForPostgresParamsToSQLiteParams = regexp.MustCompile(`\$[0-9]+`)
}

// Role holds the data about a role in the adm_roles table.  The roles
// are already created.  The adm_roles table has a lot of fields but we
// can ignore most of them.  We only need the rol_id and the rol_name
// fields.
type Role struct {
	ID   int64  `json:"rol_id"`
	Name string `json:"rol_name"`
}

/*
CREATE TABLE public.adm_roles (
    rol_id integer NOT NULL,
    rol_cat_id integer NOT NULL,
    rol_lst_id integer,
    rol_uuid character varying(36) NOT NULL,
    rol_name character varying(100) NOT NULL,
    rol_description character varying(4000),
    rol_assign_roles boolean DEFAULT false NOT NULL,
    rol_approve_users boolean DEFAULT false NOT NULL,
    rol_announcements boolean DEFAULT false NOT NULL,
    rol_events boolean DEFAULT false NOT NULL,
    rol_documents_files boolean DEFAULT false NOT NULL,
    rol_edit_user boolean DEFAULT false NOT NULL,
    rol_guestbook boolean DEFAULT false NOT NULL,
    rol_guestbook_comments boolean DEFAULT false NOT NULL,
    rol_mail_to_all boolean DEFAULT false NOT NULL,
    rol_mail_this_role smallint DEFAULT 0 NOT NULL,
    rol_photo boolean DEFAULT false NOT NULL,
    rol_profile boolean DEFAULT false NOT NULL,
    rol_weblinks boolean DEFAULT false NOT NULL,
    rol_all_lists_view boolean DEFAULT false NOT NULL,
    rol_default_registration boolean DEFAULT false NOT NULL,
    rol_leader_rights smallint DEFAULT 0 NOT NULL,
    rol_view_memberships smallint DEFAULT 0 NOT NULL,
    rol_view_members_profiles smallint DEFAULT 0 NOT NULL,
    rol_start_date date,
    rol_start_time time without time zone,
    rol_end_date date,
    rol_end_time time without time zone,
    rol_weekday smallint,
    rol_location character varying(100),
    rol_max_members integer,
    rol_cost double precision,
    rol_cost_period smallint,
    rol_usr_id_create integer,
    rol_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    rol_usr_id_change integer,
    rol_timestamp_change timestamp without time zone,
    rol_valid boolean DEFAULT true NOT NULL,
    rol_system boolean DEFAULT false NOT NULL,
    rol_administrator boolean DEFAULT false NOT NULL
);
*/

// User holds the contents of a row from the adm_users table.
type User struct {
	ID                    int       `json:"usr_id"`
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

func (db *Database) DeleteUser(u *User) error {
	sql := `
		DELETE FROM adm_users
		WHERE usr_id = $1
		RETURNING usr_id;
	`

	var usr_id int

	err := db.QueryRow(sql, u.ID).Scan(&usr_id)

	if err != nil {
		msg := fmt.Sprintf("DeleteUser: error deleting user with ID %d", u.ID)
		return errors.New(msg)
	}
	if usr_id != u.ID {
		msg := fmt.Sprintf("DeleteUser: returned ID %d not %d", usr_id, u.ID)
		return errors.New(msg)
	}

	// Success!
	u.ID = 0
	return nil
}

// Member holds data from an adm_members record.  There are
// many members for each user, one per role (Member, Admin etc)
type Member struct {
	ID        int    `json:"mem_id"`
	UserID    int    `json:"mem_usr_id"`
	RoleID    int    `json:"mem_rol_id"`
	UUID      string `json:"mem_uuid"`
	StartDate string `json:"mem_begin"`
	EndDate   string `json:"mem_end"`
	Approved  int    `json:"mem_approved"`
}

func NewMember(userID int, startDate, endDate string) *Member {

	m := Member{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	return &m
}

func (db *Database) DeleteMember(m *Member) error {
	sql := `
		DELETE FROM adm_members
		WHERE mem_id = $1
		RETURNING mem_id;
	`

	var mem_id int

	err := db.QueryRow(sql, m.ID).Scan(&mem_id)

	if err != nil {
		msg := fmt.Sprintf("DeleteMember: error deleting member with ID %d", m.ID)
		return errors.New(msg)
	}
	if mem_id == 1 {
		msg := fmt.Sprintf("DeleteUser: returned ID %d not %d", mem_id, m.ID)
		return errors.New(msg)
	}

	// Success!
	m.ID = 0
	return nil
}

// MembershipSale represents the payment of a membership sale - the annual
// membership fee.
type MembershipSale struct {
	ID                       int
	PaymentService           string  // The payment processor eg "Stripe".
	PaymentStatus            string  // "pending", "complete" or "cancelled"
	PaymentID                string  // The transaction Id from the payment processor.
	TransactionType          string  // The transaction type, eg 'membership renewal'
	MembershipYear           int     // The membership year paid for.
	OrdinaryMemberID         int     // The user ID of the member
	OrdinaryMemberFee        float64 // The fee paid for ordinary membership.
	OrdinaryMemberIsFriend   bool    // True if the ordinary member is a friend of the museum.
	OrdinaryMemberFriendFee  float64 // The fee paid for the ordinary member to be a friend.
	OrdinaryMemberFirstName  string  // First name (for new members)
	OrdinaryMemberLastName   string  // Last name (for new members)
	OrdinaryMemberEmail      string  // Email address (for new members)
	DonationToSociety        float64 // donation to the society.
	DonationToMuseum         float64 // donation to the museum.
	Giftaid                  bool    // True if the member consents to Giftaid.
	AssociateMemberID        int     // The user ID of the associate member.
	AssociateMemberFee       float64 // the fee paid for associate membership.
	AssocMemberIsFriend      bool    // True if the associate member is a friend of the museum.
	AssociateMemberFriendFee float64 // The fee paid for associate member to be a fiend.
	AssociateMemberFirstName string  // First name (for new members)
	AssociateMemberLastName  string  // Last name (for new members)
	AssociateMemberEmail     string  // Email address (for new members)
}

func NewMembershipSale(ordinaryMemberFee float64) *MembershipSale {

	sale := MembershipSale{
		OrdinaryMemberFee: ordinaryMemberFee,
	}

	return &sale
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
		// Initially the payment_id may be an empty string.   It will be set to the
		// Stripe transaction ID later.
		const createStatement = `
	INSERT INTO membership_sales (
		ms_id, 
		ms_payment_service,
		ms_payment_status,
		ms_payment_id,
		ms_transaction_type,
		ms_membership_year,
		ms_usr1_id,
		ms_usr1_fee,
		ms_usr1_friend,
		ms_usr1_friend_fee,
		ms_usr1_first_name,
		ms_usr1_last_name,
		ms_usr1_email,
		ms_usr2_id,
		ms_usr2_fee,
		ms_usr2_friend,
		ms_usr2_friend_fee,
		ms_usr2_first_name,
		ms_usr2_last_name,
		ms_usr2_email,
		ms_donation,
		ms_donation_museum,
		ms_giftaid
	) 
	VALUES
	(
	 	nextval('membership_sales_ms_id_seq'), 
	 	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	)
	RETURNING ms_id;
	`
		createError = db.QueryRow(
			createStatement,
			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.OrdinaryMemberID,
			ms.OrdinaryMemberFee,
			ms.OrdinaryMemberIsFriend,
			ms.OrdinaryMemberFriendFee,
			ms.OrdinaryMemberFirstName,
			ms.OrdinaryMemberLastName,
			ms.OrdinaryMemberEmail,
			ms.AssociateMemberID,
			ms.AssociateMemberFee,
			ms.AssocMemberIsFriend,
			ms.AssociateMemberFriendFee,
			ms.AssociateMemberFirstName,
			ms.AssociateMemberLastName,
			ms.AssociateMemberEmail,
			ms.DonationToSociety,
			ms.DonationToMuseum,
			ms.Giftaid,
		).Scan(&id)
	} else {
		// createStatement is the template to create a record in MembershipSales
		// with a null ms_usr2_id (no associate member).
		// Initially the payment_id may be an empty string.   It will be set to the
		// Stripe transaction ID later.
		const createStatement = `
			INSERT INTO membership_sales (
				ms_id,
				ms_payment_service,
				ms_payment_status,
				ms_payment_id,
				ms_transaction_type,
				ms_membership_year,
				ms_usr1_id,
				ms_usr1_fee,
				ms_usr1_friend,
				ms_usr1_friend_fee,
				ms_usr1_first_name,
				ms_usr1_last_name,
				ms_usr1_email,
				ms_donation,
				ms_donation_museum,
				ms_giftaid,
				ms_usr2_fee,
				ms_usr2_friend_fee,
				ms_usr2_friend
			) 
			Values
			(
				nextval('membership_sales_ms_id_seq'), 
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 0.0, 0.0, 'f'
			)
			RETURNING ms_id;
		`
		// The payment ID may be an empty string, to be set later.
		createError = db.QueryRow(
			createStatement,
			ms.PaymentService,
			ms.PaymentStatus,
			ms.PaymentID,
			ms.TransactionType,
			ms.MembershipYear,
			ms.OrdinaryMemberID,
			ms.OrdinaryMemberFee,
			ms.OrdinaryMemberIsFriend,
			ms.OrdinaryMemberFriendFee,
			ms.OrdinaryMemberFirstName,
			ms.OrdinaryMemberLastName,
			ms.OrdinaryMemberEmail,
			ms.DonationToSociety,
			ms.DonationToMuseum,
			ms.Giftaid,
		).Scan(&id)

	}

	if createError == nil {
		// Set the id in the membership sales object.
		ms.ID = id
	}

	return id, createError
}

// Update() updates a MembershipSale record in the database,
// setting the payment status (not null) and the payment Id (nullable).
func (ms *MembershipSale) Update(db *Database, status, paymentID string) error {

	sql := `
		UPDATE MEMBERSHIP_SALES 
		SET ms_payment_status=$1, ms_payment_id=$2
		WHERE ms_id=$3
		RETURNING ms_id;
	`

	var returnedID int
	execAndScanError := db.QueryRow(sql, status, paymentID, ms.ID).Scan(&returnedID)
	if execAndScanError != nil {
		return execAndScanError
	}

	if returnedID == 0 {
		em := fmt.Sprintf("MembershipSale.Update: zero return updating ID %d", ms.ID)
		return errors.New(em)
	}

	if returnedID != ms.ID {
		em := fmt.Sprintf(
			"MembershipSale.update: updating ID %d, got ID %d  back", ms.ID, returnedID)
		return errors.New(em)
	}

	// Success.
	return nil
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
	total := ms.OrdinaryMemberFee +
		ms.DonationToSociety +
		ms.DonationToMuseum

	if ms.OrdinaryMemberIsFriend {
		total += ms.OrdinaryMemberFriendFee
	}

	if ms.AssociateMemberID > 0 {
		total += ms.AssociateMemberFee

		if ms.AssocMemberIsFriend {
			total += ms.AssociateMemberFriendFee
		}
	}

	return total
}

// TotalPaymentForDisplay adds up the fees and returns the total
// as a string showing a number to two decimal places.
func (ms *MembershipSale) TotalPaymentForDisplay() string {
	return fmt.Sprintf("%.2f", ms.TotalPayment())
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (ms *MembershipSale) OrdinaryMembershipFeeForDisplay() string {
	return fmt.Sprintf("%.2f", ms.OrdinaryMemberFee)
}

// DonationToMuseumForDisplay gets the donation to museum
// for a display - a number to two decimal places.
func (ms *MembershipSale) DonationToMuseumForDisplay() string {
	return fmt.Sprintf("%.2f", ms.DonationToMuseum)
}

// DonationToSocietyForDisplay gets the donation to the society
// for a display - a number to two decimal places.
func (ms *MembershipSale) DonationToSocietyForDisplay() string {
	return fmt.Sprintf("%.2f", ms.DonationToSociety)
}

// AssociateMembershipFeeForDisplay gets the associate membership fee
// for display - a number to two decimal places.  If there is no
// associate, it returns "0.0".
func (ms *MembershipSale) AssociateMembershipFeeForDisplay() string {
	if ms.AssociateMemberID == 0 {
		return "0.00"
	}

	return fmt.Sprintf("%.2f", ms.AssociateMemberFee)
}

// OrdinaryMemberFriendFeeForDisplay gets the ordinary member's
// museum friend fee for display - a number to two decimal places.
// If the member is not a friend, it returns "0.0".
func (ms *MembershipSale) OrdinaryMemberFriendFeeForDisplay() string {

	if !ms.OrdinaryMemberIsFriend {
		return "0.00"
	}

	return fmt.Sprintf("%.2f", ms.OrdinaryMemberFriendFee)
}

// AssociateMemberFriendFeeForDisplay gets the associate member's
// museum friend fee for display - a number to two decimal places.
// If there is no associate or the associate is not a friend, it
// returns "0.0".
func (ms *MembershipSale) AssociateMemberFriendFeeForDisplay() string {

	if ms.AssociateMemberID == 0 {
		return "0.00"
	}

	if !ms.AssocMemberIsFriend {
		return "0.00"
	}

	return fmt.Sprintf("%.2f", ms.AssociateMemberFriendFee)
}

// getIDOfMemberRole fetches the role named 'Member' and returns
// its ID.
func (db *Database) getIDOfMemberRole() (int, error) {

	const sql = `
		select rol_id from adm_roles where rol_name='Member'
	`

	var rol_id int

	err := db.QueryRow(sql).Scan(&rol_id)
	if err != nil {
		return 0, err
	}

	return rol_id, nil
}

// CreateUser creates a user with the valid flag set.  The password is
// locked so they need to use the password change mechanism to log in.
// It's assumed that a transaction is set up in the db object.
func (db *Database) CreateUser(loginName string) (*User, error) {

	const sql = `
		insert into adm_users
		(usr_uuid, usr_login_name, usr_password, usr_valid)
		values($1, $2, '*LK*', 't')
		RETURNING usr_id
	`

	uid, uuidError := CreateUuid(db.Transaction, "usr_uuid", "adm_users")
	if uuidError != nil {
		return nil, uuidError
	}

	var usr_id int
	var createError error
	usr := NewUser(loginName)

	createError = db.QueryRow(sql, uid, usr.LoginName).Scan(&usr_id)

	if createError != nil {
		return nil, createError
	}

	usr.ID = usr_id

	return usr, nil
}

func (db *Database) GetUser(id int) (*User, error) {
	const sql = `
		SELECT usr_id, usr_uuid, usr_login_name, usr_password, usr_valid
		FROM adm_users
		WHERE usr_id = $1;
	`

	u := NewUser("name to be overwritten")

	err := db.QueryRow(sql, id).Scan(&u.ID, &u.UUID, &u.LoginName, &u.Password, &u.Valid)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// CreateMember creates a new record in the adm_members table and
// returns it.  The member has associated role 'Member'.  The format
// of the dates is 'YYYYMMDD'.  The mem_approved flag is set, which
// allows the user to log in.  It's assumed that a transaction is
// set up in the db object.
func (db *Database) CreateMember(userID int, startDate, endDate string) (*Member, error) {

	sqlPostgres := `
		insert into adm_members
		(mem_uuid, mem_rol_id, mem_usr_id, mem_begin, mem_end, mem_approved)
		values($1, $2, $3, TO_DATE($4, 'YYYY-MM-DD'), TO_DATE($5, 'YYYY-MM-DD'), 1)
		RETURNING mem_id;
	`

	sqlSQLite := `
		insert into adm_members
		(mem_uuid, mem_rol_id, mem_usr_id, mem_begin, mem_end, mem_approved)
		values(?, ?, ?, ?, ?, 1)
		RETURNING mem_id;
	`

	var sql string
	switch db.Type {
	case "postgres":
		sql = sqlPostgres
	default:
		sql = sqlSQLite
	}

	member_role_id, fetchRoleError := db.getIDOfMemberRole()
	if fetchRoleError != nil {
		return nil, fetchRoleError
	}

	uid, uuidError := CreateUuid(db.Transaction, "mem_uuid", "adm_members")
	if uuidError != nil {
		return nil, uuidError
	}

	mbr := NewMember(userID, startDate, endDate)
	mbr.RoleID = member_role_id

	var mem_id int

	createError := db.QueryRow(sql, uid, member_role_id, mbr.UserID, mbr.StartDate, mbr.EndDate).Scan(&mem_id)

	if createError != nil {
		return nil, createError
	}

	mbr.ID = mem_id

	return mbr, nil
}

func (db *Database) GetMember(id int) (*Member, error) {
	const sql = `
		SELECT mem_id, mem_uuid, mem_rol_id, mem_usr_id, mem_begin, mem_end, mem_approved
		FROM adm_members
		WHERE mem_id = $1;
	`

	// Create a new member object.  Any values passed will be overwritten by the scan.
	m := NewMember(0, "1970-01-01", "1970-01-01")

	err := db.QueryRow(sql, id).Scan(&m.ID, &m.UUID, &m.RoleID, &m.UserID, &m.StartDate, &m.EndDate)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// GetMemberOfUser gets the member record with role 'Member' associated
// with the user with the given ID.
func (db *Database) GetMemberOfUser(id int) (*Member, error) {
	const sql = `
		SELECT m.mem_id, m.mem_uuid, m.mem_rol_id, m.mem_usr_id, 
			m.mem_begin, m.mem_end, m.mem_approved
		FROM adm_members as m
		LEFT JOIN adm_roles as r
			ON r.rol_id = m.mem_rol_id
			AND r.rol_name = 'Member'
		WHERE m.mem_usr_id = $1;
	`

	// Create a new member object.  Any values passed will be overwritten by the scan.
	m := NewMember(0, "1970-01-01", "1970-01-01")

	err := db.QueryRow(sql, id).Scan(&m.ID, &m.UUID, &m.RoleID, &m.UserID, &m.StartDate, &m.EndDate, &m.Approved)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// CreateAccounts creates accounts for an ordinary member and for an
// associate member, if given.  Each account is represented by a record in
// the adm_users table, a linked record in adm_members with role Member
// and, if possible, a linked record in adm_user_data giving the user's
// email address (which is required to change their password).  The given
// sale record supplies the data.  It's assumed that the db object
// contains a transaction.  The ID of the ordinary user is returned.
//
// The ordinary member's account name is their email address, which
// simplifies the password change process.  The associate member may not
// have given their email address, in which case the login name is formed
// from their first and last name ("first.last").  HOWEVER without an
// email address they can't get control of their account by setting their
// password.  It's just a record for us that they are a paid-up member.
func (db *Database) CreateAccounts(sale *MembershipSale, startTime time.Time) (int, int, error) {

	// The accounts start today and end at the end of the year given in
	// the sale record.  The format of the date strings is YYYY-MM-DD.
	start := startTime.Format("2006-01-02")
	end := fmt.Sprintf("%d-12-31", sale.MembershipYear)

	// Get the login name(s) from the sale.
	name := getLoginNames(sale)

	// Create the user for the ordinary account.
	ordUser, createUserError := db.CreateUser(name[0])
	if createUserError != nil {
		return 0, 0, createUserError
	}

	// Update the sale.
	sale.OrdinaryMemberID = ordUser.ID

	// Create the member record for the ordinary account.
	_, createMemberError := db.CreateMember(ordUser.ID, start, end)
	if createMemberError != nil {
		return 0, 0, createMemberError
	}

	// Set the date last paid.
	dlpError := db.SetDateLastPaid(sale.OrdinaryMemberID, startTime)
	if dlpError != nil {
		return 0, 0, dlpError
	}

	// If the sale includes payment for an associate member, set up a record for
	// them too.

	var assocUser *User
	var assocUserID int

	if len(name) > 1 {

		var createUserError error
		assocUser, createUserError = db.CreateUser(name[1])
		if createUserError != nil {
			return 0, 0, createUserError
		}

		// Set the result.
		assocUserID = assocUser.ID
		// Update the sale.
		sale.AssociateMemberID = assocUser.ID

		// Create the member record for the associate account.
		_, createMemberError := db.CreateMember(assocUser.ID, start, end)
		if createMemberError != nil {
			return 0, 0, createMemberError
		}
	}

	return ordUser.ID, assocUserID, nil
}

// getLoginNames get the ordinary member's login name (their email
// address) and, if there is an associate, their login name (
// email if given, otherwise their name in the form  "first.last" )
func getLoginNames(sale *MembershipSale) []string {

	result := make([]string, 0, 2)

	result = append(result, sale.OrdinaryMemberEmail)

	if sale.AssociateMemberID > 0 {
		// The sale includes an associate member, who may or may not
		// have an email address.

		if len(sale.AssociateMemberEmail) > 0 {
			result = append(result, sale.AssociateMemberEmail)
		} else {
			// No email address.  Use the name - "first.last".
			loginName := sale.AssociateMemberFirstName + "." +
				sale.AssociateMemberLastName
			result = append(result, loginName)
		}
	}

	return result
}

// GetMembershipSale gets the membership_sale record for the user with
// the given ID.
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
		ms_donation_museum,
		ms_giftaid
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
		&ms.OrdinaryMemberID,
		&ms.OrdinaryMemberIsFriend,
		&ms.AssociateMemberID,
		&ms.AssocMemberIsFriend,
		&ms.OrdinaryMemberFee,
		&ms.AssociateMemberFee,
		&ms.OrdinaryMemberFriendFee,
		&ms.AssociateMemberFriendFee,
		&ms.DonationToSociety,
		&ms.DonationToMuseum,
		&ms.Giftaid,
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

	defer row.Close()

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

// SetMemberEndDate sets the end date of a member to the end
// of the current year.  It returns an error if the user does
// not exist or has no member record with role 'Member'.
func (db *Database) SetMemberEndDate(userID int, year int) error {

	// This query gets the member ID, start and end date of a
	// member, given their user id.  A user can have many members,
	// one per role (admin, member etc).  We need the one with
	// role "Member".

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

	// Ensure that rows is closed if there is an error.
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
		var setDateError error
		var returnedID int
		var updateSQL string

		if db.Type == "sqlite" {

			// SQLite has no special date or timestamp format.  We store
			// timestamps as strings in the format "YYYY-MM-DD HH:MM:SS.SSS".
			// It doesn't support rowsAffected, so we use RETURNING.
			updateSQL = `
				UPDATE adm_members
				SET mem_end = ?
				WHERE mem_id =?
				RETURNING mem_id;
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

		setDateError = db.QueryRow(updateSQL, endDate, id).Scan(&returnedID)

		if setDateError != nil {
			em := fmt.Sprintf("%s: %v", funcName, setDateError)
			return errors.New(em)
		}

		if returnedID == 0 {
			em := fmt.Sprintf("%s: ID zero returned updating ID %d", funcName, id)
			return errors.New(em)
		}

		if returnedID != id {
			em := fmt.Sprintf("%s: updating ID %d, got ID %d  back",
				funcName, id, returnedID)
			return errors.New(em)
		}
	}

	// Success!
	return nil
}

// SetEmailField sets the email address field for the user in
// adm_user_data.
func (db *Database) SetEmailField(userID int, emailAddress string) error {

	fieldID, fieldError := db.getEmailID()
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField[string](db, fieldID, userID, emailAddress)
}

// SetLastPayment sets the date of last payment field in adm_user_data.
func (db *Database) SetLastPayment(userID int, payment float64) error {
	fieldID, fieldError := db.getLastPaymentID()
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField[float64](db, fieldID, userID, payment)
}

// SetDonationToSociety sets the donation to society field in adm_user_data.
func (db *Database) SetDonationToSociety(userID int, payment float64) error {
	fieldID, fieldError := db.getDonationToSocietyID()
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField[float64](db, fieldID, userID, payment)
}

// SetDonationToMuseum sets the donation to museum field in adm_user_data.
func (db *Database) SetDonationToMuseum(userID int, payment float64) error {
	fieldID, fieldError := db.getDonationToMuseumID()
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField[float64](db, fieldID, userID, payment)
}

// SetDateLastPaid sets the date last paid field in adm_user_data.
func (db *Database) SetDateLastPaid(userID int, d time.Time) error {

	fieldID, fieldError := db.getFieldID("DATE_LAST_PAID")
	if fieldError != nil {
		return fieldError
	}

	return db.SetDateFieldInUserData(fieldID, userID, d)
}

// SetFriendField sets the friend of the museum field for the user in
// adm_user_data.  Tick box fields are set to 0 or 1.
func (db *Database) SetFriendField(userID int, ticked bool) error {
	fieldID, fieldError := db.getFriendID()
	if fieldError != nil {
		return fieldError
	}

	if ticked {
		return SetUserDataField[int](db, fieldID, userID, 1)
	} else {
		return SetUserDataField[int](db, fieldID, userID, 0)
	}

}

// SetGiftaidField sets the giftaid field for the user in
// adm_user_data.  In the DB, tick box fields are set to 0 or 1.
func (db *Database) SetGiftaidField(userID int, ticked bool) error {

	fieldID, fieldError := db.getGiftaidID()
	if fieldError != nil {
		return fieldError
	}
	// If the member consents to giftaid, fill in the box.  In case
	// it's already set from last year but not this year, ensure that
	// the value in the DB record is reset.
	if ticked {
		return SetUserDataField[int](db, fieldID, userID, 1)
	} else {
		return SetUserDataField[int](db, fieldID, userID, 0)
	}
}

// GetGiftaidField gets the giftaid field for the user from
// adm_user_data.  In the DB, Tick box fields are set to 0 or 1.
func (db *Database) GetGiftaidField(userID int) (bool, error) {
	fieldID, fieldError := db.getGiftaidID()
	if fieldError != nil {
		em := fmt.Sprintf("GetGiftaidField: %v", fieldError)
		return false, errors.New(em)
	}

	fetchedValue, fetchError := db.GetUserDataIntField(fieldID, userID)
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
func (db *Database) SetMembersAtAddress(userID int, members int) error {
	fieldID, fieldError := db.getMembersAtAddressID()
	if fieldError != nil {
		return fieldError
	}
	return SetUserDataField[int](db, fieldID, userID, members)
}

// SetFriendsAtAddress sets the number of friends of the museum at the
// user's address in adm_user_data.
func (db *Database) SetFriendsAtAddress(userID int, members int) error {
	fieldID, fieldError := db.getFriendsAtAddressID()
	if fieldError != nil {
		return fieldError
	}

	return SetUserDataField[int](db, fieldID, userID, members)
}

// SetUserDataFloatField sets the field with ID fieldID in adm_user_data to a
// float64 value.  If a record for the field is missing, one is created.
func (db *Database) SetUserDataFloatField(fieldID, userID int, floatValue float64) error {

	f := "SetUserDataFloatField"
	var returnedID int
	var sqlCommand string

	if db.FieldSet(fieldID, userID) {

		// There is already a record for this field.  Update the value.
		sqlCommand = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_id;;
		`
	} else {

		// There is no record for that field.  Create one.
		if db.Type == "sqlite" {
			sqlCommand = `
				INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
				VALUES ($1, $2, $3)
				RETURNING usd_id;
			`
		} else {
			sqlCommand = `
				INSERT INTO adm_user_data(usd_id, usd_usr_id, usd_usf_id, usd_value)
				VALUES (nextval('public.adm_user_data_usd_id_seq'), $1, $2, $3)
				RETURNING usd_id;
			`
		}
	}

	execAndScanError := db.QueryRow(sqlCommand, floatValue, userID, fieldID).Scan(&returnedID)
	if execAndScanError != nil {
		return execAndScanError
	}

	if returnedID == 0 {
		em := fmt.Sprintf("MembershipSale.%s: zero return updating ID %d", f, returnedID)
		return errors.New(em)
	}

	return nil
}

// SetUserDataIntField sets the field with ID fieldID in adm_user_data to an
// int value.  (This includes tick boxes, which are set to 0 or 1.)  If a record
// for the field is missing, one is created.
func (db *Database) SetUserDataIntField(fieldID, userID int, intValue int) error {

	f := "SetUserDataIntField"

	var returnedID int
	var sqlCommand string

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update the value.
		sqlCommand = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_id;
		`
	} else {

		// There is no record for that field.  Create one.

		if db.Type == "sqlite" {

			sqlCommand = `
				INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
				VALUES ($1, $2, $3)
				RETURNING usd_id;
			`

		} else {

			sqlCommand = `
				INSERT INTO adm_user_data(usd_id, usd_value, usd_usr_id, usd_usf_id)
				VALUES (nextval('public.adm_user_data_usd_id_seq'), $1, $2, $3)
				RETURNING usd_id;
			`
		}
	}

	execAndScanError := db.QueryRow(sqlCommand, intValue, userID, fieldID).Scan(&returnedID)
	if execAndScanError != nil {
		return execAndScanError
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: zero return updating ID %d", f, returnedID)
		return errors.New(em)
	}

	return nil
}

// GetUserDataIntField gets the int value from the field with ID fieldID from
// adm_user_data.  (This includes tick boxes, which are set to 0 or 1.)  If
// a record for the user is not present, an error is returned.
func (db *Database) GetUserDataIntField(fieldID, userID int) (int, error) {

	f := "getUserDataIntField"

	const sqlCommand = `
		SELECT usd_value
		FROM adm_user_data
		WHERE usd_usr_id = $1
		AND usd_usf_id = $2;
	`

	var fetchedValue int
	queryAndScanError := db.QueryRow(sqlCommand, userID, fieldID).Scan(&fetchedValue)
	if queryAndScanError != nil {
		em := fmt.Sprintf("%s: %v", f, queryAndScanError)
		return 0, errors.New(em)
	}

	return fetchedValue, nil
}

// SetUserDataStringField sets the field with ID fieldID in adm_user_data to a
// string value.  If a record for the field is missing, one is created.
func SetUserDataField[T int | float64 | bool | string](db *Database, fieldID, userID int, val T) error {

	f := "SetUserDataField"

	var returnedID int
	var sqlCommand string
	var execAndScanError error
	switch {
	case db.FieldSet(fieldID, userID):
		// There is already a record for this field.  Update the value.
		sqlCommand = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_id;
		`
		execAndScanError = db.QueryRow(sqlCommand, val, userID, fieldID).Scan(&returnedID)

	case db.Type == "sqlite":
		// SQLite - There is no record for that field.  Create one.
		sqlCommand = `
			INSERT INTO adm_user_data(usd_id, usd_usr_id, usd_usf_id, usd_value)
			VALUES (?, ?, ?, ?)
			RETURNING usd_id;
		`
		nextSQLiteID++
		execAndScanError = db.QueryRow(sqlCommand, nextSQLiteID, userID, fieldID, val).Scan(&returnedID)

	default:
		// Postgres - There is no record for that field.  Create one.
		sqlCommand = `
			INSERT INTO adm_user_data(usd_id, usd_usr_id, usd_usf_id, usd_value)
			VALUES (nextval('public.adm_user_data_usd_id_seq'), $1, $2, $3)
			RETURNING usd_id;
		`
		execAndScanError = db.QueryRow(sqlCommand, userID, fieldID, val).Scan(&returnedID)
	}

	if execAndScanError != nil {
		return execAndScanError
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: zero return updating ID %d", f, returnedID)
		return errors.New(em)
	}

	return nil
}

// GetUserDataField gets the value of type T from the field with ID
// fieldID from adm_user_data.  If a record for the user is not present, an
// error is returned.
// func Print[T any](s []T) {
func GetUserDataField[T int | float64 | bool | string](db *Database, fieldID, userID int) (T, error) {

	f := "getUserDataField"

	const sqlCommand = `
		SELECT usd_value
		FROM adm_user_data
		WHERE usd_usr_id = $1
		AND usd_usf_id = $2;
	`

	var fetchedValue T
	queryAndScanError := db.QueryRow(sqlCommand, userID, fieldID).Scan(&fetchedValue)
	if queryAndScanError != nil {
		em := fmt.Sprintf("%s: %v", f, queryAndScanError)
		// Return the zero value of the type.
		var t T
		return t, errors.New(em)
	}

	return fetchedValue, nil
}

// SetDateFieldInUserData sets the field with ID fieldID in adm_user_data to an
// date value, eg '2025-10-30'.
func (db *Database) SetDateFieldInUserData(fieldID, userID int, t time.Time) error {

	f := "SetDateFieldInUserData"

	dateStr := t.Format("2006-01-02")

	// Neither Postgres nor SQLite support rowsAffected.  Use RETURNING.
	var sql string

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update it.
		sql = `
			UPDATE adm_user_data
			SET usd_value = $1
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_usr_id;
		`
	} else {

		// There is no record for this field.  Create and set it.
		sql = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES ($1, $2, $3)
			RETURNING usd_usr_id;
		`
	}

	var returnedID int

	execAndScanError := db.QueryRow(sql, dateStr, userID, fieldID).Scan(&returnedID)
	if execAndScanError != nil {
		em := fmt.Sprintf("%s: %v", f, execAndScanError)
		return errors.New(em)
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: ID zero returned updating ID %d", f, userID)
		return errors.New(em)
	}

	if returnedID != userID {
		em := fmt.Sprintf("%s: updating ID %d, got ID %d  back",
			f, userID, returnedID)
		return errors.New(em)
	}

	return nil
}

// SetTimeFieldInUserData sets the field with ID fieldID in adm_user_data to an
// time value.
func (db *Database) SetTimeFieldInUserData(fieldID, userID int, t time.Time) error {

	f := "SetTimeFieldInUserData"

	// Neither Postgres nor SQLite support rowsAffected.  Use RETURNING.
	var sqlCommand string

	if db.FieldSet(fieldID, userID) {
		// There is already a record for this field.  Update it.
		if db.Type == "postgres" {
			// Postgres uses to_timestamp() to set the timestamp from a string.
			sqlCommand = `
			UPDATE adm_user_data
			SET usd_value = to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS')
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3
			RETURNING usd_usr_id;
		`
		} else {
			sqlCommand = `
			UPDATE adm_user_data
			SET usd_value = ?
			WHERE usd_usr_id = ?
			AND usd_usf_id = ?
			RETURNING usd_usr_id;
		`
		}
	} else {

		// There is no record for this field.  Create and set it.
		if db.Type == "postgres" {
			sqlCommand = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES (to_timestamp($1, 'YYYY-MM-DD HH24:MI:SS+HH'), $2, $3)
			RETURNING usd_usr_id;
		`
		} else {
			sqlCommand = `
			INSERT INTO adm_user_data(usd_value, usd_usr_id, usd_usf_id)
			VALUES (?, ?, ?)
			RETURNING usd_usr_id;
		`
		}
	}

	var execAndScanError error
	var returnedID int

	timeStr := t.Format("2006-01-02 15:04:05")

	execAndScanError = db.QueryRow(sqlCommand, timeStr, userID, fieldID).Scan(&returnedID)
	if execAndScanError != nil {
		em := fmt.Sprintf("%s: %v", f, execAndScanError)
		return errors.New(em)
	}

	if returnedID == 0 {
		em := fmt.Sprintf("%s: ID zero returned updating ID %d", f, userID)
		return errors.New(em)
	}

	if returnedID != userID {
		em := fmt.Sprintf("%s: updating ID %d, got ID %d  back",
			f, userID, returnedID)
		return errors.New(em)
	}

	return nil
}

// FieldSet checks whether the given field in adm_user_data
// for the given user.
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

	defer rows.Close()

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
	fieldID, fetchError := db.getFieldID(firstNameFieldName)
	if fetchError != nil {
		return 0, errors.New("getFirstNameID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getLastNameID return the Id of the last name field.
func (db *Database) getLastNameID() (int, error) {
	const lastNameFieldName = "LAST_NAME"
	fieldID, fetchError := db.getFieldID(lastNameFieldName)
	if fetchError != nil {
		return 0, errors.New("getLastNameID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getEmailID gets the ID of the Email field.
func (db *Database) getEmailID() (int, error) {
	const emailFieldName = "EMAIL"
	fieldID, fetchError := db.getFieldID(emailFieldName)
	if fetchError != nil {
		return 0, errors.New("getEmailID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getLastPaymentID gets the ID of the last payment field.
func (db *Database) getLastPaymentID() (int, error) {
	const fieldName = "VALUE_OF_LAST_PAYMENT"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getLastPaymentID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getDonationToSocietyID gets the ID of the donation to museum field.
func (db *Database) getDonationToSocietyID() (int, error) {
	const fieldName = "VALUE_OF_DONATION_TO_LDLHS"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getDonationToSocietyID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getDonationMuseumID gets the ID of the donation to museum field.
func (db *Database) getDonationToMuseumID() (int, error) {
	const fieldName = "VALUE_OF_DONATION_TO_THE_MUSEUM"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getDonationToMuseumID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getFriendID gets the ID of the donation to museum field.
func (db *Database) getFriendID() (int, error) {
	const fieldName = "FRIEND_OF_THE_MUSEUM"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getFriendID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getGiftaidID gets the ID of the giftaid field in adm_user_fields.
func (db *Database) getGiftaidID() (int, error) {
	const fieldName = "GIFT_AID"
	fieldID, fetchError := db.getFieldID(fieldName)
	if fetchError != nil {
		return 0, errors.New("getGiftaidID: " + fetchError.Error())
	}
	return fieldID, nil
}

// getMembersAtAddressID gets the ID of the members at this address field.
func (db *Database) getMembersAtAddressID() (int, error) {
	return db.getFieldID("MEMBERS_AT_ADDRESS")
}

// getFriendsAtAddressID gets the ID of the friends at this address field.
func (db *Database) getFriendsAtAddressID() (int, error) {
	return db.getFieldID("NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS")
}

// getFieldID gets the ID of the given field.
func (db *Database) getFieldID(name string) (int, error) {

	const sql = `SELECT usf_id from adm_user_fields where usf_name_intern = $1`

	var id int
	scanError := db.QueryRow(sql, name).Scan(&id)
	if scanError != nil {
		return 0, scanError
	}

	return id, nil
}

// CreateUuid creates and returns a UUID which is unique
// in the given row of the given table.
func CreateUuid(tx *sql.Tx, field, table string) (string, error) {

	// Do this up to ten times until you get a UUID that's not already
	// used.  Each attempt is very unlikely to fail.
	for i := 0; i < 10; i++ {

		// Create a UUID.
		u := uuid.New().String()

		// Check that the UUID is not already in the table.
		// (This is theoretically possible but unlikely.)
		sql := fmt.Sprintf("select %s from %s where %s = '%s'",
			field, table, field, u)

		resultSet, err := tx.Query(sql)

		if err != nil {
			return "", err
		}

		defer resultSet.Close()

		// Get the results - a list of uuids.  Should be no items
		// or one item.  If there are no items, u is unique so
		// return it.  If we fnd any items, the uuid is already
		// in the table, it's not unique.  We try again until we
		// find one that is unique.
		if !resultSet.Next() {

			// Success!
			return u, nil
		}

		var fetchedUUID string
		scanError := resultSet.Scan(&fetchedUUID)
		if scanError != nil {
			return "", scanError
		}
	}

	// All attempts have failed.  This is very very unlikely but
	// possible.
	return "", errors.New("CreateUUID: clash creating Id for table " + table)
}
