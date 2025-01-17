package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"

	"github.com/goblimey/go-stripe-payments/config"
	"github.com/goblimey/go-stripe-payments/database"
)

var paymentPageTemplate *template.Template

func init() {
	// Check the response HTML templates.
	paymentPageTemplate = template.Must(template.New("formTemplate").
		Parse(paymentPageTemplateStr))
}

// PaymentFormData holds the submitted form data for validation and display.
type PaymentFormData struct {

	// Valid is set false during validation if the form data is invalid.
	Valid bool

	// Reference Data.
	PaymentYear           int    // The payment year displayed in the title.
	OrdinaryMemberFeeStr  string // The fee for ordinary membership.
	AssociateMemberFeeStr string // The fee for associate membership.
	FriendFeeStr          string // The fee for a friend of the museum.

	// Data for validation.
	FirstName            string
	LastName             string
	Email                string
	FriendStr            string // tickbox - "on", "off", "checked" or "unchecked"
	DonationToSocietyStr string // number
	DonationToMuseumStr  string // number
	GiftaidStr           string // tickbox - "on", "off", "checked" or "unchecked"
	AssocFirstName       string
	AssocLastName        string
	AssocEmail           string
	AssocFriendStr       string // tickbox - "on", "off", "checked" or "unchecked"

	//  Values set during validation.
	Friend            bool    // True if the ordinary member's Friend tickbox is valid and true.
	DonationToSociety float64 // Set during validation
	DonationToMuseum  float64 // Set during validation
	Giftaid           bool    // True if the giftaid tickbox is valid and true.
	AssocFriend       bool    // True if the associate member's friend tickbox is valid and true.

	OrdinaryMemberFee  float64 // The ordinary member fee converted to float.
	AssociateMemberFee float64 // The associate member fee converted to float.
	FriendFee          float64 // The friend fee converted to float.

	UserID      int // The ID of the ordinary member in the database (> zero).
	AssocUserID int // The ID of the associate member in the database (zero if no associate).

	// Error messages set if the form data is invalid.
	GeneralErrorMessage           string // Set on a fatal error, eg database connection failure.
	FirstNameErrorMessage         string
	LastNameErrorMessage          string
	EmailErrorMessage             string
	DonationToSocietyErrorMessage string
	DonationToMuseumErrorMessage  string
	AssocFirstNameErrorMessage    string
	AssocLastNameErrorMessage     string
}

// NewPaymentForm finds the membership year we are currently selling
// and creates a payment form.
func NewPaymentForm(ordinaryMembershipFeeStr, associateMembershipFeeStr, friendMembershipFeeStr string) *PaymentFormData {
	paymentYear := database.GetPaymentYear(time.Now())
	f := createPaymentForm(
		paymentYear,
		ordinaryMembershipFeeStr,
		associateMembershipFeeStr,
		friendMembershipFeeStr,
	)

	return f
}

// createPaymentForm sets the payment form with the reference data
// and the given payment year.  Factored out to support unit testing.
func createPaymentForm(
	paymentYear int,
	ordinaryMembershipFeeStr,
	associateMembershipFeeStr,
	friendMembershipFeeStr string,
) *PaymentFormData {

	f := PaymentFormData{
		PaymentYear:           paymentYear,
		OrdinaryMemberFeeStr:  ordinaryMembershipFeeStr,
		AssociateMemberFeeStr: associateMembershipFeeStr,
		FriendFeeStr:          friendMembershipFeeStr,
	}

	return &f
}

// MarkMandatoryFields marks the mandatory parameters in a
// payment form by setting error messages containing asterisks.
// This drives the first view of the payment page.
func (f *PaymentFormData) MarkMandatoryFields() {
	f.FirstNameErrorMessage = "*"
	f.LastNameErrorMessage = "*"
	f.EmailErrorMessage = "*"
}

// MustSetFees converts the fees as strings to floats.  The app can't
// work without this so any error is fatal.
func MustSetFees(ordinaryMembershipFeeStr, associateMembershipFeeStr, friendMembershipFeeStr string) (float64, float64, float64) {

	var ordinaryMembershipFee float64
	n1, ordinaryMembershipFeeError :=
		fmt.Sscanf(ordinaryMembershipFeeStr, "%f", &ordinaryMembershipFee)
	if ordinaryMembershipFeeError != nil {
		fmt.Println("MustSetFees: illegal ordinaryMembershipFee ", ordinaryMembershipFeeStr)
		os.Exit(-1)
	}
	if n1 < 1 {
		fmt.Println("MustSetFees: Failed to convert ordinaryMembershipFee ", ordinaryMembershipFeeStr)
		os.Exit(-1)
	}

	var associateMembershipFee float64
	n2, associateMembershipFeeError :=
		fmt.Sscanf(associateMembershipFeeStr, "%f", &associateMembershipFee)
	if associateMembershipFeeError != nil {
		fmt.Println("MustSetFees: illegal associateMembershipFee ", associateMembershipFeeStr)
		os.Exit(-1)
	}
	if n2 < 1 {
		fmt.Println("MustSetFees: Failed to convert associateMembershipFee ", associateMembershipFeeStr)
		os.Exit(-1)
	}

	var friendMembershipFee float64
	n3, friendMembershipFeeError :=
		fmt.Sscanf(friendMembershipFeeStr, "%f", &friendMembershipFee)
	if friendMembershipFeeError != nil {
		fmt.Println("MustSetFees: illegal friendMembershipFee ", friendMembershipFeeStr)
		os.Exit(-1)
	}
	if n3 < 1 {
		fmt.Println("MustSetFees: Failed to convert friendMembershipFee ", friendMembershipFeeStr)
		os.Exit(-1)
	}

	return ordinaryMembershipFee, associateMembershipFee, friendMembershipFee
}

type Handler struct {
	Conf                   *config.Config // The config.
	PaymentYear            int            // The membership year we are currently selling.
	OrdinaryMembershipFee  float64
	AssociateMembershipFee float64
	FriendMembershipFee    float64
	// The display versions of the fees, eg "24".
	OrdinaryMembershipFeeStr  string
	AssociateMembershipFeeStr string
	FriendMembershipStr       string
	Protocol                  string // "http" or "https"
}

func New(
	conf *config.Config,
	ordinaryMembershipFee float64,
	associateMembershipFee float64,
	friendMembershipFee float64,
) *Handler {

	var protocol string
	if len(conf.SSLCertificateFile) > 0 {
		protocol = "https"
	} else {
		protocol = "http"
	}

	h := Handler{

		Conf:                      conf,
		OrdinaryMembershipFee:     ordinaryMembershipFee,
		AssociateMembershipFee:    associateMembershipFee,
		FriendMembershipFee:       friendMembershipFee,
		OrdinaryMembershipFeeStr:  fmt.Sprintf("%.2f", ordinaryMembershipFee),
		AssociateMembershipFeeStr: fmt.Sprintf("%.2f", associateMembershipFee),
		FriendMembershipStr:       fmt.Sprintf("%.2f", friendMembershipFee),
		Protocol:                  protocol,
	}

	return &h
}

// GetPaymentData handles the /displayPaymentForm request.
// It validates the incoming payment data form.  If the data
// is valid it displays the cost breakdown, otherwise it
// displays the payment data form again with error messages.
func (hdlr *Handler) GetPaymentData(w http.ResponseWriter, r *http.Request) {

	form := NewPaymentForm(
		hdlr.OrdinaryMembershipFeeStr,
		hdlr.AssociateMembershipFeeStr,
		hdlr.FriendMembershipStr,
	)

	dbConfig := database.GetDBConfigFromTheEnvironment()
	db := database.New(dbConfig)
	connectionError := db.Connect()
	if connectionError != nil {
		fmt.Println(connectionError.Error())
		form.GeneralErrorMessage = fmt.Sprintf("Fatal error - %v", connectionError)
		form.Valid = false
		return
	}

	defer db.Close()

	// The helper does the work.
	hdlr.paymentDataHelper(w, r, form, db)
}

// GetPaymentDataHelper validates the form and prepares the response.
// It's separated out to support unit testing.
func (hdlr *Handler) paymentDataHelper(w http.ResponseWriter, r *http.Request, form *PaymentFormData, db *database.Database) {

	form.FirstName = r.PostFormValue("first_name")
	form.LastName = r.PostFormValue("last_name")
	form.Email = r.PostFormValue("email")
	form.FriendStr = r.PostFormValue("friend")
	form.DonationToSocietyStr = r.PostFormValue("donation_to_society")
	form.DonationToMuseumStr = r.PostFormValue("donation_to_museum")
	form.GiftaidStr = r.PostFormValue("giftaid")

	form.AssocFirstName = r.PostFormValue("assoc_first_name")
	form.AssocLastName = r.PostFormValue("assoc_last_name")
	form.AssocEmail = r.PostFormValue("assoc_email")
	form.AssocFriendStr = r.PostFormValue("assoc_friend")

	// Validate the form data.  On the first call the form is empty.
	// The validator sets error messages containing asterisks against
	// the mandatory fields.  On calls with incoming data, it validates
	// that data and sets error messages.

	valid := validate(form, db)

	if !valid {

		// There are errors, display the form again
		// with any supplied fields filled in.
		displayPaymentForm(w, form)

		return
	}

	// If we get to here, the form data is valid and the user details in
	// it are for real users.  Build and display the payment confirmation
	// page.

	ms := database.MembershipSale{
		MembershipYear:    form.PaymentYear,
		OrdinaryMemberID:  form.UserID,            // Always present.
		AssociateMemberID: form.AssocUserID,       // 0 if no associated member.
		OrdinaryMemberFee: form.OrdinaryMemberFee, // Always present.
		DonationToSociety: form.DonationToSociety, // 0.0 if no donation given.
		DonationToMuseum:  form.DonationToMuseum,  // 0.0 if no donation given.
		// The tick boxes are dealt with below.
	}
	// Create a list of hidden variables to drive the next request.
	hiddenVars := `
			<input type='hidden' name='user_id' value='{{.OrdinaryMemberID}}'>
`

	if form.Friend {
		// The ordinary member wants to be a friend of the museum.
		ms.OrdinaryMemberIsFriend = true
		hiddenVars += `
	<input type='hidden' name='friend' value='on'>
`
	}

	if form.Giftaid {
		// The ordinary member consents to Giftaid.
		ms.Giftaid = true
		hiddenVars += `
		<input type='hidden' name='giftaid' value='on'>
`
	}

	if form.AssocUserID > 0 {
		// There is an associate member.
		ms.AssociateMemberFee = form.AssociateMemberFee
		hiddenVars += `
			<input type='hidden' name='assoc_user_id' value='{{.AssociateMemberID}}'>
`
	}

	if form.AssocFriend {
		// The associate member wants to be a friend of the museum.
		ms.AssocMemberIsFriend = true
		ms.AssociateMemberFriendFee = form.FriendFee
		hiddenVars += `
			<input type='hidden' name='assoc_friend' value='on'>
`
	}

	if form.DonationToSociety > 0.0 {
		hiddenVars += `
			<input type='hidden' name='donation_to_society' value='{{.DonationToSociety}}'>
`
	}

	if form.DonationToMuseum > 0.0 {
		hiddenVars += `
			<input type='hidden' name='donation_to_museum' value='{{.DonationToMuseum}}'>
`
	}

	insert := hdlr.createCostBreakdown(&ms) + hiddenVars

	// Insert the cost breakdown and the hidden variables into the
	// shopping trolley page temlate.
	paymentConfirmationPageTemplateBody := fmt.Sprintf(paymentConfirmationPageTemplateStr, insert)

	// Check the template.
	paymentConfirmationPageTemplate, templateError :=
		template.New("PaymentConfirmationPage").Parse(paymentConfirmationPageTemplateBody)
	if templateError != nil {
		errorHTML := fmt.Sprintf(errorHTMLTemplate, templateError.Error())
		w.Write([]byte(errorHTML))
	}

	// Write the response.
	executeError := paymentConfirmationPageTemplate.Execute(w, ms)

	if executeError != nil {
		errorHTML := fmt.Sprintf(errorHTMLTemplate, executeError.Error())
		w.Write([]byte(errorHTML))
	}
}

// CreateCheckoutSession is the handler for the /create-checkout-session
// request.  It prepares the Stripe session and invokes
func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	successURL :=
		fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}",
			h.Protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", h.Protocol, r.Host)

	invoiceEnabled := true
	year := database.GetPaymentYear(time.Now())
	description := fmt.Sprintf("Leatherhead & District Local History Society membership renewal %d", year)
	invoiceData := stripe.CheckoutSessionInvoiceCreationInvoiceDataParams{
		Description: &description,
	}
	// Create a checkout session containing the client reference ID.
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled:     &invoiceEnabled,
			InvoiceData: &invoiceData,
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("gbp"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Service"),
					},
					UnitAmount: stripe.Int64(3000),
				},
				Quantity: stripe.Int64(1),
			},
		},
		// This ID will be returned in the session.
		// ClientReferenceID: &userID,
		// Stripe will request this URL if the payment is
		// successful. The {CHECKOUT_SESSION_ID} placeholder will
		// be replaced by the session ID, which allows the handler
		// to retrieve the session.
		SuccessURL: stripe.String(successURL),
		// Stripe will request this URL if the payment if cancelled.
		CancelURL: stripe.String(cancelURL),
	}

	// Create the checkout session.
	s, err := session.New(params)
	if err != nil {
		log.Printf("/create-checkout-session: error creating stripe session: %v", err)
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// Success is the handler for the /success request.  On a successful
// payment, the Stripe system issues that request, filling in the
// {CHECKOUT_SESSION_ID} placeholder with the session ID.  The
// handler uses that to retrieve the checkout session, extract the
// client reference and complete the sale.
func (hdlr *Handler) Success(w http.ResponseWriter, r *http.Request) {

	// year := database.GetPaymentYear(time.Now())
	sessionID := r.URL.Query().Get("session_id")
	params := stripe.CheckoutSessionParams{}
	stripeSession, sessionGetError := session.Get(sessionID, &params)
	if sessionGetError != nil {
		fmt.Println("/success: error - no session")
		reportError(w, sessionGetError)
	}

	dbConfig := database.GetDBConfigFromTheEnvironment()
	db := database.New(dbConfig)
	connectError := db.Connect()
	if connectError != nil {
		fmt.Println(connectError.Error())
		return
	}

	defer db.Close()

	hdlr.successHelper(stripeSession.ClientReferenceID, sessionID, w, db)

}

// successHelper completes the sale.  It's separated out to support
// unit testing.
func (hdlr *Handler) successHelper(salesIDstr, sessionID string, w http.ResponseWriter, db *database.Database) {

	var salesID int
	_, salesIDError := fmt.Sscanf(salesIDstr, "%d", &salesID)
	if salesIDError != nil {
		fmt.Println("successHelper: ", salesIDError.Error())
		reportError(w, salesIDError)
	}

	// Get the membership sales record.  The ClientReferenceID in the payment
	// session is the ID of the sales record.
	ms, msFetchError := db.GetMembershipSale(salesID)
	if msFetchError != nil {
		fmt.Println("successHelper: ", msFetchError.Error())
		reportError(w, msFetchError)
	}

	// The userID of the full member and maybe their associate
	// member is in the sales record.

	// Set the end date of the full member.  This is the most
	// important change.  If it fails, warn the user.
	fmError := db.SetMemberEndDate(ms.OrdinaryMemberID, ms.MembershipYear)
	if fmError != nil {
		fmt.Println("successHelper: ", fmError.Error())
		email := "treasurer@leatherheadhistory.org"
		em := "Something went wrong.  Please contact " + email
		reportError(w, errors.New(em))
	}

	// Any errors after this point are less important.  Log them
	// but don't show the user an error message.

	// Set the giftaid tick box, true or false.  The box may
	// already be set from last year and this year it may be
	// set differently, so don't just set it if the given
	// value is true.
	giftAidErr := db.SetGiftaidField(ms.OrdinaryMemberID, ms.Giftaid)
	if giftAidErr != nil {
		fmt.Println("successHelper: ", giftAidErr.Error())
	}

	// Set date last paid.
	now := time.Now()
	dlpError := db.SetDateLastPaid(ms.OrdinaryMemberID, now)
	if dlpError != nil {
		fmt.Println("successHelper: ", dlpError.Error())
	}

	// Set members at address - 2 if there is an associate, otherwise 1.
	membersAtAddress := 1
	var friendsAtAddress int

	if ms.OrdinaryMemberIsFriend {
		friendsAtAddress++
	}

	// If the member is a friend, tick the box.  In case it's
	// already set from last year but not this year, ensure that the
	// value in the DB record is reset.
	friendError := db.SetFriendField(ms.OrdinaryMemberID, ms.OrdinaryMemberIsFriend)
	if friendError != nil {
		fmt.Println("successHelper: friend ", friendError.Error())
	}

	if ms.AssociateMemberID > 0 {

		// Set the Friend field for the associate.
		err := db.SetFriendField(ms.AssociateMemberID, ms.AssocMemberIsFriend)
		if err != nil {
			fmt.Println("successHelper: associate friend", err.Error())
		}

		// Set the end date of the associate member.
		setError := db.SetMemberEndDate(ms.AssociateMemberID, ms.MembershipYear)
		if setError != nil {
			fmt.Println("successHelper: ", setError.Error())
		}

		membersAtAddress++
		if ms.AssocMemberIsFriend {
			friendsAtAddress++
		}

		setMembersError := db.SetMembersAtAddress(ms.AssociateMemberID, membersAtAddress)
		if setMembersError != nil {
			fmt.Println("successHelper: ", setMembersError.Error())
		}

		setFriendsError := db.SetFriendsAtAddress(ms.AssociateMemberID, friendsAtAddress)
		if setFriendsError != nil {
			fmt.Println("successHelper: ", setFriendsError.Error())
		}
	}

	setMembersError := db.SetMembersAtAddress(ms.OrdinaryMemberID, membersAtAddress)
	if setMembersError != nil {
		fmt.Println("successHelper: members ", setMembersError.Error())
	}

	setFriendsError := db.SetFriendsAtAddress(ms.OrdinaryMemberID, friendsAtAddress)
	if setFriendsError != nil {
		fmt.Println("successHelper: friends ", setFriendsError.Error())
	}

	// Update the last payment.
	paymentError := db.SetLastPayment(ms.OrdinaryMemberID, ms.TotalPayment())
	if paymentError != nil {
		fmt.Printf("successHelper: error setting last payment for %d - %v",
			ms.OrdinaryMemberID, paymentError)
	}

	// Update the user's donation to society.
	dsError := db.SetDonationToSociety(ms.OrdinaryMemberID, ms.DonationToSociety)
	if dsError != nil {
		fmt.Printf("successHelper: error setting donation to society for %d - %v",
			ms.OrdinaryMemberID, dsError)
	}

	// Update the user's donation to museum.
	dmError := db.SetDonationToMuseum(ms.OrdinaryMemberID, ms.DonationToMuseum)
	if dmError != nil {
		fmt.Printf("successHelper: error setting donation to museum for %d - %v",
			ms.OrdinaryMemberID, dmError)
	}

	if ms.AssociateMemberID > 0 {
		assocFriendError := db.SetFriendField(ms.AssociateMemberID, ms.AssocMemberIsFriend)
		if assocFriendError != nil {
			fmt.Printf("successHelper: error setting friend value for %d - %v",
				ms.AssociateMemberID, assocFriendError)
		}
	}

	// Update the membership sale record.
	updateError := ms.Update(db, "complete", sessionID)
	if updateError != nil {
		log.Printf("successHelper: failed to update membershipsales record %d", ms.ID)
	}

	// Create the response page.

	insert := hdlr.createCostBreakdown(ms)

	// Insert the cost breakdown and the hidden variables into the template.
	successPageTemplateBody := fmt.Sprintf(successPageTemplateStr, insert)

	// Check the template.
	successPageTemplate, parseError :=
		template.New("SuccessPage").Parse(successPageTemplateBody)
	if parseError != nil {
		errorHTML := fmt.Sprintf(errorHTMLTemplate, parseError.Error())
		w.Write([]byte(errorHTML))
	}

	// Write the response.
	executeError := successPageTemplate.Execute(w, ms)

	if executeError != nil {
		errorHTML := fmt.Sprintf(errorHTMLTemplate, executeError.Error())
		w.Write([]byte(errorHTML))
	}
}

// Cancel is the handler for the /cancel request.  Stripe makes that
// request when the payment is cancelled.
func (hdlr *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(cancelHTML))
}

func (hdlr *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	form := PaymentFormData{}
	userIDParam := r.PostFormValue("user_id")
	assocUserIDParam := r.PostFormValue("assoc_user_id")
	form.FriendStr = r.PostFormValue("friend")
	form.AssocFriendStr = r.PostFormValue("assoc_friend")
	form.DonationToSocietyStr = r.PostFormValue("donation_to_society")
	form.DonationToMuseumStr = r.PostFormValue("donation_to_museum")
	form.GiftaidStr = r.PostFormValue("giftaid")

	if len(userIDParam) == 0 {
		// Somebody has bypassed the form that we gave them.
		// Send them back to the start.
		hdlr.displayInitialPaymentForm(w)
		return
	}

	var userID, assocUserID int
	var friendFee, assocFee, assocFriendFee float64

	_, userErr := fmt.Sscanf(userIDParam, "%d", &userID)
	if userErr != nil {
		fmt.Println("checkout:", userErr.Error())
		reportError(w, userErr)
	}

	if len(assocUserIDParam) > 0 {
		_, assocUserErr := fmt.Sscanf(assocUserIDParam, "%d", &assocUserID)
		if assocUserErr != nil {
			fmt.Println("checkout:", assocUserErr.Error())
			reportError(w, assocUserErr)
		}
	}

	if assocUserID > 0 {
		assocFee = hdlr.AssociateMembershipFee
	}

	if len(form.FriendStr) > 0 {
		form.Friend = getTickBox(form.FriendStr)
		if form.Friend {
			friendFee = hdlr.FriendMembershipFee
		}
	}

	if len(form.AssocFriendStr) > 0 {
		form.AssocFriend = getTickBox(form.AssocFriendStr)
		if form.AssocFriend {
			assocFriendFee = hdlr.AssociateMembershipFee
		}
	}

	if len(form.GiftaidStr) > 0 {
		form.Giftaid = getTickBox(form.GiftaidStr)
	}

	if len(form.DonationToSocietyStr) > 0 {

		e, v := checkDonation(form.DonationToSocietyStr)
		if len(e) > 0 {
			fmt.Println("checkout:", "donationToSociety - ", e)
			reportError(w, errors.New("donation to society - "+e))
		}

		form.DonationToSociety = v
	}

	if len(form.DonationToMuseumStr) > 0 {

		e, v := checkDonation(form.DonationToMuseumStr)
		if len(e) > 0 {
			fmt.Println("checkout:", "DonationToMuseum - ", e)
			reportError(w, errors.New("donation to museum - "+e))
		}

		form.DonationToMuseum = v
	}

	// The payment ID is initially an empty string.  It will be supplied
	// by the payment processor later if the payment is successful.
	now := time.Now()
	ms := database.MembershipSale{
		PaymentService:           "Stripe",
		PaymentStatus:            "Pending",
		MembershipYear:           database.GetPaymentYear(now),
		OrdinaryMemberID:         userID,
		OrdinaryMemberFee:        hdlr.OrdinaryMembershipFee,
		OrdinaryMemberIsFriend:   form.Friend,
		OrdinaryMemberFriendFee:  friendFee,
		AssociateMemberID:        assocUserID,
		AssocMemberIsFriend:      form.AssocFriend,
		AssociateMemberFee:       assocFee,
		AssociateMemberFriendFee: assocFriendFee,
		DonationToSociety:        form.DonationToSociety,
		DonationToMuseum:         form.DonationToMuseum,
		Giftaid:                  form.Giftaid,
	}

	dbConfig := database.GetDBConfigFromTheEnvironment()
	db := database.New(dbConfig)
	connectError := db.Connect()
	if connectError != nil {
		fmt.Println(connectError.Error())
		return
	}

	defer db.Close()

	salesID, createError := ms.Create(db)
	if createError != nil {
		fmt.Println("checkout:", "CreateError: ", createError.Error())
		reportError(w, createError)
	}

	salesIDStr := fmt.Sprintf("%d", salesID)

	successURL := fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}", hdlr.Protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", hdlr.Protocol, r.Host)

	priceInPennies := int64(ms.TotalPayment()*100 + 0.5)

	invoicingEnabled := true

	description := fmt.Sprintf(
		"Leatherhead & District Local History Society membership %d", hdlr.PaymentYear)

	invoiceData := stripe.CheckoutSessionInvoiceCreationInvoiceDataParams{
		Description: &description,
	}

	invoiceCreation := stripe.CheckoutSessionInvoiceCreationParams{
		Enabled:     &invoicingEnabled,
		InvoiceData: &invoiceData,
	}

	// Create a checkout session containing a client reference ID
	// that gives the ms_id of the sales record.
	params := &stripe.CheckoutSessionParams{
		Mode:            stripe.String(string(stripe.CheckoutSessionModePayment)),
		InvoiceCreation: &invoiceCreation,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("gbp"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Service"),
					},
					UnitAmount: stripe.Int64(priceInPennies),
				},
				Quantity: stripe.Int64(1),
			},
		},
		// This ID will be returned in the session.
		ClientReferenceID: &salesIDStr,
		// Stripe will request this URL if the payment is
		// successful. The {CHECKOUT_SESSION_ID} placeholder will
		// be replaced by the session ID, which allows the handler
		// to retrieve the session.
		SuccessURL: stripe.String(successURL),
		// Stripe will request this URL if the payment if cancelled.
		CancelURL: stripe.String(cancelURL),
	}

	// Create the checkout session.
	s, err := session.New(params)
	if err != nil {
		log.Printf("/checkout: error creating session: %v", err)
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// DisplayPaymentForm displays the given payment form
func displayPaymentForm(w io.Writer, form *PaymentFormData) {

	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		fmt.Println(executeError)
	}
}

// DisplayInitialPaymentForm displays an empty payment form
// with the mandatory parameters marked with asterisks.
func (hdlr *Handler) displayInitialPaymentForm(w io.Writer) {

	form := hdlr.createPaymentFormData(hdlr.PaymentYear)
	form.MarkMandatoryFields()
	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		fmt.Println(executeError)
	}
}

// createPaymentForm sets the payment form with the reference data
// and the given payment year.  Factored out to support unit testing.
func (h *Handler) createPaymentFormData(paymentYear int) *PaymentFormData {

	f := PaymentFormData{
		PaymentYear:          paymentYear,
		OrdinaryMemberFeeStr: h.OrdinaryMembershipFeeStr,
		AssociateMemberFee:   h.AssociateMembershipFee,
		FriendFee:            h.FriendMembershipFee,
	}

	return &f
}

// createCostBreakdown creates an HTML table showing the cost of the
// membership sale.  The table is inserted into a template.
func (hdlr *Handler) createCostBreakdown(ms *database.MembershipSale) string {

	table := `
<table>
	<tr>
	    <td style='border: 0'>ordinary membership</td>
		<td style='border: 0'>£{{.OrdinaryMemberFee}}</td>
	</tr>
`

	if ms.OrdinaryMemberIsFriend {
		table += `
	<tr>
	    <td style='border: 0'>friend of the museum</td>
		<td style='border: 0'>£{{.OrdinaryMemberFriendFee}}</td>
	</tr>
`
	}

	if ms.AssociateMemberID > 0 {
		table += `
	<tr>
	    <td style='border: 0'>associate member</td>
		<td style='border: 0'>£{{.AssociateMemberFee}}</td>
	</tr>		
`
	}

	if ms.AssocMemberIsFriend {
		table += `
	<tr>
	    <td style='border: 0'>associate is friend of the museum</td>
		<td style='border: 0'>£{{.AssociateMemberFriendFee}}</td>
	</tr>
`
	}

	if ms.DonationToSociety > 0 {
		table += `
	<tr>
	    <td style='border: 0'>donation to the society</td>
		<td style='border: 0'>£{{.DonationToSociety}}</td>
	</tr>		
`
	}

	if ms.DonationToMuseum > 0 {
		table += `
	<tr>
	    <td style='border: 0'>donation to the museum</td>
		<td style='border: 0'>£{{.DonationToMuseum}}</td>
	</tr>
	`
	}

	totalTemplate := `
	<tr>
	    <td style='border: 0'><b>Total</b></td>
		<td style='border: 0'>£%.2f</td>
	</tr>
`

	table += fmt.Sprintf(totalTemplate, ms.TotalPayment())

	table += `</table>
`
	return table
}

// Validation error messages - factored out to support unit testing.
const firstNameErrorMessage = "You must fill in the first name"
const lastNameErrorMessage = "You must fill in the last name"
const emailErrorMessage = "You must fill in the email address"
const assocFirstNameErrorMessage = "If you fill in anything in this section, you must fill in the first name"
const assocLastNameErrorMessage = "If you fill in anything in this section, you must fill in the last name"
const invalidNumber = "must be a number"
const negativeNumber = "must be a 0 or greater"
const noSuchMember = "cannot find this member"

// simpleValidate takes the form parameters as arguments.  It returns true
// and all empty strings if the form is valid, false and the error messages set
// if it's invalid.  It doesn't check that the user(s) exist - that requires a
// database connection.  It's called by the validation .
func simpleValidate(form *PaymentFormData) bool {

	// Set the fees in the form.
	form.OrdinaryMemberFee, form.AssociateMemberFee, form.FriendFee =
		MustSetFees(form.OrdinaryMemberFeeStr, form.AssociateMemberFeeStr, form.FriendFeeStr)

	// form.Valid starts true and is set false if any of the form data is invalid.
	form.Valid = true

	if len(form.FirstName) == 0 &&
		len(form.LastName) == 0 &&
		len(form.Email) == 0 &&
		len(form.FriendStr) == 0 &&
		len(form.AssocFirstName) == 0 &&
		len(form.AssocLastName) == 0 &&
		len(form.AssocEmail) == 0 &&
		len(form.AssocFriendStr) == 0 &&
		len(form.DonationToSocietyStr) == 0 &&
		len(form.DonationToMuseumStr) == 0 &&
		len(form.GiftaidStr) == 0 {

		// On the first call the form is empty.  Mark the mandatory fields.
		// Return false and the handler will display the form
		// again, with the marks.
		form.MarkMandatoryFields()
		form.Valid = false
		return form.Valid
	}

	form.FirstName = strings.TrimSpace(form.FirstName)
	form.LastName = strings.TrimSpace(form.LastName)
	form.Email = strings.TrimSpace(form.Email)
	form.FriendStr = strings.TrimSpace(form.FriendStr)
	form.DonationToSocietyStr = strings.TrimSpace(form.DonationToSocietyStr)
	form.DonationToMuseumStr = strings.TrimSpace(form.DonationToMuseumStr)
	form.GiftaidStr = strings.TrimSpace(form.GiftaidStr)

	form.AssocFirstName = strings.TrimSpace(form.AssocFirstName)
	form.AssocLastName = strings.TrimSpace(form.AssocLastName)
	form.AssocEmail = strings.TrimSpace(form.AssocEmail)
	form.AssocFriendStr = strings.TrimSpace(form.AssocFriendStr)

	if len(form.FirstName) == 0 {
		form.FirstNameErrorMessage = firstNameErrorMessage
		form.Valid = false
	}

	if len(form.LastName) == 0 {
		form.LastNameErrorMessage = lastNameErrorMessage
		form.Valid = false
	}

	if len(form.Email) == 0 {
		form.EmailErrorMessage = emailErrorMessage
		form.Valid = false
	}

	form.Friend = getTickBox(form.FriendStr)
	if form.Friend {
		form.FriendStr = "on"
	} else {
		form.FriendStr = "off"
	}

	form.Giftaid = getTickBox(form.GiftaidStr)
	if form.Giftaid {
		form.GiftaidStr = "on"
	} else {
		form.GiftaidStr = "off"
	}

	// The associate fields are optional but if you fill in any
	// of them, you must fill in the first and last name.
	if len(form.AssocFirstName) > 0 ||
		len(form.AssocLastName) > 0 ||
		len(form.AssocEmail) > 0 ||
		len(form.AssocFriendStr) > 0 {

		if len(form.AssocFirstName) == 0 {
			form.AssocFirstNameErrorMessage = assocFirstNameErrorMessage
			form.Valid = false
		}
		if len(form.AssocLastName) == 0 {
			form.AssocLastNameErrorMessage = assocLastNameErrorMessage
			form.Valid = false
		}
	}

	form.AssocFriend = getTickBox(form.AssocFriendStr)
	if form.AssocFriend {
		form.AssocFriendStr = "on"
	} else {
		form.AssocFriendStr = "off"
	}

	// The mandatory parameters are all submitted.  Now check the contents.

	// If donation values are submitted, they must be numbers and not
	// negative.

	if len(form.DonationToSocietyStr) > 0 {

		errorMessage, dts := checkDonation(form.DonationToSocietyStr)
		if len(errorMessage) > 0 {
			form.DonationToSocietyErrorMessage = errorMessage
			form.Valid = false
		} else {
			form.DonationToSociety = dts
		}
	}

	if len(form.DonationToMuseumStr) > 0 {

		errorMessage, dtm := checkDonation(form.DonationToMuseumStr)
		if len(errorMessage) > 0 {
			form.DonationToMuseumErrorMessage = invalidNumber
			form.Valid = false
		} else {
			form.DonationToMuseum = dtm
		}
	}

	return form.Valid
}

// checkDonation checks a donation value - must be a valid float
// and not negative.  Returns an empty error message and the donation
// as a float64 OR an error message and 0.0.
func checkDonation(str string) (string, float64) {
	var v float64
	if len(str) > 0 {

		scannedItems, scanError := fmt.Sscanf(str, "%f", &v)

		if scanError != nil {
			return invalidNumber, 0.0
		}

		if scannedItems < 1 {
			return invalidNumber, 0.0
		}

		// The number must not be negative!
		if v < 0 {
			return negativeNumber, 0.0
		}
	}

	// Success!
	return "", v
}

// validate does a complete validation of the form.  Ir calls simpleValidate takes the form parameters as arguments.  It looks up the
// user in the database with the name and/or email address given in the form to
// check that it exists.  If the details of the associate are given, it checks
// that too.  It returns true and all empty strings if the user(s) exist, false
// and error messages set otherwise.
// It's separated out from the first stage of validation because that doesn't
// need to connect to the database and so can be more easily tested.
func validate(form *PaymentFormData, db *database.Database) bool {

	valid := simpleValidate(form)
	if !valid {
		return false
	}

	// There are no simple errors in the form.  Check that the user(s)
	// exist.

	var userIDError error
	form.UserID, userIDError = db.GetUserIDofMember(form.FirstName, form.LastName, form.Email)
	if userIDError != nil {
		form.FirstNameErrorMessage = invalidNumber
		form.LastNameErrorMessage = invalidNumber
		form.EmailErrorMessage = invalidNumber
		form.Valid = false
	}

	if len(form.AssocFirstName) > 0 {
		var assocUserIDError error
		form.AssocUserID, assocUserIDError =
			db.GetUserIDofMember(form.AssocFirstName, form.AssocLastName, form.AssocEmail)
		if assocUserIDError != nil {
			form.AssocFirstNameErrorMessage = noSuchMember
			form.AssocLastNameErrorMessage = noSuchMember
			form.Valid = false
		}
	}
	return form.Valid
}

// getTickBox returns true if the tickbox is ticked ("on" or "checked"),
// false otherwise.
func getTickBox(value string) bool {
	switch {
	case len(value) == 0:
		return false
	case value == "on" || value == "checked":
		return true
	default:
		return false
	}
}

func reportError(w http.ResponseWriter, err error) {
	errorHTML := fmt.Sprintf(errorHTMLTemplate, err)
	w.Write([]byte(errorHTML))
}

const paymentPageTemplateStr = `
<html>
<head>
    <title>Membership Renewal</title>
</head>
	<body style='font-size: 100%'>
		<h2>Leatherhead & District local history Society</h2>

		<h3>Membership Renewal {{.PaymentYear}}</h3>

		<span style="color:red;">{{.GeneralErrorMessage}}</span>
		<p>
			To renew your membership for the year
			using a credit or debit card,
			please fill in the form below and press the Submit button.
			You will then be transferred to the Stripe payment system
			to make the payment.
			The History Society will not see your card details,
			just the fact that you have paid.
		</p>
		<p>
			If you are also paying for an Associate Member
			(a second member at the same address)
			please supply their details too,
			otherwise leave those boxes blank.
		</p>
		<p>
			Our fees this year are:
		<ul>
			<li>Ordinary member: £{{.OrdinaryMemberFee}}</li>
			<li>Associate member at the same address: £{{.AssociateMemberFee}}</li>
			<li>Friend of the Leatherhead museum: £{{.FriendFee}}</li>
		</ul>
		</p>
		<p>
			&nbsp;
		</p>
		<form action="/displayPaymentForm" method="POST">	
			<table style='font-size: 100%'>
				<tr>
					<td style='border: 0'>First Name:</td>
					<td style='border: 0'><input type='text' size='40' name='first_name' value='{{.FirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.FirstNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Last Name:</td>
					<td style='border: 0'><input type='text' size='40' name='last_name' value='{{.LastName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.LastNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Email Address:</td>
					<td style='border: 0'><input type='text' size='40' name='email' value='{{.Email}}'></td>
					<td style='border: 0'><span style="color:red;">{{.EmailErrorMessage}}</span></td>
				</tr>
				<tr>
					<td style='border: 0'>Friend of the Museum (£5):</td>
					<td style='border: 0; '>
						<input style='transform: scale(1.5);' type='checkbox' name='friend' {{.Friend}}>
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0'>Donation to the Society:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_society' value='{{.DonationToSocietyStr}}'></td>
					<td style='border: 0;'><span style="color:red;">{{.DonationToSocietyErrorMessage}}</span></td>
				</tr>
				<tr>
					<td style='border: 0'>Donation to the Museum:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_museum' value='{{.DonationToMuseumStr}}'></td>
					<td style='border: 0'><span style="color:red;">{{.DonationToMuseumErrorMessage}}</span></td>
				</tr>
				<tr>
					<td style='border: 0'>Giftaid:</td>
					<td style='border: 0 '>
						<input style='transform: scale(1.5);' type='checkbox' name='giftaid' {{.Giftaid}}>
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>
						Tick the Giftaid box if you are currently a UK tax payer and 
						consent to Gift Aid.
						If you pay less income tax and/or capital gains tax 
						than the amount of Gift Aid paid on all your donations, 
						you are liable to pay the difference to HMRC.
					</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>

				<tr>
					<td style='border: 0' colspan='3'>
						If there are two members at your address, 
						fill in the other member's details below:
					</td>
				</tr>
				<tr>
					<td style='border: 0'>Associate First Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_first_name' value='{{.AssocFirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocFirstNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Associate Last Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_last_name' value='{{.AssocLastName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocLastNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Email Address (optional):</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_email' value='{{.AssocEmail}}'></td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0'>Friend of the Museum (£5):</td>
					<td style='border: 0'>
						<input style='transform: scale(1.5);' type='checkbox' name='assoc_friend' {{.Friend}}>
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>

				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>

			</table>
			<input type="submit" value="Submit">
		</form>
	</body>
</html>
`

const paymentConfirmationPageTemplateStr = `
<html>
    <head><title>payment confirmation</title></head>
	<body style='font-size: 100%%'>
		<h2>Leatherhead & District Local History Society</h2>
		<h3>Membership Renewal {{.MembershipYear}}</h3>
		<p>
			If you are happy with the total,
			please press the submit button.
			You will be transferred to the Stripe payment system
			to make the payment.
			The History Society will not see your card details,
			just the fact that you have paid.
		</p>
		<form action="/checkout" method="POST">
			%s
			<input type="submit" value="Submit">
		</form>
	</body>
`

const successPageTemplateStr = `
<html>
	<head><title>Payment Successful</title></head>
    <body style='font-size: 100%%'>
        <h1>Thank you</h1>
		<p>
			Your membership has been renewed until the end of {{.MembershipYear}}.
		</p>
		<p>
			%s
		</p>
		<p>
		    If you have any questions, please email
		    <a href="mailto:chairman@leatherheadhistory.org">
			    chairman@leatherheadhistory.org
			</a>.
		</p>
    </body>
</html>
`

const cancelHTML = `
<html>
  <head><title>cancelled</title></head>
  <body style='font-size: 100%'>
    <h1>Payment cancelled</h1>
  </body>
</html>
`

const errorHTMLTemplate = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%%'>
		<p>
			An error occurred.
			<br>
			%s
		</p>
    </body>
</html>
`
