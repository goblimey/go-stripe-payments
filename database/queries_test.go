package database

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// databaseList is a list of database types that will be used in
// integration tests.  (Exhaustive tests are done using only SQLite.)
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

			defer db.Connection.Close()

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

func TestGetFirstNameID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Connection.Close()

		gotID, gotErr := db.getFirstNameID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

func TestGetLastNameID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Connection.Close()

		gotID, gotErr := db.getLastNameID()

		if gotErr != nil {
			t.Error(db.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Type, gotID)
		}
	}
}

func TestGetEmailID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := SetupDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Connection.Close()

		gotID, gotErr := db.getEmailID()

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

		defer db.Connection.Close()

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
		{"all match", "luiGi", "SchmidT", "Foo@bar.com", true},
		{"email matches", "uiGi", "chmidT", "Foo@bar.com", true},
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

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	defer db.Connection.Close()

	userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
	if searchErr != nil {
		t.Error(" - expected Schmidt to exist")
	}

	err := db.SetLastPayment(userID, 2.5)
	if err != nil {
		t.Error(err)
	}
}

// TestSetMembersAtAddress checks SetMembersAtAddress.
func TestSetMembersAtAddress(t *testing.T) {

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	defer db.Connection.Close()

	// userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
	userID, searchErr := db.GetUserIDofMember("simon", "ritchie", "simonritchie.uk@gmail.com")
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
		return
	}

	if want != got {
		t.Errorf("want %d got %d", want, got)
	}
}

// TestSetDateLastPaid checks SetDateLastPaid.
func TestSetDateLastPaid(t *testing.T) {

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	defer db.Connection.Close()

	userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
	if searchErr != nil {
		t.Error(" - expected Schmidt to exist")
	}

	london, _ := time.LoadLocation("Europe/London")
	tm := time.Date(2024, time.February, 14, 0, 0, 0, 0, london)
	const want = "2024-02-14 00:00:00+00"
	setError := db.SetDateLastPaid(userID, tm)
	if setError != nil {
		t.Error(setError)
	}

	const sql = `select usd_value from adm_user_data
		where usd_usr_id = $1
		AND usd_usf_id = $2`
	var got string
	err := db.QueryRow(sql, userID, db.dateLastPaidID).Scan(&got)
	if err != nil {
		t.Error(err)
		return
	}

	if want != got {
		t.Errorf("want %s got %s", want, got)
	}
}

// TestMemberExists checks MemberExists.
func TestSetFriend(t *testing.T) {

	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	defer db.Connection.Close()

	userID, searchErr := db.GetUserIDofMember("luiGi", "SchmidT", "Foo@bar.com")
	if searchErr != nil {
		t.Error(" - expected Schmidt to exist")
	}

	err := db.SetFriendTickBox(userID, true)
	if err != nil {
		t.Error(err)
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

		defer db.Connection.Close()

		var userID int
		if db.Type == "sqlite" {
			userID = TestUserIDSQLite
		} else {
			userID = TestUserIDPostgres
		}

		checkError := checkMemberEndYear(db, userID, 2025)
		if checkError != nil {
			t.Error(checkError)
		}
	}
}

func TestMembershipSale(t *testing.T) {
	db, connError := SetupDBForTesting("postgres")

	if connError != nil {
		t.Error(connError)
		return
	}

	defer db.Connection.Close()

	var testData = []struct {
		description string
		input       MembershipSale
		want        MembershipSale
	}{
		{
			"all",
			MembershipSale{0, "a", "b", "x", 2025, 1, 24.0, true, 5, 2, 6.0, true, 7, 42.0, 43.0},
			MembershipSale{0, "a", "b", "", 2025, 1, 24.0, true, 5, 2, 6.0, true, 7, 42.0, 43.0},
		},
		{
			"no associate",
			MembershipSale{0, "c", "d", "x", 2025, 1, 24.0, true, 5, 0, 6.0, true, 7, 42.0, 43.0},

			MembershipSale{0, "c", "d", "", 2025, 1, 24.0, true, 5, 0, 0.0, false, 0, 42.0, 43.0},
		},
		{
			"associate, no donations",
			MembershipSale{0, "e", "f", "x", 2025, 1, 24.0, true, 5, 2, 6.0, true, 7, 0.0, 0.0},
			MembershipSale{0, "e", "f", "", 2025, 1, 24.0, true, 5, 2, 6.0, true, 7, 0.0, 0.0},
		},
		{
			"no associate, no donations",
			MembershipSale{0, "g", "h", "x", 2025, 1, 24.0, true, 5, 0, 6.0, true, 7, 0.0, 0.0},

			MembershipSale{0, "g", "h", "", 2025, 1, 24.0, true, 5, 0, 0.0, false, 0, 0.0, 0.0},
		},
	}

	for _, td := range testData {
		id, createError := td.input.Create(db)
		if createError != nil {
			t.Error(createError)
			return
		}

		if id == 0 {
			t.Error("expected the returned ID to be non-zero")
		}
		if td.input.ID != id {
			t.Error("expected the ID in the supplied object to be updated")
		}

		got, fetchError := db.GetMembershipSale(id)
		if fetchError != nil {
			t.Error(fetchError)
			return
		}

		// The id has been set in the stored record.  Set the ID in
		// the want to match.

		td.want.ID = got.ID

		if td.want != *got {
			t.Errorf("want %v\ngot  %v", td.want, got)
		}

		const wantPaymentID = "some very long text"
		const wantPaymentStatus = "complete"
		got.Update(db, wantPaymentStatus, wantPaymentID)

		fetchedMS, fetchError := db.GetMembershipSale(got.ID)
		if fetchError != nil {
			t.Error(fetchError)
			return
		}

		if wantPaymentID != fetchedMS.PaymentID {
			t.Errorf("want id %s got %s", wantPaymentID, fetchedMS.PaymentID)
			return
		}

		if wantPaymentStatus != fetchedMS.PaymentStatus {
			t.Errorf("want id %s got %s", wantPaymentStatus, fetchedMS.PaymentStatus)
			return
		}

		// Delete and check that it's deleted.

		savedID := got.ID

		deleteError := got.Delete(db)
		if deleteError != nil {
			t.Error(deleteError)
			return
		}

		if got.ID != 0 {
			t.Errorf("want id of deleted record to be 0, got %d", got.ID)
			return
		}

		// This should fail and ms should be nil.
		ms, expectedError := db.GetMembershipSale(savedID)

		if expectedError == nil {
			t.Error("expected an error")
		}

		if ms != nil {
			t.Error("expected nil")
		}

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
	{
		gotYear, err := db.GetMembershipYear(userID)

		if err != nil {
			return err
		}

		if gotYear != startingYear {
			em := fmt.Sprintf("setup - want starting year %d got %d", startingYear, gotYear)
			return errors.New(em)
		}

	}

	// To test, set the year to the given year and check.

	db.SetMemberEndDate(userID, targetYear)
	{
		gotYear, err := db.GetMembershipYear(userID)

		if err != nil {
			return err
		}

		if gotYear != targetYear {
			em := fmt.Sprintf("want starting year %d got %d", startingYear, gotYear)
			return errors.New(em)
		}
	}

	// Success!
	return nil
}
