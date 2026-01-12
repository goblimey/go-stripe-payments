package handler

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kylelemons/godebug/diff"
	"github.com/stripe/stripe-go/v81"

	"github.com/goblimey/go-tools/dailylogger"

	"github.com/goblimey/go-stripe-payments/code/pkg/config"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/forms"
)

// databaseList is a list of database types that will be used in
// integration tests.
var databaseList = []string{"postgres", "sqlite"}

type TestResponseWriter struct {
	Body io.ReadCloser
	W    io.Writer
	Code int
}

func (trw *TestResponseWriter) Header() http.Header {
	return http.Header{}
}

func (trw *TestResponseWriter) Write(data []byte) (int, error) {
	// trw.Body = io.NopCloser(bytes.NewReader(data))
	// return 0, fmt.Errorf("always errors")
	n, err := trw.W.Write(data)
	return n, err
}

func (trw *TestResponseWriter) WriteHeader(statusCode int) {
	trw.Code = statusCode
}

// NewTestReponeWriter takes the given writer and cteates and returns a
// TestResponseWriter.  If you need to see what is written, use a
// bytes.Buffer as the writer.
func NewTestResponseWriter(w io.Writer) *TestResponseWriter {
	trw := TestResponseWriter{W: w}
	return &trw
}

var testConfig = config.Config{
	OrganisationName:         "org",
	OrdinaryMemberFee:        24,
	EnableOtherMemberTypes:   true,
	EnableGiftaid:            true,
	EmailAddressForQuestions: "a@b.com",
	EmailAddressForFailures:  "c@d.com",
}

func TestSuccess(t *testing.T) {
	body := strings.NewReader(`{"username": "admin","password":"secret"}`)
	http.NewRequest("POST", "/success", body)
}

// TestSetMemberDetails checks the setMemberDetails handler method.  It's a
// very important part of the system because it runs after a succesful payment
// and creates the data that the user has paid for - setting or updating the
// membership end date and so on.
func TestSetMemberDetails(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := database.ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		// Create a structured logger that writes to the dailyLogWriter.
		dailyLogWriter := dailylogger.New("..", "test.", ".log")
		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
		db.Logger = logger

		h := New(&testConfig)
		h.DB = db
		h.Logger = logger

		membershipStart := time.Date(2024, time.January, 1, 0, 0, 0, 0, h.TZ)
		membershipEnd := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		now := time.Date(2024, time.October, 1, 0, 0, 0, 0, h.TZ)
		endDate := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		startDate := time.Date(2024, time.July, 31, 10, 0, 0, 0, h.TZ)

		roleMember, re := db.GetRole("Member")
		if re != nil {
			// This should never fail.
			t.Fatal(re)
		}

		// This should never fail so on any error, stop the test suite.
		u1LoginName, ue1 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ue1 != nil {
			t.Fatal(ue1)
		}
		u2LoginName, uie2 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if uie2 != nil {
			t.Fatal(uie2)
		}
		u3LoginName, ue3 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ue3 != nil {
			t.Fatal(ue3)
		}
		u4 := createTestUser(db, t)
		m1 := database.NewMember(u4, roleMember, membershipStart, membershipEnd)
		me1 := db.CreateMember(m1)
		if me1 != nil {
			t.Fatal(me1)
		}
		u5 := createTestUser(db, t)
		m2 := database.NewMember(u5, roleMember, membershipStart, membershipEnd)
		me2 := db.CreateMember(m2)
		if me2 != nil {
			t.Fatal(me2)
		}
		u6 := createTestUser(db, t)
		m3 := database.NewMember(u6, roleMember, membershipStart, membershipEnd)
		me3 := db.CreateMember(m3)
		if me3 != nil {
			t.Fatal(me3)
		}
		u7LoginName, ue7 := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ue7 != nil {
			t.Fatal(ue7)
		}

		var testData = []struct {
			description       string
			transactionType   string
			omUser            *database.User
			omTitle           string
			omFirstName       string
			omLastName        string
			omEmail           string
			omFriend          bool
			donationToSociety float64
			donationToMuseum  float64
			giftaid           bool
			assocUser         *database.User
			assocUserID       int64
			assocTitle        string
			assocFirstName    string
			assocLastName     string
			assocEmail        string
			assocFriend       bool

			wantLoginName      string
			wantAssocLoginName string
		}{
			{
				description:       "new - all",
				transactionType:   database.TransactionTypeNewMember,
				omTitle:           "a",
				omFirstName:       "b",
				omLastName:        "c",
				omEmail:           u1LoginName,
				omFriend:          true,
				donationToSociety: 1.1,
				donationToMuseum:  2.2,
				giftaid:           true,
				assocTitle:        "d",
				assocFirstName:    "e",
				assocLastName:     "f",
				assocEmail:        u2LoginName,
				assocFriend:       true,

				wantLoginName:      u1LoginName,
				wantAssocLoginName: u2LoginName,
			},
			{
				description:     "new - one",
				transactionType: database.TransactionTypeNewMember,
				omTitle:         "g",
				omFirstName:     "h",
				omLastName:      "i",
				omEmail:         u3LoginName,

				wantLoginName: u3LoginName,
			},
			{
				description:     "renewing - all",
				transactionType: database.TransactionTypeRenewal,
				omUser:          u4,
				omTitle:         "aa",
				omFirstName:     "bb",
				omLastName:      "cc",
				omEmail:         u4.LoginName,
				assocUser:       u5,
				assocTitle:      "dd",
				assocFirstName:  "ee",
				assocLastName:   "ff",
				assocEmail:      u5.LoginName,

				wantLoginName:      u4.LoginName,
				wantAssocLoginName: u5.LoginName,
			},
			{
				description:     "renewing - one",
				transactionType: database.TransactionTypeNewMember,
				omUser:          u6,
				omTitle:         "gg",
				omFirstName:     "hh",
				omLastName:      "ii",
				omEmail:         u6.LoginName,

				wantLoginName: u6.LoginName,
			},
			{
				description:       "new - assoc no email",
				transactionType:   database.TransactionTypeNewMember,
				omTitle:           "jj",
				omFirstName:       "kk",
				omLastName:        "ll",
				omEmail:           u7LoginName,
				omFriend:          true,
				donationToSociety: 1.1,
				donationToMuseum:  2.2,
				giftaid:           true,
				assocTitle:        "mm",
				assocFirstName:    "nn",
				assocLastName:     "oo",

				wantLoginName:      u7LoginName,
				wantAssocLoginName: "nn.oo",
			},
		}

		for _, td := range testData {

			var userID int64
			if td.omUser != nil {
				userID = td.omUser.ID
			}
			var assocUserID int64
			if td.assocUser != nil {
				assocUserID = td.assocUser.ID
			}

			ms := database.MembershipSale{
				PaymentStatus:     database.PaymentStatusPending,
				TransactionType:   td.transactionType,
				MembershipYear:    2025,
				UserID:            userID,
				Title:             td.omTitle,
				FirstName:         td.omFirstName,
				LastName:          td.omLastName,
				Email:             td.omEmail,
				Friend:            td.omFriend,
				DonationToSociety: td.donationToSociety,
				DonationToMuseum:  td.donationToMuseum,
				Giftaid:           td.giftaid,
				AssocUserID:       assocUserID,
				AssocTitle:        td.assocTitle,
				AssocFirstName:    td.assocFirstName,
				AssocLastName:     td.assocLastName,
				AssocEmail:        td.assocEmail,
				AssocFriend:       td.assocFriend,
			}

			_, se := ms.Create(db)
			if se != nil {
				t.Error(se)
				continue
			}

			var buf bytes.Buffer
			w := NewTestResponseWriter(&buf)

			// Test
			smdError := h.setMemberDetails(w, &ms, startDate, endDate, now, 2025)
			if smdError != nil {
				t.Error(smdError)
			}

			// Check that the success helper has updated the membership end dates.
			if td.transactionType == database.TransactionTypeNewMember {
				var err error
				td.omUser, err = db.GetUserByLoginName(td.omEmail)
				if err != nil {
					t.Errorf("%s: %v", dbType, err)
				}
			}

			fetchedM1, me1 := db.GetMemberOfUser(td.omUser)
			if me1 != nil {
				t.Error(me1)
				continue
			}

			// Get the associate and their member record.  If there is no associate,
			// td.wantAssocLoginName will be empty.
			var fetchedM2 *database.Member
			if len(td.wantAssocLoginName) > 0 {
				var ue error
				td.assocUser, ue = db.GetUserByLoginName(td.wantAssocLoginName)
				if ue != nil {
					t.Errorf("%s: %v", dbType, ue)
					continue
				}

				var me error
				fetchedM2, me = db.GetMemberOfUser(td.assocUser)
				if me != nil {
					t.Errorf("%s: %v", dbType, me)
					continue
				}
			}

			// Check that the membership end date has been set.  (Note that date
			// formats are different in Postgress from those in SQLite.)
			if dbType == "postgres" {
				if fetchedM1.EndDate != "2025-12-31T00:00:00Z" {
					t.Errorf("%s: expected end date 2025-12-31T00:00:00Z, got %s",
						dbType, fetchedM2.EndDate)
				}
				if fetchedM2 != nil && fetchedM2.EndDate != "2025-12-31T00:00:00Z" {
					t.Errorf("%s: expected end date 2025-12-31T00:00:00Z, got %s",
						dbType, fetchedM2.EndDate)
				}
			} else {
				if fetchedM1.EndDate != "2025-12-31 23:59:59 999999 +00" {
					t.Errorf("%s: expected end date 2025-12-31 23:59:59 999999 +00, got %s",
						dbType, fetchedM1.EndDate)
				}

				if fetchedM2 != nil && fetchedM2.EndDate != "2025-12-31 23:59:59 999999 +00" {
					t.Errorf("%s: expected end date 2025-12-31 23:59:59 999999 +00, got %s",
						dbType, fetchedM2.EndDate)
				}
			}

			// Check the fields in adm_user_data.

			ttl, ttle := db.GetTitle(td.omUser.ID)
			if ttle != nil {
				t.Errorf("%s %v", dbType, ttle)
			}

			if ttl != td.omTitle {
				t.Errorf("%s: want %s got %s", dbType, td.omTitle, ttl)
			}

			fn, fne := db.GetFirstName(td.omUser.ID)
			if fne != nil {
				t.Errorf("%s %v", dbType, fne)
			}

			if fn != td.omFirstName {
				t.Errorf("%s: want %s got %s", dbType, td.omFirstName, fn)
			}

			ln, lne := db.GetLastName(td.omUser.ID)
			if lne != nil {
				t.Errorf("%s %v", dbType, lne)
			}

			if ln != td.omLastName {
				t.Errorf("%s: want %s got %s", dbType, td.omLastName, ln)
			}

			em, eme := db.GetEmail(td.omUser.ID)
			if eme != nil {
				t.Errorf("%s %v", dbType, eme)
			}

			if em != td.omEmail {
				t.Errorf("%s: want %s got %s", dbType, td.omEmail, em)
			}

			// Only the ordinary member pays so only they should have Giftaid set.
			g, ge := db.GetGiftaid(td.omUser.ID)
			if ge != nil {
				t.Errorf("%s %v", dbType, ge)
			}

			if g != td.giftaid {
				t.Errorf("%s: want %v got %v", dbType, td.giftaid, g)
			}

			// If this is a new membership sale with an associate, we set the email
			// permission for for both users even if we don't have both email addresses -
			// if they give their addresses, it's safe to assume that they are prepared
			// to receive emails.
			//
			// On a renewal, we leave the permissions as they are because the members may
			// have revoked one or both of them.

			emp, empe := db.GetReceiveEmailField(td.omUser.ID)
			if empe != nil {
				t.Errorf("%s %v", dbType, empe)
			}

			if !emp {
				t.Errorf("%s: email permssion should be true", dbType)
			}

			if td.assocUser != nil {

				attl, attle := db.GetTitle(td.assocUser.ID)
				if attle != nil {
					t.Errorf("%s %v", dbType, attle)
				}

				if attl != td.assocTitle {
					t.Errorf("%s: want %s got %s", dbType, td.assocTitle, attl)
				}

				afn, afne := db.GetFirstName(td.assocUser.ID)
				if afne != nil {
					t.Errorf("%s %v", dbType, afne)
				}

				if afn != td.assocFirstName {
					t.Errorf("%s: want %s got %s", dbType, td.assocFirstName, afn)
				}

				aln, alne := db.GetLastName(td.assocUser.ID)
				if alne != nil {
					t.Errorf("%s %v", dbType, alne)
				}

				if aln != td.assocLastName {
					t.Errorf("%s: want %s got %s", dbType, aln, td.assocLastName)
				}

				aem, aeme := db.GetReceiveEmailField(td.assocUser.ID)
				if aeme != nil {
					t.Errorf("%s %v", dbType, aeme)
				}

				if !aem {
					t.Errorf("%s: email perm not set", dbType)
				}
			}
		}
	}
}

// TestCheckDonation checks the checkDonation validation method -
// a donation string value must contain a 0 or positive float.
func TestCheckDonation(t *testing.T) {
	var testData = []struct {
		str              string
		wantErrorMessage string
		wantValue        float64
	}{

		{"1.3", "", 1.3},
		{"0.1", "", 0.1},
		{"junk", invalidNumber, 0.0},
		{"-0.1", negativeNumber, 0.0},
		{"", "", 0.0},
	}

	for _, td := range testData {
		gotErrorMessage, gotValue := checkNonNegativeNumber(td.str)
		if td.wantValue != gotValue {
			t.Errorf("%s: want %f got %f", td.str, td.wantValue, gotValue)
		}
		if td.wantErrorMessage != gotErrorMessage {
			t.Errorf("%s: want %s got %s", td.str, td.wantErrorMessage, gotErrorMessage)
		}
	}
}

// TestGetTickBox checks the getTickBox function.
func TestGetTickBox(t *testing.T) {
	var testData = []struct {
		input            string
		wantStatus       bool
		wantUpdatedInput string
		wantOutput       string
	}{
		{"on", true, "on", "checked"},
		{"off", false, "off", "unchecked"},
		{"foo", false, "off", "unchecked"},
		{"", false, "off", "unchecked"},
	}

	for _, td := range testData {
		gotStatus, gotUpdatedInput, gotOutput := getTickBox(td.input)

		if td.wantStatus != gotStatus {
			t.Errorf("%s: want %v got %v", td.input, td.wantStatus, gotStatus)
		}

		if td.wantUpdatedInput != gotUpdatedInput {
			t.Errorf("%s: want %v got %v", td.input, td.wantUpdatedInput, gotUpdatedInput)
		}

		if td.wantOutput != gotOutput {
			t.Errorf("%s: want %v got %v", td.input, td.wantOutput, gotOutput)
		}
	}
}

// TestSimpleValidation checks the first stage of validation.
func TestSimpleValidation(t *testing.T) {

	var testData = []struct {
		description                string
		form                       forms.SaleForm
		wantValid                  bool
		wantForm                   forms.SaleForm
		wantOrdinaryMembershipFee  string
		wantAssociateMembershipFee string
		wantFriendFeePaid          string
		wantDonationToSociety      string
		wantDonationToMuseum       string
	}{

		{
			"valid - all",
			// This also checks the Trimspace calls.
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "	Mr  ", FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 78.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			true,

			forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "Mr", FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "7.83", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: "f", AssocLastName: "l",
				AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.83", "£8.90",
		},

		{
			"valid - not a friend, associate is a friend",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: " 1.5\t", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			true,
			forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true, AssocFriend: true,
				GiftaidOutput: "checked", AssocFriendOutput: "checked", FriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£24.00", "£3.40", "£5.60", "£1.50", "£2.50",
		},
		{
			"valid - no associate",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 1.5\t", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			true,
			forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: true, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true, AssocFriend: false,
				GiftaidOutput: "checked", AssocFriendOutput: "unchecked", FriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£24.00", "£6.00", "£5.00", "£1.50", "£2.50",
		},
		{
			"valid - ordinary member is not a friend, associate is a friend",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: " 1.5\t", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "d", AssocLastName: "e", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			true,
			forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "d", AssocLastName: "e", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true, AssocFriend: true,
				GiftaidOutput: "checked", AssocFriendOutput: "checked", FriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£24.00", "£6.00", "£5.00", "£1.50", "£2.50",
		},
		{
			description: "valid - associate friend tickbox is empty (associate is not a friend)",
			form: forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 1.5\t", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				Friend: true, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true,
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com",
				AssocFriendInput: "off",
				AssocFriend:      false, AssocFriendFeeToPay: 0,
				GiftaidOutput: "checked", AssocFriendOutput: "unchecked", FriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "£24.00",
			wantAssociateMembershipFee: "£6.00",
			wantFriendFeePaid:          "£5.00",
			wantDonationToSociety:      "£1.50",
			wantDonationToMuseum:       "£2.50",
		},
		{
			description: "valid - friend tick box on, others off",
			form: forms.SaleForm{
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "",
				Friend: false, DonationToSociety: 1.234, DonationToMuseum: 5.678, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:          true,
				MembershipYear: 2024, AssocFeeToPay: 0, FriendFeeToPay: 0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: false, GiftaidOutput: "unchecked", AssocFriend: false, AssocFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "",
			wantAssociateMembershipFee: "",
			wantFriendFeePaid:          "",
			wantDonationToSociety:      "£7.80",
			wantDonationToMuseum:       "£8.90",
		},
		{
			description: "valid - assoc friend tick box on, others off",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.23, AssocMemberFee: 3.46, FriendFee: 5.678,
				MembershipYear: 2024, AssocFeeToPay: 3.456, FriendFeeToPay: 5.678,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 1.23, AssocMemberFee: 3.46, FriendFee: 5.678,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				GiftaidInput:           "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9",
				Giftaid: false, GiftaidOutput: "unchecked", AssocFriend: true, AssocFriendOutput: "checked",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssocFriendInput: "on",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "£1.23",
			wantAssociateMembershipFee: "£3.46",
			wantFriendFeePaid:          "£5.68",
			wantDonationToSociety:      "£7.80",
			wantDonationToMuseum:       "£8.90",
		},
		{
			description: "valid - no associate, giftaid tick box on, others off",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.5, FriendFeeToPay: 5.7,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 0.0, FriendFeeToPay: 0.0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: false, FriendOutput: "unchecked",
				DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: false,
				AssocFriendOutput: "unchecked",
				UserID:            0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "",
				EmailErrorMessage:          "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "£1.20",
			wantAssociateMembershipFee: "£3.40",
			wantFriendFeePaid:          "£5.60",
			wantDonationToSociety:      "£7.80",
			wantDonationToMuseum:       "£8.90",
		},
		{
			description: "valid - member and associate, both friends, no giftaid",
			form: forms.SaleForm{
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5.0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocTitle: "Dr", AssocFirstName: "c", AssocLastName: "d",
				AssocEmail: "c@d.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocTitle: "Dr", AssocFirstName: "c", AssocLastName: "d",
				AssocEmail: "c@d.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: false, GiftaidOutput: "unchecked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},

			wantOrdinaryMembershipFee:  "£24.00",
			wantAssociateMembershipFee: "£6.00",
			wantFriendFeePaid:          "£5.00",
			wantDonationToSociety:      "£7.80",
			wantDonationToMuseum:       "£8.90",
		},
		{
			description: "invalid - ordinary member first name missing",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "99.989", DonationToMuseumInput: "11.1111", GiftaidInput: "",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: false,
			wantForm: forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "99.989", DonationToMuseumInput: "11.1111", GiftaidInput: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: true, FriendOutput: "checked", DonationToSociety: 99.989, DonationToMuseum: 11.1111,
				Giftaid: false, GiftaidOutput: "unchecked", AssocFriend: false, AssocFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: firstNameErrorMessage, LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},

			wantOrdinaryMembershipFee:  "£1.20",
			wantAssociateMembershipFee: "£3.40",
			wantFriendFeePaid:          "£5.60",
			wantDonationToSociety:      "£99.99",
			wantDonationToMuseum:       "£11.11",
		},

		{
			"ordinary member last name missing",
			forms.SaleForm{

				MembershipYear: 2024, OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: " a\t", LastName: "", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				MembershipYear: 2024, OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				FirstName: "a", LastName: "", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},
		{
			"ordinary member email missing",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: " a\t", LastName: "b", Email: "", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},
		{
			"associate member first name missing",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: " a\t", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},
		{
			"associate member last name missing",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},

		{
			"associate member but no ordinary member",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "", LastName: "", Email: "", FriendInput: "",
				DonationToSocietyInput: " 7.8\t", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "", LastName: "", Email: "", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: false, GiftaidOutput: "unchecked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: firstNameErrorMessage,
				LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},

		{
			"associate email address but associate member's name missing",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},
		{
			"associate friend tick box but associate member's name missing",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"£1.20", "£3.40", "£5.60", "£7.80", "£8.90",
		},

		{
			description: "donation to society invalid number",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "junk", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: false,
			wantForm: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "junk", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: true, FriendOutput: "checked", DonationToSociety: 0.0, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: false, AssocFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
				DonationToSocietyErrorMessage: invalidNumber,
			},
			wantOrdinaryMembershipFee:  "£1.20",
			wantAssociateMembershipFee: "£3.40",
			wantFriendFeePaid:          "£5.60",
			wantDonationToSociety:      "",
			wantDonationToMuseum:       "£8.90",
		},
		{
			"donation to museum invalid number",
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "6.7", DonationToMuseumInput: "junk", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				FirstName:      "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "6.7", DonationToMuseumInput: "junk", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: true, FriendOutput: "checked", DonationToSociety: 6.7, DonationToMuseum: 0.0,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: false, AssocFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
				DonationToMuseumErrorMessage: invalidNumber,
			},
			"£1.20", "£3.40", "£5.60", "£6.70", "",
		},
		{
			"invalid - negative donation to society",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				Title: "Mr ", FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " -7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			false,

			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "Mr", FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "-7.83", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 0, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
				DonationToSocietyErrorMessage: negativeNumber,
			},
			"£1.20", "£3.40", "£5.60", "", "£8.90",
		},
		{
			"invalid - donation to museum is negative",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				Title: "Mr", FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " -8.9 ", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
				DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			false,

			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "Mr", FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "7.83", DonationToMuseumInput: "-8.9", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: "f", AssocLastName: "l",
				AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 0, Giftaid: true,
				AssocFriend: true,
				UserID:      0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
				DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: negativeNumber,
			},
			"£1.20", "£3.40", "£5.60", "£7.83", "",
		},
		{
			"invalid - title but no name",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 0.0, FriendFeeToPay: 5.6,
				Title: "Mr", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: " f ", AssocLastName: " l ",
				AssocEmail:       "  a@l.com  ",
				AssocFriendInput: "on",
				Friend:           true, DonationToSociety: 7.83, DonationToMuseum: 78.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			false,

			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "Mr", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "7.83", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com",
				AssocFriendInput: "on",
				Friend:           true, DonationToSociety: 7.83, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: firstNameErrorMessage, LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.83", "£8.90",
		},
		{
			"invalid - associate title but no name",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				Title: "Mr ", FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 78.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			false,

			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024,
				Title:          "Mr", FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "7.83", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocTitle: "Lord High Admiral", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 8.9, Giftaid: true,
				AssocFriend: true,
				UserID:      0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
			"£1.20", "£3.40", "£5.60", "£7.83", "£8.90",
		},
	}

	for _, td := range testData {
		valid := ValidateSaleForm(&td.form)

		// The tick box outputs have been set so the got and want forms are no longer equal.
		// Check the values and then unset them.

		if td.wantForm.FriendOutput != td.form.FriendOutput {
			t.Errorf("%s: want %s got %s", td.description, td.wantForm.FriendOutput, td.form.FriendOutput)
		}

		if td.wantForm.AssocFriendOutput != td.form.AssocFriendOutput {
			t.Errorf("%s: want %s got %s",
				td.description, td.wantForm.AssocFriendOutput, td.form.AssocFriendOutput)
		}

		if td.wantForm.GiftaidOutput != td.form.GiftaidOutput {
			t.Errorf("%s: want %s got %s", td.description, td.wantForm.GiftaidOutput, td.form.GiftaidOutput)
		}

		if td.wantValid != valid {
			t.Errorf("%s want %v got %v", td.description, td.wantValid, valid)
		}

		if td.wantForm != td.form {
			t.Errorf("%s:\nwant %v\n got %v\n", td.description, td.wantForm, td.form)
		}

		if td.wantOrdinaryMembershipFee != td.form.OrdinaryMemberFeeForDisplay() {
			t.Errorf("%s:\nordinary fee - want %v\n got %v\n", td.description, td.wantOrdinaryMembershipFee, td.form.OrdinaryMemberFeeForDisplay())
		}

		if td.wantAssociateMembershipFee != td.form.AssocFeeForDisplay() {
			t.Errorf("%s:\nassoc fee - want %v\n got %v\n", td.description, td.wantAssociateMembershipFee, td.form.AssocFeeForDisplay())
		}

		if td.wantFriendFeePaid != td.form.FriendFeeForDisplay() {
			t.Errorf("%s:\nfriend fee - want %v\n got %v\n", td.description, td.wantFriendFeePaid, td.form.FriendFeeForDisplay())
		}

		if td.wantDonationToSociety != td.form.DonationToSocietyForDisplay() {
			t.Errorf("%s:\ndonation to society - want %v\n got %v\n", td.description, td.wantDonationToSociety, td.form.DonationToSocietyForDisplay())
		}

		if td.wantDonationToMuseum != td.form.DonationToMuseumForDisplay() {
			t.Errorf("%s:\ndonation to museum - want %v\n got %v\n", td.description, td.wantDonationToMuseum, td.form.DonationToMuseumForDisplay())
		}
	}
}

func TestUsersExistWithAssociate(t *testing.T) {
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

		oLN, ole := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ole != nil {
			t.Error(ole)
		}

		oFN, ofe := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ofe != nil {
			t.Error(ofe)
		}

		oEmail, oee := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if oee != nil {
			t.Error(oee)
		}

		assocFN, afe := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if afe != nil {
			t.Error(afe)
		}

		assocLN, ale := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ale != nil {
			t.Error(ale)
		}

		assocEmail, aee := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if aee != nil {
			t.Error(aee)
		}

		sale := database.MembershipSale{
			OrdinaryMemberFeePaid: 1.2, AssocFeePaid: 3.4, FriendFeePaid: 5.6, MembershipYear: 2024,
			FirstName: oFN, LastName: oLN, Email: oEmail,
			AssocFirstName: assocFN, AssocLastName: assocLN, AssocEmail: assocEmail,
			UserID: 0, AssocUserID: 0,
		}

		// Create an ordinary user and an associate.
		oID, aID, createError := db.CreateAccounts(&sale, time.Now(), time.Now())
		if createError != nil {
			t.Error(createError)
			return
		}

		// Test.
		fetchedOID, fetchedAID, lookupError := usersExist(&sale, db)
		if lookupError != nil {
			t.Error(lookupError)
			return
		}

		// Check - usersExist() should give back the userIDs of the two users.

		if fetchedOID != oID {
			t.Errorf("want ID %d got %d", oID, fetchedOID)
		}
		if fetchedAID != aID {
			t.Errorf("want ID %d got %d", aID, fetchedAID)

		}

		db.Rollback()
	}
}

// TestUsersExistWhenAssociateHasNoEmailAddress checks UsersExist when there is an
// associate with no email address.  (The function must use the first and last name.)
func TestUsersExistWhenAssociateHasNoEmailAddress(t *testing.T) {
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

		oLN, ole := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ole != nil {
			t.Error(ole)
		}

		oFN, ofe := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ofe != nil {
			t.Error(ofe)
		}

		oEmail, oee := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if oee != nil {
			t.Error(oee)
		}

		assocFN, afe := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if afe != nil {
			t.Error(afe)
		}

		assocLN, ale := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ale != nil {
			t.Error(ale)
		}

		wantAssocLoginName := assocFN + "." + assocLN

		sale := database.MembershipSale{
			OrdinaryMemberFeePaid: 1.2, AssocFeePaid: 3.4, FriendFeePaid: 5.6, MembershipYear: 2024,
			FirstName: oFN, LastName: oLN, Email: oEmail,
			AssocFirstName: assocFN, AssocLastName: assocLN,
			UserID: 0, AssocUserID: 0,
		}

		// Create an ordinary user and an associate.  Thevassociate has no email address
		// so the account name will be "firstname.lastname" and there will be first name
		// and last name fields in adm_user_data.
		oID, aID, createError := db.CreateAccounts(&sale, time.Now(), time.Now())
		if createError != nil {
			t.Error(createError)
			return
		}

		// The test - the associate user has no email address but usersExist() can still
		// find it using the first name and the last name.
		fetchedOID, fetchedAID, lookupError := usersExist(&sale, db)
		if lookupError != nil {
			t.Error(lookupError)
			return
		}

		if fetchedOID != oID {
			t.Errorf("want ID %d got %d", oID, fetchedOID)
			return
		}
		if fetchedAID != aID {
			t.Errorf("want ID %d got %d", aID, fetchedAID)
			return
		}

		assoc, fetchAssocError := db.GetUser(aID)
		if fetchAssocError != nil {
			t.Error(fetchAssocError)
			return
		}

		if assoc.LoginName != wantAssocLoginName {
			t.Errorf("want %s got %s", wantAssocLoginName, assoc.LoginName)
			return
		}
		db.Rollback()
	}
}

func TestUsersExistWithNoAssociate(t *testing.T) {
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

		oLN, ole := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ole != nil {
			t.Error(ole)
		}

		oFN, ofe := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if ofe != nil {
			t.Error(ofe)
		}

		oEmail, oee := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
		if oee != nil {
			t.Error(oee)
		}

		sale := database.MembershipSale{
			OrdinaryMemberFeePaid: 1.2, AssocFeePaid: 3.4, FriendFeePaid: 5.6, MembershipYear: 2024,
			FirstName: oFN, LastName: oLN, Email: oEmail,
			UserID: 0, AssocUserID: 0,
		}

		// Create users.
		oID, aID, createError := db.CreateAccounts(&sale, time.Now(), time.Now())
		if createError != nil {
			t.Error(createError)
			return
		}

		if aID != 0 {
			t.Errorf("want ID 0 (no associate) got %d", aID)
			return
		}

		// The test - usersExist() should give back the userIDs of the ordinary user
		// but no associate.
		fetchedOID, fetchedAID, lookupError := usersExist(&sale, db)
		if lookupError != nil {
			t.Error(lookupError)
			return
		}

		if fetchedOID != oID {
			t.Errorf("want ID %d got %d", oID, fetchedOID)
			return
		}
		if fetchedAID != 0 {
			t.Errorf("want ID 0 (no associate) got %d", fetchedAID)
			return
		}

		db.Rollback()
	}
}

func TestGetMembershipSaleOnSuccess(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := database.ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		// Create a structured logger that writes to the dailyLogWriter.
		dailyLogWriter := dailylogger.New("..", "test.", ".log")
		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
		db.Logger = logger

		h := New(&testConfig)
		h.DB = db
		h.Logger = logger

		roleMember, re := db.GetRole("Member")
		if re != nil {
			// This should never fail.
			t.Fatal(re)
		}

		u1 := createTestUser(db, t)
		membershipStart := time.Date(2024, time.January, 1, 0, 0, 0, 0, h.TZ)
		membershipEnd := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		m1 := database.NewMember(u1, roleMember, membershipStart, membershipEnd)
		db.CreateMember(m1)
		u2 := createTestUser(db, t)
		m2 := database.NewMember(u2, roleMember, membershipStart, membershipEnd)
		db.CreateMember(m2)

		ms := database.MembershipSale{
			PaymentStatus:  database.PaymentStatusPending,
			MembershipYear: 2025,
			UserID:         u1.ID,
			Title:          "a",
			FirstName:      "b",
			LastName:       "c",
			Email:          u1.LoginName,
			Friend:         true,  // One friend at this address.
			AssocUserID:    u2.ID, // Two members at this address.
			AssocTitle:     "d",
			AssocFirstName: "e",
			AssocLastName:  "f",
			AssocEmail:     u2.LoginName,
			AssocFriend:    false,
			CountryCode:    "ABW",
		}

		id, se := ms.Create(db)
		if se != nil {
			t.Error(se)
			continue
		}

		customer := stripe.Customer{ID: "foo", Email: "a@b.com"}
		session := stripe.CheckoutSession{
			PaymentStatus:     "paid",
			ClientReferenceID: fmt.Sprintf("%d", id),
			Customer:          &customer,
		}

		var buf bytes.Buffer
		w := NewTestResponseWriter(&buf)

		now := time.Date(2024, time.October, 1, 0, 0, 0, 0, h.TZ)
		endDate := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		startDate := time.Date(2024, time.July, 31, 10, 0, 0, 0, h.TZ)

		fetchedMS, msError := h.getMembershipSaleOnSuccess(
			w, &session, startDate, endDate, now, 2025)
		if msError != nil {
			t.Error(msError)
			continue
		}

		if fetchedMS.ID != id {
			t.Errorf("want %d got %d", id, fetchedMS.ID)
		}
	}
}

// TestSetAccountingRecordsForMembers checks setAccountingRecordsForMembers.
func TestSetAccountingRecordsForMembers(t *testing.T) {

	for _, dbType := range databaseList {

		db, connError := database.ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		// Create a structured logger that writes to the dailyLogWriter.
		dailyLogWriter := dailylogger.New("..", "test.", ".log")
		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
		db.Logger = logger

		h := New(&testConfig)
		h.DB = db
		h.Logger = logger

		roleMember, re := db.GetRole("Member")
		if re != nil {
			// This should never fail.
			t.Fatal(re)
		}

		u := createTestUser(db, t)
		membershipStart := time.Date(2024, time.January, 1, 0, 0, 0, 0, h.TZ)
		membershipEnd := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		paymentDate := time.Date(2024, time.February, 14, 12, 0, 0, 0, h.TZ)

		m1 := database.NewMember(u, roleMember, membershipStart, membershipEnd)
		db.CreateMember(m1)

		assocU := createTestUser(db, t)
		m2 := database.NewMember(assocU, roleMember, membershipStart, membershipEnd)
		db.CreateMember(m2)

		ms := database.MembershipSale{
			PaymentStatus:     database.PaymentStatusPending,
			MembershipYear:    2025,
			UserID:            u.ID,
			Title:             "a",
			FirstName:         "b",
			LastName:          "c",
			Email:             u.LoginName,
			CountryCode:       "ABW",
			Friend:            true, // One friend at this address.
			Giftaid:           true,
			DonationToSociety: 1.1,
			DonationToMuseum:  2.2,
			AssocUserID:       assocU.ID, // Two members at this address.
			AssocTitle:        "d",
			AssocFirstName:    "e",
			AssocLastName:     "f",
			AssocEmail:        assocU.LoginName,
			AssocFriend:       false,
		}

		_, se := ms.Create(db)
		if se != nil {
			t.Error(se)
			continue
		}

		var buf bytes.Buffer
		w := NewTestResponseWriter(&buf)

		now := time.Date(2024, time.October, 1, 0, 0, 0, 0, h.TZ)
		const wantDateLastPaid = "2024-02-14"
		endDate := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, h.TZ)
		startDate := time.Date(2024, time.July, 31, 10, 0, 0, 0, h.TZ)

		smdError := h.setMemberDetails(w, &ms, startDate, endDate, now, 2025)
		if smdError != nil {
			t.Error(smdError)
		}

		// Test
		h.setAccountingRecordsForMembers(&ms, paymentDate)

		// Check.

		// The ordinary member paid, the assocaite member didn't, so only the
		// ordinary member has this feld set.
		dlp, dlpError := db.GetDateLastPaid(u.ID)
		if dlpError != nil {
			t.Error(dbType + ": " + dlpError.Error())
		}
		if dlp != wantDateLastPaid {
			t.Errorf("%s: expected end date %s, got %s",
				dbType, wantDateLastPaid, dlp)
		}

		maa1, maae1 := db.GetMembersAtAddress(u.ID)
		if maae1 != nil {
			t.Error(maae1)
			continue
		}

		if maa1 != 2 {
			t.Errorf("want 2 got %d", maa1)
		}

		maa2, maae2 := db.GetMembersAtAddress(assocU.ID)
		if maae2 != nil {
			t.Error(maae2)
			continue
		}

		if maa2 != 2 {
			t.Errorf("%s: want 2 got %d", dbType, maa2)
		}

		faa1, faae1 := db.GetFriendsAtAddress(u.ID)
		if faae1 != nil {
			t.Error(faae1)
			continue
		}

		if faa1 != 1 {
			t.Errorf("want 1 got %d", faa1)
		}

		faa2, faae2 := db.GetFriendsAtAddress(assocU.ID)
		if faae2 != nil {
			t.Error(faae2)
			continue
		}

		if faa2 != 1 {
			t.Errorf("%s want 1 got %d", dbType, faa2)
		}

		giftaid, gaError := db.GetGiftaid(u.ID)
		if gaError != nil {
			t.Error(dbType + " " + gaError.Error())
		}

		if !giftaid {
			t.Errorf("%s: want giftaid tickbox set", dbType)
		}

		dts, dtsError := db.GetDonationToMuseum(u.ID)
		if dtsError != nil {
			t.Error(dbType + " " + dtsError.Error())
		}

		if dts != 2.2 {
			t.Errorf("%s: want 2.2 got %f", dbType, dts)
		}

		dtm, dtmError := db.GetDonationToMuseum(u.ID)
		if dtmError != nil {
			t.Error(dtmError)
			continue
		}

		if dtm != 2.2 {
			t.Errorf("%s: want 2.2 got %f", dbType, dtm)
		}
	}
}

// func TestSuccessHelper(t *testing.T) {

// 	london, le := time.LoadLocation("Europe/London")
// 	if le != nil {
// 		t.Fatal(le)
// 	}

// 	for _, dbType := range databaseList {

// 		db, connError := database.ConnectForTesting(dbType)

// 		if connError != nil {
// 			t.Error(connError)
// 			return
// 		}

// 		defer db.Rollback()
// 		defer db.CloseAndDelete()

// 		roleMember, re := db.GetRole("Member")
// 		if re != nil {
// 			// This should never fail.
// 			t.Fatal(re)
// 		}

// 		u1 := createTestUser(db, t)
// 		membershipStart := time.Date(2024, time.January, 1, 0, 0, 0, 0, london)
// 		membershipEnd := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, london)
// 		m1 := database.NewMember(u1, roleMember, membershipStart, membershipEnd)
// 		db.CreateMember(m1)
// 		u2 := createTestUser(db, t)
// 		m2 := database.NewMember(u2, roleMember, membershipStart, membershipEnd)
// 		db.CreateMember(m2)

// 		ms := database.MembershipSale{
// 			PaymentStatus:  database.PaymentStatusPending,
// 			MembershipYear: 2025,
// 			UserID:         u1.ID,
// 			Title:          "a",
// 			FirstName:      "b",
// 			LastName:       "c",
// 			Email:          u1.LoginName,
// 			Friend:         true,  // One friend at this address.
// 			AssocUserID:    u2.ID, // Two members at this address.
// 			AssocTitle:     "d",
// 			AssocFirstName: "e",
// 			AssocLastName:  "f",
// 			AssocEmail:     u2.LoginName,
// 			AssocFriend:    false,
// 			CountryCode:    "ABW",
// 		}

// 		id, se := ms.Create(db)
// 		if se != nil {
// 			t.Error(se)
// 			continue
// 		}

// 		customer := stripe.Customer{ID: "foo", Email: "a@b.com"}
// 		session := stripe.CheckoutSession{
// 			PaymentStatus:     "paid",
// 			ClientReferenceID: fmt.Sprintf("%d", id),
// 			Customer:          &customer,
// 		}

// 		var buf bytes.Buffer
// 		w := NewTestResponseWriter(&buf)

// 		dailyLogWriter := dailylogger.New("..", "test.", ".log")

// 		// Create a structured logger that writes to the dailyLogWriter.
// 		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
// 		db.Logger = logger
// 		h := Handler{DB: db, Logger: logger, Conf: &testConfig}

// 		now := time.Date(2024, time.October, 1, 0, 0, 0, 0, london)
// 		endDate := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, london)
// 		h.successHelper(w, &session, now, endDate, now, 2024)

// 		// Check that the success helper has updated the membership end dates.
// 		fetchedM1, me1 := db.GetMemberOfUser(u1)
// 		if me1 != nil {
// 			t.Error(me1)
// 			continue
// 		}
// 		fetchedM2, me2 := db.GetMemberOfUser(u2)
// 		if me2 != nil {
// 			t.Error(me2)
// 			continue
// 		}

// 		// date formats are different in Postgress from those in SQLite.
// 		if dbType == "postgres" {
// 			if fetchedM1.EndDate != "2025-12-31T00:00:00Z" {
// 				t.Errorf("%s: expected end date 2025-12-31T00:00:00Z, got %s",
// 					dbType, fetchedM2.EndDate)
// 			}
// 			if fetchedM2.EndDate != "2025-12-31T00:00:00Z" {
// 				t.Errorf("%s: expected end date 2025-12-31T00:00:00Z, got %s",
// 					dbType, fetchedM2.EndDate)
// 			}
// 		} else {
// 			if fetchedM1.EndDate != "2025-12-31 23:59:59 999999 +00" {
// 				t.Errorf("%s: expected end date 2025-12-31 23:59:59 999999 +00, got %s",
// 					dbType, fetchedM1.EndDate)
// 			}

// 			if fetchedM2.EndDate != "2025-12-31 23:59:59 999999 +00" {
// 				t.Errorf("%s: expected end date 2025-12-31 23:59:59 999999 +00, got %s",
// 					dbType, fetchedM2.EndDate)
// 			}
// 		}

// 		maa1, maae1 := db.GetMembersAtAddress(u1.ID)
// 		if maae1 != nil {
// 			t.Error(maae1)
// 			continue
// 		}

// 		if maa1 != 2 {
// 			t.Errorf("want 2 got %d", maa1)
// 		}

// 		maa2, maae2 := db.GetMembersAtAddress(u2.ID)
// 		if maae2 != nil {
// 			t.Error(maae2)
// 			continue
// 		}

// 		if maa2 != 2 {
// 			t.Errorf("want 2 got %d", maa2)
// 		}

// 		faa1, faae1 := db.GetFriendsAtAddress(u1.ID)
// 		if faae1 != nil {
// 			t.Error(faae1)
// 			continue
// 		}

// 		if faa1 != 1 {
// 			t.Errorf("want 1 got %d", faa1)
// 		}

// 		faa2, faae2 := db.GetFriendsAtAddress(u2.ID)
// 		if faae2 != nil {
// 			t.Error(faae2)
// 			continue
// 		}

// 		if faa2 != 1 {
// 			t.Errorf("want 1 got %d", faa2)
// 		}
// 	}
// }

// TestExtraDetailsHelper tests the extra details helper.
func TestExtraDetailsHelper(t *testing.T) {
	for _, dbType := range databaseList {

		db, connError := database.ConnectForTesting(dbType)

		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		u1 := createTestUser(db, t)
		u2 := createTestUser(db, t)

		// Get the interests.  If there are any, assign the first one to u1.
		interests, ie := db.GetInterests()
		if ie == nil {
			if len(interests) > 0 {
				interest := database.NewMembersInterest(u1.ID, interests[0].ID)
				db.CreateMembersInterest(interest)
			}
		}

		values := make(url.Values, 0)
		values.Add("title", "Mr")
		values.Add("account_name", u1.LoginName)
		values.Add("assoc_account_name", u2.LoginName)
		values.Add("phone", "01")
		values.Add("mobile", "+44 1")
		values.Add("address_line_2", "Flat 3")
		values.Add("address_line_3", "1 High Street")
		values.Add("town", "Leatherhead")
		values.Add("postcode", "A11 1AA")
		values.Add("county", "Rutshire")
		values.Add("country_code", "GBR")
		values.Add("location_of_interest", "pachesam")
		values.Add("interest", "1")
		values.Add("interest", "2")
		values.Add("other_topics_of_interest", "barfoobar,barfoo")
		values.Add("assoc_mobile", "02")

		r := http.Request{PostForm: values}

		var buf bytes.Buffer
		w := NewTestResponseWriter(&buf)

		dailyLogWriter := dailylogger.New("..", "test.", ".log")

		// Create a structured logger that writes to the dailyLogWriter.
		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
		db.Logger = logger
		h := New(&testConfig)
		h.DB = db
		h.Logger = logger

		// Run the test.
		edhe := h.ExtraDetailsHelper(w, &r, 2024, time.Now())
		if edhe != nil {
			t.Error(edhe)
			continue
		}

		// Check the results

		// ordinary user's address line 1.
		u1a1, u1a1e := db.GetAddressLine1(u1.ID)
		if u1a1e != nil {
			t.Errorf("%s: %v", dbType, u1a1e)
			continue
		}

		al2 := values.Get("address_line_2")
		// In the request parameter "address_line_1" is not set so "address_line_2" is used.
		if u1a1 != al2 {
			t.Errorf("want %s got %s", al2, u1a1)
		}

		// Associate user's address line 1.
		u2a1, u2a1e := db.GetAddressLine1(u2.ID)
		if u2a1e != nil {
			t.Error(u2a1e)
		}

		if u2a1 != al2 {
			t.Errorf("want %s got %s", al2, u2a1)
		}

		// Ordinary user's address line 2.
		u1a2, u1a2e := db.GetAddressLine2(u1.ID)
		if u1a2e != nil {
			t.Error(u1a2e)
		}

		if u1a2 != values.Get("address_line_3") {
			t.Errorf("want %s got %s", values.Get("address_line_3"), u1a2)
		}

		// Associate uesr's address line 2.
		u2a2, u2a2e := db.GetAddressLine2(u2.ID)
		if u2a2e != nil {
			t.Error(u2a2e)
		}

		if u2a2 != values.Get("address_line_3") {
			t.Errorf("want %s got %s", values.Get("town"), u2a2)
		}

		// Ordinary user's town.
		u1t, u1te := db.GetTown(u1.ID)
		if u1te != nil {
			t.Error(u1te)
		}

		if u1t != values.Get("town") {
			t.Errorf("want %s got %s", values.Get("town"), u1t)
		}

		// Associate users town.
		u2t, u2te := db.GetTown(u2.ID)
		if u2te != nil {
			t.Error(u2te)
		}

		if u2t != values.Get("town") {
			t.Errorf("want %s got %s", values.Get("town"), u2t)
		}

		// Ordinary user's county.
		u1ct, u1cte := db.GetCounty(u1.ID)
		if u1cte != nil {
			t.Error(u1cte)
		}

		if u1ct != values.Get("county") {
			t.Errorf("want %s got %s", values.Get("county"), u1ct)
		}

		// Associate uset's county.
		u2ct, u2cte := db.GetCountry(u2.ID)
		if u2cte != nil {
			t.Error(u2cte)
		}

		if u2ct != values.Get("country_code") {
			t.Errorf("want %s got %s", values.Get("country_code"), u2ct)
		}

		// Ordinary user's country code.
		u1cnt, u1cnte := db.GetCountryCode(u1.ID)
		if u1cnte != nil {
			t.Error(u1cnte)
		}

		if u1cnt != values.Get("country_code") {
			t.Errorf("want %s got %s", values.Get("country_code"), u1cnt)
		}

		// Associate user's country_code.
		u2cnt, u2pc1e := db.GetCountryCode(u2.ID)
		if u2pc1e != nil {
			t.Error(u2pc1e)
		}

		if u2cnt != values.Get("country_code") {
			t.Errorf("want %s got %s", values.Get("country_code"), u2cnt)
		}

		// Ordinary user's postcode.
		u1pc1, u1pc1e := db.GetPostcode(u1.ID)
		if u1pc1e != nil {
			t.Error(u1pc1e)
		}

		if u1pc1 != values.Get("postcode") {
			t.Errorf("want %s got %s", values.Get("postcode"), u1pc1)
		}

		// Associate user's postcode.
		u2pc1, u2pc1e := db.GetPostcode(u2.ID)
		if u2pc1e != nil {
			t.Error(u2pc1e)
		}

		if u2pc1 != values.Get("postcode") {
			t.Errorf("want %s got %s", values.Get("postcode"), u2pc1)
		}

		// Ordinary user's phone number.
		ph, phe := db.GetPhone(u1.ID)
		if phe != nil {
			t.Error(phe)
		}

		if ph != values.Get("phone") {
			t.Errorf("want %s got %s", values.Get("phone"), ph)
		}

		// Associate user's phone number.  (Should be the same as the user's number.)
		aph, aphe := db.GetPhone(u1.ID)
		if aphe != nil {
			t.Error(phe)
		}

		if aph != values.Get("phone") {
			t.Errorf("want %s got %s", values.Get("assoc_phone"), aph)
		}

		// Ordinary user's mobile number.
		m, me := db.GetMobile(u1.ID)
		if me != nil {
			t.Error(me)
		}

		wantMob := values.Get("mobile")
		if m != wantMob {
			t.Errorf("want %s got %s", values.Get("mobile"), m)
		}

		// The associate's mobile number.
		am, ame := db.GetMobile(u2.ID)
		if ame != nil {
			t.Error(ame)
		}

		wantAmob := values.Get("assoc_mobile")
		if am != wantAmob {
			t.Errorf("want %s got %s", wantAmob, am)
		}

		// Ordinary user's other interests.
		moi, oie := db.GetMembersOtherInterests(u1.ID)
		if oie != nil {
			t.Error(oie)
		}

		if moi.UserID != u1.ID {
			t.Errorf("want %d, got %d", u1.ID, moi.UserID)
			break
		}

		if moi.Interests != values.Get("other_topics_of_interest") {
			t.Errorf("want %s got %s", values.Get("other_topics_of_interest"), u1pc1)
			break
		}

		// Ordinary user's selected interests.  The adm_interests table contains
		// the interests - {id, name}.  The supplied interest values should be ids
		// from that table.
		iList, ie := db.GetInterests()
		if ie != nil {
			t.Error(ie)
			break
		}
		interestList := make(map[int64]database.Interest)

		for _, interest := range iList {
			interestList[interest.ID] = interest
		}
		nt, nError := db.GetMembersInterests(u1.ID)
		if nError != nil {
			t.Error(nError)
		}

		if len(nt) != 2 {
			t.Errorf("want 2 interests, got %d", len(nt))
			// don't break just yet.
		}

		// Check the userIDs.
		if nt[0].UserID != u1.ID {
			t.Errorf("want %d, got %d", u1.ID, nt[0].UserID)
		}
		if nt[1].UserID != u1.ID {
			t.Errorf("want %d, got %d", u1.ID, nt[1].UserID)
		}
		// Check the interestIDs

		for _, idStr := range values["interest"] {
			var id int64
			n, err := fmt.Sscanf(idStr, "%d", &id)
			if err != nil {
				t.Error(err)
			}
			if n != 1 {
				t.Errorf("cannot convert %s to integer", idStr)
				break
			}
			_, ok := interestList[id]
			if !ok {
				t.Errorf("interest %d not in list", id)
			}
		}
	}
}

// TestFetchCurrentExtraDetails checks fetchCurrentExtraDetails.
func TestFetchCurrentExtraDetails(t *testing.T) {
	for _, dbType := range databaseList {
		db, connError := database.ConnectForTesting(dbType)
		if connError != nil {
			t.Error(connError)
			return
		}

		defer db.Rollback()
		defer db.CloseAndDelete()

		u1 := createTestUser(db, t)
		u2 := createTestUser(db, t)

		ms := database.MembershipSale{UserID: u1.ID, AssocUserID: u2.ID}

		// Get the interests.  If there are any, assign the ones with ID 0 and 2 to u1.
		interests, ie := db.GetInterests()
		if ie == nil {
			if len(interests) > 0 {
				interest1 := database.NewMembersInterest(u1.ID, interests[0].ID)
				db.CreateMembersInterest(interest1)
				interest2 := database.NewMembersInterest(u1.ID, interests[2].ID)
				db.CreateMembersInterest(interest2)
			}
		}

		db.SetAddressLine1(u1.ID, "a1")
		db.SetAddressLine1(u2.ID, "a1")

		db.SetAddressLine2(u1.ID, "a2")
		db.SetAddressLine2(u2.ID, "a2")

		db.SetAddressLine3(u1.ID, "a3")
		db.SetAddressLine1(u2.ID, "a3")

		db.SetTown(u1.ID, "t")
		db.SetTown(u2.ID, "t")

		db.SetPostcode(u1.ID, "pc")
		db.SetPostcode(u2.ID, "pc")

		db.SetCounty(u1.ID, "cty")
		db.SetCounty(u2.ID, "cty")

		db.SetCountryCode(u1.ID, "GBR")
		db.SetCountryCode(u2.ID, "GBR")

		db.SetPhone(u1.ID, "+44 1")
		db.SetPhone(u2.ID, "+44 1")

		db.SetMobile(u1.ID, "+44 2")

		db.SetMobile(u2.ID, "+44 3")

		db.SetLocationOfInterest(u1.ID, "Ashtead")

		moi := database.NewMembersOtherInterests(u1.ID, "foobar")

		db.CreateMembersOtherInterests(moi)

		// Create a structured logger that writes to the dailyLogWriter.
		dailyLogWriter := dailylogger.New("..", "test.", ".log")
		logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
		db.Logger = logger
		h := Handler{DB: db, Logger: logger, Conf: &testConfig}

		// Test.
		h.fetchCurrentExtraDetails(&ms)

		// Check.
		if ms.AddressLine1 != "a1" {
			t.Errorf("%s - want a1 got %s", dbType, ms.AddressLine1)
		}

		if ms.AddressLine2 != "a2" {
			t.Errorf("%s - want a2 got %s", dbType, ms.AddressLine2)
		}

		if ms.AddressLine3 != "a3" {
			t.Errorf("%s - want a3 got %s", dbType, ms.AddressLine3)
		}

		if ms.Town != "t" {
			t.Errorf("%s - want t got %s", dbType, ms.Town)
		}

		if ms.County != "cty" {
			t.Errorf("%s - want cty got %s", dbType, ms.County)
		}

		if ms.Postcode != "pc" {
			t.Errorf("%s - want pc got %s", dbType, ms.County)
		}

		if ms.CountryCode != "GBR" {
			t.Errorf("%s - want GBR got %s", dbType, ms.CountryCode)
		}

		if ms.LocationOfInterest != "Ashtead" {
			t.Errorf("%s - want Ashtead got %s", dbType, ms.LocationOfInterest)
		}

		if ms.Phone != "+44 1" {
			t.Errorf("%s - want +44 1 got %s", dbType, ms.Phone)
		}

		if ms.Mobile != "+44 2" {
			t.Errorf("%s - want +44 2 got %s", dbType, ms.Mobile)
		}

		if ms.AssocMobile != "+44 3" {
			t.Errorf("%s - want +44 3 got %s", dbType, ms.AssocMobile)
		}

		fetchedMOI, moie := db.GetMembersOtherInterests(u1.ID)
		if moie != nil {
			t.Error(moie)
			continue
		}
		if fetchedMOI.Interests != "foobar" {
			t.Errorf("%s - want foobar got %s", dbType, fetchedMOI.Interests)
		}

		if len(ms.TopicsOfInterest) != 2 {
			t.Errorf("%s want 2 got %d", dbType, len(ms.TopicsOfInterest))
			continue
		}

		p1, ok1 := ms.TopicsOfInterest[1]
		if !ok1 {
			t.Errorf("%s - topic[1] should exist", dbType)
		}
		if p1 != nil {
			t.Errorf("%s - topic[1] should be nil", dbType)
		}

		p2, ok2 := ms.TopicsOfInterest[3]
		if !ok2 {
			t.Errorf("%s - topic[3] should exist", dbType)
		}
		if p2 != nil {
			t.Errorf("%s - topic[3] should be nil", dbType)
		}
	}
}

// TestMakeCountrySelectionList checks MakeCountrySelectionList.
func TestMakeCountrySelectionList(t *testing.T) {

	// The main list of countries.
	countries := []database.Country{
		database.Country{Code: "1", Name: "country 1"},
		database.Country{Code: "2", Name: "country 2"},
		database.Country{Code: "3", Name: "country 3"},
		database.Country{Code: "4", Name: "country 4"},
		database.Country{Code: "5", Name: "country 5"},
	}

	// Leaders and wanted results:

	// A leader list containg no countries.
	var noLeaders []database.Country

	const wantNoLeaders = `
		<select name='country_code' id='countries' size='5'>
			<option value='1'>country 1</option>
			<option value='2'>country 2</option>
			<option value='3'>country 3</option>
			<option value='4'>country 4</option>
			<option value='5'>country 5</option>
		</select>
	`

	// A leader list containing all of the countries.  (Note that only the
	// code has to match.  Making the country name different from the ones
	// in countries allows us to see where the resulting values came from)
	all := []database.Country{
		database.Country{Code: "1", Name: "country 11"},
		database.Country{Code: "2", Name: "country 22"},
		database.Country{Code: "3", Name: "country 33"},
		database.Country{Code: "4", Name: "country 44"},
		database.Country{Code: "5", Name: "country 55"},
	}

	const wantAll = `
		<select name='country_code' id='countries' size='5'>
			<option value='1'>country 11</option>
			<option value='2'>country 22</option>
			<option value='3'>country 33</option>
			<option value='4'>country 44</option>
			<option value='5'>country 55</option>
		</select>
	`

	one := []database.Country{
		database.Country{Code: "2", Name: "country 22"},
	}

	const wantOne = `
		<select name='country_code' id='countries' size='5'>
			<option value='2'>country 22</option>
			<option value='1'>country 1</option>
			<option value='3'>country 3</option>
			<option value='4'>country 4</option>
			<option value='5'>country 5</option>
		</select>
	`

	// countries scattered through the list - not at the start, not at the end
	// and not next to each other.
	scattered := []database.Country{
		database.Country{Code: "2", Name: "country 22"},
		database.Country{Code: "4", Name: "country 44"},
	}

	const wantScattered = `
		<select name='country_code' id='countries' size='5'>
			<option value='2'>country 22</option>
			<option value='4'>country 44</option>
			<option value='1'>country 1</option>
			<option value='3'>country 3</option>
			<option value='5'>country 5</option>
		</select>
	`

	var testData = []struct {
		description string
		leaders     []database.Country
		want        string
	}{

		{"none", noLeaders, wantNoLeaders},
		{"all", all, wantAll},
		{"one", one, wantOne},
		{"scattered", scattered, wantScattered},
	}

	for _, td := range testData {
		got := MakeCountrySelectionList(td.leaders, countries)
		got = cleanUpHTML(got)
		want := cleanUpHTML(td.want)

		if got != want {
			t.Errorf("%s\nwant %s\ngot  %s\n%s", td.description, want, got, diff.Diff(want, got))
		}
	}
}

func TestMakeCountrySelectionListFavouring(t *testing.T) {
	db, connError := database.OpenDBForTesting("sqlite")

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

	// Create a structured logger that writes to the dailyLogWriter.
	dailyLogWriter := dailylogger.New("..", "test.", ".log")
	logger := slog.New(slog.NewTextHandler(dailyLogWriter, nil))
	db.Logger = logger
	h := Handler{DB: db, Logger: logger, Conf: &testConfig}

	want := `
		<select name='country_code' id='countries' size='5'>
			<option value='GBR'>United Kingdom</option>
			<option value='ABW'>Aruba</option>
			<option value='ZWE'>Zimbabwe</option>
		</select>
	`
	want = cleanUpHTML(want)

	got, err := h.MakeCountrySelectionListFavouring("GBR")
	if err != nil {
		t.Error(err)
		return
	}

	got = cleanUpHTML(got)

	if got != want {
		t.Errorf("want %s\ngot  %s\n%s", want, got, diff.Diff(want, got))
	}

}

func cleanUpHTML(s string) string {
	// Regular expressions for cleaning up the HTML.
	// The very first runes in the string are white space
	leadingWhiteSpaceRX := `^[\n\t ]*`
	leadingWhiteSpace := regexp.MustCompile(leadingWhiteSpaceRX)
	// The very last runes in the string are white space.
	trailingWhiteSpaceRX := `[\n\t ]*$`
	trailingWhiteSpace := regexp.MustCompile(trailingWhiteSpaceRX)
	// leading spaces on a line.
	leadingSpacesRX := `\n[ \t]*`
	leadingSpaces := regexp.MustCompile(leadingSpacesRX)

	s = leadingWhiteSpace.ReplaceAllString(s, "")
	s = trailingWhiteSpace.ReplaceAllString(s, "")
	s = leadingSpaces.ReplaceAllString(s, "\n")
	s = strings.ReplaceAll(s, "\n\n", "\n")

	return s
}

// createtestuser creates a user for testing.
func createTestUser(db *database.Database, t *testing.T) *database.User {

	// This shoud never fail so on any error, stop the test suite.
	u, uie := database.CreateUuid(db.Transaction, "usr_login_name", "adm_users")
	if uie != nil {
		t.Fatal(uie)
	}

	user := database.NewUser(u)

	e := db.CreateUser(user)
	if e != nil {
		t.Fatal(e)
	}

	return user
}
