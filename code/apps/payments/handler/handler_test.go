package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/forms"
)

// databaseList is a list of database types that will be used in
// integration tests.
var databaseList = []string{"postgres", "sqlite"}

func TestSuccess(t *testing.T) {
	body := strings.NewReader(`{"username": "admin","password":"secret"}`)
	http.NewRequest("POST", "/success", body)
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
		gotErrorMessage, gotValue := checkDonation(td.str)
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
			forms.SaleForm{ // This also checks the Trimspace calls.
				Valid:             false,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendInput: "on",
				DonationToSocietyInput: " 7.83\t", DonationToMuseumInput: " 8.9 ", GiftaidInput: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 78.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			true,

			forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 1.2, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				FriendOutput: "checked", AssocFriendOutput: "checked", GiftaidOutput: "checked",
				DonationToSocietyInput: "7.83", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, DonationToSociety: 7.83, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£1.20", "£3.40", "£5.60", "£7.83", "£8.90",
		},

		{
			"valid - not a friend",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 3.4, FriendFee: 5.6,
				MembershipYear: 2024, AssocFeeToPay: 6, FriendFeeToPay: 5,
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
				MembershipYear: 2024, AssocFeeToPay: 6, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true, AssocFriend: true,
				GiftaidOutput: "checked", AssocFriendOutput: "checked", FriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			"£24.00", "£6.00", "£5.00", "£1.50", "£2.50",
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
				MembershipYear: 2024, AssocFeeToPay: 6, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
			"valid - ordinary member is not a friend",
			forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
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
				MembershipYear: 2024, AssocFeeToPay: 6, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
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
			description: "valid - associate friend is empty (not a friend)",
			form: forms.SaleForm{
				Valid:             false,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5,
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
				MembershipYear: 2024, AssocFeeToPay: 6, FriendFeeToPay: 5,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "1.5", DonationToMuseumInput: "2.5", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendInput: "off",
				Friend: true, DonationToSociety: 1.5, DonationToMuseum: 2.5, Giftaid: true, AssocFriend: false,
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
				MembershipYear: 2024, AssocFeeToPay: 3.456, FriendFeeToPay: 5.678,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off", GiftaidInput: "off",
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
			description: "valid - giftaid tick box on, others off",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
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
				OrdinaryMemberFee: 1.2, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "off",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendInput: "off",
				Friend: false, FriendOutput: "unchecked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: false, AssocFriendOutput: "unchecked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantOrdinaryMembershipFee:  "£1.20",
			wantAssociateMembershipFee: "£3.40",
			wantFriendFeePaid:          "£5.60",
			wantDonationToSociety:      "£7.80",
			wantDonationToMuseum:       "£8.90",
		},
		{
			description: "valid - no giftaid",
			form: forms.SaleForm{
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5.0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "",
				AssocFirstName: "c", AssocLastName: "d", AssocEmail: "c@d.com", AssocFriendInput: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			wantForm: forms.SaleForm{
				Valid:             true,
				OrdinaryMemberFee: 24, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 6.0, FriendFeeToPay: 5.0,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "off",
				AssocFirstName: "c", AssocLastName: "d", AssocEmail: "c@d.com", AssocFriendInput: "on",
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
			description: "ordinary member first name missing",
			form: forms.SaleForm{
				OrdinaryMemberFee: 1.2, AssocMemberFee: 6, FriendFee: 5,
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
				OrdinaryMemberFee: 1.2, AssocMemberFee: 6, FriendFee: 5,
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "", LastName: "b", Email: "a@b.com", FriendInput: "on",
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

				MembershipYear: 2024, OrdinaryMemberFee: 1.2, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
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
				MembershipYear: 2024, OrdinaryMemberFee: 1.2, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "", FriendInput: "on",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
				DonationToSocietyInput: "7.8", DonationToMuseumInput: "8.9", GiftaidInput: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendInput: "on",
				Friend: true, FriendOutput: "checked", DonationToSociety: 7.8, DonationToMuseum: 8.9,
				Giftaid: true, GiftaidOutput: "checked", AssocFriend: true, AssocFriendOutput: "checked",
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "", LastName: "", Email: "", FriendInput: "off",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
				MembershipYear: 2024, AssocFeeToPay: 3.4, FriendFeeToPay: 5.6,
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendInput: "on",
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
	}

	for _, td := range testData {
		valid := Validate(&td.form)

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

		// The test - usersExist() should give back the userIDs of the two users.
		fetchedOID, fetchedAID, lookupError := usersExist(&sale, db)
		if lookupError != nil {
			t.Error(lookupError)
			return
		}

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

		// Create an ordinary user.
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
