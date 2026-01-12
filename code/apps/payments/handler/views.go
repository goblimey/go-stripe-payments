package handler

// HTML views, separated out for clarity.

// paymentPageTemplateStr defines the initial payment form that collects
// the data for the sale.  Data is taken from a MembershipSale object.
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
			<li>Full price member: {{.OrdinaryMemberFeeForDisplay}}</li>
		{{if .EnableOtherMemberTypes}}
			<li>Associate member at the same address: {{.AssocFeeForDisplay}}</li>
			<li>Friend of the Leatherhead museum: {{.FriendFeeForDisplay}}</li>
		{{end}}
		</ul>
		</p>
		<p>
			&nbsp;
		</p>
		<form action="/subscribe/" method="POST">	
			<table style='font-size: 100%'>

				<tr>
					<td style='border: 0'>Title (Mr, Mrs, Ms, Dr etc):</td>
					<td style='border: 0'><input type='text' size='40' name='title' value='{{.Title}}'></td>
					<td style='border: 0'><span style="color:red;">{{.TitleErrorMessage}}</span></td>
				</tr>

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
						<input style='transform: scale(1.5);' type='checkbox' name='friend' {{.FriendOutput}}>
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
					<td style='border: 0'>
						<input type='text' size='40' name='donation_to_museum' value='{{.DonationToMuseumForDisplay}}'>
					</td>
					<td style='border: 0'><span style="color:red;">{{.DonationToMuseumErrorMessage}}</span></td>
				</tr>
			{{end}}

			{{if .EnableGiftaid}}
				<tr>
					<td style='border: 0'>Gift Aid:</td>
					<td style='border: 0 '>
						<input style='transform: scale(1.5);' type='checkbox' name='giftaid' {{.GiftaidOutput}}>
					</td>
					<td style='border: 0'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>&nbsp;</td>
				</tr>
				<tr>
					<td style='border: 0' colspan='3'>
						Tick the Gift Aid box if you are currently a UK tax payer and 
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
						fill in the other member's details below.
						If they don't want to receive emails,
						leave their email address blank.
					</td>
				</tr>

				<tr>
					<td style='border: 0'>Associate member's Title:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_title' value='{{.AssocFirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocTitleErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Associate member's First Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_first_name' value='{{.AssocFirstName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocFirstNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Associate member's Last Name:</td>
					<td style='border: 0'><input type='text' size='40' name='assoc_last_name' value='{{.AssocLastName}}'></td>
					<td style='border: 0'><span style="color:red;">{{.AssocLastNameErrorMessage}}</span></td>
				</tr>

				<tr>
					<td style='border: 0'>Associate member's email Address (optional):</td>
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

// paymentConfirmationPageTemplate defines the payment confirmation page.
// page.  Data is taken from a SaleForm object.
const paymentConfirmationPageTemplate = `
<html>
    <head><title>payment confirmation</title></head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>
		<h3>Membership payment for {{.MembershipYear}}</h3>
		<p>
			If you are happy with the total,
			please press the submit button.
			You will be transferred to the Stripe payment system
			to make the payment.
		</p>
		<form action="/checkout" method="POST">
			<input type='hidden' name='title' value={{.Title}}>
			<input type='hidden' name='first_name' value={{.FirstName}}>
			<input type='hidden' name='last_name' value={{.LastName}}>
			<input type='hidden' name='email' value={{.Email}}>
			<input type='hidden' name='donation_to_society' value='{{.DonationToSociety}}'>
			<input type='hidden' name='donation_to_museum' value='{{.DonationToMuseum}}'>
			<input type='hidden' name='assoc_title' value={{.AssocTitle}}>
			<input type='hidden' name='assoc_first_name' value={{.AssocFirstName}}>
			<input type='hidden' name='assoc_last_name' value={{.AssocLastName}}>
			<input type='hidden' name='assoc_email' value={{.AssocEmail}}>
		{{if .Friend}}
			<input type='hidden' name='friend' value='on'>
		{{end}}
		{{if .Giftaid}}
			<input type='hidden' name='giftaid' value='on'>
		{{end}}
		{{if .AssocFriend}}
			<input type='hidden' name='assoc_friend' value='on'>
		{{end}}

			<table>
				<tr>
					<td style='border: 0'>
						Full price membership for 
						{{.Title}} {{.FirstName}} {{.LastName}}
					</td>
					<td style='border: 0' align='right'>
						{{.OrdinaryMemberFeeForDisplay}}
					</td>
				</tr
			{{if .Friend}}
				<tr>
					<td style='border: 0'>
						{{.Title}} {{.FirstName}} {{.LastName}} 
						is a friend of the Museum
					</td>
					<td style='border: 0' align='right'>
						{{.FriendFeeForDisplay}}
					</td>
				</tr>
			{{end}}

			{{if gt (len .DonationToSocietyForDisplay) 0}}
				<tr>
					<td style='border: 0'>Donation to the Society</td>
					<td style='border: 0' align='right'>
						{{.DonationToSocietyForDisplay}}
					</td>
				</tr>
			{{end}}

			{{if gt (len .DonationToMuseumForDisplay) 0}}
				<tr>
					<td style='border: 0'>Donation to the Museum</td>
					<td style='border: 0' align='right'>
						{{.DonationToMuseumForDisplay}}
					</td>
				</tr>
			{{end}}

			{{if ne .AssocFirstName ""}}
				<tr>
					<td style='border: 0'>
						Associate membership for 
						{{.AssocTitle}} {{.AssocFirstName}} {{.AssocLastName}}
					</td>
					<td style='border: 0' align='right'>
						{{.AssocFeeForDisplay}}
					</td>
				</tr>
				{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>
						{{.AssocTitle}} {{.AssocFirstName}} {{.AssocLastName}} 
						is a friend of the Museum 
					</td>
					<td style='border: 0' align='right'>
						{{.FriendFeeForDisplay}}
					</td>
				</tr>
				{{end}}
			{{end}}
				<tr>
					<td style='border: 0'><b>Total</b></td>
					<td style='border: 0' align='right'>
						{{.TotalForDisplay}}
					</td>
				</tr>
			</table>
			<input type="submit" value="Submit">
		</form>
	</body>
`

// successPageFormatString defines two pages, the one shown on a successful sale
// that displays the payment breakdown and starts the process of collecting the
// user's extra details, and then the page that handles validation errors when
// the form on the first page is submitted.  The page contains a selection list
// of countries and a selection list of member's interests which is populated
// from the database.  To support that, the Go HTML template is put together
// from three strings, one created on the fly containing the selection list (or
// which is empty if there is not adm_interests table).
//
// The resulting template takes data from a MembershipSale object.  Immediately
// after a succesful sale the PaymentStatus contains a value, which turns on the
// payment breakdown at the start of the page.  In subsequent calls, the
// PaymentStatus value is empty, which turns off the payment breakdown.
const successPageTemplateString1 = `
<html>
	<head><title>Payment Successful</title></head>
    <body style='font-size: 100%'>
	<h2>{{.OrganisationName}}</h2>
	{{if gt (len .PaymentStatus) 0}}
        <p>
			Thank you for your payment.
			You are now a member until the end of {{.MembershipYear}}.
		</p>
		<p>
			<table>
				<tr>
					<td style='border: 0'>
						Full price Membership for {{.Title}} {{.FirstName}} {{.LastName}} {{.Email}}
					</td>
					<td style='border: 0' align='right'>
						{{.OrdinaryMemberFeeForDisplay}}
					</td>
				</tr>
			{{if .Friend}}
				<tr>
					<td style='border: 0'>Friend of the Museum</td>
					<td style='border: 0' align='right'>
						{{.FriendFeeForDisplay}}
					</td>
				</tr>
			{{end}}
			{{if gt .DonationToSociety 0.0}}
				<tr>
					<td style='border: 0'>Donation to the Society</td>
					<td style='border: 0' align='right'>
						{{.DonationToSocietyForDisplay}}
					</td>
				</tr>
			{{end}}

			{{if gt .DonationToMuseum 0.0}}
				<tr>
					<td style='border: 0'>Donation to the museum</td>
					<td style='border: 0' align='right'>
						{{.DonationToMuseumForDisplay}}
					</td>
				</tr>
			{{end}}

			{{if gt (len .AssocLastName) 0}}
				<tr>
					<td style='border: 0'>Associate Membership for 
						{{.AssocTitle}} {{.AssocFirstName}} {{.AssocLastName}} {{.AssocEmail}}</td>
					<td style='border: 0' align='right'>
						{{.AssocFeeForDisplay}}
					</td>
				</tr>
			{{end}}
			
			{{if .AssocFriend}}
				<tr>
					<td style='border: 0'>Associate is a friend of the Museum</td>
					<td style='border: 0' align='right'>
						{{.FriendFeeForDisplay}}
					</td>
				</tr>
			{{end}}
				<tr>
					<td style='border: 0'><b>Total</b></td>
					<td style='border: 0' align='right'>
						{{.TotalForDisplay}}
					</td>
				</tr>
			</table>
		</p>
		<p>
		    If you have any questions, please email
		    <a href="mailto:{{.EmailAddressForQuestions}}">
			    {{.EmailAddressForQuestions}}
			</a>.
		</p>
		{{end}}
		<p>&nbsp;</p>
	{{if eq .TransactionType "new member"}}
		<p>
			We would appreciate it if you would fill in as much of this information
			as you care to.  
			It will be held in our membership database.
		</p>
		
	{{else}}
		<p>
			Please check that this information is up to date:
		</p>
		<p>
			<form action="/completion" method="POST">
				<input type='hidden' name='organisation_name' value='{{.OrganisationName}}'>
				<input type="submit" value="No Changes">
			</form>
		</p>
	{{end}}
		
		
		<p>
		<form action="/extradetails" method="POST">
			<input type='hidden' name='account_name' value='{{.AccountName}}'>
		<input type='hidden' name='title' value='{{.Title}}'>
		<input type='hidden' name='first_name' value='{{.FirstName}}'>
		<input type='hidden' name='last_name' value='{{.LastName}}'>
		{{if gt (len .AssocAccountName) 0}}
			<input type='hidden' name='assoc_account_name' value='{{.AssocAccountName}}'>
			<input type='hidden' name='assoc_title' value='{{.AssocTitle}}'>
			<input type='hidden' name='assoc_first_name' value='{{.AssocFirstName}}'>
			<input type='hidden' name='assoc_last_name' value='{{.AssocLastName}}'>
		{{end}}

			<table style='font-size: 100%'>
				<tr>
					<td style='border: 0'><b>Address</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='address_line_1' value='{{.AddressLine1}}'>
					</td>
					<td style="color:red;">{{.AddressLine1Error}}</td>
				</tr>
				<tr>
					<td style='border: 0'>&nbsp;</td>
					<td style='border: 0'>
						<input type='text' size='40' name='address_line_2' value='{{.AddressLine2}}'>
					</td>
					<td style="color:red;">{{.AddressLine2Error}}</td>
				</tr>
				<tr>
					<td style='border: 0'>&nbsp;</td>
					<td style='border: 0'>
						<input type='text' size='40' name='address_line_3' value='{{.AddressLine3}}'>
					</td>
					<td style="color:red;">{{.AddressLine3Error}}</td>
				</tr>
				<tr>
					<td style='border: 0'><b>Town</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='town' value='{{.Town}}'>
					</td>
					<td style="color:red;">{{.TownError}}</td>
				</tr>
				<tr>
					<td style='border: 0'><b>County</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='county' value='{{.County}}'>
					</td>
					<td style="color:red;">{{.CountyError}}</td>
				</tr>
				<tr>
					<td style='border: 0'><b>Postcode</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='postcode' value='{{.Postcode}}'>
					</td>
					<td style="color:red;">{{.PostcodeError}}</td>
				</tr>	
`

// The country selection list is added here.
const successPageTemplateString2 = `
				<tr>
					<td style='border: 0'><b>Landline phone number</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='phone' value='{{.Phone}}'>
					</td>
					<td style="color:red;">{{.PhoneError}}</td>
				</tr>
				<tr>
					<td style='border: 0'><b>Mobile number for {{.FirstName}} {{.LastName}} </b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='mobile' value='{{.Mobile}}'>
					</td>
					<td style="color:red;">{{.MobileError}}</td>
				</tr>
				{{if gt (len .AssocAccountName) 0}}
				<tr>
					<td style='border: 0'><b>Mobile number for {{.AssocFirstName}} {{.AssocLastName}}</b></td>
					<td style='border: 0'>
						<input type='text' size='40' name='assoc_mobile' value='{{.AssocMobile}}'>
					</td>
					<td style="color:red;">{{.AssocMobileError}}</td>
				</tr>
				{{end}}
				<tr>
					<td style='border: 0'>
						<b>Parish of Interest</b>
						<br>
						(Ashtead, Fetcham, Bookham,
						Leatherhead or All Areas)
					</td>
					<td style='border: 0'>
						<input type='text' size='40' name='location_of_interest' value='{{.LocationOfInterest}}'>
					</td>
					<td style="color:red;">{{.LocationOfInterestError}}</td>
				</tr>
`

// The interests selection list and other interests box are added here if the
// adm_interests table exists and contains some rows.
const successPageTemplateString3 = `
			</table>
			<input type="submit" value="Update">
		</form>
		</p>
    </body>
</html>
`

// completionPageTemplate defines the page shown on completion.
// Data is taken from a MembershipSale object.
const completionPageTemplateString = `
<html>
    <head><title>payment complete</title></head>
	<body style='font-size: 100%'>
		<h2>{{.OrganisationName}}</h2>
		<h3>Membership for the year {{.MembershipYear}}</h3>
		<p>
			Thank you for your payment.
		</p>
	</body>
</html>
`

// cancelHTML defines the cancel page, called when the payment is cancelled
// on the Stripe system.  Not sure under what circumstances this happens or
// how to provoke it in the test environment.
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
// text but it may be used when all sorts of things are going wrong, including
// the template system itsef, so we make the rendering of the page as simple as
// possible and not done via the template system.  We just expand the pattern
// using Sprintf (which means that "100%" must be presented as "100%%").
//
// None of the configurable data is submitted by the user so there is no chance
// of an injection attack.
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
// customer has paid.  It contains configurable text but it may be used when
// all sorts of things are going wrong, including the template system itsef, so
// we make the rendering of the page as simple as possible and not done via the
// template system.  We just expand the pattern using Sprintf (which means that
// "100%" must be presented as "100%%").  None of the configurable data is
// submitted by the user so there is no chance of an injection attack.
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
