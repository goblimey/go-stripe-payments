package handler

import (
	"fmt"
	"testing"
)

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

// TestGetTickBox checks the getTextBox function.
func TestGetTickBox(t *testing.T) {
	var testData = []struct {
		input string
		want  bool
	}{
		{"on", true},
		{"checked", true},
		{"off", false},
		{"unchecked", false},
		{"foo", false},
		{"", false},
	}

	for _, td := range testData {
		got := getTickBox(td.input)

		if td.want != got {
			t.Errorf("%s: want %v got %v", td.input, td.want, got)
		}
	}
}

// TestValidation1 checks the first stage of validation.
func TestSimpleValidation1(t *testing.T) {

	var testData = []struct {
		description string
		form        PaymentFormData
		wantValid   bool
		want        PaymentFormData
	}{
		{
			"valid - all",
			PaymentFormData{ // This also checks the Trimspace calls.
				Valid:       false,
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ", GiftaidStr: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",

				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "", DonationToSocietyErrorMessage: "", DonationToMuseumErrorMessage: "",
			},
			true,
			PaymentFormData{
				true, 2024, "1.2", "3.4", "5.6",
				"a", "b", "a@b.com", "on",
				"7.8", "8.9", "on",
				"f", "l", "a@l.com", "on",
				true, 7.8, 8.9, true, true,
				1.2, 3.4, 5.6, 0, 0, "", "", "", "", "", "", "", "",
			},
		},

		{
			"valid - no friends",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "off",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "off",
				false, 1.5, 2.5, true, false,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - no associate",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"", "", "", "",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"", "", "", "off",
				true, 1.5, 2.5, true, false,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"valid - ordinary member is not a friend",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "on",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "on",
				false, 1.5, 2.5, true, true,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			description: "valid - associate is not friend",
			form: PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			wantValid: true,
			want: PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "on",
				"f", "l", "a@l.com", "off",
				true, 1.5, 2.5, true, false,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			description: "valid - friend tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			want: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "off",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "off",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: false, AssocFriend: false,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},
		{
			description: "valid - assoc friend tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "off",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssocFriendStr: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			want: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "off",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "off",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "", AssocFriendStr: "on",
				Friend: false, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: false, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},
		{
			description: "valid - giftaid tick box on, others off",
			form: PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "off",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			wantValid: true,
			want: PaymentFormData{
				Valid:       true,
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "off",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "off",
				Friend: false, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: false,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},
		{
			"valid - no giftaid",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "on",
				false, 0.0, 0.0, true, false, // These should all updated.
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			true,
			PaymentFormData{
				true, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "on",
				true, 1.5, 2.5, false, true,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", "", "", "",
			},
		},
		{
			"ordinary member first name missing",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"", "b", "a@b.com", "on",
				"1.5", "2.5", "off",
				"f", "l", "a@l.com", "off",
				true, 1.5, 2.5, false, false,
				24.0, 6.0, 5, 0, 0, "", firstNameErrorMessage, "", "", "", "", "", "",
			},
		},

		// PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
		// FirstName: " a\t", LastName: " b ", Email: " a@b.com ", FriendStr: "on",
		// DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ", GiftaidStr: "on",
		// AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",
		// Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
		// UserID: 0, AssocUserID: 0,
		// GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
		// AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",

		{
			"ordinary member last name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: " a\t", LastName: "", Email: " a@b.com ", FriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ", GiftaidStr: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},
		{
			"ordinary member email missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: " a\t", LastName: "b", Email: "", FriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ", GiftaidStr: "on",
				AssocFirstName: " f ", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},
		{
			"associate member first name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: " a\t", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: " 8.9 ", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: " l ", AssocEmail: "  a@l.com  ", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: "",
			},
		},
		{
			"associate member last name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "f", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
		},

		{
			"associate member but no ordinary member",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "", LastName: "", Email: "", FriendStr: "",
				DonationToSocietyStr: " 7.8\t", DonationToMuseumStr: "8.9", GiftaidStr: "",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "", LastName: "", Email: "", FriendStr: "off",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "off",
				AssocFirstName: "f", AssocLastName: "l", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: false, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: false, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: firstNameErrorMessage,
				LastNameErrorMessage: lastNameErrorMessage, EmailErrorMessage: emailErrorMessage,
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
		},

		{
			"associate email address but associate member's name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: true, AssocFriend: true,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "a@l.com", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
		},
		{
			"associate friend tick box but associate member's name missing",
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "on",
				Friend: false, DonationToSociety: 0.0, DonationToMuseum: 0.0, Giftaid: false, AssocFriend: false,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: "", AssocLastNameErrorMessage: "",
			},
			false,
			PaymentFormData{
				PaymentYear: 2024, OrdinaryMemberFeeStr: "1.2", AssociateMemberFeeStr: "3.4", FriendFeeStr: "5.6",
				FirstName: "a", LastName: "b", Email: "a@b.com", FriendStr: "on",
				DonationToSocietyStr: "7.8", DonationToMuseumStr: "8.9", GiftaidStr: "on",
				AssocFirstName: "", AssocLastName: "", AssocEmail: "", AssocFriendStr: "on",
				Friend: true, DonationToSociety: 7.8, DonationToMuseum: 8.9, Giftaid: true, AssocFriend: true,
				OrdinaryMemberFee: 1.2, AssociateMemberFee: 3.4, FriendFee: 5.6,
				UserID: 0, AssocUserID: 0,
				GeneralErrorMessage: "", FirstNameErrorMessage: "", LastNameErrorMessage: "", EmailErrorMessage: "",
				AssocFirstNameErrorMessage: assocFirstNameErrorMessage, AssocLastNameErrorMessage: assocLastNameErrorMessage,
			},
		},

		{
			"donation to society invalid number",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"junk", "4.5", "on",
				"f", "l", "a@l.com", "on",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "on",
				"junk", "4.5", "on",
				"f", "l", "a@l.com", "on",
				true, 0.0, 4.5, true, true,
				24.0, 6.0, 5, 0, 0, "", "", "", "", invalidNumber, "", "", "",
			},
		},
		{
			"donation to museum invalid number",
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "junk", "",
				"f", "l", "a@l.com", "",
				false, 0.0, 0.0, false, false,
				0.0, 0.0, 0.0, 0, 0, "", "", "", "", "", "", "", "",
			},
			false,
			PaymentFormData{
				false, 2024, "24.0", "6.0", "5.0",
				"a", "b", "a@b.com", "off",
				"1.5", "junk", "off",
				"f", "l", "a@l.com", "off",
				false, 1.5, 0.0, false, false,
				24.0, 6.0, 5, 0, 0, "", "", "", "", "", invalidNumber, "", "",
			},
		},
	}

	for _, td := range testData {
		valid := simpleValidate(&td.form)
		if td.wantValid != valid {
			t.Errorf("%s want %v got %v", td.description, td.wantValid, valid)
		}

		if td.want != td.form {
			fmt.Printf("%s:\nwant %v\n got %v\n", td.description, td.want, td.form)
		}

	}
}
