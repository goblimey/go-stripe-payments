package database

import (
	"reflect"
	"testing"
)

func TestConnectSQLite(t *testing.T) {

	db := New(&DBConfigForTestingWithSQLite)

	err := db.Connect()
	if err != nil {
		t.Error(err)
	}

	defer db.CloseAndDelete()

	if db == nil {
		t.Error("should return non-nil")
	}
}

// TestPostgresParamsToSQLiteParams checks the postgresParamsToSQLiteParams.
func TestPostgresParamsToSQLiteParams(t *testing.T) {

	const multilineQuery = `abc$21
		def$21`
	const multilineResult = `abc?
		def?`

	var testData = []struct {
		query string
		want  string
	}{
		{"abc$1def$2", "abc?def?"},
		{"$1$2$3", "???"},
		{"noparams", "noparams"},
		{"abc$21def$21", "abc?def?"},
		{multilineQuery, multilineResult},
	}

	for _, td := range testData {
		got := postgresParamsToSQLiteParams(td.query)
		if got != td.want {
			t.Errorf("got %s want %s", got, td.want)
		}
	}
}

// TestUpdateRowWithAssociate checks ms.UpdateRow when there is an ordinary member and an
// associate member.  (The Update separate logic and SQL for this.)
func TestUpdateRowWithAssociate(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.Logger = createLoggerForTesting()

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		// We need two users.
		u1Name, uuidError1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uuidError1 != nil {
			t.Error(uuidError1)
		}
		u1 := NewUser(u1Name)
		u1Err := db.CreateUser(u1)
		if u1Err != nil {
			t.Error(u1Err)
			return
		}

		u2Name, uuidError2 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uuidError2 != nil {
			t.Error(uuidError2)
		}
		u2 := NewUser(u2Name)
		u2Err := db.CreateUser(u2)
		if u2Err != nil {
			t.Error(u2Err)
			return
		}

		u3Name, uuidError3 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uuidError3 != nil {
			t.Error(uuidError3)
		}
		u3 := NewUser(u3Name)
		u3Err := db.CreateUser(u3)
		if u3Err != nil {
			t.Error(u3Err)
			return
		}

		// The update should only affect one records so ms1 should not be touched during this test.

		ms1 := MembershipSale{
			PaymentService:  "a",
			PaymentStatus:   "b",
			PaymentID:       "c",
			TransactionType: "d",
			MembershipYear:  2024,
		}

		ms1ID, ms1Err := ms1.Create(db)
		if ms1Err != nil {
			t.Error(ms1Err)
			return
		}
		_ = ms1ID

		ms2Orig := MembershipSale{
			PaymentService:        "e",
			PaymentStatus:         "f",
			PaymentID:             "g",
			TransactionType:       "h",
			MembershipYear:        2024,
			UserID:                u1.ID,
			OrdinaryMemberFeePaid: 1.2,
			Friend:                true,
			FriendFeePaid:         3.4,
			FirstName:             "i",
			LastName:              "j",
			Email:                 "k",
			DonationToSociety:     5.6,
			DonationToMuseum:      7.8,
			Giftaid:               true,
			AssocUserID:           u2.ID,
			AssocFeePaid:          9.1,
			AssocFriend:           true,
			AssocFriendFeePaid:    2.3,
			AssocFirstName:        "l",
			AssocLastName:         "m",
			AssocEmail:            "n",

			// Reference Data from the config - not stored in the DB, so must set false
			// for the later comparisons to work.
			EnableOtherMemberTypes: false,
			EnableGiftaid:          false,
		}

		ms2ID, ms2Err := ms2Orig.Create(db)
		if ms2Err != nil {
			t.Error(ms2Err)
			return
		}

		// ms2 should be a copy of ms2Orig.
		ms2, fetchMSError1 := db.GetMembershipSale(ms2ID)
		if fetchMSError1 != nil {
			t.Error(fetchMSError1)
			return
		}

		if !reflect.DeepEqual(*ms2, ms2Orig) {
			t.Error("fetched ms does not match original")
			return
		}

		// Now test the update with an associate.
		// Change all of the non-associate fields in ms2 that are stored in the database and save it.
		ms2.PaymentService = "eb"
		ms2.PaymentStatus = "fb"
		ms2.PaymentID = "gb"
		ms2.TransactionType = "hb"
		ms2.MembershipYear = 2025
		ms2.UserID = u2.ID
		ms2.OrdinaryMemberFeePaid = 1.25
		ms2.Friend = true
		ms2.FriendFeePaid = 3.45
		ms2.FirstName = "ib"
		ms2.LastName = "jb"
		ms2.Email = "kb"
		ms2.DonationToSociety = 5.65
		ms2.DonationToMuseum = 7.85
		ms2.Giftaid = false
		ms2.AssocUserID = u3.ID
		ms2.AssocFeePaid = 9.1
		ms2.AssocFriend = false
		ms2.AssocFriendFeePaid = 2.35
		ms2.AssocFirstName = "lb"
		ms2.AssocLastName = "mb"
		ms2.AssocEmail = "nb"

		ms2Copy := *ms2

		// Update
		ms2Err2 := ms2.Update(db)
		if ms2Err2 != nil {
			t.Error(ms2Err2)
			return
		}

		// Fetch.
		ms3, fetchMSError2 := db.GetMembershipSale(ms2ID)
		if fetchMSError2 != nil {
			t.Error(fetchMSError2)
			return
		}

		// Check
		if !reflect.DeepEqual(*ms3, ms2Copy) {
			t.Error("fetched ms does not match original")
			return
		}

		// The update should only affect one record.  A fairly simple mistake in the SQL
		// would make it update all records.  To guard against that, check that ms1 has not
		// been touched.
		ms1Fetched, fetchMS1Error := db.GetMembershipSale(ms1ID)
		if fetchMS1Error != nil {
			t.Error(fetchMS1Error)
			return
		}

		if !reflect.DeepEqual(*ms1Fetched, ms1) {
			t.Error("fetched ms does not match original - that record that should not be touched.")
			return
		}

	}
}

// TestUpdateRowWithNoAssociate checks ms.UpdateRow when there is just an ordinary member and no
// associate.  (The Update has separate logic and SQL for this.)
func TestUpdateRowWithNoAssociate(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.Logger = createLoggerForTesting()

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			return
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		// We need two users.
		u1Name, uuidError1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uuidError1 != nil {
			t.Error(uuidError1)
		}
		u1 := NewUser(u1Name)
		u1Err := db.CreateUser(u1)
		if u1Err != nil {
			t.Error(u1Err)
			return
		}

		u2Name, uuidError2 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uuidError2 != nil {
			t.Error(uuidError2)
		}
		u2 := NewUser(u2Name)
		u2Err := db.CreateUser(u2)
		if u2Err != nil {
			t.Error(u2Err)
			return
		}

		// The update should only affect one records so ms1 should not be touched during this test.
		ms1 := MembershipSale{
			PaymentService:  "a",
			PaymentStatus:   "b",
			PaymentID:       "c",
			TransactionType: "d",
			MembershipYear:  2024,
		}

		ms1ID, ms1Err := ms1.Create(db)
		if ms1Err != nil {
			t.Error(ms1Err)
			return
		}
		_ = ms1ID

		// Create a a sale for an an ordinary user (user1) with no associate.
		ms2Orig := MembershipSale{
			PaymentService:        "e",
			PaymentStatus:         "f",
			PaymentID:             "g",
			TransactionType:       "h",
			MembershipYear:        2024,
			UserID:                u1.ID,
			OrdinaryMemberFeePaid: 1.2,
			Friend:                true,
			FriendFeePaid:         3.4,
			FirstName:             "i",
			LastName:              "j",
			Email:                 "k",

			// Reference Data from the config - not stored in the DB, so must be false
			// for the later comparisons to work.  (That's the default but we set it
			// for clarity.))
			EnableOtherMemberTypes: false,
			EnableGiftaid:          false,
		}

		ms2ID, ms2Err := ms2Orig.Create(db)
		if ms2Err != nil {
			t.Error(ms2Err)
			return
		}

		// ms2 should be a copy of ms2Orig.
		ms2, fetchMSError1 := db.GetMembershipSale(ms2ID)
		if fetchMSError1 != nil {
			t.Error(fetchMSError1)
			return
		}

		if !reflect.DeepEqual(*ms2, ms2Orig) {
			t.Error("fetched ms does not match original")
			return
		}

		// Setup done.  Now test the update with no associate.
		// Change all of the fields in ms2 that are stored in the database and save it.  That
		// includes changing the user ID.
		ms2.PaymentService = "eb"
		ms2.PaymentStatus = "fb"
		ms2.PaymentID = "gb"
		ms2.TransactionType = "hb"
		ms2.MembershipYear = 2025
		ms2.UserID = u2.ID
		ms2.OrdinaryMemberFeePaid = 1.25
		ms2.Friend = true
		ms2.FriendFeePaid = 3.45
		ms2.FirstName = "ib"
		ms2.LastName = "jb"
		ms2.Email = "kb"
		ms2.DonationToSociety = 5.65
		ms2.DonationToMuseum = 7.85
		ms2.Giftaid = true
		ms2.EnableOtherMemberTypes = false
		ms2.EnableGiftaid = false

		ms2Copy := *ms2

		// Update
		ms2Err2 := ms2.Update(db)
		if ms2Err2 != nil {
			t.Error(ms2Err2)
			return
		}

		// Fetch.
		ms3, fetchMSError2 := db.GetMembershipSale(ms2ID)
		if fetchMSError2 != nil {
			t.Error(fetchMSError2)
			return
		}

		// Check
		if !reflect.DeepEqual(*ms3, ms2Copy) {
			t.Error("fetched ms does not match original")
			return
		}

		// The yupdate should only affect one record.  A fairly simple mistake in the SQL
		// would make it update all records.  To guard against that, check that ms1 has not
		// been touched.
		ms1Fetched, fetchMS1Error := db.GetMembershipSale(ms1ID)
		if fetchMS1Error != nil {
			t.Error(fetchMS1Error)
			return
		}

		if !reflect.DeepEqual(*ms1Fetched, ms1) {
			t.Error("fetched ms does not match original - that record that should not be touched.")
			return
		}

	}
}

// TestMembershipSalesDisplays checks the display functions of the
// MembershipSale type.
func TestMembershipSaleDisplays(t *testing.T) {

	var testData = []struct {
		description               string
		form                      MembershipSale
		wantOrdinaryMembershipFee string
		wantFriendFeePaid         string
		wantDonationToSociety     string
		wantDonationToMuseum      string
		wantAssociateMembersFee   string
		wantAssociateFriendFee    string
		wantTotalForDisplay       string
	}{

		{
			"all",
			MembershipSale{
				OrdinaryMemberFeePaid: 1.23,
				Friend:                true,
				FriendFeePaid:         3.46,
				DonationToSociety:     3.456,
				DonationToMuseum:      4.567,
				AssocFeePaid:          2.35,
				AssocUserID:           1,
				AssocFriend:           true,
				AssocFriendFeePaid:    6.789,
			},
			"£1.23", "£3.46", "£3.46", "£4.57", "2.35", "£6.79", "£21.85",
		},
		{
			"ordinary only",
			MembershipSale{
				OrdinaryMemberFeePaid: 1.23,
				FriendFeePaid:         0,
				AssocFeePaid:          0,
				AssocFriend:           true,
				AssocFriendFeePaid:    0,
			},
			"£1.23", "", "", "", "", "", "£1.23",
		},
		{
			"ordinary member is friend",
			MembershipSale{
				OrdinaryMemberFeePaid: 1.23,
				Friend:                true,
				FriendFeePaid:         2.345,
			},
			"£1.23", "£2.35", "", "", "", "", "£3.58",
		},
		{
			"associate member",
			MembershipSale{
				OrdinaryMemberFeePaid: 1.23,
				Friend:                false,
				FriendFeePaid:         0,
				AssocUserID:           1,
				AssocFeePaid:          5.678,
				AssocFriend:           false,
				AssocFriendFeePaid:    0,
			},
			"£1.23", "", "", "", "5.68", "", "£6.91",
		},
		{
			"associate member who is friend",
			MembershipSale{
				OrdinaryMemberFeePaid: 1.23,
				FriendFeePaid:         0,
				AssocUserID:           1,
				AssocFeePaid:          5.678,
				AssocFriend:           true,
				AssocFriendFeePaid:    6.789,
			},
			"£1.23", "", "", "", "5.68", "£6.79", "£13.70",
		},
	}

	for _, td := range testData {

		if td.wantOrdinaryMembershipFee != td.form.OrdinaryMemberFeeForDisplay() {
			t.Errorf("%s: want ordinary member fee for display %s got %s",
				td.description,
				td.wantOrdinaryMembershipFee,
				td.form.OrdinaryMemberFeeForDisplay())
		}

		if td.wantFriendFeePaid != td.form.FriendFeeForDisplay() {
			t.Errorf("%s: want ordinary member friend fee paid %s got %s",
				td.description,
				td.wantFriendFeePaid,
				td.form.FriendFeeForDisplay())
		}

		if td.wantDonationToSociety != td.form.DonationToSocietyForDisplay() {
			t.Errorf("%s: want donationToSociety %s got %s",
				td.description,
				td.wantDonationToSociety,
				td.form.DonationToSocietyForDisplay())
		}

		if td.wantFriendFeePaid != td.form.FriendFeeForDisplay() {
			t.Errorf("%s: want ordinary member friend fee %s got %s",
				td.description,
				td.wantFriendFeePaid,
				td.form.FriendFeeForDisplay())
		}

		if td.wantDonationToMuseum != td.form.DonationToMuseumForDisplay() {
			t.Errorf("%s: want donation to museum %s got %s",
				td.description,
				td.wantDonationToMuseum,
				td.form.DonationToMuseumForDisplay())
		}

		if td.wantAssociateFriendFee != td.form.AssocFriendFeeForDisplay() {
			t.Errorf("%s: want associate member friend fee %s got %s",
				td.description,
				td.wantAssociateFriendFee,
				td.form.AssocFriendFeeForDisplay())
		}

		if td.wantTotalForDisplay != td.form.TotalForDisplay() {
			t.Errorf("%s: want total %s got %s",
				td.description,
				td.wantTotalForDisplay,
				td.form.TotalForDisplay())
		}
	}
}
