package database

import (
	"errors"
	"fmt"
	"math"
	"reflect"
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

// TestCreateMembersInterest checks CreateMembersInterest.
func TestCreateMembersInterest(t *testing.T) {
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
			t.Errorf("%s: %v", dbType, prepError)
			continue
		}

		user, cue := CreateUser(db)
		if cue != nil {
			t.Errorf("%s: %v", dbType, cue)
			continue
		}

		mis, me := db.GetInterests()
		if me != nil {
			t.Errorf("%s: %v", dbType, me)
		}

		if len(mis) == 0 {
			t.Errorf("%s: %s", dbType, "no interests")
			continue
		}

		// we have at least one interest, which is enough for this test.

		mi := NewMembersInterest(user.ID, mis[0].ID)
		ce := db.CreateMembersInterest(mi)
		if ce != nil {
			t.Errorf("%s: %v", dbType, ce)
			continue
		}

		fetchedInterests, fetchError := db.GetMembersInterests(user.ID)
		if fetchError != nil {
			t.Errorf("%s: %v", dbType, fetchError)
			continue
		}

		// Expect just one interest, same as mi.
		if len(fetchedInterests) != 1 {
			t.Errorf("%s: want 1 interest got %d", dbType, len(fetchedInterests))
		}

		if fetchedInterests[0] != *mi {
			t.Errorf("%s: want %v got %v", dbType, mi, fetchedInterests[0])
		}
	}

}

// TestGetMembersInterests checks GetMembersInterests.
func TestGetMembersInterests(t *testing.T) {

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

		// The setup is long-winded.

		//  Create a user.
		loginName, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Error(e1)
			return
		}
		user := NewUser(loginName)
		createUserError := db.CreateUser(user)
		if createUserError != nil {
			t.Errorf("%s: %v", dbType, createUserError)
			return
		}

		// Get the ID of two interests - any will do.  (If there are not at
		// least two rows, the test will fail.)

		interest, ie := db.GetInterests()
		if ie != nil {
			t.Errorf("%s: %v", dbType, ie)
			return
		}

		if len(interest) < 2 {
			t.Errorf("%s: need at least two interest, got %d", dbType, len(interest))
			return
		}

		// Create two MembersInterest rows in the database
		mi1 := NewMembersInterest(user.ID, interest[0].ID)
		createError1 := db.CreateMembersInterest(mi1)
		if createError1 != nil {
			t.Errorf("%s: %v", db.Config.Type, createError1)
			continue
		}
		mi2 := NewMembersInterest(user.ID, interest[1].ID)
		createError2 := db.CreateMembersInterest(mi2)
		if createError2 != nil {
			t.Errorf("%s: %v", db.Config.Type, createError2)
			continue
		}

		// Test GetMembersInterests.  The interest objects were fetched in order
		// and the results are ordered.
		got, gotError := db.GetMembersInterests(user.ID)
		if gotError != nil {
			t.Errorf("%s: %v", db.Config.Type, gotError)
			continue
		}
		if len(got) != 2 {
			t.Errorf("%s: want 2 rows got %d", db.Config.Type, len(got))
			continue
		}
		if got[0].UserID != user.ID {
			t.Errorf("%s: want id %d got %d", db.Config.Type, user.ID, got[0].UserID)
			continue
		}
		if got[0].InterestID != interest[0].ID {
			t.Errorf("%s: want id %d got %d", db.Config.Type, interest[0].ID, got[0].InterestID)
			continue
		}
		if got[1].UserID != user.ID {
			t.Errorf("%s: want id %d got %d", db.Config.Type, user.ID, got[1].UserID)
			continue
		}
		if got[1].InterestID != interest[1].ID {
			t.Errorf("%s: want id %d got %d", db.Config.Type, interest[1].ID, got[1].InterestID)
			continue
		}

		// Check the unique constraint handling - attempt to create a copy
		// of mi1.
		mi3 := NewMembersInterest(user.ID, interest[0].ID)
		createError3 := db.CreateMembersInterest(mi3)
		if createError3 != nil {
			t.Errorf("%s: %v", db.Config.Type, createError3)
			continue
		}

		// There should still only be two interests.
		got2, got2Error := db.GetMembersInterests(user.ID)
		if got2Error != nil {
			t.Errorf("%s: %v", db.Config.Type, got2Error)
			continue
		}
		if len(got2) != 2 {
			t.Errorf("%s: want 2 rows got %d", db.Config.Type, len(got2))
			continue
		}
	}
}

// TestGetMembersInterestForUserWithJunkUser checks that GetMembersInterestForUserWithJunkUser
// returns an empty list if the user doesn't exist.
func TestGetMembersInterestsWithJunkUser(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := ConnectForTesting(dbType)
		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		// In theory the maximum user ID is the maximum 64-bit number, but on my
		// system they are limited to 32 bits.  No matter, that should be big
		// enough to ensure that there is no user with an ID this big.
		const junkUserID = math.MaxInt32
		interests, err := db.GetMembersInterests(junkUserID)
		if err != nil {
			t.Errorf("%s: %v", dbType, err)
			continue
		}

		if len(interests) != 0 {
			t.Errorf("%s: expected an empty list, got %d items", dbType, len(interests))
			continue
		}
	}
}

// TestUpdateMembersOtherInterests checkss UpdateMembersOtherInterests.
func TestUpdateMembersOtherInterests(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Errorf("%s: %v", dbType, connError)
			continue
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Errorf("%s: %v", dbType, prepError)
			continue
		}

		// Set up.
		user, cue := CreateUser(db)
		if cue != nil {
			t.Errorf("%s: %v", dbType, cue)
			continue
		}

		const interests = "stuff"
		moi := NewMembersOtherInterests(user.ID, interests)
		coi := db.CreateMembersOtherInterests(moi)
		if coi != nil {
			t.Errorf("%s: %v", dbType, coi)
			continue
		}

		// Test.
		const newInterest = "other stuff"
		moi.Interests = newInterest
		db.UpdateMembersOtherInterests(moi)

		got, fetchError := db.GetMembersOtherInterests(user.ID)
		if fetchError != nil {
			t.Errorf("%s: %v", dbType, fetchError)
			continue
		}

		if *got != *moi {
			t.Errorf("%s: want %v, got %v", dbType, *moi, *got)
			continue
		}
	}
}

// TestUpsertMembersOtherInterests checks UpsertMembersOtherInterests.
func TestUpsertMembersOtherInterests(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := OpenDBForTesting(dbType)

		if connError != nil {
			t.Errorf("%s: %v", dbType, connError)
			continue
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := PrepareTestTables(db)
		if prepError != nil {
			t.Errorf("%s: %v", dbType, prepError)
			continue
		}

		// Set up.
		user1, cue := CreateUser(db)
		if cue != nil {
			t.Errorf("%s: %v", dbType, cue)
			continue
		}

		user2, cue2 := CreateUser(db)
		if cue2 != nil {
			t.Errorf("%s: %v", dbType, cue2)
			continue
		}

		interests1, e1 := CreateUuid(db.Transaction, "moi_interests", "adm_members_other_interests")
		if e1 != nil {
			t.Error(e1)
			return
		}
		want1 := NewMembersOtherInterests(user1.ID, interests1)
		coi := db.CreateMembersOtherInterests(want1)
		if coi != nil {
			t.Errorf("%s: %v", dbType, coi)
			continue
		}

		// We have two users.  User1 has expressed some interests.  User2
		// has not.

		// Test.

		// Upsert user 1, relacing original interests.
		interests2, e2 := CreateUuid(db.Transaction, "moi_interests", "adm_members_other_interests")
		if e2 != nil {
			t.Error(e2)
			return
		}
		want1.Interests = interests2
		ue1 := db.UpsertMembersOtherInterests(want1)
		if ue1 != nil {
			t.Errorf("%s: %v", dbType, ue1)
			continue
		}

		got1, fetchError1 := db.GetMembersOtherInterests(user1.ID)
		if fetchError1 != nil {
			t.Errorf("%s: %v", dbType, fetchError1)
			continue
		}

		if *got1 != *want1 {
			t.Errorf("%s: want %v, got %v", dbType, *want1, *got1)
			continue
		}

		// Upsert user 2 creating a new record.
		interests3, e3 := CreateUuid(db.Transaction, "moi_interests", "adm_members_other_interests")
		if e3 != nil {
			t.Error(e3)
			return
		}
		want2 := NewMembersOtherInterests(user2.ID, interests3)
		ue2 := db.UpsertMembersOtherInterests(want2)
		if ue2 != nil {
			t.Errorf("%s: %v", dbType, ue2)
			continue
		}
		got2, fetchError2 := db.GetMembersOtherInterests(user2.ID)
		if fetchError2 != nil {
			t.Errorf("%s: %v", dbType, fetchError2)
			continue
		}

		if *got2 != *want2 {
			t.Errorf("%s: want %v, got %v", dbType, *want2, *got2)
			continue
		}
	}
}

// TestGetMembersOtherInterestsForUser checks GetMembersOtherInterestsForUser.
func TestGetMembersOtherInterestsForUser(t *testing.T) {

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

		//  Create a user.
		loginName, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Error(e1)
			return
		}
		user := NewUser(loginName)
		createUserError := db.CreateUser(user)
		if createUserError != nil {
			t.Error(createUserError)
			return
		}

		// Create an other interest.
		interests, e2 := CreateUuid(db.Transaction, "moi_interests", "adm_members_other_interests")
		if e2 != nil {
			t.Error(e2)
			return
		}

		moi := NewMembersOtherInterests(user.ID, interests)

		createInterestError := db.CreateMembersOtherInterests(moi)
		if createInterestError != nil {
			t.Error(createInterestError)
			return
		}

		// Test GetOMembersOtherInterestsForUser.
		got, fetchError := db.GetMembersOtherInterests(user.ID)
		if fetchError != nil {
			t.Error(fetchError)
			return
		}

		if got.ID <= 0 {
			t.Errorf("%s: want id greater than zero got %d", db.Config.Type, got.ID)
			continue
		}

		if got.UserID != user.ID {
			t.Errorf("%s: want id %d got %d", db.Config.Type, user.ID, got.UserID)
			continue
		}

		if got.Interests != interests {
			t.Errorf("%s: want interests %s got %s", db.Config.Type, interests, got.Interests)
			continue
		}
	}
}

// TestGetAndSetExtraDetails checkes SetExtraDetails and GetExtraDetails.
func TestSetAndGetExtraDetails(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		u, ue := CreateUser(db)
		if ue != nil {
			// This ahould never fail.
			t.Fatal(ue)
		}

		toi, toie := db.GetInterests()
		if toie != nil {
			t.Fatal(toie)
		}

		var topics = make(map[int64]interface{})
		for _, interest := range toi {
			topics[interest.ID] = nil
		}

		ms := MembershipSale{
			UserID:                u.ID,
			AddressLine1:          "a",
			AddressLine2:          "b",
			AddressLine3:          "c",
			Town:                  "d",
			County:                "e",
			Postcode:              "f",
			Country:               "GBR",
			Phone:                 "g",
			Mobile:                "h",
			TopicsOfInterest:      topics,
			OtherTopicsOfInterest: "i",
		}

		saveError := db.SaveExtraDetails(&ms)
		if saveError != nil {
			t.Error(saveError)
			break
		}

		got := MembershipSale{UserID: u.ID}
		fetchError := db.GetExtraDetails(&got)
		if fetchError != nil {
			t.Error(fetchError)
			break
		}

		if reflect.DeepEqual(got, ms) {
			t.Errorf("want %v got %v", ms, got)
			break
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
				AssocTitle:     "Prof",
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
				AssocTitle:     "Dr",
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
				AssocTitle:     "Mr",
				AssocFirstName: "Fred",
				AssocLastName:  "Smith",
			},
			2,
			"foo@example.com",
			"Fred.Smith",
		},
		{
			"ordinary member and associate have the same email address",
			MembershipSale{
				AssocUserID:    42,
				Email:          "foo@example.com",
				AssocEmail:     "foo@example.com",
				AssocTitle:     "Mr",
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

		loginName1, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Error(e1)
			return
		}

		loginName2, e2 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e2 != nil {
			t.Error(e2)
			return
		}

		loginName3, e3 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e3 != nil {
			t.Error(e3)
			return
		}

		loginName4, e4 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e4 != nil {
			t.Error(e4)
			return
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
					Email:           loginName1,
					Title:           "Mr",
					FirstName:       "c",
					LastName:        "d",
				},
				loginName1,
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
					Email:           loginName2,
					Title:           "Prof",
					FirstName:       "a",
					LastName:        "b",
					AssocTitle:      "Dr",
					AssocFirstName:  "John",
					AssocLastName:   "Smith",
					AssocEmail:      loginName3,
				},
				loginName2,
				"2024-07-04",
				"2024-07-04T00:00:00Z",
				"2024-12-31",
				"2024-12-31T00:00:00Z",
				loginName3,
			},
			{
				"ordinary member and associate without email",
				time.Date(2025, time.February, 14, 12, 35, 15, 0, time.UTC),
				time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC),
				MembershipSale{
					TransactionType: TransactionTypeNewMember,
					MembershipYear:  2025,
					Email:           loginName4,
					Title:           "Mr",
					AssocUserID:     4,
					AssocTitle:      "Professor",
					AssocFirstName:  "Fred",
					AssocLastName:   "Smith",
				},
				loginName4,
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
				t.Errorf("%s: %s - %v", dbType, td.description, createError)
			}

			// Check that the database records have been created, for each
			// member: an adm_user record, an adm_member record and an
			// adm_user_data record with field name 'DATE_LAST_PAID' and
			// containing the given time in the format 'YYYY-MM-DD'

			user1, fetchUser1Error := db.GetUser(u1)
			if fetchUser1Error != nil {
				t.Errorf("%s: %s - %v", dbType, td.description, fetchUser1Error)
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
		}

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

		user, _, wantTitle, wantFirstName, wantLastName, umError := createTestUserEtc(db)
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

		it, te := db.GetUserDataFieldIDByNameIntern("SALUTATION")
		if te != nil {
			t.Error(te)
			continue
		}

		tt, tError := GetUserDataField[string](db, it, user.ID)
		if tError != nil {
			t.Error(tError)
			continue
		}

		if tt != wantTitle {
			t.Errorf("want %s got %s", wantTitle, tt)
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

		user, _, tl, fn, ln, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
		}

		var testData = []struct {
			description string
			title       string
			firstName   string
			lastName    string
			email       string
			shouldWork  bool
		}{
			{"no match", "junk", "junk", "junk", "junk", false},
			{"all match", tl, fn, ln, user.LoginName, true},
			{"email matches", "junk", "junk", "more junk", user.LoginName, true},
			{"names match", tl, fn, ln, "junk", true},
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		db, connError := ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		user, cuError := CreateUser(db)
		if cuError != nil {
			t.Error(cuError)
			continue
		}

		var testData = []struct {
			Want      string
			FieldName string
		}{
			{user.LoginName, "EMAIL"},
			{"Dr", "SALUTATION"},
			{"Peter", "FIRST_NAME"},
			{"Smith", "LAST_NAME"},
			{"1 The High Street", "STREET"},
			{"Little Hampton on the Marsh", "ADDRESS_LINE_2"},
			{"Some Bigger Place", "ADDRESS_LINE_3"},
			{"Some County", "COUNTY"},
			{"AA1 1AA", "POSTCODE"},
			{"ABC", "COUNTRY"},
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

			fieldID, ide := db.GetUserDataFieldIDByNameIntern(td.FieldName)
			if ide != nil {
				t.Error(ide)
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		user, _, _, _, _, umError := createTestUserEtc(db)
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

		user, _, _, _, _, ue := createTestUserEtc(db)
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

// TestSetGiftaid checks the SetGiftaid method.
func TestSetGiftaid(t *testing.T) {

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

		user, _, _, _, _, ue := createTestUserEtc(db)
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
		createErr1 := db.SetGiftaid(user.ID, true)
		if createErr1 != nil {
			t.Errorf("%s: %v", db.Config.Type, createErr1)
			break
		}

		// Check the field - should be true.
		got1, err1 := db.GetGiftaid(user.ID)

		if err1 != nil {
			t.Error(err1)
			return
		}

		if !got1 {
			t.Errorf("%s: expected giftaid to be set true for user %d", dbType, user.ID)
		}

		// Update the giftaid field to false.
		createErr := db.SetGiftaid(user.ID, false)
		if createErr != nil {
			t.Errorf("%s: %v", db.Config.Type, createErr)
		}

		got2, err2 := db.GetGiftaid(user.ID)

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

		user, _, _, _, _, ue := createTestUserEtc(db)
		if ue != nil {
			t.Errorf("%s: %v", dbType, ue)
			continue
		}

		assoc, _, _, _, _, ae := createTestUserEtc(db)
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
				"all fields set except User IDs.",
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: 0, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocTitle: "Mr", AssocFirstName: "George", AssocLastName: "Harrison", AssocEmail: "c@d.com",
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: 0, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocTitle: "Mr", AssocFirstName: "George", AssocLastName: "Harrison", AssocEmail: "c@d.com",
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},
			{
				"all fields set",
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID,
					Title: "Prof", FirstName: "Jane", LastName: "Smith",
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5,
					DonationToSociety: 2,
					DonationToMuseum:  6.0, Giftaid: true,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID,
					Title: "Prof", FirstName: "Jane", LastName: "Smith",
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},

			{
				"no associate",
				MembershipSale{
					ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "Jane", LastName: "Smith",
					Friend: true, FriendFeePaid: 5,
					DonationToSociety: 2,
					DonationToMuseum:  6.0, Giftaid: true,
					AssocUserID: 0,
					AssocTitle:  "Dr", AssocFirstName: "john", AssocLastName: "Jones",
					AssocFeePaid: 42.0,
					AssocFriend:  true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "c", PaymentStatus: "d", PaymentID: "e",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Title:                 "Prof", FirstName: "Jane", LastName: "Smith",
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "john", AssocLastName: "Jones",
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
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "f", PaymentStatus: "g", PaymentID: "h",
					MembershipYear: 2025, UserID: user.ID,
					OrdinaryMemberFeePaid: 24.0,
					Friend:                true, FriendFeePaid: 5, DonationToSociety: 0,
					DonationToMuseum: 0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
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
					AssocUserID: 0, AssocFeePaid: 42.0,
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},
			{
				"ordinary member is friend", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: false, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: true, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: false, AssocFriendFeePaid: 43.0,
				},
			},
			{
				"Gifaid", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: false, AssocFriendFeePaid: 0.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: true,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: false, AssocFriendFeePaid: 0.0,
				},
			},
			{
				"associate member is friend", // Set just one bool value.
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
				MembershipSale{
					ID: 0, PaymentService: "a", PaymentStatus: "b", PaymentID: "x",
					MembershipYear: 2025, UserID: user.ID, OrdinaryMemberFeePaid: 24.0,
					Title: "Prof", FirstName: "John", LastName: "Lennon", Email: "a@b.com",
					Friend: false, FriendFeePaid: 5, DonationToSociety: 2,
					DonationToMuseum: 6.0, Giftaid: false,
					AssocUserID: assoc.ID, AssocFeePaid: 42.0,
					AssocTitle: "Dr", AssocFirstName: "Vivien", AssocLastName: "Jones",
					AssocFriend: true, AssocFriendFeePaid: 43.0,
				},
			},
		}

		for _, td := range testData {
			id, createError := td.input.Create(db)
			if createError != nil {
				t.Errorf("%s: %s %v", dbType, td.description, createError)
				continue
			}

			if id == 0 {
				t.Error("expected the returned ID to be non-zero")
				continue
			}
			if td.input.ID != id {
				t.Error("expected the ID in the supplied object to be updated")
				continue
			}

			got, fetchError := db.GetMembershipSale(id)
			if fetchError != nil {
				t.Errorf("%s %s - %v", dbType, td.description, fetchError)
				continue
			}

			// The id has been set in the stored record.  Set the ID in
			// the want to match.

			td.want.ID = got.ID

			if !reflect.DeepEqual(td.want, *got) {
				t.Errorf("%s %s\nwant %v\ngot  %v", dbType, td.description, td.want, *got)
				break
			}

			const wantPaymentID = "some very long text"
			const wantPaymentStatus = "complete"

			td.want.PaymentID = wantPaymentID
			td.want.PaymentStatus = wantPaymentStatus
			got.PaymentID = wantPaymentID
			got.PaymentStatus = wantPaymentStatus

			// If the user ID is set then we can update
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

			if !reflect.DeepEqual(td.want, *updatedMS) {
				t.Errorf("%s %s\nwant %v\ngot  %v", dbType, td.description, td.want, *updatedMS)
				break
			}

			// Tidy up - delete the membership_sales record and check that it's deleted.

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

// TestMembershipSaleUpdateFailsWithUnknownID checks that a membeship sale
// update fails when the ID does not match anything in the database.
func TestMembershipSaleUpdateFailsWithUnknownID(t *testing.T) {

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

		user, _, _, _, _, ue := createTestUserEtc(db)
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
			continue
		}

		if id == 0 {
			t.Error("expected the returned ID to be non-zero")
			continue
		}

		// Set the Id to a non-existent record.
		sale.ID++

		// Expect the update to fail.
		err := sale.Update(db)

		if err == nil {
			t.Error("expected an error")
		}

		// The defers earlier ensure that the transaction is now rolled back
		// and the connection closed.
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

		user, _, _, _, _, ue := createTestUserEtc(db)
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

// CreateUser creates a user for testing.
func CreateUser(db *Database) (*User, error) {
	//  Create a user.
	loginName, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
	if e1 != nil {
		return nil, e1
	}
	user := NewUser(loginName)
	e2 := db.CreateUser(user)
	if e2 != nil {
		return nil, e2
	}

	return user, nil
}

// CreateTestUserEtc is a helper that creates a user with member
// and adm_user_data records.
func createTestUserEtc(db *Database) (*User, *Member, string, string, string, error) {
	startTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)

	roleMember, re := db.GetRole(RoleNameMember)
	if re != nil {
		return nil, nil, "", "", "", re
	}
	// Use different names each time to avoid clashes in the postgres
	// database.

	email, e1 := CreateUuid(db.Transaction, "usr_login_name", "adm_users")
	if e1 != nil {
		return nil, nil, "", "", "", e1
	}

	title, e2 := CreateUuid(db.Transaction, "usd_value", "adm_user_data")
	if e2 != nil {
		return nil, nil, "", "", "", e2
	}

	firstName, e3 := CreateUuid(db.Transaction, "usd_value", "adm_user_data")
	if e3 != nil {
		return nil, nil, "", "", "", e3
	}

	lastName, e4 := CreateUuid(db.Transaction, "usd_value", "adm_user_data")
	if e4 != nil {
		return nil, nil, "", "", "", e4
	}
	user, member, umError := db.CreateUserAndMember(email, title, firstName, lastName, roleMember, startTime, endTime)

	return user, member, title, firstName, lastName, umError
}
