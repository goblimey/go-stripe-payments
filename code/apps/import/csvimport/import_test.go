package csvimport

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/goblimey/go-stripe-payments/code/pkg/database"

	_ "modernc.org/sqlite"
)

var databaseList = []string{"postgres", "sqlite"}

// TestAdmidioEmailFilter checks the query used by the Admidio system to get the list of members to
// email - used by message_write.php in /var/www/html/members.sihg.org.uk/admidio/adm_program/modules/messages.
func TestAdmidioEmailFilter(t *testing.T) {

	ukTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("failed to get london timezone - %v", err)
	}

	// Year start is 1st April.
	membershipStart := time.Date(2025, time.April, 1, 0, 0, 0, 0, ukTime)
	// Year end is 31st March in the next year.
	membershipEnd := time.Date(2026, time.March, 31, 23, 59, 59, 999999999, ukTime)

	// Year start is 1st April.
	lastYearMembershipStart := time.Date(2024, time.April, 1, 0, 0, 0, 0, ukTime)
	// Year end is 31st March in the next year.
	lastYearMembershipEnd := time.Date(2025, time.March, 31, 23, 59, 59, 999999999, ukTime)

	// The day before the membership lapses.
	dayBefore := time.Date(2026, time.March, 30, 23, 59, 59, 999999999, ukTime)
	// The very end of the last day of membership.
	lastDay := time.Date(2026, time.March, 31, 23, 59, 59, 999999999, ukTime)
	// The start of the day after the membership lapses.
	dayAfter := time.Date(2026, time.April, 1, 0, 0, 0, 0, ukTime)

	for _, dbType := range databaseList {

		db, connError := database.OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := database.PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		un1, un1e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un1e != nil {
			t.Errorf("%s - %v", "un7", un1e)
			return
		}

		un2, un2e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un2e != nil {
			t.Errorf("%s - %v", "un7", un2e)
			return
		}

		un3, un3e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un3e != nil {
			t.Errorf("%s - %v", "un7", un3e)
			return
		}

		un4, un4e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un4e != nil {
			t.Errorf("%s - %v", "un7", un4e)
			return
		}

		un5, un5e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un5e != nil {
			t.Errorf("%s - %v", "un7", un5e)
			return
		}

		un6, un6e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un6e != nil {
			t.Errorf("%s - %v", "un7", un6e)
			return
		}

		un7, un7e := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if un7e != nil {
			t.Errorf("%s - %v", "un7", un7e)
			return
		}

		firstname, fne := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if fne != nil {
			t.Errorf("%s - %v", "firstname", fne)
			return
		}

		lastname, psee := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if psee != nil {
			t.Errorf("%s - %v", "lastname", psee)
			return
		}

		var testData = []struct {
			now  time.Time
			line CSVLine
		}{
			{
				// This one should be in the list.
				dayBefore,
				CSVLine{
					UserName:        un1,
					Email:           un1,
					FirstName:       "john",
					Surname:         "smith",
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},

			{
				// This one should be in the list.
				lastDay,
				CSVLine{
					UserName:        un2,
					Email:           un2,
					FirstName:       "jack",
					Surname:         "smith",
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
			{
				// This one will have its valid flag unset (below) so it should not be in the list.
				dayBefore,
				CSVLine{
					UserName:        un3,
					Email:           un3,
					FirstName:       "harry",
					Surname:         "jones",
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
			{
				// This one will have its member record removed (below) so should not be in the list.
				dayBefore,
				CSVLine{
					UserName:        un4,
					Email:           un4,
					FirstName:       "jarvis",
					Surname:         "cocker",
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
			{
				// This one should not be in the list - lapsed member.
				dayAfter,
				CSVLine{
					UserName:        un5,
					Email:           un5,
					FirstName:       "mick",
					Surname:         "jagger",
					MembershipStart: lastYearMembershipStart,
					MembershipEnd:   lastYearMembershipEnd,
				},
			},
			{
				// This one should not be in the list - email is empty.
				dayBefore,
				CSVLine{
					UserName:        un6,
					Email:           "",
					FirstName:       firstname,
					Surname:         lastname,
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
			{
				// This one should not be in the list - email is not suplied at all.
				dayBefore,
				CSVLine{
					UserName:        firstname + "." + lastname,
					Email:           "",
					FirstName:       firstname,
					Surname:         lastname,
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
			{
				// This one should not be in the list.  Its permission to send emails will be removed (see below).
				lastDay,
				CSVLine{
					UserName:        un7,
					Email:           un7,
					FirstName:       "David",
					Surname:         "Bowie",
					MembershipStart: membershipStart,
					MembershipEnd:   membershipEnd,
				},
			},
		}

		// Create the users, members etc.
		for _, td := range testData {
			id, err := ProcessRecord(db, &td.line)
			if err != nil {
				t.Error(td.line.UserName + ": " + err.Error())
				return
			}
			if id <= 0 {
				t.Errorf("%s: want positive ID got %d", td.line.UserName, id)
				return
			}
		}

		// Unset the valid flag for un3.
		u3, ue3 := db.GetUsersByLoginName(un3)
		if ue3 != nil {
			t.Error(ue3)
			return
		}

		if len(u3) != 1 {
			t.Errorf("want 1 user got %d", len(u3))
			return
		}
		u3[0].Valid = false
		uu3e := db.UpdateUser(&u3[0])
		if uu3e != nil {
			t.Error(uu3e)
			return
		}

		// Delete the member record for user un4.
		u4, ue4 := db.GetUsersByLoginName(un4)
		if ue4 != nil {
			t.Error(ue4)
			return
		}

		if len(u4) != 1 {
			t.Errorf("want 1 user got %d", len(u4))
			return
		}

		m, me := db.GetMemberOfUser(&u4[0])
		if me != nil {
			t.Error(me)
		}

		mde := db.DeleteMember(m)
		if mde != nil {
			t.Error(mde)
		}

		// Remove the email permission from un7.
		u7, ue7 := db.GetUsersByLoginName(un7)
		if ue7 != nil {
			t.Error(ue7)
			return
		}

		if len(u7) != 1 {
			t.Errorf("want 1 user got %d", len(u7))
			return
		}

		pse, psee := db.GetUserDataFieldIDByNameIntern(database.EmailPermNameIntern)
		if psee != nil {
			t.Error(psee)
		}

		// The permission field is 1 (allowed) or 0 (not allowed).
		// Set it to not allowed.
		snError := database.SetUserDataField(db, pse, u7[0].ID, 0)
		if snError != nil {
			t.Error(snError)
			return
		}

		// sql1, sql2 etc filter the users. Each applies all the filters from the preceding
		// query plus one more.

		// sql1 fetches all the users, including the System user set up by PrepareTestTables.
		const sqlTemplate1 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				 ON fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				 ON firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				 ON fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				 ON lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id;
		`
		var sql1 string
		switch db.Config.Type {
		case "postgres":
			sql1 = fmt.Sprintf(sqlTemplate1, "COALESCE", "COALESCE")
		default:
			sql1 = fmt.Sprintf(sqlTemplate1, "IFNULL", "IFNULL")
		}

		rows1, selectError1 := db.Query(sql1)
		if selectError1 != nil {
			t.Error(selectError1)
			return
		}

		// Count the returned rows.
		var n1 int
		for {
			if !rows1.Next() {
				break
			}
			n1++

		}
		rows1.Close()

		if n1 != 9 {
			t.Errorf("want 9 users got %d", n1)
			return
		}

		// sql2 filters out users whose valid flag is not set - e3.
		const sqlTemplate2 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				 ON fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				 ON firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				 ON fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				 ON lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id
			where u.usr_valid = 't';
		`

		var sql2 string
		switch db.Config.Type {
		case "postgres":
			sql2 = fmt.Sprintf(sqlTemplate2, "COALESCE", "COALESCE")
		default:
			sql2 = fmt.Sprintf(sqlTemplate2, "IFNULL", "IFNULL")
		}

		rows2, selectError2 := db.Query(sql2)
		if selectError2 != nil {
			t.Error(selectError2)
		}

		// Count the returned rows and watch the names - un3 should not appear.
		var n2 int
		for {
			if !rows2.Next() {
				break
			}

			var u database.User
			var fn, ln string

			ue := rows2.Scan(&u.ID, &u.LoginName, &fn, &ln)
			if ue != nil {
				t.Error(ue)
				return
			}
			if u.LoginName == un3 {
				t.Error("invalid user filter failed")
				return
			}
			n2++
		}

		rows2.Close()

		if n2 != 8 {
			t.Errorf("want 8 users got %d", n2)
			return
		}

		// sql3 also filters out users who do have Member status - System and un4
		const sqlTemplate3 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				 ON fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				 ON firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				 ON fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				 ON lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id
			INNER JOIN adm_members as m
					 ON u.usr_id = m.mem_usr_id
			inner join adm_roles as r
				 ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			where u.usr_valid = 't';
		`

		var sql3 string
		switch db.Config.Type {
		case "postgres":
			sql3 = fmt.Sprintf(sqlTemplate3, "COALESCE", "COALESCE")
		default:
			sql3 = fmt.Sprintf(sqlTemplate3, "IFNULL", "IFNULL")
		}

		rows3, selectError3 := db.Query(sql3)
		if selectError3 != nil {
			t.Error(selectError3)
		}

		// Count the returned rows and watch the names - System and un4 should not appear.
		var n3 int
		for {
			if !rows3.Next() {
				break
			}

			var u database.User
			var fn, ln string

			ue := rows3.Scan(&u.ID, &u.LoginName, &fn, &ln)
			if ue != nil {
				t.Error(ue)
				return
			}
			if u.LoginName == "System" || u.LoginName == un4 {
				t.Errorf("non member filter failed %s", u.LoginName)
				return
			}
			n3++
		}

		rows3.Close()

		if n3 != 6 {
			t.Errorf("want 6 users got %d", n3)
			return
		}

		// sql4 also filters out users whose membership has lapsed.
		const sqlTemplate4 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				on fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				on firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				on fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				on lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id
			INNER JOIN adm_members as m
				 ON u.usr_id = m.mem_usr_id
			inner join adm_roles as r
				 ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			WHERE u.usr_valid = 't'
			AND m.mem_end >= $1;
		`

		var sql4 string
		switch db.Config.Type {
		case "postgres":
			sql4 = fmt.Sprintf(sqlTemplate4, "COALESCE", "COALESCE")
		default:
			sql4 = fmt.Sprintf(sqlTemplate4, "IFNULL", "IFNULL")
		}

		rows4, selectError4 := db.Query(sql4, membershipStart.Format("2006-01-02"))
		if selectError4 != nil {
			t.Error(selectError4)
		}

		// Count the returned rows and watch the names - un5 should not appear.
		var n4 int
		for {
			if !rows4.Next() {
				break
			}

			var u database.User
			var fn, ln string

			ue := rows4.Scan(&u.ID, &u.LoginName, &fn, &ln)
			if ue != nil {
				t.Error(ue)
			}
			if u.LoginName == un5 {
				t.Errorf("lapsed member filter failed %s", u.LoginName)
			}
			n4++
		}

		rows4.Close()

		if n4 != 5 {
			t.Errorf("want 5 users got %d", n4)
			return
		}

		// sql5 also filters out people who have an email entry in adm_user_data but it's empty.
		const sqlTemplate5 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, ''),
			%s(email.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				 ON fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				 ON firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				 ON fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				 ON lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id
			INNER JOIN adm_members as m
				 ON u.usr_id = m.mem_usr_id
			inner join adm_roles as r
				  ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			LEFT JOIN adm_user_fields as fieldEmail
				ON fieldEmail.usf_name_intern = 'EMAIL'
				AND fieldEmail.usf_type = 'EMAIL'
			INNER JOIN adm_user_data as email
				 ON email.usd_usr_id = u.usr_id
				AND fieldEmail.usf_id = email.usd_usf_id
				AND length(email.usd_value) > 0
			WHERE u.usr_valid = 't'
				AND m.mem_end >= $1;
		`

		var sql5 string
		switch db.Config.Type {
		case "postgres":
			sql5 = fmt.Sprintf(sqlTemplate5, "COALESCE", "COALESCE", "COALESCE")
		default:
			sql5 = fmt.Sprintf(sqlTemplate5, "IFNULL", "IFNULL", "IFNULL")
		}

		rows5, selectError5 := db.Query(sql5, membershipStart.Format("2006-01-02"))
		if selectError5 != nil {
			t.Error(selectError5)
			return
		}

		// Count the returned rows and watch the names - un6 and firstname.lastname should not appear.
		var n5 int
		for {
			if !rows5.Next() {
				break
			}

			var u database.User
			var fn, ln, email string

			ue := rows5.Scan(&u.ID, &u.LoginName, &fn, &ln, &email)
			if ue != nil {
				t.Error(ue)
			}
			if u.LoginName == un6 || u.LoginName == firstname+"."+lastname {
				t.Errorf("empty email filter failed %s", u.LoginName)
			}
			n5++
		}

		rows5.Close()

		if n5 != 3 {
			t.Errorf("want 3 users got %d", n5)
			return
		}

		// sql6 also filters out people who have no email entry in adm_user_data.
		// This is the query used in the Admidio system, except for the last line,
		// where m.mem_end is compared with CURRENT_DATE.
		const sqlTemplate6 = `
			SELECT u.usr_id, u.usr_login_name, %s(firstName.usd_value, ''), %s(lastName.usd_value, ''),
			%s(email.usd_value, '')
			FROM adm_users as u
			LEFT JOIN adm_user_fields as fieldFirstName
				 ON fieldFirstName.usf_name_intern = 'FIRST_NAME'
			LEFT JOIN adm_user_data as firstName
				 ON firstName.usd_usr_id = u.usr_id
				AND fieldFirstName.usf_id = firstName.usd_usf_id
			LEFT JOIN adm_user_fields as fieldLastName
				 ON fieldLastName.usf_name_intern = 'LAST_NAME'
			LEFT JOIN adm_user_data as lastName
				 ON lastName.usd_usr_id = u.usr_id
				AND fieldLastName.usf_id = lastName.usd_usf_id
			INNER JOIN adm_members as m
				 ON u.usr_id = m.mem_usr_id
			inner join adm_roles as r
				 ON r.rol_id = m.mem_rol_id
				AND r.rol_name = 'Member'
			LEFT JOIN adm_user_fields as fieldEmail
				ON fieldEmail.usf_name_intern = 'EMAIL'
				AND fieldEmail.usf_type = 'EMAIL'
			INNER JOIN adm_user_data as email
				 ON email.usd_usr_id = u.usr_id
				AND fieldEmail.usf_id = email.usd_usf_id
				AND length(email.usd_value) > 0
			LEFT JOIN adm_user_fields  as fieldPerm
				ON fieldPerm.usf_name_intern = '` + database.EmailPermNameIntern + `'
			INNER join adm_user_data as perm
				ON perm.usd_usr_id = u.usr_id
				AND perm.usd_usf_id = fieldPerm.usf_id
				AND perm.usd_value = '1'
			WHERE u.usr_valid = 't' 
			AND m.mem_end >= $1;
		`

		var sql6 string
		switch db.Config.Type {
		case "postgres":
			sql6 = fmt.Sprintf(sqlTemplate6, "COALESCE", "COALESCE", "COALESCE")
		default:
			sql6 = fmt.Sprintf(sqlTemplate6, "IFNULL", "IFNULL", "IFNULL")
		}

		rows6, selectError6 := db.Query(sql6, membershipStart.Format("2006-01-02"))
		if selectError6 != nil {
			t.Error(selectError6)
			return
		}

		// Count the returned rows and watch the names - un7 should not appear.
		var n6 int
		for {
			if !rows6.Next() {
				break
			}

			var u database.User
			var fn, ln, email string

			ue := rows6.Scan(&u.ID, &u.LoginName, &fn, &ln, &email)
			if ue != nil {
				t.Error(ue)
			}

			if u.LoginName == un7 {
				t.Errorf("no permission filter failed %s", u.LoginName)
			}
			n6++
		}

		rows6.Close()

		if n6 != 2 {
			t.Errorf("want 2 users got %d", n6)
			return
		}
	}
}

func TestImport(t *testing.T) {

	records, err := Import("./members.test.csv", 2025)
	if err != nil {
		m := err.Error()
		t.Error(m)
	}

	if len(records) == 0 {
		t.Error("expected some records")
		return
	}

	if len(records) != 6 {
		t.Errorf("want 6 records got %d", len(records))
		return
	}

	if records[0].UserName != "a@gmail.com" {
		t.Errorf("wantuser name  a@gmail.com got %s", records[0].UserName)
	}

	if records[0].Email != "A@gmail.com" {
		t.Errorf("want A@gmail.com got %s", records[0].Email)
	}

	if records[0].Mobile != "" {
		t.Errorf("want empty string got %s", records[0].Mobile)
	}

	if records[1].UserName != "b@outlook.com" {
		t.Errorf("want user name b@outlook.com got %s", records[0].UserName)
	}

	if records[1].Email != "B@OUTLOOK.COM" {
		t.Errorf("want B@OUTLOOK.COM got %s", records[1].Email)
	}

	if records[1].Mobile != "07748 111111" {
		t.Errorf("want 07748 111111 got %s", records[1].Mobile)
	}

	if records[2].UserName != "c@outlook.com" {
		t.Errorf("want user name c@outlook.com got %s", records[2].UserName)
	}

	if records[3].UserName != "dennis.finch.hatton" {
		t.Errorf("want user name dennis.finch.hatton got %s", records[3].UserName)
	}

	if records[4].UserName != "dennis" {
		t.Errorf("want user name dennis got %s", records[4].UserName)
	}

	if records[5].UserName != "hatton" {
		t.Errorf("want user name hatton got %s", records[5].UserName)
	}
}

// TestImportEmailOnly tests an import of a file containing only email addresses.
func TestImportEmailOnly(t *testing.T) {

	records, err := Import("./members.email.only.csv", 2025)
	if err != nil {
		m := err.Error()
		t.Error(m)
	}

	if len(records) == 0 {
		t.Error("expected some records")
	}

	if len(records) != 1 {
		t.Errorf("want 1 record got %d", len(records))
	}

	if records[0].UserName != "c@example.com" {
		t.Errorf("wantuser name  c@example.com got %s", records[0].UserName)
	}

	if records[0].Email != "c@example.com" {
		t.Errorf("want c@example.com got %s", records[0].Email)
	}
}

// TestImportJustEmail tests an import of a file containing only emails.
func TestImportJustEmail(t *testing.T) {

	records, err := Import("./members.email.only.csv", 2025)
	if err != nil {
		m := err.Error()
		t.Error(m)
	}

	if len(records) == 0 {
		t.Error("expected some records")
	}

	if len(records) != 1 {
		t.Errorf("want 1 record got %d", len(records))
	}

	if records[0].UserName != "c@example.com" {
		t.Errorf("wantuser name  c@example.com got %s", records[0].UserName)
	}

	if records[0].Email != "c@example.com" {
		t.Errorf("wantuser name  c@example.com got %s", records[0].Email)
	}
}

func TestImportToDatabase(t *testing.T) {

	ukTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("failed to get london timezone - %v", err)
	}

	// Year end is 31st March.
	membershipEnd := time.Date(2026, time.March, 31, 23, 59, 59, 999999999, ukTime)
	// Year start is 1st April in the previous year.
	membershipStart := time.Date(2025, time.April, 1, 0, 0, 0, 0, ukTime)

	for _, dbType := range databaseList {
		db, connError := database.OpenDBForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			continue
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := database.PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		if dbType == "sqlite" {
			log.Printf("%s\\%s", db.SQLiteTempDir+"\\sqlite.db", db.Config.Name)
		}

		email1, e1 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e1 != nil {
			t.Errorf("%s - %v", email1, e1)
			continue
		}

		email2, e2 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if e2 != nil {
			t.Errorf("%s - %v", email2, e2)
			continue
		}

		firstName, e3 := database.CreateUuid(db.Transaction, "usd_value", "adm_user_data")
		if e3 != nil {
			t.Errorf("%s - %v", firstName, e3)
			continue
		}

		lastName, e4 := database.CreateUuid(db.Transaction, "usd_value", "adm_user_data")
		if e4 != nil {
			t.Errorf("%s - %v", firstName, e4)
			continue
		}
		wantLine := []CSVLine{
			{
				UserName:        email1,
				Email:           email1,
				Title:           "Mr",
				FirstName:       "Alan",
				Surname:         "Smith",
				AddressLine1:    "Flat2",
				AddressLine2:    "Apsley Mansion",
				AddressLine3:    "1 The High Street",
				Town:            "Farnham",
				County:          "Surrey",
				Postcode:        "GU9 0AB",
				Country:         "UK",
				Phone:           "01234 567890",
				Mobile:          "07890123456",
				MembershipStart: membershipStart,
				MembershipEnd:   membershipEnd,
			},
			{
				UserName:        email2,
				Email:           email2,
				Title:           "Dr",
				FirstName:       "Roger",
				Surname:         "Jones",
				AddressLine1:    "Flat 1",
				AddressLine2:    "Bookham Towers",
				AddressLine3:    "The High Street",
				Town:            "Bookham",
				County:          "Hampshire",
				Postcode:        "KT21 1AA",
				Country:         "UK",
				Phone:           "01234 567890",
				Mobile:          "07748 111111",
				MembershipStart: membershipStart,
				MembershipEnd:   membershipEnd,
			},
			{
				UserName:        firstName + "." + lastName,
				Email:           "",
				Title:           "Dr",
				FirstName:       "Roger",
				Surname:         "Jones",
				AddressLine1:    "Flat 1",
				AddressLine2:    "Bookham Towers",
				AddressLine3:    "The High Street",
				Town:            "Bookham",
				County:          "Hampshire",
				Postcode:        "KT21 1AA",
				Country:         "UK",
				Phone:           "01234 567890",
				Mobile:          "07748 111111",
				MembershipStart: membershipStart,
				MembershipEnd:   membershipEnd,
			},
		}

		// Create the users, members etc.
		for _, record := range wantLine {
			id, err := ProcessRecord(db, &record)
			if err != nil {
				t.Error(record.UserName + ": " + err.Error())
				continue
			}
			if id <= 0 {
				t.Errorf("%s: want positive ID got %d", record.UserName, id)
				continue
			}
		}

		users, getUsersError := db.GetUsers()

		if getUsersError != nil {
			t.Error(getUsersError)
			continue
		}

		// There is already the system user and we created three more.
		if len(users) < 4 {
			t.Errorf("want 4 users got %d", len(users))
			continue
		}

		if len(users) > 4 {
			t.Errorf("want 4 users got too many - %d", len(users))
			continue
		}

		// The system user doesn't need a member so there are only three memers.
		members, getMembersError := db.GetMembers()
		if getMembersError != nil {
			t.Error(getMembersError)
			continue
		}

		if len(members) != 3 {
			t.Errorf("want 3 members got %d", len(members))
			continue
		}

		for _, member := range members {
			// Date fields in sqlite are like "2025-04-01", in Postgres "2025-04-01T00:00:00Z".
			ms := "2025-04-01"
			if dbType == "postgres" {
				ms = ms + "T00:00:00Z"
			}
			me := "2026-03-31"
			if dbType == "postgres" {
				me = me + "T00:00:00Z"
			}

			if member.StartDate != ms {
				t.Errorf("member %d want start %s got %s", member.ID, ms, member.StartDate)
			}
			if member.EndDate != me {
				t.Errorf("member %d want end %s got %s", member.ID, me, member.EndDate)
			}
		}

		for i := range wantLine {

			// Get the users with the given email address - should be exactly one.
			users, getUserError := db.GetUsersByLoginName(wantLine[i].UserName)
			if getUserError != nil {
				t.Error(getUserError)
				continue
			}

			if len(users) < 1 {
				t.Error("got no users")
				continue
			}

			if len(users) > 1 {
				t.Errorf("want 1 user got %d", len(users))
				continue
			}

			fetchedUser := users[0]

			if fetchedUser.LoginName != wantLine[i].UserName {
				t.Errorf("want login name %s got %s", wantLine[i].UserName, fetchedUser.LoginName)
				continue
			}

			if fetchedUser.ID != members[i].UserID {
				t.Errorf("want %d got %d", users[i].ID, members[i].UserID)
				continue
			}

			if len(wantLine[i].Email) > 0 {
				// Check fields in adm_user_data.
				emID, emError :=
					db.GetUserDataFieldIDByNameIntern("EMAIL")
				if emError != nil {
					t.Error(emError)
					continue
				}

				email, emailError := database.GetUserDataField[string](db, emID, fetchedUser.ID)
				if emailError != nil {
					t.Error(emailError)
					continue
				}

				if email != wantLine[i].Email {
					t.Errorf("%s: want %s got %s", dbType, wantLine[i].Email, email)
					continue
				}
			} else {
				// Expect no rows in the results.
				emID, emError :=
					db.GetUserDataFieldIDByNameIntern("EMAIL")
				if emError != nil {
					t.Error(emError)
					continue
				}

				_, emailError := database.GetUserDataField[string](db, emID, fetchedUser.ID)

				if !strings.Contains(emailError.Error(), "no rows") {
					t.Errorf("%s: %s expected a 'no rows'  error", dbType, wantLine[i].Email)
					continue
				}
			}

			title, titleError :=
				database.GetUserDataField[string](db, db.UserField["SALUTATION"].ID, fetchedUser.ID)
			if titleError != nil {
				t.Error(titleError)
				continue
			}

			if title != wantLine[i].Title {
				t.Errorf("got %s want %s", title, wantLine[i].Title)
				continue
			}

			ifn, fne := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
			if fne != nil {
				t.Error(fne)
				continue
			}

			val, fnError := database.GetUserDataField[string](db, ifn, fetchedUser.ID)
			if fnError != nil {
				t.Error(fnError)
				continue
			}
			if val != wantLine[i].FirstName {
				t.Errorf("want %s got %s", wantLine[i].FirstName, val)
				continue
			}

			iln, lne := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
			if lne != nil {
				t.Error(lne)
				continue
			}

			val, lnError := database.GetUserDataField[string](db, iln, fetchedUser.ID)
			if lnError != nil {
				t.Error(lnError)
				continue
			}
			if val != wantLine[i].Surname {
				t.Errorf("want %s got %s", wantLine[i].Surname, val)
				continue
			}

			street, sError :=
				database.GetUserDataField[string](db, db.UserField["STREET"].ID, fetchedUser.ID)
			if sError != nil {
				t.Error(sError)
				continue
			}

			if street != wantLine[i].AddressLine1 {
				t.Errorf("got %s want %s", street, wantLine[i].AddressLine1)
				continue
			}

			al2, al2Error :=
				database.GetUserDataField[string](db, db.UserField["ADDRESS_LINE_2"].ID, fetchedUser.ID)
			if al2Error != nil {
				t.Error(al2Error)
				continue
			}

			if al2 != wantLine[i].AddressLine2 {
				t.Errorf("got %s want %s", al2, wantLine[i].AddressLine2)
				continue
			}

			al3, al3Error :=
				database.GetUserDataField[string](db, db.UserField["ADDRESS_LINE_3"].ID, fetchedUser.ID)
			if al3Error != nil {
				t.Error(al3Error)
				continue
			}

			if al3 != wantLine[i].AddressLine3 {
				t.Errorf("got %s want %s", al3, wantLine[i].AddressLine3)
				continue
			}

			city, cError :=
				database.GetUserDataField[string](db, db.UserField["CITY"].ID, fetchedUser.ID)
			if cError != nil {
				t.Error(cError)
				continue
			}

			if city != wantLine[i].Town {
				t.Errorf("got %s want %s", city, wantLine[i].Town)
				continue
			}

			county, cError :=
				database.GetUserDataField[string](db, db.UserField["COUNTY"].ID, fetchedUser.ID)
			if cError != nil {
				t.Error(cError)
				continue
			}

			if county != wantLine[i].County {
				t.Errorf("got %s want %s", county, wantLine[i].County)
				continue
			}

			postcode, pcError :=
				database.GetUserDataField[string](db, db.UserField["POSTCODE"].ID, fetchedUser.ID)
			if pcError != nil {
				t.Error(pcError)
				continue
			}

			if postcode != wantLine[i].Postcode {
				t.Errorf("got %s want %s", postcode, wantLine[i].Postcode)
				continue
			}

			country, ctrError :=
				database.GetUserDataField[string](db, db.UserField["COUNTRY"].ID, fetchedUser.ID)
			if ctrError != nil {
				t.Error(ctrError)
				continue
			}

			if country != wantLine[i].Country {
				t.Errorf("got %s want %s", country, wantLine[i].Country)
				continue
			}

			phone, pError :=
				database.GetUserDataField[string](db, db.UserField["PHONE"].ID, fetchedUser.ID)
			if pError != nil {
				t.Error(pError)
				continue
			}

			if phone != wantLine[i].Phone {
				t.Errorf("got %s want %s", phone, wantLine[i].Phone)
				continue
			}

			mobile, mError :=
				database.GetUserDataField[string](db, db.UserField["MOBILE"].ID, fetchedUser.ID)
			if mError != nil {
				t.Error(mError)
				continue
			}

			if mobile != wantLine[i].Mobile {
				t.Errorf("got %s want %s", mobile, wantLine[i].Mobile)
				continue
			}
		}
	}
}

// TestProcesRecordEmailOnly checks that ProcessRecord handles a record
// where only an Email address is given/
func TestProcesRecordEmailOnly(t *testing.T) {

	ukTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("failed to get london timezone - %v", err)
	}

	// Year end is 31st March.
	membershipEnd := time.Date(2026, time.March, 31, 23, 59, 59, 999999999, ukTime)
	// Year start is 1st April in the previous year.
	membershipStart := time.Date(2025, time.April, 1, 0, 0, 0, 0, ukTime)

	line := CSVLine{
		UserName:        "a@gmail.com",
		Email:           "a@gmail.com",
		MembershipStart: membershipStart,
		MembershipEnd:   membershipEnd,
	}

	for _, dbType := range databaseList {
		db, connError := database.OpenDBForTesting(dbType)
		if connError != nil {
			t.Error(connError)
			return
		}

		db.BeginTx()

		defer db.Rollback()
		defer db.CloseAndDelete()

		prepError := database.PrepareTestTables(db)
		if prepError != nil {
			t.Error(prepError)
		}

		log.Printf("%s\\%s", db.SQLiteTempDir+"\\sqlite.db", db.Config.Name)

		// Create the users, members etc.
		id, createError := ProcessRecord(db, &line)
		if createError != nil {
			t.Error(createError)
			return
		}

		// Get the users with the given email address - should be exactly one.
		users, getUserError := db.GetUsersByLoginName(line.UserName)
		if getUserError != nil {
			t.Error(getUserError)
		}

		if len(users) < 1 {
			t.Error("got no users")
			return
		}

		if len(users) > 1 {
			t.Errorf("want 1 user got %d", len(users))
		}

		fetchedUser := users[0]

		if fetchedUser.ID != id {
			t.Errorf("want id %d got %d", id, fetchedUser.ID)
			return
		}

		if fetchedUser.LoginName != line.UserName {
			t.Errorf("want login name %s got %s", line.UserName, fetchedUser.LoginName)
			return
		}
	}
}
