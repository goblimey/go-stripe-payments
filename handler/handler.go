package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"

	"github.com/goblimey/go-stripe-payments/config"
	"github.com/goblimey/go-stripe-payments/database"
)

// Membership fees - should really be in the database and time-sensitive.
const ordinaryMembershipFee = float64(24)
const associateMembershipFee = float64(6)
const friendMembershipFee = float64(5)

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
	PaymentYear            int     // The payment year displayed in the title.
	OrdinaryMembershipFee  float64 // The ordinary member fee
	AssociateMembershipFee float64 // The associate member fee. (0 if no associate.)
	FriendMembershipFee    float64 // The friend fee.  (0 if no associate or not friend).

	// Data for validation.
	FirstName              string
	LastName               string
	Email                  string
	FriendInput            string // tickbox - "on", "off", "checked" or "unchecked"
	DonationToSocietyInput string // number
	DonationToMuseumInput  string // number
	GiftaidInput           string // tickbox - "on", "off", "checked" or "unchecked"
	AssocFirstName         string
	AssocLastName          string
	AssocEmail             string
	AssociateIsFriendInput string // tickbox - "on", "off", "checked" or "unchecked"

	//  Values set during validation.
	Friend                  bool    // True if the ordinary member's Friend tickbox is valid and true.
	FriendOutput            string  // Checkbox setting - "checked" or "unchecked"
	DonationToSociety       float64 // Donation to the society.
	DonationToMuseum        float64 // Donation to the museum.
	Giftaid                 bool    // True if the giftaid tickbox is valid and true.
	GiftaidOutput           string  // Checkbox setting - "checked" or "unchecked"
	AssociateIsFriend       bool    // True if the associate member's friend tickbox is valid and true.
	AssociateIsFriendOutput string  // Checkbox setting - "checked" or "unchecked"

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
func NewPaymentForm() *PaymentFormData {

	paymentYear := database.GetPaymentYear(time.Now())
	f := createPaymentForm(paymentYear)

	return f
}

// createPaymentForm sets the payment form with the reference data
// and the given payment year.  Factored out to support unit testing.
func createPaymentForm(paymentYear int) *PaymentFormData {

	f := PaymentFormData{
		PaymentYear:            paymentYear,
		OrdinaryMembershipFee:  ordinaryMembershipFee,
		AssociateMembershipFee: associateMembershipFee,
		FriendMembershipFee:    friendMembershipFee,
	}

	return &f
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (form *PaymentFormData) OrdinaryMembershipFeeForDisplay() string {
	return fmt.Sprintf("%.2f", form.OrdinaryMembershipFee)
}

// AssociateMembershipFeeForDisplay gets the associate membership fee
// for a display - a number to two decimal places.
func (form *PaymentFormData) AssociateMembershipFeeForDisplay() string {
	return fmt.Sprintf("%.2f", form.AssociateMembershipFee)
}

// DonationToMuseumForDisplay gets the donation to museum
// for a display - a number to two decimal places.
func (form *PaymentFormData) DonationToMuseumForDisplay() string {
	if form.DonationToMuseum == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", form.DonationToMuseum)
}

// DonationToSocietyForDisplay gets the donation to the society
// for a display - a number to two decimal places.
func (form *PaymentFormData) DonationToSocietyForDisplay() string {
	if form.DonationToSociety == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", form.DonationToSociety)
}

// FriendMembershipFeeForDisplay gets the friend membership fee
// for display - a number to two decimal places.
func (form *PaymentFormData) FriendMembershipFeeForDisplay() string {
	return fmt.Sprintf("%.2f", form.FriendMembershipFee)
}

// MarkMandatoryFields marks the mandatory parameters in a
// payment form by setting error messages containing asterisks.
// This drives the first view of the payment page.
func (f *PaymentFormData) MarkMandatoryFields() {
	f.FirstNameErrorMessage = "*"
	f.LastNameErrorMessage = "*"
	f.EmailErrorMessage = "*"
}

type Handler struct {
	Conf                   *config.Config // The config.
	PaymentYear            int            // The membership year we are currently selling.
	OrdinaryMembershipFee  float64
	AssociateMembershipFee float64
	FriendMembershipFee    float64
	Protocol               string // "http" or "https"
}

// OrdinaryMembershipFeeForDisplay gets the ordinary membership fee
// for a display - a number to two decimal places.
func (hdlr *Handler) OrdinaryMembershipFeeForDisplay() string {
	return fmt.Sprintf("%.2f", hdlr.OrdinaryMembershipFee)
}

// AssociateMembershipFeeForDisplay gets the associate membership fee
// for display - a number to two decimal places.  If there is no
// associate, it returns "0.0".
func (hdlr *Handler) AssociateMembershipFeeForDisplay() string {

	return fmt.Sprintf("%.2f", hdlr.AssociateMembershipFee)
}

// OrdinaryMemberFriendFeeForDisplay gets the ordinary member's
// museum friend fee for display - a number to two decimal places.
// If the member is not a friend, it returns "0.0".
func (hdlr *Handler) FriendFeeForDisplay() string {

	return fmt.Sprintf("%.2f", hdlr.FriendMembershipFee)
}

func New(conf *config.Config) *Handler {

	var protocol string
	if len(conf.SSLCertificateFile) > 0 {
		protocol = "https"
	} else {
		protocol = "http"
	}

	h := Handler{

		Conf:                   conf,
		OrdinaryMembershipFee:  ordinaryMembershipFee,
		AssociateMembershipFee: associateMembershipFee,
		FriendMembershipFee:    friendMembershipFee,
		Protocol:               protocol,
	}

	return &h
}

// GetPaymentData handles the /displayPaymentForm request.
// It validates the incoming payment data form.  If the data
// is valid it displays the cost breakdown, otherwise it
// displays the payment data form again with error messages.
func (hdlr *Handler) GetPaymentData(w http.ResponseWriter, r *http.Request) {

	form := NewPaymentForm()

	dbConfig := database.GetDBConfigFromTheEnvironment()
	db := database.New(dbConfig)
	connectionError := db.Connect()
	if connectionError != nil {
		fmt.Println(connectionError.Error())
		form.GeneralErrorMessage = fmt.Sprintf("Fatal error - %v", connectionError)
		form.Valid = false
		return
	}

	// Start a transaction, stored in the database object.
	txError := db.BeginTx()

	if txError != nil {
		reportError(w, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer db.Rollback()

	defer db.Close()

	// The helper does the work.
	hdlr.paymentDataHelper(w, r, form, db)

	// paymentDataHelper doesn't change the database so we can just
	// close the transaction via a rollback.
	db.Rollback()
}

// GetPaymentDataHelper validates the form and prepares the response.
// It's separated out to support unit testing.
func (hdlr *Handler) paymentDataHelper(w http.ResponseWriter, r *http.Request, form *PaymentFormData, db *database.Database) {

	form.FirstName = r.PostFormValue("first_name")
	form.LastName = r.PostFormValue("last_name")
	form.Email = r.PostFormValue("email")
	form.FriendInput = r.PostFormValue("friend")
	form.DonationToSocietyInput = r.PostFormValue("donation_to_society")
	form.DonationToMuseumInput = r.PostFormValue("donation_to_museum")
	form.GiftaidInput = r.PostFormValue("giftaid")

	form.AssocFirstName = r.PostFormValue("assoc_first_name")
	form.AssocLastName = r.PostFormValue("assoc_last_name")
	form.AssocEmail = r.PostFormValue("assoc_email")
	form.AssociateIsFriendInput = r.PostFormValue("assoc_friend")

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

	ms := database.NewMembershipSale(form.OrdinaryMembershipFee)
	ms.TransactionType = database.TransactionTypeRenewal
	ms.MembershipYear = form.PaymentYear
	ms.OrdinaryMemberID = form.UserID                 // Always present.
	ms.AssociateMemberID = form.AssocUserID           // 0 if no associated member.
	ms.OrdinaryMemberFee = form.OrdinaryMembershipFee // Always present.
	ms.DonationToSociety = form.DonationToSociety     // 0.0 if no donation given.
	ms.DonationToMuseum = form.DonationToMuseum       // 0.0 if no donation given.

	// Create a list of hidden variables to drive the next request.
	hiddenVars := `
			<input type='hidden' name='user_id' value='{{.OrdinaryMemberID}}'>
`

	if form.Friend {
		// The ordinary member wants to be a friend of the museum.
		ms.OrdinaryMemberIsFriend = true
		ms.OrdinaryMemberFriendFee = form.FriendMembershipFee
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
		ms.AssociateMemberFee = form.AssociateMembershipFee
		hiddenVars += `
			<input type='hidden' name='assoc_user_id' value='{{.AssociateMemberID}}'>
`

		if form.AssociateIsFriend {
			// The associate member wants to be a friend of the museum.
			ms.AssocMemberIsFriend = true
			ms.AssociateMemberFriendFee = form.FriendMembershipFee
			hiddenVars += `
		<input type='hidden' name='assoc_friend' value='on'>
`
		}
	}

	if form.DonationToSociety > 0.0 {
		ms.DonationToSociety = form.DonationToSociety
		hiddenVars += `
			<input type='hidden' name='donation_to_society' value='{{.DonationToSocietyForDisplay}}'>
`
	}

	if form.DonationToMuseum > 0.0 {
		ms.DonationToMuseum = form.DonationToMuseum
		hiddenVars += `
			<input type='hidden' name='donation_to_museum' value='{{.DonationToMuseumForDisplay}}'>
`
	}

	insert := hdlr.createCostBreakdown(ms) + hiddenVars

	// Insert the cost breakdown and the hidden variables into the
	// shopping trolley page template.
	paymentConfirmationPageTemplateBody := fmt.Sprintf(paymentConfirmationPageTemplateStr, insert)

	// Check the template.
	paymentConfirmationPageTemplate, templateError :=
		template.New("PaymentConfirmationPage").Parse(paymentConfirmationPageTemplateBody)
	if templateError != nil {
		w.Write([]byte(errorHTML))
		return
	}

	// Write the response.
	executeError := paymentConfirmationPageTemplate.Execute(w, ms)

	if executeError != nil {
		w.Write([]byte(errorHTML))
		return
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

	txError := db.BeginTx()

	if txError != nil {
		reportError(w, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer db.Rollback()

	defer db.Close()

	startTime := time.Now()

	sale, fatal, warning := successHelper(db, stripeSession.ClientReferenceID, sessionID, startTime)

	if fatal != nil {
		// A fatal error is something that needs to be reported to the user -
		// the payment has been taken but we can't update the database records.

		reportError(w, fatal)
		db.Rollback()
		return
	}

	// There was no fatal error, so commit the transaction.
	db.Commit()

	if warning != nil {
		// A warning is for non-fatal issues.  Make a log entry but don't
		// show the user an error message.
		fmt.Printf("success:  %v", warning)
	}

	// Create the response page.

	insert := hdlr.createCostBreakdown(sale)

	// Insert the cost breakdown and the hidden variables into the template.
	successPageTemplateBody := fmt.Sprintf(successPageTemplateStr, insert)

	// Check the template.
	successPageTemplate, parseError :=
		template.New("SuccessPage").Parse(successPageTemplateBody)
	if parseError != nil {
		w.Write([]byte(errorHTML))
		return
	}

	// Write the response.
	executeError := successPageTemplate.Execute(w, &sale)

	if executeError != nil {
		w.Write([]byte(errorHTML))
		return
	}
}

// successHelper completes the sale.  It's separated out to support
// unit testing.
func successHelper(db *database.Database, salesIDstr, sessionID string, startDate time.Time) (sale *database.MembershipSale, fatal, warning error) {

	var salesID int
	_, salesIDError := fmt.Sscanf(salesIDstr, "%d", &salesID)
	if salesIDError != nil {
		fmt.Printf("successHelper: user ID %s - %v", salesIDstr, salesIDError.Error())
		return nil, salesIDError, nil
	}

	// Get the membership sales record.  The ClientReferenceID in the payment
	// session is the ID of the sales record.

	sale, fetchError := db.GetMembershipSale(salesID)
	if fetchError != nil {
		// The user has paid but we can't fulfill the sale so this
		// error is fatal.
		return nil, fetchError, nil
	}

	if sale.TransactionType == database.TransactionTypeNewMember {

		// This sale is for a new ordinary member and possibly a new associate member.
		// Create records for them.

		ordinaryMemberID, associateMemberID, createUserError := db.CreateAccounts(sale, time.Now())
		if createUserError != nil {
			return nil, createUserError, nil
		}

		sale.OrdinaryMemberID = ordinaryMemberID
		sale.AssociateMemberID = associateMemberID

	} else {
		// This sale is an ordinary member renewing, possibly with an associate.

		// For a new member, various fields are set at this point, so set them
		// for the renewing member(s) too.  The most important change is setting
		// the member end date, because that's what marks them as a paid up
		// member, which is what they've just paid for.

		// Set the end date for the ordinary member.
		omError := db.SetMemberEndDate(sale.OrdinaryMemberID, sale.MembershipYear)
		if omError != nil {
			return nil, omError, nil
		}

		if sale.AssociateMemberID > 0 {
			// Set the end date for the associate member.
			assocError := db.SetMemberEndDate(sale.AssociateMemberID, sale.MembershipYear)
			if assocError != nil {
				return nil, assocError, nil
			}
		}

		// Set the date last paid field.  It's for our accounting, so not so important.
		// If it fails, return a warning not a fatal error.
		now := time.Now()
		dlpError := db.SetDateLastPaid(sale.OrdinaryMemberID, now)
		if dlpError != nil {
			return sale, nil, dlpError
		}
	}

	// The ID of the user record of the ordinary member is in the sale record.  If
	// the sale includes an associate member, the ID of their user record is there too.

	// Count members and friends at this address.  Those values
	// will be written later.
	membersAtAddress := 1
	var friendsAtAddress int

	if sale.OrdinaryMemberIsFriend {
		friendsAtAddress++
	}

	if sale.AssociateMemberID > 0 {
		membersAtAddress++
		if sale.AssocMemberIsFriend {
			friendsAtAddress++
		}
	}

	// Changes after this point are for our own accounting records, so
	// the user doesn't need to know about any errors.  Return them as
	// warnings.

	// Update the date of the last payment.  This is important because
	// it's used for the giftaid calculation.
	paymentError := db.SetLastPayment(
		sale.OrdinaryMemberID, sale.TotalPayment())
	if paymentError != nil {
		em := fmt.Sprintf("error setting last payment for %d - %v",
			sale.OrdinaryMemberID, paymentError)

		return sale, nil, errors.New(em)
	}

	// Set the giftaid tick box, true or false.
	giftAidErr := db.SetGiftaidField(sale.OrdinaryMemberID, sale.Giftaid)
	if giftAidErr != nil {
		return sale, nil, giftAidErr
	}

	// Set the members at address and friends at address in the
	// ordinary member's record.
	setMembersError := db.SetMembersAtAddress(
		sale.OrdinaryMemberID, membersAtAddress)
	if setMembersError != nil {
		return sale, nil, setMembersError
	}

	setFriendsError := db.SetFriendsAtAddress(
		sale.OrdinaryMemberID, friendsAtAddress)
	if setFriendsError != nil {
		return sale, nil, setFriendsError
	}

	// If the member is a friend, tick the box.  In case it's
	// already set from last year but not this year, ensure that the
	// value in the DB record is reset.
	friendError := db.SetFriendField(
		sale.OrdinaryMemberID, sale.OrdinaryMemberIsFriend)
	if friendError != nil {
		return sale, nil, friendError
	}

	if sale.AssociateMemberID > 0 {

		// Set the Friend field for the associate member.
		friendError := db.SetFriendField(sale.AssociateMemberID, sale.AssocMemberIsFriend)
		if friendError != nil {
			em := fmt.Sprintf("error setting friend value for %d - %v",
				sale.AssociateMemberID, friendError)
			return sale, nil, errors.New(em)
		}

		setMembersError := db.SetMembersAtAddress(
			sale.AssociateMemberID, membersAtAddress)
		if setMembersError != nil {
			return sale, nil, setMembersError
		}

		setFriendsError := db.SetFriendsAtAddress(
			sale.AssociateMemberID, friendsAtAddress)
		if setFriendsError != nil {
			return sale, nil, setFriendsError
		}
	}

	// Update the user's donation to society.
	dsError := db.SetDonationToSociety(
		sale.OrdinaryMemberID, sale.DonationToSociety)
	if dsError != nil {
		em := fmt.Sprintf("error setting donation to society for %d - %v",
			sale.OrdinaryMemberID, dsError)
		return sale, nil, errors.New(em)
	}

	// Update the user's donation to museum.
	dmError := db.SetDonationToMuseum(
		sale.OrdinaryMemberID, sale.DonationToMuseum)
	if dmError != nil {
		em := fmt.Sprintf("error setting donation to museum for %d - %v",
			sale.OrdinaryMemberID, dmError)
		return sale, nil, errors.New(em)
	}

	// Update the membership sale record.  Do this last because
	updateError := sale.Update(db, "complete", sessionID)
	if updateError != nil {
		em := fmt.Sprintf("failed to update membership sales record %d", sale.ID)
		return sale, nil, errors.New(em)
	}

	// Success!
	return sale, nil, nil
}

// Cancel is the handler for the /cancel request.  Stripe makes that
// request when the payment is cancelled.
func (hdlr *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(cancelHTML))
}

// Checkout is the handler for the /checkout request.  It validates the
// HTTP parameters and, if valid, creates a MembershipSale record and
// redirects to the Stripe payment website.
func (hdlr *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	form := PaymentFormData{}
	userIDParam := r.PostFormValue("user_id")
	assocUserIDParam := r.PostFormValue("assoc_user_id")
	form.FriendInput = r.PostFormValue("friend")
	form.AssociateIsFriendInput = r.PostFormValue("assoc_friend")
	form.DonationToSocietyInput = r.PostFormValue("donation_to_society")
	form.DonationToMuseumInput = r.PostFormValue("donation_to_museum")
	form.GiftaidInput = r.PostFormValue("giftaid")

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

	if len(form.FriendInput) > 0 {
		// form.FriendOutput is marked as not used.  It's used in the
		// paymentPageTemplateStr but the compiler can't see that.
		form.Friend, form.FriendOutput = getTickBox(form.FriendInput)
		if form.Friend {
			friendFee = hdlr.FriendMembershipFee
		}
	}

	if len(form.AssociateIsFriendInput) > 0 {
		// form.AssociateIsFriendOutput is marked as not used.  It's used
		// in the paymentPageTemplateStr but the compiler can't see that.
		form.AssociateIsFriend, form.AssociateIsFriendOutput =
			getTickBox(form.AssociateIsFriendInput)
		if form.AssociateIsFriend {
			assocFriendFee = hdlr.FriendMembershipFee
		}
	}

	if len(form.GiftaidInput) > 0 {
		// form.GiftaidOutput is marked as not used.  It's used in the
		// paymentPageTemplateStr but the compiler can't see that.
		form.Giftaid, form.GiftaidOutput = getTickBox(form.GiftaidInput)
	}

	if len(form.DonationToSocietyInput) > 0 {

		e, v := checkDonation(form.DonationToSocietyInput)
		if len(e) > 0 {
			fmt.Println("checkout:", "donationToSociety - ", e)
			reportError(w, errors.New("donation to society - "+e))
		}

		form.DonationToSociety = v
	}

	if len(form.DonationToMuseumInput) > 0 {

		e, v := checkDonation(form.DonationToMuseumInput)
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
		AssocMemberIsFriend:      form.AssociateIsFriend,
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

	// Start a transaction, stored in the database object.
	txError := db.BeginTx()

	if txError != nil {
		reportError(w, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer db.Rollback()

	defer db.Close()

	salesID, createError := ms.Create(db)
	if createError != nil {
		db.Rollback()
		fmt.Println("checkout:", "CreateError: ", createError.Error())
		reportError(w, createError)
	}

	// We have all we need from the database - commit the transaction.
	db.Commit()

	// Prepare to pass control to the Stripe payment page.

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

	salesIDStr := fmt.Sprintf("%d", salesID)

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

	form := NewPaymentForm()
	form.MarkMandatoryFields()
	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		fmt.Println(executeError)
	}
}

// createCostBreakdown creates an HTML table showing the cost of the
// membership sale.  The table is inserted into a template.
func (hdlr *Handler) createCostBreakdown(ms *database.MembershipSale) string {

	table := `
<table>
	<tr>
	    <td style='border: 0'>ordinary membership</td>
		<td style='border: 0'>£{{.OrdinaryMembershipFeeForDisplay}}</td>
	</tr>
`

	if ms.OrdinaryMemberIsFriend {
		table += `
	<tr>
	    <td style='border: 0'>friend of the museum</td>
		<td style='border: 0'>£{{.OrdinaryMemberFriendFeeForDisplay}}</td>
	</tr>
`
	}

	if ms.AssociateMemberID > 0 {
		table += `
	<tr>
	    <td style='border: 0'>associate member</td>
		<td style='border: 0'>£{{.AssociateMembershipFeeForDisplay}}</td>
	</tr>		
`
	}

	if ms.AssocMemberIsFriend {
		table += `
	<tr>
	    <td style='border: 0'>associate is friend of the museum</td>
		<td style='border: 0'>£{{.AssociateMemberFriendFeeForDisplay}}</td>
	</tr>
`
	}

	if ms.DonationToSociety > 0 {
		table += `
	<tr>
	    <td style='border: 0'>donation to the society</td>
		<td style='border: 0'>£{{.DonationToSocietyForDisplay}}</td>
	</tr>		
`
	}

	if ms.DonationToMuseum > 0 {
		table += `
	<tr>
	    <td style='border: 0'>donation to the museum</td>
		<td style='border: 0'>£{{.DonationToMuseumForDisplay}}</td>
	</tr>
	`
	}

	table += `
	<tr>
	    <td style='border: 0'><b>Total</b></td>
		<td style='border: 0'>£{{.TotalPaymentForDisplay}}</td>
	</tr>
`

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
	// form.OrdinaryMemberFee, form.AssociateMemberFee, form.FriendFee =
	// 	MustSetFees(form.OrdinaryMemberFeeStr, form.AssociateMemberFeeStr, form.FriendFeeStr)

	// form.Valid starts true and is set false if any of the form data is invalid.
	form.Valid = true

	if len(form.FirstName) == 0 &&
		len(form.LastName) == 0 &&
		len(form.Email) == 0 &&
		len(form.FriendInput) == 0 &&
		len(form.AssocFirstName) == 0 &&
		len(form.AssocLastName) == 0 &&
		len(form.AssocEmail) == 0 &&
		len(form.AssociateIsFriendInput) == 0 &&
		len(form.DonationToSocietyInput) == 0 &&
		len(form.DonationToMuseumInput) == 0 &&
		len(form.GiftaidInput) == 0 {

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
	form.FriendInput = strings.TrimSpace(form.FriendInput)
	form.DonationToSocietyInput = strings.TrimSpace(form.DonationToSocietyInput)
	form.DonationToMuseumInput = strings.TrimSpace(form.DonationToMuseumInput)
	form.GiftaidInput = strings.TrimSpace(form.GiftaidInput)

	form.AssocFirstName = strings.TrimSpace(form.AssocFirstName)
	form.AssocLastName = strings.TrimSpace(form.AssocLastName)
	form.AssocEmail = strings.TrimSpace(form.AssocEmail)
	form.AssociateIsFriendInput = strings.TrimSpace(form.AssociateIsFriendInput)

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

	// The associate fields are optional but if you fill in any
	// of them, you must fill in the first and last name.
	if len(form.AssocFirstName) > 0 ||
		len(form.AssocLastName) > 0 ||
		len(form.AssocEmail) > 0 ||
		len(form.AssociateIsFriendInput) > 0 {

		if len(form.AssocFirstName) == 0 {
			form.AssocFirstNameErrorMessage = assocFirstNameErrorMessage
			form.Valid = false
		}
		if len(form.AssocLastName) == 0 {
			form.AssocLastNameErrorMessage = assocLastNameErrorMessage
			form.Valid = false
		}
	}

	form.Friend, form.FriendOutput = getTickBox(form.FriendInput)

	form.Giftaid, form.GiftaidOutput = getTickBox(form.GiftaidInput)

	form.AssociateIsFriend, form.AssociateIsFriendOutput =
		getTickBox(form.AssociateIsFriendInput)

	// The mandatory parameters are all present.  Now check the contents.

	// If donation values are submitted, they must be numbers and not
	// negative.

	if len(form.DonationToSocietyInput) > 0 {

		errorMessage, dts := checkDonation(form.DonationToSocietyInput)
		if len(errorMessage) > 0 {
			form.DonationToSocietyErrorMessage = errorMessage
			form.Valid = false
		} else {
			form.DonationToSociety = dts
		}
	}

	if len(form.DonationToMuseumInput) > 0 {

		errorMessage, dtm := checkDonation(form.DonationToMuseumInput)
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

// validate does a complete validation of the form.  It calls simpleValidate takes the form parameters as arguments.  It looks up the
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

// getTickBox returns true and "checked" if the tickbox is ticked ("on"),
// false and "unchecked" otherwise.
func getTickBox(value string) (bool, string) {
	switch {
	case len(value) == 0:
		return false, "unchecked"
	case value == "on":
		return true, "checked"
	default:
		return false, "unchecked"
	}
}

func reportError(w http.ResponseWriter, err error) {

	fmt.Printf("error %v", err.Error())
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
			<li>Ordinary member: £{{.OrdinaryMembershipFeeForDisplay}}</li>
			<li>Associate member at the same address: £{{.AssociateMembershipFeeForDisplay}}</li>
			<li>Friend of the Leatherhead museum: £{{.FriendMembershipFeeForDisplay}}</li>
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
					<td style='border: 0'>Friend of the Museum:</td>
					<td style='border: 0; '>
						<input style='transform: scale(1.5);' type='checkbox' name='friend' {{.FriendOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0'>Donation to the Society:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_society' value='{{.DonationToSocietyForDisplay}}'></td>
					<td style='border: 0;'><span style="color:red;">{{.DonationToSocietyErrorMessage}}</span></td>
				</tr>
				<tr>
					<td style='border: 0'>Donation to the Museum:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_museum' value='{{.DonationToMuseumForDisplay}}'></td>
					<td style='border: 0'><span style="color:red;">{{.DonationToMuseumErrorMessage}}</span></td>
				</tr>
				<tr>
					<td style='border: 0'>Giftaid:</td>
					<td style='border: 0 '>
						<input style='transform: scale(1.5);' type='checkbox' name='giftaid' {{.GiftaidOutput}} />
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
					<td style='border: 0'>Friend of the Museum:</td>
					<td style='border: 0'>
						<input style='transform: scale(1.5);' type='checkbox' name='assoc_friend' {{.AssociateIsFriendOutput}} />
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

const errorHTML = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%'>
		<p>
			Something went wrong after you had paid.
			Please
			<a href="mailto:treasurer@leatherheadhistory.org">
				email the treasurer
			</a>
		</p>
    </body>
</html>
`
