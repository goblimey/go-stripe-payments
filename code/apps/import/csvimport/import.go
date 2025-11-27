package csvimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/goblimey/go-stripe-payments/code/pkg/database"
)

// CSVLine holds the data about a record in the DB extract from SyAS,
// supplied as a CSV file.
type CSVLine struct {
	UserName,
	Email,
	Title,
	Initials,
	FirstName,
	MiddleName,
	Surname,
	Position,
	Organisation,
	AddressLine1,
	AddressLine2,
	AddressLine3,
	Town,
	County,
	Postcode,
	Country,
	Phone,
	Mobile,
	MatchingSIHGLlistEmail string
	MembershipEnd,
	MembershipStart time.Time
}

var compressSpaceRegex *regexp.Regexp

func init() {

	// A regex to capture a stream of white space.
	compressSpaceRegex = regexp.MustCompile("[ \t\n]+")
}

func Import(file fs.File, lastYearOfMembership int) ([]CSVLine, error) {

	// records is the returned object.
	records := make([]CSVLine, 0)

	ukTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("failed to get london timezone - %v", err)
	}

	// Year end is 31st March in the given year.
	membershipEnd := time.Date(lastYearOfMembership, time.March, 31, 23, 59, 59, 999999999, ukTime)
	// Year start is 1st April in the previous year.
	membershipStart := time.Date(lastYearOfMembership-1, time.April, 1, 0, 0, 0, 0, ukTime)

	// Closes the file
	defer file.Close()

	// Open a reader reading the CSV file.
	reader := csv.NewReader(file)

	// The reader returns the contents of the CSV file as a
	// slice of slices of strings.
	lines, readError := reader.ReadAll()

	if readError != nil {
		fmt.Println("Error reading records")
		return records, readError
	}

	// Create a slice of CSVline objects.
	for l, fields := range lines {

		record, getError := getLine(fields)

		if getError != nil {
			// l starts at 0, not 1.
			m := fmt.Sprintf("line %d: ", l+1)
			for _, field := range fields {
				m += field + ","
			}
			slog.Error(m)
		}

		if record == nil {
			// If the first line is the heading, we get a (nil, nil)
			// return which we ignore.
			continue
		}

		record.MembershipStart = membershipStart
		record.MembershipEnd = membershipEnd

		records = append(records, *record)
	}

	return records, nil
}

func getLine(field []string) (*CSVLine, error) {

	// Each line should contain the same number of fields, either one (an email
	// address) or 17.
	const expectedNumberOfFields = 17

	if len(field) > 1 && len(field) < expectedNumberOfFields {
		m := fmt.Sprintf("expected at least %d fields, got %d",
			expectedNumberOfFields, len(field))
		return nil, errors.New(m)
	}

	if field[0] == "Email" {
		// The first field of the line is the column heading.  Ignore it.
		return nil, nil
	}

	// Trim leading and trailing white space from all fields.
	for i := range field {
		//  Must use the real field, not a copy.
		field[i] = strings.TrimSpace(field[i])
	}

	var record CSVLine
	switch len(field) {
	case 1:
		record.Email = field[0]
	default:
		// Admidio insists that the Country is the three-letter version, if set.
		// There is no contry specified in the data, so assume GBR.
		record = CSVLine{
			Email:                  field[0],
			Title:                  field[1],
			Initials:               field[2],
			FirstName:              field[3],
			MiddleName:             field[4],
			Surname:                field[5],
			Position:               field[6],
			Organisation:           field[7],
			AddressLine1:           field[8],
			AddressLine2:           field[9],
			AddressLine3:           field[10],
			Town:                   field[11],
			County:                 field[12],
			Postcode:               field[13],
			Country:                "GBR",
			Phone:                  field[14],
			Mobile:                 field[15],
			MatchingSIHGLlistEmail: field[16],
		}

	}

	var userNameError error
	record.UserName, userNameError = database.GetUserName(record.Email, record.FirstName, record.Surname)
	if userNameError != nil {
		return nil, userNameError
	}

	return &record, nil
}

func CreateRecords(db *database.Database, file fs.File, membershipYearEnd int) {

	records, importError := Import(file, membershipYearEnd)

	if importError != nil {
		slog.Error(importError.Error())
		os.Exit(-1)
	}

	for _, record := range records {
		line := fmt.Sprintf("%d %s %s: %s",
			record.MembershipEnd.Year(),
			record.FirstName,
			record.Surname,
			record.Email)

		slog.Info(line)
		ProcessRecord(db, &record)
	}
}

// ProcessRecord creates a user, member etc from the given CSV line.  On
// success it returns the ID of the user.
func ProcessRecord(db *database.Database, line *CSVLine) (int64, error) {

	// Check that the account name is not already in use - sometimes two  members use the
	// same email address.

	memberExists, checkMemberError := db.MemberExists(line.UserName, line.Email)
	if checkMemberError != nil {
		return 0, checkMemberError
	}

	if memberExists {
		m := fmt.Sprintf("already joined: %s %s", line.UserName, line.Surname)
		slog.Error(m)
		return 0, nil
	}

	// The LoginName is available.  Create the records.

	role, roleError := db.GetRole("Member")
	if roleError != nil {
		return 0, roleError
	}

	if len(line.UserName) == 0 {
		return 0, errors.New("ProcessRecord: no user name")
	}

	user := database.NewUser(line.UserName)
	createUserError := db.CreateUser(user)
	if createUserError != nil {
		slog.Error(createUserError.Error())
		db.Rollback()
		return 0, createUserError
	}

	member := database.NewMember(user, role, line.MembershipStart, line.MembershipEnd)
	memberError := db.CreateMember(member)
	if memberError != nil {
		slog.Error(memberError.Error())
		db.Rollback()
		return 0, memberError
	}

	// A member ca nly be in the download if they have given this permission.
	idp, idpe := db.GetUserDataFieldIDByNameIntern(database.DataStoragePermNameIntern)
	if idpe != nil {
		slog.Error(database.DataStoragePermNameIntern + ": " + idpe.Error())
		db.Rollback()
		return 0, idpe
	}

	// This "boolean" field contains 1 (yes) or 0 (no).
	err := database.SetUserDataField(db, idp, user.ID, 1)
	if err != nil {
		slog.Error(err.Error())
		db.Rollback()
		return 0, err
	}

	if len(line.Email) > 0 {
		// Also, assume permission to send emails and leave it to the user to turn it off.
		ipe, ipee := db.GetUserDataFieldIDByNameIntern(database.EmailPermNameIntern)
		if ipee != nil {
			slog.Error(database.EmailPermNameIntern + ": " + ipee.Error())
			db.Rollback()
			return 0, ipee
		}
		// This "boolean" field contains 1 (yes) or 0 (no).
		err := database.SetUserDataField(db, ipe, user.ID, 1)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}

	}

	if len(line.Title) > 0 {
		is, se := db.GetUserDataFieldIDByNameIntern("SALUTATION")
		if se != nil {
			slog.Error("SALUTATION: " + se.Error())
			db.Rollback()
			return 0, se
		}
		err := database.SetUserDataField(db, is, user.ID, line.Title)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}

	}

	if len(line.FirstName) > 0 {
		ifn, fne := db.GetUserDataFieldIDByNameIntern("FIRST_NAME")
		if fne != nil {
			slog.Error("FIRST_NAME: " + fne.Error())
			db.Rollback()
			return 0, fne
		}
		err := database.SetUserDataField(db, ifn, user.ID, line.FirstName)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Surname) > 0 {
		if len(line.Surname) > 0 {
			iln, lne := db.GetUserDataFieldIDByNameIntern("LAST_NAME")
			if lne != nil {
				slog.Error("LAST_NAME: " + lne.Error())
				db.Rollback()
				return 0, lne
			}
			snError := database.SetUserDataField(db, iln, user.ID, line.Surname)
			if snError != nil {
				slog.Error(snError.Error())
				db.Rollback()
				return 0, snError
			}
		}
	}

	if len(line.Email) > 0 {
		// We use the email address as the account name in adm_users but that's our choice.
		// We also have a separate email entry in adm_user_data.
		ie, ee := db.GetUserDataFieldIDByNameIntern("EMAIL")
		if ee != nil {
			slog.Error("EMAIL: " + ee.Error())
			db.Rollback()
			return 0, ee
		}
		emailError := database.SetUserDataField(db, ie, user.ID, line.Email)
		if emailError != nil {
			slog.Error(emailError.Error())
			db.Rollback()
			return 0, emailError
		}
	}

	if len(line.AddressLine1) > 0 {
		ist, ste := db.GetUserDataFieldIDByNameIntern("STREET")
		if ste != nil {
			slog.Error("STREET: " + ste.Error())
			db.Rollback()
			return 0, ste
		}

		a1Error :=
			database.SetUserDataField(db, ist, user.ID, line.AddressLine1)
		if a1Error != nil {
			slog.Error(a1Error.Error())
			db.Rollback()
			return 0, a1Error
		}
	}

	if len(line.AddressLine2) > 0 {
		ial2, al2e := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_2")
		if al2e != nil {
			slog.Error("ADDRESS_LINE_2: " + al2e.Error())
			db.Rollback()
			return 0, al2e
		}
		err := database.SetUserDataField(db, ial2, user.ID, line.AddressLine2)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.AddressLine3) > 0 {
		ial3, al3e := db.GetUserDataFieldIDByNameIntern("ADDRESS_LINE_3")
		if al3e != nil {
			slog.Error("ADDRESS_LINE_3: " + al3e.Error())
			db.Rollback()
			return 0, al3e
		}
		err := database.SetUserDataField(db, ial3, user.ID, line.AddressLine3)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Town) > 0 {
		ic, ce := db.GetUserDataFieldIDByNameIntern("CITY")
		if ce != nil {
			slog.Error("CITY: " + ce.Error())
			db.Rollback()
			return 0, ce
		}
		err := database.SetUserDataField(db, ic, user.ID, line.Town)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.County) > 0 {
		ict, cte := db.GetUserDataFieldIDByNameIntern("COUNTY")
		if cte != nil {
			slog.Error("COUNTY: " + cte.Error())
			db.Rollback()
			return 0, cte
		}
		err := database.SetUserDataField(db, ict, user.ID, line.County)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Postcode) > 0 {
		ipc, pce := db.GetUserDataFieldIDByNameIntern("POSTCODE")
		if pce != nil {
			slog.Error("POSTCODE: " + pce.Error())
			db.Rollback()
			return 0, pce
		}
		err := database.SetUserDataField(db, ipc, user.ID, line.Postcode)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Country) > 0 {
		ictr, ctre := db.GetUserDataFieldIDByNameIntern("COUNTRY")
		if ctre != nil {
			slog.Error("COUNTRY: " + ctre.Error())
			db.Rollback()
			return 0, ctre
		}
		err := database.SetUserDataField(db, ictr, user.ID, line.Country)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Phone) > 0 {
		ip, fpe := db.GetUserDataFieldIDByNameIntern("PHONE")
		if fpe != nil {
			slog.Error("PHONE: " + fpe.Error())
			db.Rollback()
			return 0, fpe
		}
		err := database.SetUserDataField(db, ip, user.ID, line.Phone)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}

	if len(line.Mobile) > 0 {
		im, fme := db.GetUserDataFieldIDByNameIntern("MOBILE")
		if fme != nil {
			slog.Error("MOBILE: " + fme.Error())
			db.Rollback()
			return 0, fme
		}
		err := database.SetUserDataField(db, im, user.ID, line.Mobile)
		if err != nil {
			slog.Error(err.Error())
			db.Rollback()
			return 0, err
		}
	}
	return user.ID, nil
}
