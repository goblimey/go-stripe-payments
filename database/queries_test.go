package database

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// databaseList is a list of database types that will be used in
// integration tests.  (Exhaustive tests may be done using only SQLite.)
var databaseList = []string{"postgres", "sqlite"}

// TestGetPaymentYear checks that GetSellingYear correctly identifies
// the membership year that we should be selling on a given date.
func TestGetPaymentYear(t *testing.T) {

	locationLondon, locationError := time.LoadLocation("Europe/London")
	if locationError != nil {
		t.Error(locationError)
		return
	}

	var testData = []struct {
		description string
		timeForTest time.Time
		want        int
	}{
		{
			"just before start of selling next year",
			time.Date(2024, time.September, 30, 23, 59, 59, 999999999, locationLondon),
			2024,
		},
		{
			"just after start of selling next year",
			time.Date(2024, time.October, 1, 0, 0, 0, 0, locationLondon),
			2025,
		},
		{
			"well after start of selling next year",
			time.Date(2024, time.November, 5, 12, 30, 32, 500000, locationLondon),
			2025,
		},

		{
			"well into the year we are selling",
			time.Date(2024, time.February, 14, 1, 2, 3, 4, locationLondon),
			2024,
		},
	}

	for _, td := range testData {

		got := GetPaymentYear(td.timeForTest)

		if td.want != got {
			t.Errorf("%s want %d got %d", td.description, td.want, got)
			continue
		}
	}
}

// TestClose checks that Close does not throw an error if the database
// is open and quiet.
func TestClose(t *testing.T) {
	for _, dbType := range databaseList {
		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		err := db.Close()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGetMemberRole(t *testing.T) {
	db, connError := SetupDBForTesting("sqlite")

	if connError != nil {
		t.Error(connError)
		return
	}

	db.BeginTx()

	defer db.Rollback()
	defer db.Close()

	got, err := db.getIDOfMemberRole()
	if err != nil {
		t.Error(err)
		return
	}

	if got <= 0 {
		t.Errorf("want ID greater than 0, got %d", got)
	}
}

func TestCreateUser(t *testing.T) {
	for _, dbType := range databaseList {
		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.Close()

		const want = "foo"

		got, createError := db.CreateUser(want)

		if createError != nil {
			t.Errorf("%s: %v", dbType, createError)
			return
		}

		if got.ID == 0 {
			t.Error("expected ID to be non-zero")
		}

		if want != got.LoginName {
			t.Errorf("want %s got %s", want, got.LoginName)
		}

		deleteError := db.DeleteUser(got)

		if deleteError != nil {
			t.Error(deleteError)
		}

		// Expect the ID of the user object to be 0 after the
		// database record has been deleted.
		if got.ID != 0 {
			t.Errorf("want ID of 0 got %d", got.ID)
		}
	}
}

func TestCreateMember(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.Close()

		var wantUserID int
		switch dbType {
		case "postgres":
			wantUserID = TestUserIDPostgres
		default:
			wantUserID = TestUserIDSQLite
		}

		const wantStartDate = "1970-01-01"
		const wantEndDate = "2025-12-31"
		got, err := db.CreateMember(wantUserID, wantStartDate, wantEndDate)

		if err != nil {
			t.Error(err)
			return
		}

		if got.ID == 0 {
			t.Error("expected ID to be non-zero")
		}

		if wantUserID != got.UserID {
			t.Errorf("want %d got %d", wantUserID, got.UserID)
		}

		if wantStartDate != got.StartDate {
			t.Errorf("want %s got %s", wantStartDate, got.StartDate)
		}

		if wantEndDate != got.EndDate {
			t.Errorf("want %s got %s", wantEndDate, got.EndDate)
		}

		// Tidy up.
		deleteError := db.DeleteMember(got)
		if deleteError != nil {
			t.Error(deleteError)
		}

		// Expect the ID of the member object to be 0 after the
		// database record has been deleted.
		if got.ID != 0 {
			t.Errorf("want ID of 0 got %d", got.ID)
		}
	}
}

func TestGetLoginNames(t *testing.T) {
	var testData = []struct {
		description              string
		sale                     MembershipSale
		wantLen                  int
		wantOrdinaryAccountName  string
		wantAssociateAccountName string
	}{
		{
			"ordinary member and associate without email",
			MembershipSale{
				OrdinaryMemberEmail:      "foo@example.com",
				AssociateMemberID:        42,
				AssociateMemberFirstName: "Fred",
				AssociateMemberLastName:  "Smith",
			},
			2,
			"foo@example.com",
			"Fred.Smith",
		},
		{
			"ordinary member only",
			MembershipSale{
				OrdinaryMemberEmail: "foo@example.com",
			},
			1,
			"foo@example.com",
			"",
		},
		{
			"ordinary member and associate with email",
			MembershipSale{
				AssociateMemberID:    1,
				OrdinaryMemberEmail:  "foo@example.com",
				AssociateMemberEmail: "bar@example.com",
			},
			2,
			"foo@example.com",
			"bar@example.com",
		},
		{
			"ordinary member and associate without email",
			MembershipSale{
				AssociateMemberID:        42,
				OrdinaryMemberEmail:      "foo@example.com",
				AssociateMemberFirstName: "Fred",
				AssociateMemberLastName:  "Smith",
			},
			2,
			"foo@example.com",
			"Fred.Smith",
		},
	}

	for _, td := range testData {

		name := getLoginNames(&td.sale)

		if td.wantLen != len(name) {
			t.Errorf("%s: want %d got %d",
				td.description, td.wantLen, len(name))
		}

		if td.wantOrdinaryAccountName != name[0] {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantOrdinaryAccountName, name[0])
		}

		if len(name) > 1 {

			if td.wantAssociateAccountName != name[1] {
				t.Errorf("%s: want %s got %s",
					td.description, td.wantOrdinaryAccountName, name[1])
			}

			if len(name) != 2 {
				t.Errorf("%s: want 2 got %d",
					td.description, len(name))
			}
		}
	}
}

// TestCreateAccounts checks CreateAccounts
func TestCreateAccounts(t *testing.T) {

	// PaymentService
	// PaymentStatus
	// PaymentID
	// TransactionType
	// MembershipYear
	// OrdinaryMemberID
	// OrdinaryMemberFee
	// OrdinaryMemberIsFriend
	// OrdinaryMemberFriendFee
	// OrdinaryMemberFirstName
	// OrdinaryMemberLastName
	// OrdinaryMemberEmail
	// DonationToSociety
	// DonationToMuseum
	// Giftaid
	// AssociateMemberID
	// AssociateMemberFee
	// AssocMemberIsFriend
	// AssociateMemberFriendFee
	// AssociateMemberLastName
	// AssociateMemberEmail

	const wantPassword = "*LK*"

	var testData = []struct {
		description              string
		now                      time.Time // This controls the start date of the new member.
		sale                     MembershipSale
		wantOrdinaryAccountName  string
		wantStartDate            string
		wantEndDate              string
		wantAssociateAccountName string
	}{
		{
			"ordinary member only",
			time.Date(2024, time.October, 1, 12, 35, 15, 0, time.UTC),
			MembershipSale{
				TransactionType:     TransactionTypeNewMember,
				MembershipYear:      2025,
				OrdinaryMemberEmail: "foo@example.com",
			},
			"foo@example.com",
			"2024-10-01",
			"2025-12-31",
			"",
		},
		{
			"ordinary member and associate with email",
			time.Date(2024, time.July, 4, 8, 9, 10, 0, time.UTC),
			MembershipSale{
				TransactionType:      TransactionTypeNewMember,
				MembershipYear:       2024,
				AssociateMemberID:    42,
				OrdinaryMemberEmail:  "foo@example.com",
				AssociateMemberEmail: "bar@example.com",
			},
			"foo@example.com",
			"2024-07-04",
			"2024-12-31",
			"bar@example.com",
		},
		{
			"ordinary member and associate without email",
			time.Date(2025, time.February, 14, 12, 35, 15, 0, time.UTC),
			MembershipSale{
				TransactionType:          TransactionTypeNewMember,
				MembershipYear:           2025,
				OrdinaryMemberEmail:      "foo@example.com",
				AssociateMemberID:        4,
				AssociateMemberFirstName: "Fred",
				AssociateMemberLastName:  "Smith",
			},
			"foo@example.com",
			"2025-02-14",
			"2025-12-31",
			"Fred.Smith",
		},
	}

	for _, td := range testData {
		// CreateAccounts calls other functions to
		// do the work and they are tested with
		// Postgres.  So we can test this only using
		// SQLite.  We can also just create and then
		// roll back.
		db, connError := SetupDBForTesting("sqlite")

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.Close()

		u1, u2, createError := db.CreateAccounts(&td.sale, td.now)
		if createError != nil {
			t.Errorf("%s: %v", td.description, createError)
		}

		// Check that the database records have been created, for each
		// member: an adm_user record, an adm_member record and an
		// adm_user_data record with field name 'DATE_LAST_PAID' and
		// containing the given time in the format 'YYYY-MM-DD'

		user1, fetchUser1Error := db.GetUser(u1)
		if fetchUser1Error != nil {
			t.Errorf("%s: %v", td.description, fetchUser1Error)
		}

		// We don't know what the uuid is supposed to be so we can only check that
		// there is one.
		if len(user1.UUID) == 0 {
			t.Error("uuid is empty")
		}

		if td.wantOrdinaryAccountName != user1.LoginName {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantOrdinaryAccountName, user1.LoginName)
		}

		if wantPassword != user1.Password {
			t.Errorf("%s: want %s got %s",
				td.description, wantPassword, user1.Password)
		}

		if !user1.Valid {
			t.Errorf("%s: want valid", td.description)
		}

		member1, fetchMemberError := db.GetMemberOfUser(user1.ID)
		if fetchMemberError != nil {
			t.Error(fetchMemberError)
		}

		if member1.ID <= 0 {
			t.Errorf("%s: want ID > 0: %d", td.description, member1.ID)
		}

		if len(member1.UUID) == 0 {
			t.Error("uuid is empty")
		}

		if user1.ID != member1.UserID {
			t.Errorf("%s: want %d got %d",
				td.description, user1.ID, member1.UserID)
		}

		if td.wantStartDate != member1.StartDate {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantStartDate, member1.StartDate)
		}

		if td.wantEndDate != member1.EndDate {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantEndDate, member1.EndDate)
		}

		if td.sale.AssociateMemberID != 0 {

			user2, fetchUser2Error := db.GetUser(u2)

			if fetchUser2Error != nil {
				t.Errorf("%s: %v", td.description, fetchUser2Error)
			}

			// We don't know what the uuid is supposed to be so we can only check that
			// there is one.
			if len(user2.UUID) == 0 {
				t.Error("uuid is empty")
			}

			if td.wantAssociateAccountName != user2.LoginName {
				t.Errorf("%s: want %s got %s",
					td.description, td.wantAssociateAccountName, user2.LoginName)
			}

			if wantPassword != user2.Password {
				t.Errorf("%s: want %s got %s",
					td.description, wantPassword, user2.Password)
			}

			if !user2.Valid {
				t.Errorf("%s: want valid", td.description)
			}
		}

		fieldID, fieldError := db.getFieldID("DATE_LAST_PAID")
		if fieldError != nil {
			t.Error(fieldError)
		}

		sql := `
			SELECT usd_value
			FROM adm_user_data
			WHERE usd_usr_id = $2
			AND usd_usf_id = $3;
		`
		var dateLastPaid string

		fetchDateError := db.QueryRow(sql, user1.ID, fieldID).Scan(&dateLastPaid)
		if fetchDateError != nil {
			t.Error(fetchDateError)
		}

		if td.wantStartDate != dateLastPaid {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantStartDate, dateLastPaid)
		}

		// Done.
		db.Rollback()
	}
}

// TestGetFieldID tests getFieldID.
func TestGetFieldID(t *testing.T) {

	for _, dbType := range databaseList {
		var testData = []struct {
			fieldName string
		}{
			{"FIRST_NAME"},
			{"LAST_NAME"},
			{"EMAIL"},
		}

		for _, td := range testData {

			db, connError := SetupDBForTesting(dbType)

			if connError != nil {
				t.Error(connError)
				return
			}

			txError := db.BeginTx()
			if txError != nil {
				t.Error(txError)
				return
			}

			defer db.Rollback()
			defer db.Close()

			gotID, gotErr := db.getFieldID(td.fieldName)

			if gotErr != nil {
				t.Error(gotErr)
			}

			if gotID <= 0 {
				t.Errorf("%s %s want i > 0 got %d", dbType, td.fieldName, gotID)
			}
		}
	}
}

// TestGetFirstNameID checks getFirstNameID.
func TestGetFirstNameID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		gotID, gotErr := db.getFirstNameID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

// TestGetLastNameID checks getLastNameID.
func TestGetLastNameID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		gotID, gotErr := db.getLastNameID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

// TestGetEmailID checks getEmailID.
func TestGetEmailID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		gotID, gotErr := db.getEmailID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

// TestGetGiftaidID checks getGiftaidID.
func TestGetGiftaidID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		gotID, gotErr := db.getGiftaidID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

// TestMemberExists checks MemberExists.
func TestMemberExists(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		id, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		if searchErr != nil {
			t.Error(db.Type + " - expected Schmidt to exist")
		}

		if id <= 0 {
			t.Errorf("want id > 0 got %d", id)
		}
	}
}

// TestMemberExistsSQLite tests MemberExists thoroughly using an
// SQLite in-memory database.
func TestMemberExistsSQLite(t *testing.T) {

	db, connError := SetupDBForTesting("sqlite")

	if connError != nil {
		t.Error(connError)
		return
	}

	txError := db.BeginTx()
	if txError != nil {
		t.Error(txError)
		return
	}
	defer db.Rollback()
	defer db.Close()

	var testData = []struct {
		description string
		firstName   string
		lastName    string
		email       string
		shouldWork  bool
	}{
		{"no match", "junk", "junk", "junk", false},
		{"all match", "luiGi", "SchmidT", strings.ToUpper(TestAssociateEmail), true},
		{"email matches", "uiGi", "chmidT", TestAssociateEmail, true},
		{"names match", "luiGi", "SchmidT", "junk", true},
	}

	for _, td := range testData {

		id, searchError := db.GetUserIDofMember(td.firstName, td.lastName, td.email)
		if td.shouldWork {
			if searchError != nil {
				t.Error(searchError)
			}
			if id <= 0 {
				t.Errorf("%s want id > 0 got %d", td.description, id)
			}
		} else {
			// we expect the call to fail with an error.
			if searchError == nil {
				em := fmt.Sprintf("%s: expected an error", td.description)
				t.Error(em)
			}
		}
	}
}

// TestSetLastPayment checks SetLastPayment.
func TestSetLastPayment(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
		}

		err := db.SetLastPayment(userID, 2.5)
		if err != nil {
			t.Error(err)
		}
	}
}

// TestSetEmailField checks SetEmailField.
func TestSetEmailField(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()

		if txError != nil {
			t.Error(txError)
			return
		}

		defer db.Rollback()
		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
		}

		const want = "foo"

		err := db.SetEmailField(userID, want)
		if err != nil {
			t.Error(err)
		}

		fieldID, fieldError := db.getEmailID()
		if fieldError != nil {
			t.Error(fieldError)
			return
		}

		got, fetchError := GetUserDataField[string](db, fieldID, userID)
		if fetchError != nil {
			t.Error(fetchError)
			return
		}

		if want != got {
			t.Errorf("%s: want %s got %s", dbType, want, got)
			return
		}
	}
}

// TestSetMembersAtAddress checks SetMembersAtAddress.
func TestSetMembersAtAddress(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Errorf("%s: %v", dbType, connError)
			break
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Errorf("%s: %v", dbType, txError)
			break
		}
		defer db.Rollback()
		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		// userID, searchErr := db.GetUserIDofMember("simon", "ritchie", "simonritchie.uk@gmail.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
			break
		}

		const want = 5
		setError := db.SetMembersAtAddress(userID, want)
		if setError != nil {
			t.Errorf("%s: %v", dbType, setError)
			break
		}

		const sql = `
			select usd_value from adm_user_data
			where usd_usr_id = $1
			AND usd_usf_id = $2
		`

		id, fetchIDError := db.getMembersAtAddressID()
		if fetchIDError != nil {
			t.Errorf("%s: %v", dbType, fetchIDError)
			break
		}

		var got int
		err := db.QueryRow(sql, userID, id).Scan(&got)
		if err != nil {
			t.Errorf("%s %v", dbType, err)
			break
		}

		if want != got {
			t.Errorf("%s: want %d got %d", dbType, want, got)
			break
		}
	}
}

// TestSetDateLastPaid checks SetDateLastPaid.
func TestSetDateLastPaid(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			break
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
		}

		london, _ := time.LoadLocation("Europe/London")
		tm := time.Date(2024, time.February, 14, 1, 2, 3, 4, london)

		setError := db.SetDateLastPaid(userID, tm)
		if setError != nil {
			t.Errorf("%s: %v", dbType, setError)
			break
		}

		const sqlCommand = `
				select usd_value  
				from adm_user_data
				where usd_usr_id = $1
				AND usd_usf_id = $2`

		id, fieldError := db.getFieldID("DATE_LAST_PAID")
		if fieldError != nil {
			t.Errorf("%s: %v", dbType, fieldError)
			break
		}

		var got string
		queryAndScanError := db.QueryRow(sqlCommand, userID, id).Scan(&got)
		if queryAndScanError != nil {
			t.Errorf("%s: %v", dbType, queryAndScanError)
			break
		}

		// SQLite and postgres provide the resulting date/time value in slightly
		// different formats.
		const want = "2024-02-14"

		if want != got {
			t.Errorf("%s: want %s got %s", dbType, want, got)
			break
		}
	}
}

// TestSetFriendField checks SetFriendField.
func TestSetFriendField(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
		}

		err := db.SetFriendField(userID, true)
		if err != nil {
			t.Error(err)
		}
	}
}

// TestSetMemberEndDate checks SetMemberUpdate
func TestSetMemberEndDate(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		var userID int
		if db.Type == "sqlite" {
			userID = TestUserIDSQLite
		} else {
			userID = TestUserIDPostgres
		}

		// Set the member's end date to 2025 and check it.
		setError := db.SetMemberEndDate(userID, 2025)
		if setError != nil {
			t.Error(dbType + ": " + setError.Error())
			return
		}
		checkError2 := checkMemberEndYear(db, userID, 2025)
		if checkError2 != nil {
			t.Error(dbType + ": " + checkError2.Error())
			return
		}
	}
}

func TestSetGiftaidField(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		giftaidID, fetchIDErr := db.getGiftaidID()
		if fetchIDErr != nil {
			t.Errorf("%s: %v", db.Type, fetchIDErr)
		}

		var userID int
		if db.Type == "sqlite" {
			userID = TestUserIDSQLite
		} else {
			userID = TestAssociateIDPostgres
		}

		// Delete any existing giftaid field for the test user.
		const deleteGiftaidCMD = `
			DELETE FROM adm_user_data
			WHERE usd_usr_id = $1
			AND usd_usf_id = $2;
		`
		_, execError := db.Exec(deleteGiftaidCMD, userID, giftaidID)
		if execError != nil {
			t.Errorf("%s: %v", db.Type, execError)
			break
		}

		// Create a giftaid field set to true.
		createErr1 := db.SetGiftaidField(userID, true)
		if createErr1 != nil {
			t.Errorf("%s: %v", db.Type, createErr1)
			break
		}

		// Check the field - should be true.
		got1, err1 := db.GetGiftaidField(userID)

		if err1 != nil {
			t.Error(err1)
			return
		}

		if !got1 {
			t.Errorf("%s: expected giftaid to be set true for user %d", db.Type, userID)
		}

		// Update the giftaid field to false.
		createErr := db.SetGiftaidField(userID, false)
		if createErr != nil {
			t.Errorf("%s: %v", db.Type, createErr)
		}

		got2, err2 := db.GetGiftaidField(userID)

		if err2 != nil {
			t.Error(err2)
			return
		}
		// Check the field - should be false.
		if got2 {
			t.Errorf("%s: expected giftaid to be set false for user %d", db.Type, userID)
			return
		}
	}
}

func TestMembershipSale(t *testing.T) {

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	txError := db.BeginTx()
	if txError != nil {
		t.Error(txError)
		return
	}
	defer db.Rollback()
	defer db.Close()

	var testData = []struct {
		description string
		input       MembershipSale
		want        MembershipSale
	}{
		{
			"no associate, no donations",
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, Giftaid: true,
				AssociateMemberID: 0, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 0,
				DonationToMuseum: 0, Giftaid: true,
				AssociateMemberID: 0, AssociateMemberFee: 0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 0,
			},
		},
		{
			"no associate",
			MembershipSale{
				ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5,
				DonationToSociety: 2,
				DonationToMuseum:  6.0, Giftaid: true,
				AssociateMemberID: 0,
				// These values should be ignored.
				AssociateMemberFee:  42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: true,
				AssociateMemberID: 0, AssociateMemberFee: 0.0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 0.0,
			},
		},
		{
			"all",
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5,
				DonationToSociety: 2,
				DonationToMuseum:  6.0, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
		},

		{
			"associate, no donations",
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 0,
				DonationToMuseum: 0, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
		},
		{
			"no associate, no donations",
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, Giftaid: true,
				AssociateMemberID: 0, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
				OrdinaryMemberFee:      24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 0,
				DonationToMuseum: 0, Giftaid: true,
				AssociateMemberID: 0, AssociateMemberFee: 0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 0,
			},
		},
		{
			"ordinary member is friend", // Set just one bool value.
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: false,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: false,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 43.0,
			},
		},
		{
			"Gifaid", // Set just one bool value.
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: false, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 0.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: false, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: true,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: false, AssociateMemberFriendFee: 0.0,
			},
		},
		{
			"associate member is friend", // Set just one bool value.
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: false, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: false,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
			MembershipSale{
				ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
				MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres, OrdinaryMemberFee: 24.0,
				OrdinaryMemberIsFriend: false, OrdinaryMemberFriendFee: 5, DonationToSociety: 2,
				DonationToMuseum: 6.0, Giftaid: false,
				AssociateMemberID: TestAssociateIDPostgres, AssociateMemberFee: 42.0,
				AssocMemberIsFriend: true, AssociateMemberFriendFee: 43.0,
			},
		},
	}

	for _, td := range testData {
		id, createError := td.input.Create(db)
		if createError != nil {
			t.Errorf("postgres: %s %v", td.description, createError.Error())
			break
		}

		if id == 0 {
			t.Error("expected the returned ID to be non-zero")
			break
		}
		if td.input.ID != id {
			t.Error("expected the ID in the supplied object to be updated")
			break
		}

		got, fetchError := db.GetMembershipSale(id)
		if fetchError != nil {
			t.Error(fetchError)
			break
		}

		// The id has been set in the stored record.  Set the ID in
		// the want to match.

		td.want.ID = got.ID

		if td.want != *got {
			t.Errorf("%s\nwant %v\ngot  %v", td.description, td.want, *got)
			break
		}

		const wantPaymentID = "some very long text"
		const wantPaymentStatus = "complete"

		td.want.PaymentID = wantPaymentID
		td.want.PaymentStatus = wantPaymentStatus

		updateError := got.Update(db, wantPaymentStatus, wantPaymentID)
		if updateError != nil {
			t.Errorf("%s: %v", td.description, updateError)
			break
		}

		updatedMS, fetchError := db.GetMembershipSale(got.ID)
		if fetchError != nil {
			t.Errorf("%s: %v", td.description, fetchError)
			break
		}

		if td.want != *updatedMS {
			t.Errorf("%s\nwant %v\ngot  %v", td.description, td.want, *updatedMS)
			break
		}

		// Tidy up - delete the membershipsales record and check that it's deleted.

		savedID := got.ID

		deleteError := got.Delete(db)
		if deleteError != nil {
			t.Error(deleteError)
			break
		}

		if got.ID != 0 {
			t.Errorf("%s: want id of deleted record to be 0, got %d", td.description, got.ID)
			break
		}

		// This should fail and ms should be nil.
		ms, expectedError := db.GetMembershipSale(savedID)

		if expectedError == nil {
			t.Errorf("%s: expected an error", td.description)
			break
		}

		if ms != nil {
			t.Errorf("%s: expected nil", td.description)
			break
		}

	}
}

func TestMembershipSaleUpdateFailure(t *testing.T) {

	const wantPaymentID = "some very long text"
	const wantPaymentStatus = "complete"

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	txError := db.BeginTx()
	if txError != nil {
		t.Error(txError)
		return
	}
	defer db.Rollback()
	defer db.Close()

	sale := MembershipSale{
		ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
		MembershipYear: 2025, OrdinaryMemberID: TestUserIDPostgres,
		OrdinaryMemberFee:      24.0,
		OrdinaryMemberIsFriend: true, OrdinaryMemberFriendFee: 5,
		DonationToSociety: 2, DonationToMuseum: 6.0, Giftaid: true,
		AssociateMemberID: 0, AssociateMemberFee: 0.0,
		AssocMemberIsFriend: false, AssociateMemberFriendFee: 0.0,
	}

	id, createError := sale.Create(db)
	if createError != nil {
		t.Error(createError)
		return
	}

	if id == 0 {
		t.Error("expected the returned ID to be non-zero")
	}

	// Set the Id to a non-existent record.
	sale.ID++

	// Expect the update to fail.
	err := sale.Update(db, wantPaymentStatus, wantPaymentID)

	if err == nil {
		t.Error("expected an error")
	}

	// Tidy up.

	sale.ID = id
	deleteError := sale.Delete(db)
	if deleteError != nil {
		t.Error(deleteError)
		return
	}
}

// checkMemberEndYear is a helper function that checks that the year in
// the given user's member end date in the given database matches the
// given year.  It works for sqlite and postgres.
func checkMemberEndYear(db *Database, userID, targetYear int) error {

	// To prepare, set the user's end date to a year other than the
	// target, fetch it back and check it.

	startingYear := targetYear - 1

	db.SetMemberEndDate(userID, startingYear)

	gotYear1, err1 := db.GetMembershipYear(userID)

	if err1 != nil {
		return err1
	}

	if gotYear1 != startingYear {
		em := fmt.Sprintf("setup - want starting year %d got %d", startingYear, gotYear1)
		return errors.New(em)
	}

	// To test, set the year to the given year and check.
	db.SetMemberEndDate(userID, targetYear)
	gotYear2, err2 := db.GetMembershipYear(userID)

	if err2 != nil {
		return err2
	}

	if gotYear2 != targetYear {
		em := fmt.Sprintf("want starting year %d got %d", startingYear, gotYear2)
		return errors.New(em)
	}

	// Success!
	return nil
}

// TestMembershipSalesDisplays checks the display functions of the
// MembershipSale type.
func TestMembershipSalesDisplays(t *testing.T) {

	var testData = []struct {
		description                 string
		ms                          MembershipSale
		wantOrdinaryMembershipFee   string
		wantOrdinaryMemberFriendFee string
		wantDonationToSociety       string
		wantDonationToMuseum        string
		wantAssociateMembersFee     string
		wantAssociateFriendFee      string
		wantTotal                   string
	}{

		{
			"all",
			MembershipSale{
				OrdinaryMemberFee:        1.234,
				OrdinaryMemberIsFriend:   true,
				OrdinaryMemberFriendFee:  2.345,
				DonationToSociety:        3.456,
				DonationToMuseum:         4.567,
				AssociateMemberID:        1,
				AssociateMemberFee:       5.678,
				AssocMemberIsFriend:      true,
				AssociateMemberFriendFee: 6.789,
			},
			"1.23", "2.35", "3.46", "4.57", "5.68", "6.79", "24.07",
		},
		{
			"ordinary only",
			MembershipSale{
				OrdinaryMemberFee:        1.234,
				OrdinaryMemberFriendFee:  2.345,
				AssociateMemberFee:       5.678,
				AssocMemberIsFriend:      true,
				AssociateMemberFriendFee: 6.789,
			},
			"1.23", "0.00", "0.00", "0.00", "0.00", "0.00", "1.23",
		},
		{
			"ordinary member is friend",
			MembershipSale{
				OrdinaryMemberFee:       1.234,
				OrdinaryMemberIsFriend:  true,
				OrdinaryMemberFriendFee: 2.345,
			},
			"1.23", "2.35", "0.00", "0.00", "0.00", "0.00", "3.58",
		},
		{
			"associate member",
			MembershipSale{
				OrdinaryMemberFee:        1.234,
				OrdinaryMemberIsFriend:   false,
				OrdinaryMemberFriendFee:  2.345,
				AssociateMemberID:        1,
				AssociateMemberFee:       5.678,
				AssocMemberIsFriend:      false,
				AssociateMemberFriendFee: 6.789,
			},
			"1.23", "0.00", "0.00", "0.00", "5.68", "0.00", "6.91",
		},
		{
			"associate member who is friend",
			MembershipSale{
				OrdinaryMemberFee:        1.234,
				OrdinaryMemberFriendFee:  2.345,
				AssociateMemberID:        1,
				AssociateMemberFee:       5.678,
				AssocMemberIsFriend:      true,
				AssociateMemberFriendFee: 6.789,
			},
			"1.23", "0.00", "0.00", "0.00", "5.68", "6.79", "13.70",
		},
	}

	for _, td := range testData {

		if td.wantOrdinaryMembershipFee != td.ms.OrdinaryMembershipFeeForDisplay() {
			t.Errorf("%s: want ordinary member fee %s got %s",
				td.description,
				td.wantOrdinaryMembershipFee,
				td.ms.OrdinaryMembershipFeeForDisplay())
		}

		if td.wantOrdinaryMemberFriendFee != td.ms.OrdinaryMemberFriendFeeForDisplay() {
			t.Errorf("%s: want ordinary member friend fee %s got %s",
				td.description,
				td.wantOrdinaryMemberFriendFee,
				td.ms.OrdinaryMemberFriendFeeForDisplay())
		}

		if td.wantDonationToSociety != td.ms.DonationToSocietyForDisplay() {
			t.Errorf("%s: want donationToSociety %s got %s",
				td.description,
				td.wantDonationToSociety,
				td.ms.DonationToSocietyForDisplay())
		}

		if td.wantOrdinaryMemberFriendFee != td.ms.OrdinaryMemberFriendFeeForDisplay() {
			t.Errorf("%s: want ordinary member friend fee %s got %s",
				td.description,
				td.wantOrdinaryMemberFriendFee,
				td.ms.OrdinaryMemberFriendFeeForDisplay())
		}

		if td.wantDonationToSociety != td.ms.DonationToSocietyForDisplay() {
			t.Errorf("%s: want donation to society %s got %s",
				td.description,
				td.wantDonationToSociety,
				td.ms.DonationToSocietyForDisplay())
		}

		if td.wantDonationToMuseum != td.ms.DonationToMuseumForDisplay() {
			t.Errorf("%s: want donation to museum %s got %s",
				td.description,
				td.wantDonationToMuseum,
				td.ms.DonationToMuseumForDisplay())
		}

		if td.wantAssociateFriendFee != td.ms.AssociateMemberFriendFeeForDisplay() {
			t.Errorf("%s: want associate member friend fee %s got %s",
				td.description,
				td.wantAssociateFriendFee,
				td.ms.AssociateMemberFriendFeeForDisplay())
		}

		if td.wantTotal != td.ms.TotalPaymentForDisplay() {
			t.Errorf("%s: want total %s got %s",
				td.description,
				td.wantTotal,
				td.ms.TotalPaymentForDisplay())
		}
	}
}

// TestSetTimeFieldInUserData check SetTimeFieldInUserData.
func TestSetTimeFieldInUserData(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.Close()

		fieldID, fieldError := db.getFieldID("DATE_LAST_PAID")
		if fieldError != nil {
			t.Error(fieldError)
			return
		}

		utc, _ := time.LoadLocation("UTC")
		d := time.Date(2025, time.February, 14, 1, 2, 3, 4, utc)

		const wantPostgres = "2025-02-14 01:02:03+00"
		const wantSQLite = "2025-02-14 01:02:03"

		var userID int
		var want string
		if db.Type == "sqlite" {
			userID = TestUserIDSQLite
			want = wantSQLite
		} else {
			userID = TestUserIDPostgres
			want = wantPostgres
		}

		setError := db.SetTimeFieldInUserData(fieldID, userID, d)

		if setError != nil {
			t.Error(setError)
			return
		}

		// Fetch the value back and check it.
		sql := `
			select usd_value from adm_user_data 
			where usd_usr_id = $1
			and usd_usf_id = $2;
		`
		var got string
		fetchError := db.QueryRow(sql, userID, fieldID).Scan(&got)
		if fetchError != nil {
			t.Error(fetchError)
		}

		if want != got {
			t.Errorf("want %s got %s", want, got)
			return
		}
	}
}
