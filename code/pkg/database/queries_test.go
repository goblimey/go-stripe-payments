package database

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

// databaseList is a list of database types that will be used in
// integration tests.
var databaseList = []string{"postgres", "sqlite"}

// TestGetMembershipYear checks that GetSellingYear correctly identifies
// the membership year that we should be selling on a given date.
func TestGetMembershipYear(t *testing.T) {
	println("TestGetMembershipYear")
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

		got := GetMembershipYear(td.timeForTest)

		if td.want != got {
			t.Errorf("%s want %d got %d", td.description, td.want, got)
			continue
		}
	}
}

// TestClose checks that Close does not throw an error if the database
// is open and quiet.
func TestClose(t *testing.T) {
	fmt.Println("TestClose")
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		// If either of these fail, the test will fail.
		defer db.Rollback()
		defer db.CloseAndDelete()
	}
}

func TestCreateSQLiteTables(t *testing.T) {
	fmt.Println("TestCreateSQLiteTables")
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Errorf("%s: %v", dbType, connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Errorf("%s: %v", dbType, prepError)
		}
	}
}

// TestPrepareTestTables checks that PrepareTestTables runs without an error.
func TestPrepareTestTables(t *testing.T) {
	fmt.Println("TestPrepareTestTables")
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}
	}
}

func TestGetRole(t *testing.T) {

	fmt.Println("TestGetRole")

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		roleNames := []string{RoleNameAdmin, RoleNameMember}
		for _, name := range roleNames {
			got, err := db.GetRole(name)
			if err != nil {
				t.Error(dbType + ": " + err.Error())
				return
			}

			if got.ID <= 0 {
				t.Errorf("want ID greater than 0, got %d", got.ID)
			}

			if got.Name != name {
				t.Errorf("want %s got %s", name, got.Name)
			}
		}
	}
}

// TestCreateAndDeleteUser checks that we can create and then delete a user.
func TestCreateAndDeleteUser(t *testing.T) {
	fmt.Println("TestCreateUser")
	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		const want = "foo"
		user := NewUser(want)

		createError := db.CreateUser(user)

		if createError != nil {
			t.Errorf("%s: %v", dbType, createError)
			return
		}

		got := user.ID

		if got == 0 {
			t.Error("expected ID to be non-zero")
		}

		if want != user.LoginName {
			t.Errorf("want %s got %s", want, user.LoginName)
		}

		// Fetch the user back and check it.

		u, fetchUserError := db.GetUser(got)
		if fetchUserError != nil {
			t.Error(fetchUserError)
		}

		if u.ID != user.ID {
			t.Errorf("want id %d got %d", user.ID, u.ID)
		}

		if u.LoginName != want {
			t.Errorf("want %s got %s", want, u.LoginName)
		}

		if u.Password != "*LK*" {
			t.Errorf("want *LK* got %s", u.Password)
		}

		if !u.Valid {
			t.Error("expected the user to be valid")
		}

		deleteError := db.DeleteUser(user)

		if deleteError != nil {
			t.Error(deleteError)
		}

		// Expect the ID of the user object to be 0 after the
		// database record has been deleted.
		if user.ID != 0 {
			t.Errorf("want ID of 0 got %d", user.ID)
		}
	}
}

// TestGetUsersByLoginName checks GetUsersByLoginName.  Note - there is a veey small chance
// that this test will fail due to a coincidence, but then work if run again.
func TestGetUsersByLoginName(t *testing.T) {

	containsAlpha, recError := regexp.Compile("[a-z]+")
	if recError != nil {
		t.Error(recError)
		return
	}

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		// We must use a UUID for the names because the tests may run in parallel and the name
		// in the database must be unique.  The UUID is something like
		// "7b0c7609-fb64-487e-b7fc-1ce9eaa5e6fa".  The search done by GetUsersByLoginName is
		// case-insensitive so we need to try more than one version of the name.

		// Attempt to create some unique names but give up eventually if we can't.
		for i := 0; i < 20; i++ {

			n, uuidError := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
			if uuidError != nil {
				t.Error(uuidError)
				return
			}

			nameLower := strings.ToLower(n)

			// We need the UUID to contain some alphabetic characters.
			if !containsAlpha.Match([]byte(nameLower)) {
				continue
			}

			nameUpper := strings.ToUpper(n)

			// Ensure that the transformed names have not been used.
			users1, fetchError1 := db.GetUsersByLoginName(nameLower)
			if fetchError1 != nil {
				t.Error(fetchError1)
			}

			if len(users1) > 0 {
				// Name already used.  try again.
				continue
			}

			users2, fetchError2 := db.GetUsersByLoginName(nameUpper)
			if fetchError2 != nil {
				t.Error(fetchError2)
			}

			if len(users2) > 0 {
				// Name already used.  try again.
				continue
			}

			// Create a user with a lower case name.

			user := NewUser(nameLower)

			createError := db.CreateUser(user)
			if createError != nil {
				t.Error(createError)
			}

			// Finally, the test.

			// Search for the user we just created by the name we used to create it.
			userL, fetchErrorL := db.GetUsersByLoginName(nameLower)
			if fetchErrorL != nil {
				t.Error(fetchErrorL)
			}

			if len(userL) == 0 {
				t.Error("cannot find user using its name in upper case")
			}

			if userL[0].ID != user.ID {
				t.Errorf("want userID %d got %d", user.ID, userL[0].ID)
			}

			// Search for the user we just created by the name we used to create it converted
			// to upper case.
			userU, fetchErrorU := db.GetUsersByLoginName(nameUpper)
			if fetchErrorU != nil {
				t.Error(fetchErrorU)
			}

			if len(userU) == 0 {
				t.Error("cannot find user using its name in upper case")
			}

			if userU[0].ID != user.ID {
				t.Errorf("want userID %d got %d", user.ID, userU[0].ID)
			}

			db.Rollback()

			return
		}

		// If we get to here, our attempts to create unique names have failed.
		t.Error("failed to create unique user names.  Giving up")
		return
	}
}

// TestCreateAndDeleteMember creates a member and the user which
// that requires, checks the result, deletes them and checks that
// too.
func TestCreateAndDeleteMember(t *testing.T) {
	fmt.Println("TestCreateAndDeleteMember")
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		role, roleError := db.GetRole("Member") // Lazy loading.
		if roleError != nil {
			t.Error(dbType + ": " + roleError.Error())
			return
		}

		const wantStartDate = "1970-01-01"
		const wantEndDate = "2025-12-31"

		loginName, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Error(e1)
			return
		}

		user := NewUser(loginName)
		userError := db.CreateUser(user)
		if userError != nil {
			t.Error(userError)
			continue
		}
		startDate := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)
		member := NewMember(user, role, startDate, endDate)
		err := db.CreateMember(member)
		if err != nil {
			t.Error(dbType + ": " + err.Error())
			continue
		}

		gotID := member.ID

		if gotID == 0 {
			t.Error("expected ID to be non-zero")
		}

		if len(member.UUID) == 0 {
			t.Error("expected a UUID")
		}

		if wantStartDate != member.StartDate {
			t.Errorf("want %s got %s", wantStartDate, member.StartDate)
		}

		if wantEndDate != member.EndDate {
			t.Errorf("want %s got %s", wantEndDate, member.EndDate)
		}

		// Remove and check the remains.
		deleteError := db.DeleteMember(member)
		if deleteError != nil {
			t.Error(deleteError)
		}
		deleteUserError := db.DeleteUser(user)
		if deleteUserError != nil {
			t.Error(deleteUserError)
		}

		// Expect the ID of the member object to be 0 after the
		// database record has been deleted.
		if member.ID != 0 {
			t.Errorf("want ID of 0 got %d", member.ID)
		}
	}
}

func TestGetLoginNames(t *testing.T) {
	fmt.Println("TestGetLoginNames")

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
				Email:          "foo@example.com",
				AssocUserID:    42,
				AssocFirstName: "Fred",
				AssocLastName:  "Smith",
			},
			2,
			"foo@example.com",
			"Fred.Smith",
		},
		{
			"ordinary member only",
			MembershipSale{
				Email: "foo@example.com",
			},
			1,
			"foo@example.com",
			"",
		},
		{
			"ordinary member and associate with email",
			MembershipSale{
				AssocUserID:    1,
				AssocFirstName: "Fred",
				Email:          "foo@example.com",
				AssocEmail:     "bar@example.com",
			},
			2,
			"foo@example.com",
			"bar@example.com",
		},
		{
			"ordinary member and associate without email",
			MembershipSale{
				AssocUserID:    42,
				Email:          "foo@example.com",
				AssocFirstName: "Fred",
				AssocLastName:  "Smith",
			},
			2,
			"foo@example.com",
			"Fred.Smith",
		},
	}

	for _, td := range testData {

		name, getNamesError := getLoginNames(&td.sale)
		if getNamesError != nil {
			t.Errorf("%s: %v",
				td.description, getNamesError)
		}

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

	fmt.Println("TestCreateAccounts")

	for _, dbType := range databaseList {
		const wantPassword = "*LK*"

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		var testData = []struct {
			description              string
			now                      time.Time // This controls the start date of the new member.
			end                      time.Time // The end date of the new member.
			sale                     MembershipSale
			wantOrdinaryAccountName  string
			wantStartDateSQLite      string
			wantStartDatePostgres    string
			wantEndDateSQLite        string
			wantEndDatePostgres      string
			wantAssociateAccountName string
		}{
			{
				"ordinary member only",
				time.Date(2024, time.October, 1, 12, 35, 15, 0, time.UTC),
				time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC),
				MembershipSale{
					TransactionType: TransactionTypeNewMember,
					MembershipYear:  2025,
					Email:           "foo1@example.com",
				},
				"foo1@example.com",
				"2024-10-01",
				"2024-10-01T00:00:00Z",
				"2025-12-31",
				"2025-12-31T00:00:00Z",
				"",
			},
			{
				"ordinary member and associate with email",
				time.Date(2024, time.July, 4, 8, 9, 10, 0, time.UTC),
				time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
				MembershipSale{
					TransactionType: TransactionTypeNewMember,
					MembershipYear:  2024,
					AssocUserID:     42,
					Email:           "foo2@example.com",
					AssocFirstName:  "John",
					AssocEmail:      "bar@example.com",
				},
				"foo2@example.com",
				"2024-07-04",
				"2024-07-04T00:00:00Z",
				"2024-12-31",
				"2024-12-31T00:00:00Z",
				"bar@example.com",
			},
			{
				"ordinary member and associate without email",
				time.Date(2025, time.February, 14, 12, 35, 15, 0, time.UTC),
				time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC),
				MembershipSale{
					TransactionType: TransactionTypeNewMember,
					MembershipYear:  2025,
					Email:           "foo3@example.com",
					AssocUserID:     4,
					AssocFirstName:  "Fred",
					AssocLastName:   "Smith",
				},
				"foo3@example.com",
				"2025-02-14",
				"2025-02-14T00:00:00Z",
				"2025-12-31",
				"2025-12-31T00:00:00Z",
				"Fred.Smith",
			},
		}

		for _, td := range testData {

			u1, u2, createError := db.CreateAccounts(&td.sale, td.now, td.end)
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

			member1, fetchMemberError := db.GetMemberOfUser(user1)
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

			if dbType == "sqlite" {
				if td.wantStartDateSQLite != member1.StartDate {
					t.Errorf("%s: want %s got %s",
						td.description, td.wantStartDateSQLite, member1.StartDate)
				}

				if td.wantEndDateSQLite != member1.EndDate {
					t.Errorf("%s: want %s got %s",
						td.description, td.wantEndDateSQLite, member1.EndDate)
				}
			} else {
				if td.wantStartDatePostgres != member1.StartDate {
					t.Errorf("%s: want %s got %s",
						td.description, td.wantStartDatePostgres, member1.StartDate)
				}

				if td.wantEndDatePostgres != member1.EndDate {
					t.Errorf("%s: want %s got %s",
						td.description, td.wantEndDatePostgres, member1.EndDate)
				}
			}

			if td.sale.AssocUserID != 0 {

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

			fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("DATE_LAST_PAID")
			if fieldError != nil {
				t.Error(fieldError)
			}

			sql := `
			SELECT usd_value
			FROM adm_user_data
			WHERE usd_usr_id = $1
			AND usd_usf_id = $2;
		`
			var dateLastPaid string

			fetchDateError := db.QueryRow(sql, user1.ID, fieldID).Scan(&dateLastPaid)
			if fetchDateError != nil {
				t.Error(fetchDateError)
			}

			//
			if td.wantStartDateSQLite != dateLastPaid {
				t.Errorf("%s: want %s got %s",
					td.description, td.wantStartDateSQLite, dateLastPaid)
			}

			// Done.
		}

		// The deferred rollback will run.
	}
}

// TestGetFieldID tests GetFieldID.
func TestGetFieldID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		fmt.Println("TestGetFieldID")
		for _, dbType := range databaseList {
			var testData = []struct {
				fieldName string
			}{
				{"FIRST_NAME"},
				{"LAST_NAME"},
				{"EMAIL"},
			}

			for _, td := range testData {

				gotID, gotErr := db.GetUserDataFieldIDByNameIntern(td.fieldName)

				if gotErr != nil {
					t.Error(gotErr)
				}

				if gotID <= 0 {
					t.Errorf("%s %s want i > 0 got %d", dbType, td.fieldName, gotID)
				}
			}
		}
	}
}

// TestGetGiftaidID checks getGiftaidID.
func TestGetGiftaidID(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		gotID, gotErr := db.GetUserDataFieldIDByNameIntern("GIFT_AID")

		if gotErr != nil {
			t.Error(db.Config.Type + ": " + gotErr.Error())
		}

		if gotID <= 0 {
			t.Errorf("%s want i > 0 got %d", db.Config.Type, gotID)
		}
	}
}

// TestMemberExists checks MemberExists.
func TestMemberExists(t *testing.T) {
	fmt.Println("TestMemberExists")

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
			continue
		}

		user, _, wantFirstName, wantLastName, umError := createTestUserEtc(db)
		if umError != nil {
			t.Errorf("%s: %v", dbType, umError)
			continue
		}

		emID, emError :=
			db.GetUserDataFieldIDByNameIntern("EMAIL")
		if emError != nil {
			t.Error(emError)
			continue
		}

		email, emailError := GetUserDataField[string](db, emID, user.ID)
		if emailError != nil {
			t.Error(emailError)
			continue
		}

		if email != user.LoginName {
			t.Errorf("want %s got %s", user.LoginName, email)
			continue
		}

		ifn, fne := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
		if fne != nil {
			t.Error(fne)
			continue
		}

		fn, fnError := GetUserDataField[string](db, ifn, user.ID)
		if fnError != nil {
			t.Error(fnError)
			continue
		}

		if fn != wantFirstName {
			t.Errorf("want %s got %s", wantFirstName, fn)
			continue
		}

		iln, lne := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
		if lne != nil {
			t.Error(lne)
			continue
		}

		ln, lnError := GetUserDataField[string](db, iln, user.ID)
		if lnError != nil {
			t.Error(lnError)
			continue
		}

		if ln != wantLastName {
			t.Errorf("want %s got %s", wantLastName, ln)
			continue
		}

		id, searchErr := db.GetUserIDofMember(fn, ln, user.LoginName)
		if searchErr != nil {
			t.Error(db.Config.Type + " - expected member to exist")
			continue
		}

		if id <= 0 {
			t.Errorf("want id > 0 got %d", id)
			continue
		}

		if id != user.ID {
			t.Errorf("want %d got %d", user.ID, id)
			continue
		}
	}
}

// TestMemberExistsDetailed tests MemberExists thoroughly.
func TestMemberExistsDetailed(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, fn, ln, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
		}

		var testData = []struct {
			description string
			firstName   string
			lastName    string
			email       string
			shouldWork  bool
		}{
			{"no match", "junk", "junk", "junk", false},
			{"all match", fn, ln, user.LoginName, true},
			{"email matches", "junk", "more junk", user.LoginName, true},
			{"names match", fn, ln, "junk", true},
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
				// if db.Config.Type == "postgres" {
				// 	// we expect the call to fail with an error.
				// 	if searchError == nil {
				// 		t.Errorf("%s: want an error", td.description)
				// 	}
				// } else {
				// Using sqlite we expect the query to return no rows and
				// the call to return an ID of zero.
				if id != 0 {
					t.Errorf("%s: want ID 0 (no match), got %d", td.description, id)
				}
				//}
			}
		}
	}
}

// TestSetLastPayment checks SetLastPayment.
func TestSetLastPayment(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			em := fmt.Sprintf("%s: %v", db.Config.Type, umError)
			t.Error(em)
			return
		}

		// if user == nil {
		// 	em := fmt.Sprintf("%s: expected a user", db.Config.Type)
		// 	t.Error(em)
		// 	return
		// }

		setError := db.SetLastPayment(user.ID, 2.5)
		if setError != nil {
			t.Error(setError)
			continue
		}

		fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_LAST_PAYMENT")
		if fieldError != nil {
			t.Error(fieldError)
			continue
		}

		p, getError := GetUserDataField[float64](db, fieldID, user.ID)
		if getError != nil {
			t.Error(getError)
			continue
		}

		if p != 2.5 {
			t.Errorf("%s: want 2.5 got %f", db.Config.Type, p)
			continue
		}
	}
}

// TestSetLastPayment checks SetLastPayment.
func TestSetDonationToSociety(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			em := fmt.Sprintf("%s: %v", db.Config.Type, umError)
			t.Error(em)
			return
		}

		err := db.SetDonationToSociety(user.ID, 2.5)
		if err != nil {
			t.Error(err)
		}

		fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_LDLHS")
		if fieldError != nil {
			t.Error(fieldError)
			continue
		}

		p, getError := GetUserDataField[float64](db, fieldID, user.ID)
		if getError != nil {
			t.Error(getError)
			continue
		}

		if p != 2.5 {
			t.Errorf("%s: want 2.5 got %f", db.Config.Type, p)
			continue
		}
	}
}

// TestSetLastPayment checks SetLastPayment.
func TestSetDonationToMuseum(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			em := fmt.Sprintf("%s: %v", db.Config.Type, umError)
			t.Error(em)
			return
		}

		err := db.SetDonationToMuseum(user.ID, 2.5)
		if err != nil {
			t.Error(err)
		}

		fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("VALUE_OF_DONATION_TO_THE_MUSEUM")
		if fieldError != nil {
			t.Error(fieldError)
			continue
		}

		p, getError := GetUserDataField[float64](db, fieldID, user.ID)
		if getError != nil {
			t.Error(getError)
			continue
		}

		if p != 2.5 {
			t.Errorf("%s: want 2.5 got %f", db.Config.Type, p)
			continue
		}
	}
}

func TestCreateUuid(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		uid, uidError := CreateUuid(db.Transaction, "rol_uuid", "adm_roles")
		if uidError != nil {
			t.Error(uidError)
			continue
		}

		if len(uid) <= 0 {
			t.Errorf("want a uuid got empty string")
			continue
		}
	}
}

// TestSetUserDataField checks the SetUserDataField function.
func TestSetUserDataField(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		txError := db.BeginTx()

		if txError != nil {
			t.Error(txError)
			continue
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
			continue
		}

		// Use different names each time to avoid clashes in the postgres
		// database.

		loginName, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Error(e1)
		}

		// Create a user
		user := NewUser(loginName)
		cuError := db.CreateUser(user)
		if cuError != nil {
			t.Error(cuError)
			continue
		}

		var testData = []struct {
			Want      string
			FieldName string
		}{
			{loginName, "EMAIL"},
			{"Dr", "SALUTATION"},
			{"Peter", "FIRST_NAME"},
			{"Smith", "LAST_NAME"},
			{"1 The High Street", "STREET"},
			{"Little Hampton on the Marsh", "ADDRESS_LINE_2"},
			{"Some Bigger Place", "ADDRESS_LINE_3"},
			{"Some County", "COUNTY"},
			{"AA1 1AA", "POSTCODE"},
			{"Fairyland", "COUNTRY"},
			{"01234 567890", "PHONE"},
			{"07123 456789", "MOBILE"},
		}

		// We check SetUserDataField using GetUserDataField.  To detect a problem where one call
		// overwrites the results of another, we set all the fields in one go and then check
		// them all later.
		for _, td := range testData {

			fieldID, fetchError := db.GetUserDataFieldIDByNameIntern(td.FieldName)
			if fetchError != nil {
				t.Errorf("cannot find adm_user_field %s", td.FieldName)
				continue
			}

			setError := SetUserDataField(db, fieldID, user.ID, td.Want)
			if setError != nil {
				t.Error(setError)
				continue
			}

		}

		for _, td := range testData {

			f := db.UserField[td.FieldName]
			if f == nil {
				t.Errorf("%s: cannot find adm_user_field %s", dbType, td.FieldName)
				continue
			}

			fieldID := f.ID

			if fieldID == 0 {
				t.Error("want non-zero adm_user_field ID")
				continue
			}

			got, getError := GetUserDataField[string](db, fieldID, user.ID)
			if getError != nil {
				t.Error(getError)
				continue
			}

			if got != td.Want {
				t.Errorf("%s: want %s got %s", dbType, td.Want, got)
				continue
			}
		}
	}
}

// TestSetMembersAtAddress checks SetMembersAtAddress.
func TestSetMembersAtAddress(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			t.Error(umError)
			return
		}

		const want = 5
		setError := db.SetMembersAtAddress(user.ID, want)
		if setError != nil {
			t.Errorf("%s: %v", dbType, setError)
			break
		}

		const sql = `
			select usd_value from adm_user_data
			where usd_usr_id = $1
			AND usd_usf_id = $2
		`

		id, fetchIDError := db.GetUserDataFieldIDByNameIntern("MEMBERS_AT_ADDRESS")
		if fetchIDError != nil {
			t.Errorf("%s: %v", dbType, fetchIDError)
			break
		}

		var got int
		err := db.QueryRow(sql, user.ID, id).Scan(&got)
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

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			t.Error(umError)
			return
		}

		london, _ := time.LoadLocation("Europe/London")
		tm := time.Date(2024, time.February, 14, 1, 2, 3, 4, london)

		setError := db.SetDateLastPaid(user.ID, tm)
		if setError != nil {
			t.Errorf("%s: %v", dbType, setError)
			break
		}

		const sqlCommand = `
				select usd_value  
				from adm_user_data
				where usd_usr_id = $1
				AND usd_usf_id = $2`

		id, fieldError := db.GetUserDataFieldIDByNameIntern("DATE_LAST_PAID")
		if fieldError != nil {
			t.Errorf("%s: %v", dbType, fieldError)
			break
		}

		var got string
		queryAndScanError := db.QueryRow(sqlCommand, user.ID, id).Scan(&got)
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

		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, umError := createTestUserEtc(db)
		if umError != nil {
			t.Error(umError)
			return
		}

		err := db.SetFriendField(user.ID, true)
		if err != nil {
			t.Error(err)
		}
	}
}

// TestSetMemberEndDate checks SetMemberEnddate
func TestSetMemberEndDate(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			continue
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
			continue
		}

		user, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
		}

		// Set the member's end date to 2025 and check it.
		setError := db.SetMemberEndDate(user.ID, 2025)
		if setError != nil {
			t.Error(dbType + ": " + setError.Error())
			return
		}
		checkError2 := checkMemberEndYear(db, user.ID, 2025)
		if checkError2 != nil {
			t.Error(dbType + ": " + checkError2.Error())
			return
		}
	}
}

func TestSetGiftaidField(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		giftaidID, fetchIDErr := db.GetUserDataFieldIDByNameIntern("GIFT_AID")
		if fetchIDErr != nil {
			t.Errorf("%s: %v", dbType, fetchIDErr)
		}

		user, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
		}

		// Delete any existing giftaid field for the test user.
		const deleteGiftaidCMD = `
			DELETE FROM adm_user_data
			WHERE usd_usr_id = $1
			AND usd_usf_id = $2;
		`
		_, execError := db.Exec(deleteGiftaidCMD, user.ID, giftaidID)
		if execError != nil {
			t.Errorf("%s: %v", db.Config.Type, execError)
			break
		}

		// Create a giftaid field set to true.
		createErr1 := db.SetGiftaidField(user.ID, true)
		if createErr1 != nil {
			t.Errorf("%s: %v", db.Config.Type, createErr1)
			break
		}

		// Check the field - should be true.
		got1, err1 := db.GetGiftaidField(user.ID)

		if err1 != nil {
			t.Error(err1)
			return
		}

		if !got1 {
			t.Errorf("%s: expected giftaid to be set true for user %d", dbType, user.ID)
		}

		// Update the giftaid field to false.
		createErr := db.SetGiftaidField(user.ID, false)
		if createErr != nil {
			t.Errorf("%s: %v", db.Config.Type, createErr)
		}

		got2, err2 := db.GetGiftaidField(user.ID)

		if err2 != nil {
			t.Error(err2)
			return
		}
		// Check the field - should be false.
		if got2 {
			t.Errorf("%s: expected giftaid to be set false for user %d", db.Config.Type, user.ID)
			return
		}
	}
}

func TestMembershipSale(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

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
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		user, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
			continue
		}

		assoc, _, _, _, ae := createTestUserEtc(db)
		if ae != nil {
			t.Errorf("%s: %v", dbType, ue)
			continue
		}

		var testData = []struct {
			description string
			input       MembershipSale
			want        MembershipSale
		}{
			{
				"no associate, no donations",
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: true, FriendFeePaid: 5, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 0,
					DonationToMuseum: 0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 0,
					AssocFriend: false, AssocFriendFeePaid: 0,
				},
			},
			{
				"no associate",
				MembershipSale{
					ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5,
					DonationToSociety: 2,
					DonationToMuseum:  6.0, Giftaid: true,
					AssocUserID: 0,
					// These values should be ignored.
					AssocFeePaid: 42.0,
					AssocFriend:  true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 0.0,
					AssocFriend: false, AssocFriendFeePaid: 0.0,
				},
			},
			{
				"all",
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5,
					DonationToSociety: 2,
					DonationToMuseum:  6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},

			{
				"associate, no donations",
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 0,
					DonationToMuseum: 0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},
			{
				"no associate, no donations",
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 0,
					DonationToMuseum: 0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 0,
					AssocFriend: false, AssocFriendFeePaid: 0,
				},
			},
			{
				"ordinary member is friend", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: false, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: false, AssocFriendFeePaid: 43.0,
				},
			},
			{
				"Gifaid", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: false, AssocFriendFeePaid: 0.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: false, AssocFriendFeePaid: 0.0,
				},
			},
			{
				"associate member is friend", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},
		}

		for _, td := range testData {
			id, createError := td.input.Create(db)
			if createError != nil {
				t.Errorf("%s: %s %v", dbType, td.description, createError)
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
				t.Errorf("%s %s - %v", dbType, td.description, fetchError)
				break
			}

			// The id has been set in the stored record.  Set the ID in
			// the want to match.

			td.want.ID = got.ID

			if td.want != *got {
				t.Errorf("%s %s\nwant %v\ngot  %v", dbType, td.description, td.want, *got)
				break
			}

			const wantPaymentID = "some very long text"
			const wantPaymentStatus = "complete"

			td.want.PaymentID = wantPaymentID
			td.want.PaymentStatus = wantPaymentStatus
			got.PaymentID = wantPaymentID
			got.PaymentStatus = wantPaymentStatus

			updateError := got.Update(db)
			if updateError != nil {
				t.Errorf("%s %s: %v", dbType, td.description, updateError)
				break
			}

			updatedMS, fetchError := db.GetMembershipSale(got.ID)
			if fetchError != nil {
				t.Errorf("%s %s: %v", dbType, td.description, fetchError)
				break
			}

			if td.want != *updatedMS {
				t.Errorf("%s %s\nwant %v\ngot  %v", dbType, td.description, td.want, *updatedMS)
				break
			}

			// Tidy up - delete the membershipsales record and check that it's deleted.

			savedID := got.ID

			deleteError := got.Delete(db)
			if deleteError != nil {
				t.Errorf("%s %s - %v", dbType, td.description, deleteError)
				break
			}

			if got.ID != 0 {
				t.Errorf("%s %s: want id of deleted record to be 0, got %d", dbType, td.description, got.ID)
				break
			}

			// This should fail and ms should be nil.
			ms, expectedError := db.GetMembershipSale(savedID)

			if expectedError == nil {
				t.Errorf("%s %s: expected an error", dbType, td.description)
				break
			}

			if ms != nil {
				t.Errorf("%s %s: expected nil", dbType, td.description)
				break
			}
		}
	}
}

func F(t *testing.T) {

	for _, dbType := range databaseList {
		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			continue
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
			continue
		}

		user, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
			continue
		}

		sale := MembershipSale{
			ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
			MembershipYear: 2025, UserID: user.ID,
			OrdinaryMemberFeePaid: 24.0,
			Friend:                true, FriendFeePaid: 5,
			DonationToSociety: 2, DonationToMuseum: 6.0, Giftaid: true,
			AssocUserID: 0, AssocFeePaid: 0.0,
			AssocFriend: false, AssocFriendFeePaid: 0.0,
		}

		id, createError := sale.Create(db)
		if createError != nil {
			t.Errorf("%s - %v", dbType, createError)
			return
		}

		if id == 0 {
			t.Error("expected the returned ID to be non-zero")
		}

		// Set the Id to a non-existent record.
		sale.ID++

		// Expect the update to fail.
		err := sale.Update(db)

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
}

// checkMemberEndYear is a helper function that checks that the year in
// the given user's member end date in the given database matches the
// given year.  It works for sqlite and postgres.
func checkMemberEndYear(db *Database, userID int64, targetYear int) error {

	// To prepare, set the user's end date to a year other than the
	// target, fetch it back and check it.

	startingYear := targetYear - 1

	db.SetMemberEndDate(userID, startingYear)

	gotYear1, err1 := db.GetMembershipYearOfUser(userID)

	if err1 != nil {
		return err1
	}

	if gotYear1 != startingYear {
		em := fmt.Sprintf("setup - want starting year %d got %d", startingYear, gotYear1)
		return errors.New(em)
	}

	// To test, set the year to the given year and check.
	db.SetMemberEndDate(userID, targetYear)
	gotYear2, err2 := db.GetMembershipYearOfUser(userID)

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

// TestSetTimeFieldInUserData check SetTimeFieldInUserData.
func TestSetTimeFieldInUserData(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			continue
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
			continue
		}

		fieldID, fieldError := db.GetUserDataFieldIDByNameIntern("DATE_LAST_PAID")
		if fieldError != nil {
			t.Error(fieldError)
			continue
		}

		user, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
		}

		utc, _ := time.LoadLocation("UTC")
		d := time.Date(2025, time.February, 14, 1, 2, 3, 4, utc)

		const wantPostgres = "2025-02-14 01:02:03+00"
		const wantSQLite = "2025-02-14 01:02:03"

		var want string
		if dbType == "sqlite" {
			want = wantSQLite
		} else {
			want = wantPostgres
		}

		setError := db.SetTimeFieldInUserData(fieldID, user.ID, d)

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
		fetchError := db.QueryRow(sql, user.ID, fieldID).Scan(&got)
		if fetchError != nil {
			t.Error(fetchError)
		}

		if want != got {
			t.Errorf("want %s got %s", want, got)
			return
		}
	}
}

func TestCreateCategoryAndGetCategory(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		txError := db.BeginTx()
		if txError != nil {
			t.Error(txError)
			continue
		}
		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Errorf("%s - %v", dbType, prepError)
			continue
		}

		// A category with all fields filled in.
		complete := NewCategory(ourOrganisation, "a", "ni1", "n1", false, true, 1, systemUser)
		// A ctaegory with no organisation.
		noOrg := NewCategory(nil, "b", "ni2", "n2", true, false, 2, systemUser)
		// A category with no user.
		noUser := NewCategory(ourOrganisation, "c", "ni3", "n3", true, true, 3, nil)
		// A category with no organisation and no user.
		neither := NewCategory(nil, "d", "ni4", "n4", true, true, 4, nil)

		var testData = []struct {
			Description    string
			Cat            *Category
			WantOrg        *Organisation
			WantType       string
			WantNameIntern string
			WantName       string
			WantSystem     bool
			WantDefault    bool
			WantSequence   int
			WantCreateUser *User
		}{
			{"complete", complete, ourOrganisation, "a", "ni1", "n1", false, true, 1, systemUser},
			{"No org", noOrg, nil, "b", "ni2", "n2", true, false, 2, systemUser},
			{"No user", noUser, ourOrganisation, "c", "ni3", "n3", true, true, 3, nil},
			{"Neither", neither, nil, "d", "ni4", "n4", true, true, 4, nil},
		}

		for _, td := range testData {
			createCatError := db.CreateCategory(td.Cat)
			if createCatError != nil {
				t.Errorf("%s: %s %v", td.Description, dbType, createCatError)
				continue
			}

			got, fetchCatError := db.GetCategory(td.Cat.ID)
			if fetchCatError != nil {
				t.Errorf("%s: %s %v", td.Description, dbType, fetchCatError)
				continue
			}

			if got.ID != td.Cat.ID {
				t.Errorf("%s: %s want %d got %d", td.Description, dbType, td.Cat.ID, got.ID)
				continue
			}

			switch {
			case got.Org == nil && td.WantOrg == nil:

			case got.Org != nil && td.WantOrg == nil:
				t.Errorf("%s: %s want nil org", td.Description, dbType)
				continue
			case got.Org.ID != td.WantOrg.ID:
				t.Errorf("%s: %s want %v got %v", td.Description, dbType, td.WantOrg, got.Org)
				continue
			default:
			}

			if got.Type != td.Cat.Type {
				t.Errorf("%s: %s want %s got %s", td.Description, dbType, td.WantType, got.Type)
				continue
			}

			if got.NameIntern != td.WantNameIntern {
				t.Errorf("%s: %s want %s got %s", td.Description, dbType, td.WantNameIntern, got.NameIntern)
				continue
			}

			if got.Name != td.WantName {
				t.Errorf("%s: %s want %s got %s", td.Description, dbType, td.WantName, got.Name)
				continue
			}

			if got.System != td.WantSystem {
				t.Errorf("%s: %s want %v got %v", td.Description, dbType, td.WantSystem, got.System)
				continue
			}

			if got.Default != td.WantDefault {
				t.Errorf("%s: %s want %v got %v", td.Description, dbType, td.WantDefault, got.Default)
				continue
			}

			if got.Sequence != td.WantSequence {
				t.Errorf("%s: %s want %v got %v", td.Description, dbType, td.WantSequence, got.Sequence)
				continue
			}

			switch {
			case got.CreateUser == nil && td.WantCreateUser == nil:

			case got.CreateUser != nil && td.WantCreateUser == nil:
				t.Errorf("%s: %s want nil org", td.Description, dbType)
				continue
			case got.CreateUser.ID != td.WantCreateUser.ID:
				t.Errorf("%s: %s want %v got %v", td.Description, dbType, td.WantCreateUser, got.CreateUser)
				continue
			default:
			}
		}
	}
}

// TestGetUserName checks the GetUserName function.
func TestGetUserName(t *testing.T) {

	var testData = []struct {
		Description      string
		Email            string
		FirstName        string
		LastName         string
		WantUserName     string
		WantErrorMessage string
	}{

		{"email", "\tJohn@eXample.com ", "John", "Smith", "john@example.com", ""},
		{"no email", "", "John", "Smith", "john.smith", ""},
		{"spaces in first name", "", "J A", "Smith Brown", "j.a.smith.brown", ""},
		{"spaces in second name", "", "Herbert", "George Wells", "herbert.george.wells", ""},
		{"white space", "", "\tH\t  G\t", "\tWells ", "h.g.wells", ""},
		{"nothing", "", "", "", "", NoUserNameError},
	}

	for _, td := range testData {

		u, err := GetUserName(td.Email, td.FirstName, td.LastName)
		if len(td.WantErrorMessage) > 0 {
			// We want an error message.
			if err.Error() != td.WantErrorMessage {
				t.Errorf("%s: want %s got %s", td.Description, td.WantErrorMessage, err.Error())
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: %v", td.Description, err)
			continue
		}

		if u != td.WantUserName {
			t.Errorf("%s: want %s got %s", td.Description, td.WantUserName, u)
			continue
		}
	}
}

// CreateTestUserEtc is a helper that creates a user with member
// and adm_user_data records.
func createTestUserEtc(db *Database) (*User, *Member, string, string, error) {
	startTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)

	roleMember, re := db.GetRole(RoleNameMember)
	if re != nil {
		return nil, nil, "", "", re
	}
	// Use different names each time to avoid clashes in the postgres
	// database.

	email, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
	if e1 != nil {
		return nil, nil, "", "", e1
	}

	firstName, e2 := CreateUuid(db.Transaction, "usd_value", "adm_user_data")
	if e2 != nil {
		return nil, nil, "", "", e2
	}

	lastName, e3 := CreateUuid(db.Transaction, "usd_value", "adm_user_data")
	if e3 != nil {
		return nil, nil, "", "", e3
	}
	user, member, umError := db.CreateUserAndMember(email, firstName, lastName, roleMember, startTime, endTime)

	return user, member, firstName, lastName, umError
}
