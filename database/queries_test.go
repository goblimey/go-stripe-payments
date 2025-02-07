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

// TestSetMembersAtAddress checks SetMembersAtAddress.
func TestSetMembersAtAddress(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			break
		}

		defer db.Close()

		userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
		// userID, searchErr := db.GetUserIDofMember("simon", "ritchie", "simonritchie.uk@gmail.com")
		if searchErr != nil {
			t.Error(" - expected Schmidt to exist")
		}

		const want = 5
		setError := db.SetMembersAtAddress(userID, want)
		if setError != nil {
			t.Error(setError)
		}

		const sql = `select usd_value from adm_user_data
		where usd_usr_id = $1
		AND usd_usf_id = $2`
		var got int
		err := db.QueryRow(sql, userID, db.membersAtAddressID).Scan(&got)
		if err != nil {
			t.Error(err)
			break
		}

		if want != got {
			t.Errorf("want %d got %d", want, got)
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

		var got string
		queryAndScanError := db.QueryRow(sqlCommand, userID, db.dateLastPaidID).Scan(&got)
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

	defer db.Close()

	var testData = []struct {
		description string
		input       MembershipSale
		want        MembershipSale
	}{
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
			t.Error(td.description + ": " + createError.Error())
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

		td.want.PaymentID = wantPaymentID
		td.want.PaymentStatus = wantPaymentStatus

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

		defer db.Close()

		fieldID, fieldError := db.getFieldIDOnce("DATE_LAST_PAID", &db.dateLastPaidID)
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
