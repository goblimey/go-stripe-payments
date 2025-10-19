package handler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"

	"github.com/goblimey/go-stripe-payments/code/pkg/config"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
	"github.com/goblimey/go-stripe-payments/code/pkg/forms"
)

var paymentPageTemplate *template.Template

func init() {
	// Check the response HTML templates.
	paymentPageTemplate = template.Must(template.New("paymentFormTemplate").
		Parse(paymentPageTemplateStr))
}

type Handler struct {
	Conf                   *config.Config     // The incoming config.
	DBConfig               *database.DBConfig // The database config
	DB                     *database.Database // The database connection.
	MembershipYear         int                // The membership year we are currently selling.
	OrdinaryMembershipFee  float64            // The fee for ordinary membership.
	AssociateMembershipFee float64            // The fee for associate membership (0 if not enabled).
	FriendMembershipFee    float64            // The fee for friend's membership (0 if not enabled).
	Protocol               string             // "http" or "https"
	PrePaymentErrorHTML    string             // The default error message page before the customer pays.
	PostPaymentErrorHTML   string             // The default error message page after the customer has paid.
	SuccessPageHTML        string             // The page displayed on a successful sale.
	Logger                 *slog.Logger       // The daily logger.
}

func New(conf *config.Config) *Handler {

	var protocol string
	if conf.HTTP {
		protocol = "http" // Testing with HTTP.
	} else {
		protocol = "https" // HTTPS - probably production.
	}

	dbConfig := database.DBConfig{
		Type: conf.DBType,
		Host: conf.DBHostname,
		Port: conf.DBPort,
		Name: conf.DBDatabase,
		User: conf.DBUser,
		Pass: conf.DBPassword,
	}

	prePaymentErrorHTML := fmt.Sprintf(prePaymentErrorHTMLPattern, conf.EmailAddressForQuestions, conf.EmailAddressForQuestions)
	// If things go wrong after the customer has paid they should be referred to somebody
	// who can refund their payment, for example the treasurer.
	postPaymentErrorHTML := fmt.Sprintf(postPaymentErrorHTMLPattern, conf.EmailAddressForFailures, conf.EmailAddressForFailures)

	h := Handler{
		Conf:                   conf,
		DBConfig:               &dbConfig,
		OrdinaryMembershipFee:  conf.OrdinaryMemberFee,
		AssociateMembershipFee: conf.AssocMemberFee,
		FriendMembershipFee:    conf.FriendFee,
		Protocol:               protocol,
		PrePaymentErrorHTML:    prePaymentErrorHTML,
		PostPaymentErrorHTML:   postPaymentErrorHTML,
	}

	return &h
}

// Home handles the "/" request.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("home")

	w.Write([]byte("home page\n"))
}

// GetPaymentData handles the /displayPaymentForm request.
// It validates the incoming payment data form.  If the data
// is valid it displays the cost breakdown, otherwise it
// displays the payment data form again with error messages.
func (h *Handler) GetPaymentData(w http.ResponseWriter, r *http.Request) {

	h.Logger.Info("GetPaymentData")

	now := time.Now()
	membershipYear := database.GetMembershipYear(now)
	form := forms.NewSaleForm(h.Conf, membershipYear, now)

	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		fmt.Println(connectionError.Error())
		form.GeneralErrorMessage = fmt.Sprintf("Fatal error - %v", connectionError)
		form.Valid = false
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction, stored in the database object.
	txError := h.DB.BeginTx()

	if txError != nil {
		form.Valid = false
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already might be caused by some
	// sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()

	defer h.DB.Close()

	// The helper does the work.
	h.paymentDataHelper(w, r)

	// paymentDataHelper doesn't change the database so we can just
	// close the transaction via a rollback.
	h.DB.Rollback()
}

// GetPaymentDataHelper validates the form and prepares the response.
// It's separated out to support unit testing.
func (h *Handler) paymentDataHelper(w http.ResponseWriter, r *http.Request) {

	now := time.Now()
	membershipYear := database.GetMembershipYear(now)
	sf := forms.NewSaleForm(h.Conf, membershipYear, now)

	sf.OrdinaryMemberFee = h.Conf.OrdinaryMemberFee
	sf.AssocMemberFee = h.Conf.AssocMemberFee
	sf.FriendFee = h.Conf.FriendFee

	sf.FirstName = r.PostFormValue("first_name")
	sf.LastName = r.PostFormValue("last_name")
	sf.Email = r.PostFormValue("email")
	sf.FriendInput = r.PostFormValue("friend")
	sf.DonationToSocietyInput = r.PostFormValue("donation_to_society")
	sf.DonationToMuseumInput = r.PostFormValue("donation_to_museum")
	sf.GiftaidInput = r.PostFormValue("giftaid")

	sf.AssocFirstName = r.PostFormValue("assoc_first_name")
	sf.AssocLastName = r.PostFormValue("assoc_last_name")
	sf.AssocEmail = r.PostFormValue("assoc_email")
	sf.AssocFriendInput = r.PostFormValue("assoc_friend")

	if len(sf.FirstName) == 00 &&
		len(sf.LastName) == 0 &&
		len(sf.Email) == 0 &&
		len(sf.DonationToSocietyInput) == 0 &&
		len(sf.DonationToMuseumInput) == 0 &&
		len(sf.AssocFirstName) == 0 &&
		len(sf.AssocLastName) == 0 {

		// On the first call in a sequence, display an empty form with mandatory fields marked.
		h.displayInitialSaleForm(w)
		return
	}

	// Validate the form data.  On the first call the form is empty.
	// The validator sets error messages containing asterisks against
	// the mandatory fields.  On calls with incoming data, it validates
	// that data and sets error messages.

	valid := Validate(sf)

	if !valid {

		// There are errors, display the form again
		// with any supplied fields filled in.
		displaySaleForm(w, sf)

		return
	}

	// Build and display the payment confirmation page.

	if sf.Friend {
		// The ordinary members is a friend so a fee is due.
		sf.FriendFeeToPay = h.Conf.FriendFee
	}

	if len(sf.AssocFirstName) > 0 {

		// There s anassociate, so a fee is due.
		sf.AssocFeeToPay = h.Conf.AssocMemberFee

		if sf.AssocFriend {
			//The associate is a friend of the museum, so a fee is due.
			sf.AssocFriendFeeToPay = h.Conf.FriendFee
		}
	}

	// Check the template.
	paymentConfirmationPageTemplate, templateError :=
		template.New("PaymentConfirmationPage").Parse(paymentConfirmationPageTemplate)
	if templateError != nil {
		h.Logger.Info(templateError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Write the response.
	executeError := paymentConfirmationPageTemplate.Execute(w, sf)

	if executeError != nil {
		h.Logger.Info(executeError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}
}

// Checkout is the handler for the /checkout request.  It validates the
// HTTP parameters and, if valid, creates a MembershipSale record and
// redirects to the Stripe payment website.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {

	h.Logger.Info("Checkout")

	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectionError := h.DB.Connect()
	if connectionError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connectionError)
		return
	}

	// Start a transaction, stored in the database object.
	txError := h.DB.BeginTx()

	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()
	defer h.DB.Close()

	h.checkoutHelper(w, r)

	//The helper writes an HTTP response so we should never get to here.
}

// checkoutHelper creates a MembershipSale record to record progress and prepares the response.
// On success that's a redirect to the Stripe payment system.  The helper is separated out to
// support unit testing.
func (h *Handler) checkoutHelper(w http.ResponseWriter, r *http.Request) {

	// In theory everything has been validated but we can't assume that the data is the
	// result of the user following the expected flow of pages.  They may be trying to
	// pull a fast one, so o we validate some stuff again.  If there is any error, we stop
	// processing.

	ms := database.NewMembershipSale(h.Conf)

	// Convert and check the donation values - if specified they must both be POSITIVE
	// numbers.
	ds := strings.TrimSpace(r.PostFormValue("donation_to_society"))
	if len(ds) > 0 {
		// Convert the donation input string to a float.
		var em string
		em, ms.DonationToSociety = checkDonation(ds)
		if len(em) != 0 {
			// The data should already have been validated so this should never happen.
			h.logMessage("CheckoutHelper: %v\n", em)
			h.reportError(w, h.PostPaymentErrorHTML, errors.New("internal error"))
		}
	}

	dm := strings.TrimSpace(r.PostFormValue("donation_to_museum"))
	if len(dm) > 0 {
		// Convert the donation input string to a float.
		var em string
		em, ms.DonationToMuseum = checkDonation(dm)
		if len(em) > 0 {
			// The data should already have been validated so this should never happen.
			h.logMessage("CheckoutHelper: %v\n", em)
			h.reportError(w, h.PostPaymentErrorHTML, errors.New("internal error"))
		}
	}

	// The incoming data is valid.  Complete and commit the membership_sales record.
	now := time.Now()
	ms.MembershipYear = database.GetMembershipYear(now)
	ms.PaymentService = "Stripe"
	ms.PaymentStatus = "Pending"
	ms.FirstName = strings.TrimSpace(r.PostFormValue("first_name"))
	ms.LastName = strings.TrimSpace(r.PostFormValue("last_name"))
	ms.Email = strings.TrimSpace(r.PostFormValue("email"))
	ms.Friend, _, _ = getTickBox(r.PostFormValue("friend"))
	ms.AssocFriend, _, _ = getTickBox(r.PostFormValue("assoc_friend"))
	ms.Giftaid, _, _ = getTickBox(r.PostFormValue("giftaid"))
	ms.AssocFirstName = strings.TrimSpace(r.PostFormValue("assoc_first_name"))
	ms.AssocLastName = strings.TrimSpace(r.PostFormValue("assoc_last_name"))
	ms.AssocEmail = strings.TrimSpace(r.PostFormValue("assoc_email"))
	ms.PaymentStatus = database.PaymentStatusPending

	if ms.EnableOtherMemberTypes {
		ms.OrdinaryMemberFeePaid = h.OrdinaryMembershipFee
		if ms.Friend {
			// The ordinary member is a friend so must pay the friend fee.
			ms.FriendFeePaid = h.FriendMembershipFee
		}

		if len(ms.AssocFirstName) > 0 {
			// There is an associate member - another fee.
			ms.AssocFeePaid = h.AssociateMembershipFee

			if ms.AssocFriend {
				ms.AssocFriendFeePaid = h.FriendMembershipFee
			}
		}
	}

	salesID, createError := ms.Create(h.DB)
	if createError != nil {
		h.DB.Rollback()
		h.logMessage("checkout: CreateError - %v", createError)
		h.reportError(w, h.PrePaymentErrorHTML, createError)
	}

	// We have all we need from the database - commit the transaction.
	h.DB.Commit()

	// Prepare to pass control to the Stripe payment page.

	successURL := fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}", h.Protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", h.Protocol, r.Host)

	invoicingEnabled := true

	description := fmt.Sprintf(
		"%s membership %d", h.Conf.OrganisationName, h.MembershipYear)

	invoiceData := stripe.CheckoutSessionInvoiceCreationInvoiceDataParams{
		Description: &description,
	}

	invoiceCreation := stripe.CheckoutSessionInvoiceCreationParams{
		Enabled:     &invoicingEnabled,
		InvoiceData: &invoiceData,
	}

	// Create a Stripe checkout session containing a client reference ID that gives
	// the ms_id of the membership_sales record.  This allows the application to
	// pick up where it left off when control is returned from Stripe.

	salesIDStr := fmt.Sprintf("%d", salesID)

	priceInPennies := int64(ms.Total()*100 + 0.5)

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
		// Stripe will request this URL if the payment is successful. The
		// {CHECKOUT_SESSION_ID} placeholder will be replaced by the session ID, which
		// allows the handler to retrieve the session.
		SuccessURL: stripe.String(successURL),
		// Stripe will request this URL if the payment if cancelled.
		CancelURL: stripe.String(cancelURL),
	}

	// Create the checkout session.
	s, sessErr := session.New(params)
	if sessErr != nil {

		h.DB.Rollback()
		h.logMessage("error creating Stripe session - %v", sessErr)
		h.reportError(w, h.PrePaymentErrorHTML, sessErr)
	}

	// Redirect to the Stripe system.  On a successful payment, it will
	// redirect to /success and we will continue.
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// CreateCheckoutSession is the handler for the /create-checkout-session
// request.  It prepares the Stripe session and redirects the browser to
// the Stripe payment page.
func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	successURL :=
		fmt.Sprintf("%s://%s/success?session_id={CHECKOUT_SESSION_ID}",
			h.Protocol, r.Host)
	cancelURL := fmt.Sprintf("%s://%s/cancel", h.Protocol, r.Host)

	invoiceEnabled := true
	year := database.GetMembershipYear(time.Now())
	description := fmt.Sprintf("%s membership system %d", h.Conf.OrganisationName, year)
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
		h.logMessage("/create-checkout-session: error creating stripe session: %v", err)
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// Success is the handler for the /success request.  On a successful
// payment, the Stripe system issues that request, filling in the
// {CHECKOUT_SESSION_ID} placeholder with the session ID.  The
// handler uses that to retrieve the checkout session, extract the
// client reference and complete the sale.
func (h *Handler) Success(w http.ResponseWriter, r *http.Request) {

	h.logMessage("Success()\n")

	// Get the Stripe session.
	sessionID := r.URL.Query().Get("session_id")
	params := stripe.CheckoutSessionParams{}
	stripeSession, sessionGetError := session.Get(sessionID, &params)
	if sessionGetError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, sessionGetError)
		return
	}

	connError := h.connectToDB()
	if connError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, connError)
		return
	}

	// The transaction should be ether committed or rolled back before this function exits.
	// In case it isn't, set up a deferred rollback now.  We choose a rollback rather than
	// a commit because failure to close the transaction already is almost certainly caused
	// by some sort of catastrophic error.
	//
	// When the function is returning, the transaction may have already been closed, in which
	// case the second rollback may return an error, but that will be ignored.
	defer h.DB.Rollback()

	defer h.DB.Close()

	// We figure out the start and end dates here to support unit testing of the SuccessHelper.
	startTime := time.Now()
	// The end date is the end of the calendar year of payment.
	londonTime, tzError := time.LoadLocation("Europe/London")
	if tzError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, tzError)
		return
	}
	yearEnd := time.Date(startTime.Year(), time.December, 31, 23, 59, 59, 999999999, londonTime)

	h.successHelper(w, r, stripeSession, sessionID, startTime, yearEnd, time.Now())

	// The helper should send an HTTP response so we shouldn't get to here.
}

// successHelper completes the sale.  It's separated out and the start and end dates are supplied to
// support unit testing.
func (h *Handler) successHelper(w http.ResponseWriter, r *http.Request, stripeSession *stripe.CheckoutSession, sessionID string, startDate, endDate, paymentDate time.Time) {

	h.logMessage("successHelper: year ending %v", endDate)

	if stripeSession.PaymentStatus != "paid" {
		e := fmt.Errorf("payment not made")
		h.reportError(w, h.PostPaymentErrorHTML, e)
		h.DB.Rollback()
		return
	}

	var salesID int64
	_, salesIDError := fmt.Sscanf(stripeSession.ClientReferenceID, "%d", &salesID)
	if salesIDError != nil {
		// Can't fetch the sales ID from the session.  This is bad, as the user has paid for a sale
		// we can't honour.  Report it to the user.
		e := fmt.Errorf("successHelper: error converting sales ID %s - %v", stripeSession.ClientReferenceID, salesIDError.Error())
		h.reportError(w, h.PostPaymentErrorHTML, e)
		h.DB.Rollback()
		return
	}

	// Get the membership sales record.  The ClientReferenceID in the payment
	// session is the ID of the sales record.
	sale, fetchError := h.DB.GetMembershipSale(salesID)
	if fetchError != nil {
		// The user has paid but we can't fulfill the sale so this error is bad.  Report
		// it to the user.
		h.DB.Rollback()
		h.reportError(w, h.PostPaymentErrorHTML, fetchError)
		return
	}

	// Add the reference data (quoted by the HTML pages)
	sale.OrganisationName = h.Conf.OrganisationName
	sale.EnableOtherMemberTypes = h.Conf.EnableOtherMemberTypes
	sale.EnableGiftaid = h.Conf.EnableGiftaid
	sale.EmailAddressForQuestions = h.Conf.EmailAddressForQuestions
	sale.EmailAddressForFailures = h.Conf.EmailAddressForFailures

	// The user must have paid at least this fee.
	sale.OrdinaryMemberFeePaid = h.Conf.OrdinaryMemberFee

	// The sale is complete.
	sale.PaymentStatus = database.PaymentStatusComplete

	sale.PaymentID = fmt.Sprintf("%s %s", stripeSession.Customer.ID, stripeSession.Customer.Email)

	// Membership renewal or new member(s)?
	var lookupError error
	sale.UserID, sale.AssocUserID, lookupError = usersExist(sale, h.DB)
	if lookupError != nil {
		h.DB.Rollback()
		h.reportError(w, h.PrePaymentErrorHTML, lookupError)
		return
	}

	saleType := "new member"
	if sale.UserID > 0 {
		// Membership renewal
		saleType = "membership renewal"
	}
	h.logMessage("SuccessHelper: user ID %s %s %s %s %s %s %s\n",
		saleType, sale.FirstName, sale.LastName, sale.Email,
		sale.AssocFirstName, sale.AssocLastName, sale.AssocEmail)

	if sale.UserID == 0 {
		// A new member is registering.  We haven't created
		// the member account(s) yet.
		sale.TransactionType = database.TransactionTypeNewMember
		userID, assocUserID, createUserError :=
			h.DB.CreateAccounts(sale, startDate, endDate)
		if createUserError != nil {
			// Failed to create one or both of the users.  The user has paid but we
			// can't fulfill the sale so this error is bad.  Report it to the user.
			h.DB.Rollback()
			h.reportError(w, h.PostPaymentErrorHTML, createUserError)
			return
		}

		h.logMessage("created users %d %d\n", userID, assocUserID)
		sale.UserID = userID
		sale.AssocUserID = assocUserID
	} else {
		// An existing member is renewing, possibly with an associate member.
		sale.TransactionType = database.TransactionTypeRenewal

		// For a new member, various fields are set at this point, so set them
		// for the renewing member(s) too.  The most important change is setting
		// the member end date, because that's what marks them as a paid up
		// member, which is what they've just paid for.

		// Set the end date for the ordinary member.
		omError := h.DB.SetMemberEndDate(sale.UserID, sale.MembershipYear)
		if omError != nil {
			h.DB.Rollback()
			h.reportError(w, h.PostPaymentErrorHTML, omError)
			return
		}

		if h.Conf.EnableOtherMemberTypes && sale.AssocUserID > 0 {
			// Set the end date for the associate member.
			assocError := h.DB.SetMemberEndDate(sale.AssocUserID, sale.MembershipYear)
			if assocError != nil {
				h.DB.Rollback()
				h.reportError(w, h.PostPaymentErrorHTML, assocError)
				return
			}
		}
	}

	// We've done the important update.  In case something catastrophic happens later,
	// commit the changes made so far and then open a new transaction.

	updateError1 := sale.Update(h.DB)
	if updateError1 != nil {

		h.logMessage("successHelper: user ID %d - failed to update membership sales record %d - %v",
			sale.UserID, sale.ID, updateError1)
	}

	commit1Error := h.DB.Commit()
	if commit1Error != nil {
		h.DB.Rollback()
		h.reportError(w, h.PostPaymentErrorHTML, commit1Error)
		return
	}

	txError := h.DB.BeginTx()
	if txError != nil {
		h.reportError(w, h.PrePaymentErrorHTML, txError)
		return
	}

	// The rest of the data we are going to set is for our own accounting.  If we get
	// an error, the user is less bothered.  Just log it or return it as a warning.

	dlpError := h.DB.SetDateLastPaid(sale.UserID, paymentDate)
	if dlpError != nil {
		h.logMessage("successHelper: user ID %d - %v\n", sale.UserID, dlpError)
	}

	// The ID of the user record of the ordinary member is in the sale record.  If
	// the sale includes an associate member, the ID of their user record is there too.

	// Count members and friends at this address.  Those values will be written later
	// to the ordinary and (if present) the associate.
	membersAtAddress := 1
	var friendsAtAddress int

	if sale.Friend {
		friendsAtAddress++
	}

	if h.Conf.EnableOtherMemberTypes && sale.AssocUserID > 0 {
		membersAtAddress++
		if sale.AssocFriend {
			friendsAtAddress++
		}
	}

	paymentError := h.DB.SetLastPayment(sale.UserID, sale.Total())
	if paymentError != nil {
		em := fmt.Sprintf("error setting last payment for %d - %v",
			sale.UserID, paymentError)
		h.logMessage("successHelper: user ID %d - %s", sale.UserID, em)
	}

	// Set the members at address and friends at address in the ordinary member's record.
	setMembersError := h.DB.SetMembersAtAddress(sale.UserID, membersAtAddress)
	if setMembersError != nil {
		h.logMessage("successHelper: user ID %d - %v", sale.UserID, setMembersError)
	}

	if h.Conf.EnableGiftaid && sale.Giftaid {
		// Set the giftaid tick box, true or false.
		giftAidError := h.DB.SetGiftaidField(sale.UserID, sale.Giftaid)
		if giftAidError != nil {
			h.logMessage("successHelper: user ID %d - %v\n", sale.UserID, giftAidError)
		}
	}

	setFriendsError := h.DB.SetFriendsAtAddress(sale.UserID, friendsAtAddress)
	if setFriendsError != nil {
		h.logMessage("successHelper: user ID %d - %v", sale.UserID, setFriendsError)
	}

	// If the member is a friend, tick the box.  The user may have been a friend last
	// year and so the record in the DB will be ticked.  The user may not be a friend
	// this year, so always reset the value.
	friendError := h.DB.SetFriendField(
		sale.UserID, sale.Friend)
	if friendError != nil {
		h.logMessage("successHelper: user ID %d - %v\n", sale.UserID, friendError)
	}

	// Update the user's donation to society.
	dsError := h.DB.SetDonationToSociety(sale.UserID, sale.DonationToSociety)
	if dsError != nil {
		e := fmt.Errorf("error setting donation to society for %d - %v",
			sale.UserID, dsError)
		h.logMessage("successHelper: user ID %d - %v\n", sale.UserID, e)
	}

	// Update the user's donation to museum.
	dmError := h.DB.SetDonationToMuseum(sale.UserID, sale.DonationToMuseum)
	if dmError != nil {
		e := fmt.Errorf("error setting donation to museum for %d - %v",
			sale.UserID, dmError)
		h.logMessage("successHelper: user ID %d - %v\n", sale.UserID, e)
	}

	if h.Conf.EnableOtherMemberTypes && sale.AssocUserID > 0 {
		// Associate members are enabled and there is one.  Set the Friend field in the
		// associate member's record.
		setFriendsError := h.DB.SetFriendField(sale.AssocUserID, sale.AssocFriend)
		if setFriendsError != nil {
			e := fmt.Errorf("error setting friend value for %d - %v", sale.AssocUserID, friendError)
			h.logMessage("successHelper: user ID %d - %v\n", sale.AssocUserID, e)
		}

		// Set the members at address in the associate member's record.
		setMembersError := h.DB.SetMembersAtAddress(sale.AssocUserID, membersAtAddress)
		if setMembersError != nil {
			h.logMessage("successHelper: user ID %d - %v", sale.AssocUserID, setMembersError)
		}
	}

	// Commit the last few changes.
	commit2Error := h.DB.Commit()
	if commit2Error != nil {
		h.logMessage("successHelper: user ID %d - %v", sale.UserID, commit2Error)
	}

	// Create the response page.

	// Check the template.
	successPageTemplate, parseError := template.New("SuccessPage").Parse(successPageTemplateStr)
	if parseError != nil {
		h.Logger.Info(parseError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Write the response.
	executeError := successPageTemplate.Execute(w, sale)
	if executeError != nil {
		h.Logger.Info(executeError.Error())
		w.Write([]byte(h.PrePaymentErrorHTML))
		return
	}

	// Success!
}

// Cancel is the handler for the /cancel request.  Stripe makes that
// request when the payment is cancelled.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(cancelHTML))
}

func (h *Handler) connectToDB() error {
	h.DB = database.New(h.DBConfig)
	h.DB.Logger = h.Logger
	connectError := h.DB.Connect()
	if connectError != nil {
		return connectError
	}

	txError := h.DB.BeginTx()
	if txError != nil {
		return txError
	}

	return nil
}

func (hdlr *Handler) reportError(w http.ResponseWriter, errorHTML string, err error) {

	hdlr.Logger.Error(err.Error())
	w.Write([]byte(errorHTML))
}

// Fatal logs a fatal error to the structured log and exits.
func (hdlr *Handler) Fatal(err error) {
	hdlr.Logger.Error(err.Error())
	os.Exit(-1)
}

// DisplayPaymentForm displays the given payment form
func displaySaleForm(w io.Writer, form *forms.SaleForm) {

	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		fmt.Println(executeError)
	}
}

// DisplayInitialSaleForm displays an empty payment form
// with the mandatory parameters marked with asterisks.
func (h *Handler) displayInitialSaleForm(w io.Writer) {

	now := time.Now()
	membershipYear := database.GetMembershipYear(now)
	form := forms.NewSaleForm(h.Conf, membershipYear, now)
	form.MarkMandatoryFields()
	executeError := paymentPageTemplate.Execute(w, form)

	if executeError != nil {
		h.Logger.Info(executeError.Error())
		fmt.Println(executeError)
	}
}

func (h *Handler) logMessage(pattern string, a ...any) {
	str := fmt.Sprintf(pattern, a...)
	h.Logger.Info(str)
}

// Validation error messages - factored out to support unit testing.
const firstNameErrorMessage = "You must fill in the first name"
const lastNameErrorMessage = "You must fill in the last name"
const emailErrorMessage = "You must fill in the email address"
const assocFirstNameErrorMessage = "If you fill in anything in this section, you must fill in the first name"
const assocLastNameErrorMessage = "If you fill in anything in this section, you must fill in the last name"
const invalidNumber = "must be a number"
const negativeNumber = "must be a 0 or greater"

// Validate takes the form parameters as arguments.  It returns true
// and all empty strings if the form is valid, false and the error messages set
// if it's invalid.
func Validate(sf *forms.SaleForm) bool {

	// Set the fees in the form.
	// form.OrdinaryMemberFee, form.AssociateMemberFee, form.FriendFee =
	// 	MustSetFees(form.OrdinaryMemberFeeStr, form.AssociateMemberFeeStr, form.FriendFeeStr)

	// form.Valid starts true and is set false if any of the form data is invalid.
	sf.Valid = true

	if len(sf.FirstName) == 0 &&
		len(sf.LastName) == 0 &&
		len(sf.Email) == 0 &&
		len(sf.FriendInput) == 0 &&
		len(sf.AssocFirstName) == 0 &&
		len(sf.AssocLastName) == 0 &&
		len(sf.AssocEmail) == 0 &&
		len(sf.AssocFriendInput) == 0 &&
		len(sf.DonationToSocietyInput) == 0 &&
		len(sf.DonationToMuseumInput) == 0 &&
		len(sf.GiftaidInput) == 0 {

		// On the first call the form is empty.  Mark the mandatory fields.
		// Return false and the handler will display the form
		// again, with the marks.
		sf.MarkMandatoryFields()
		sf.Valid = false
		return sf.Valid
	}

	sf.FirstName = strings.TrimSpace(sf.FirstName)
	sf.LastName = strings.TrimSpace(sf.LastName)
	sf.Email = strings.TrimSpace(sf.Email)
	sf.FriendInput = strings.TrimSpace(sf.FriendInput)
	sf.DonationToSocietyInput = strings.TrimSpace(sf.DonationToSocietyInput)
	sf.DonationToMuseumInput = strings.TrimSpace(sf.DonationToMuseumInput)
	sf.GiftaidInput = strings.TrimSpace(sf.GiftaidInput)
	sf.AssocFirstName = strings.TrimSpace(sf.AssocFirstName)
	sf.AssocLastName = strings.TrimSpace(sf.AssocLastName)
	sf.AssocEmail = strings.TrimSpace(sf.AssocEmail)
	sf.AssocFriendInput = strings.TrimSpace(sf.AssocFriendInput)

	if len(sf.FirstName) == 0 {
		sf.FirstNameErrorMessage = firstNameErrorMessage
		sf.Valid = false
	}

	if len(sf.LastName) == 0 {
		sf.LastNameErrorMessage = lastNameErrorMessage
		sf.Valid = false
	}

	if len(sf.Email) == 0 {
		sf.EmailErrorMessage = emailErrorMessage
		sf.Valid = false
	}

	// The associate fields are optional but if you fill in any
	// of them, you must fill in the first and last name.
	// AssociateFriendInput may be "" or "off".
	if len(sf.AssocFirstName) > 0 ||
		len(sf.AssocLastName) > 0 ||
		len(sf.AssocEmail) > 0 ||
		sf.AssocFriendInput == "on" {

		if len(sf.AssocFirstName) == 0 {
			sf.AssocFirstNameErrorMessage = assocFirstNameErrorMessage
			sf.Valid = false
		}
		if len(sf.AssocLastName) == 0 {
			sf.AssocLastNameErrorMessage = assocLastNameErrorMessage
			sf.Valid = false
		}
	}

	sf.Friend, sf.FriendInput, sf.FriendOutput = getTickBox(sf.FriendInput)

	sf.Giftaid, sf.GiftaidInput, sf.GiftaidOutput = getTickBox(sf.GiftaidInput)

	sf.AssocFriend, sf.AssocFriendInput, sf.AssocFriendOutput = getTickBox(sf.AssocFriendInput)

	// The mandatory parameters are all present.  Now check the contents.

	// If donation values are submitted, they must be numbers and not
	// negative.

	if len(sf.DonationToSocietyInput) > 0 {

		errorMessage, dts := checkDonation(sf.DonationToSocietyInput)
		if len(errorMessage) > 0 {
			sf.DonationToSocietyErrorMessage = errorMessage
			sf.Valid = false
		} else {
			sf.DonationToSociety = dts
		}
	}

	if len(sf.DonationToMuseumInput) > 0 {

		errorMessage, dtm := checkDonation(sf.DonationToMuseumInput)
		if len(errorMessage) > 0 {
			sf.DonationToMuseumErrorMessage = invalidNumber
			sf.Valid = false
		} else {
			sf.DonationToMuseum = dtm
		}
	}

	return sf.Valid
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

// usersExist checks if the users in the payment form are already in the system -
// meaning that this is a membership renewal. It looks up the ordinary user in
// the database with the name and/or email address given in the form to check that
// they exists.  If the details of an associate member are given, it checks that
// user too.  It returns the user IDs if both user(s) exist - (42,43,nil) or just
// the ID of the ordinary user if there is no associate - (42,0, nil).  If there
// is no match, it returns (0,0,nil).  If there is an error it returns (0,0,error).
func usersExist(ms *database.MembershipSale, db *database.Database) (int64, int64, error) {

	userID, userIDError := db.GetUserIDofMember(ms.FirstName, ms.LastName, ms.Email)
	if userIDError != nil {
		return 0, 0, userIDError
	}

	var assocUserID int64

	if len(ms.AssocFirstName) > 0 {
		var assocUserIDError error
		assocUserID, assocUserIDError =
			db.GetUserIDofMember(ms.AssocFirstName, ms.AssocLastName, ms.AssocEmail)
		if assocUserIDError != nil {
			return 0, 0, assocUserIDError
		}
	}
	return userID, assocUserID, nil
}

// getTickBox returns true and "checked" if the tickbox is ticked ("on"),
// false and "unchecked" otherwise.
func getTickBox(value string) (bool, string, string) {
	switch {
	case len(value) == 0:
		return false, "off", "unchecked"
	case value == "on":
		return true, "on", "checked"
	default:
		return false, "off", "unchecked"
	}
}

const paymentPageTemplateStr = `
<html>
<head>
    <title>Membership Payment System</title>
</head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>

		<h3>Membership Year {{.MembershipYear}}</h3>

		<span style="color:red;">{{.GeneralErrorMessage}}</span>
		<p>
			To become a member or renew your membership
			using a credit or debit card,
			please fill in the form below and press the Submit button.
		</p>
	{{if .EnableOtherMemberTypes}}
		<p>
			If you are also paying for an Associate Member
			(a second member at the same address)
			please supply their details too,
			otherwise leave those boxes blank.
		</p>
	{{end}}
		<p>
			Our fees this year are:
		<ul>
			<li>Ordinary member: {{.OrdinaryMemberFeeForDisplay}}</li>
		{{if .EnableOtherMemberTypes}}
			<li>Associate member at the same address: {{.AssocFeeForDisplay}}</li>
			<li>Friend of the Leatherhead museum: {{.FriendFeeForDisplay}}</li>
		{{end}}
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
			{{if .EnableOtherMemberTypes}}
				<tr>
					<td style='border: 0'>Friend of the Museum:</td>
					<td style='border: 0; '>
						<input style='transform: scale(1.5);' type='checkbox' name='friend' {{.FriendOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
			{{end}}

				<tr>
					<td style='border: 0'>Donation:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_society' value='{{.DonationToSocietyForDisplay}}'></td>
					<td style='border: 0;'><span style="color:red;">{{.DonationToSocietyErrorMessage}}</span></td>
				</tr>

			{{if .EnableOtherMemberTypes}}
				<tr>
					<td style='border: 0'>Donation to the Museum:</td>
					<td style='border: 0'><input type='text' size='40' name='donation_to_museum' value='{{.DonationToMuseumForDisplay}}'></td>
					<td style='border: 0'><span style="color:red;">{{.DonationToMuseumErrorMessage}}</span></td>
				</tr>
			{{end}}

			{{if .EnableGiftaid}}
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
			{{end}}
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>

			{{if .EnableOtherMemberTypes}}
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
						<input style='transform: scale(1.5);' type='checkbox' name='assoc_friend' {{.AssocFriendOutput}} />
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>

				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
			{{end}}

			</table>
			<input type="submit" value="Submit">
		</form>
	</body>
</html>
`

const paymentConfirmationPageTemplate = `
<html>
    <head><title>payment confirmation</title></head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>
		<h3>Membership Fee for {{.MembershipYear}}</h3>
		<p>
			If you are happy with the total,
			please press the submit button.
			You will be transferred to the Stripe payment system
			to make the payment.
		</p>
		<form action="/checkout" method="POST">

			<table>
				<tr>
					<td style='border: 0'>ordinary membership</td>
					<td style='border: 0'>{{.OrdinaryMemberFeeForDisplay}}</td>
				</tr>
			{{if .Friend}}
				<tr>
					<td style='border: 0'>friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{if ne .AssocFirstName ""}}
				<tr>
					<td style='border: 0'>associate member</td>
					<td style='border: 0'>{{.AssocFeeForDisplay}}</td>
				</tr>
			{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>associate is a friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{end}}

			{{if gt .DonationToSociety 0.0}}
			<tr>
				<td style='border: 0'>donation to the Society</td>
				<td style='border: 0'>{{.DonationToSocietyForDisplay}}</td>
			</tr>
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the museum</td>
				<td style='border: 0'>{{.DonationToMuseumForDisplay}}</td>
			</tr>
			{{end}}

			<tr>
				<td style='border: 0'><b>Total</b></td>
				<td style='border: 0'>{{.TotalForDisplay}}</td>
			</tr>

			</table>
			<input type='hidden' name='first_name' value={{.FirstName}}>
			<input type='hidden' name='last_name' value={{.LastName}}>
			<input type='hidden' name='email' value={{.Email}}>
		{{if .Friend}}
			<input type='hidden' name='friend' value='on'>
		{{end}}
		{{if ne .AssocFirstName ""}}
			<input type='hidden' name='assoc_first_name' value={{.AssocFirstName}}>
			<input type='hidden' name='assoc_last_name' value={{.AssocLastName}}>
			<input type='hidden' name='assoc_email' value={{.AssocEmail}}>
		{{if .AssocFriend}}
			<input type='hidden' name='assoc_friend' value='on'>
		{{end}}
		{{end}}
		{{if .Giftaid}}
			<input type='hidden' name='giftaid' value='on'>
		{{end}}
		{{if gt .DonationToSociety 0.0}}
			<input type='hidden' name='donation_to_society' value='{{.DonationToSociety}}'>
		{{end}}
		{{if gt .DonationToMuseum 0.0}}
			<input type='hidden' name='donation_to_museum' value='{{.DonationToMuseum}}'>
		{{end}}
			<input type="submit" value="Submit">
		</form>
	</body>
`

const successPageTemplateStr = `
<html>
	<head><title>Payment Successful</title></head>
    <body style='font-size: 100%'>
	<h2>{{.OrganisationName}}</h2>
        <p>
			Thank you for your payment.
			You are now a member until the end of {{.MembershipYear}}.
		</p>
		<p>
			<table>
				<tr>
					<td style='border: 0'>ordinary membership</td>
					<td style='border: 0'>{{.OrdinaryMemberFeeForDisplay}}</td>
				</tr>
			{{if .Friend}}
				<tr>
					<td style='border: 0'>friend of the museum</td>
					<td style='border: 0'>{{.FriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{if ne .AssocFirstName ""}}
				<tr>
					<td style='border: 0'>associate member</td>
					<td style='border: 0'>{{.AssocFeeForDisplay}}</td>
				</tr>
			{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>associate is friend of the museum</td>
					<td style='border: 0'>{{.AssocFriendFeeForDisplay}}</td>
				</tr>
			{{end}}
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the Society</td>
				<td style='border: 0'>{{.DonationToSocietyForDisplay}}</td>
			</tr>
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
			<tr>
				<td style='border: 0'>donation to the museum</td>
				<td style='border: 0'>{{.DonationToMuseumForDisplay}}</td>
			</tr>
			{{end}}

			<tr>
				<td style='border: 0'><b>Total</b></td>
				<td style='border: 0'>{{.TotalForDisplay}}</td>
			</tr>

			</table>
		</p>
		<p>
		    If you have any questions, please email
		    <a href="mailto:{{.EmailAddressForQuestions}}">
			    {{.EmailAddressForQuestions}}
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

// prePaymentErrorHTMLPattern defines the default error message page before the
// customer has paid (not as bad as a failure after).  It contains configurable
// text but it may be used when things are going badly wrong.  It may be the
// template system that's going wrong, so we don't use the template system to
// render the page.  We just expand the pattern using Sprintf (which means that
// "100%" must be presented as "100%%").
const prePaymentErrorHTMLPattern = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%%'>
		<p>
			Internal error.  Please try again.
			If the error persists, please email
			<a href="mailto:%s">
				%s
			</a>
		</p>
    </body>
</html>
`

// postPaymentErrorHTMLPattern defines the default error message page after the
// customer has paid.  It contains configurable text but it may be used when things
// are going badly wrong.  It may be the template system that's going wrong, so we
// don't use the template system to render the page.  We just expand the pattertn
// using Sprintf (which means that "100%" must be presented as "100%%").
//
// The customer is now due a refund so the given email address should be somebody
// who can arrange that, eg the treasurer.
const postPaymentErrorHTMLPattern = `
<html>
    <head><title>error</title></head>
    <body style='font-size: 100%%'>
		<p>
			Something went wrong after you had paid.
			Please
			<a href="mailto:%s">
				%s
			</a>
		</p>
    </body>
</html>
`
