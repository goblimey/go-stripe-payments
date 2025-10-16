# A web solution that updates a member's end date in an Admidio system

Admidio is an open-source web-based member contact system.  Each member record has a start and end date.  This solution updates a member's end date on payment of a fee.  The fee is collected via the Stripe online payment system

The Admidio software is distributed [here](https://www.admidio.org/) 

Stripe's Go API is [here](https://docs.stripe.com/api?lang=go).

Admdio uses an underlying database which can be
MySQL, postgres or SQLite.
This application works with the postgres or the SQLite version. 


## Building the software

To run this software you need to open a Stripe account.
Once you've done that,
log in and visit https://dashboard.stripe.com/test/apikeys.
That will display a secret key that you can use for system testing
and another for a live application.
They are specially manufactured for you.

When the program starts
it looks for a file "config.json"
in the current directory.
This should contain:

```
{
    "stripe_secret_key": "your Stripe secret key",
    "ssl_certificate_file": "public https certificate file",
    "ssl_certificate_key_file": "private https certificate file"
}

```

You also need some environment variables to specify the database:

```
export DBType='{type}'          # Database type - postgres or sqlite.
export DBHost='{host}' 		    # Host machine running the database
export DBPort={port} 			# Database port
export DBDatabase='{name}' 		# Database name
export DBUser='{user}' 		    # database user name
export DBPassword='{password}'  # database Password
```


Build the software:

```
go build

go test
```

Run the software on your test server:
```
. config.sh

./go-stripe-payments
```

The server runs on port 4242.  If it's running on the server
example.com, navigate your web browser to

http://example.com:4242/displayPaymentForm


## How does it work?


Stripe offers an API that allows software written in various languages
to use their service to take payments.
Go is one of the supported languages.
They also offer some very useful mock payment services that allows you
to test your software thoroughly.

For the society that uses this software,
membership runs from the 1st January to December 31st.
New members who join from the 1st October onwards
get membership from that date to the end of the next year. 

If two members live at the same address,
one can be an associate and get cheaper membership.

At present the system only handles membership renewals,
not new membership.

The system collects the data needed to figure out the membership price to charge.
It creates a membership_sales record in the database,
uses the Stripe API to create a checkout session containing the price
and the ID
of the database record,
then redirects the customer's browser to the Stripe payment system.
Stripe collects the money and returns control to this website
by calling the success request specified in the session.
It adds a parameter to the call which is the ID of the checkout session,
something like

```
https://example.com/success?session_id={something}
```

The success handler recovers the checkout session and thereby the 
membership_sale record.
This tells it which user record(s) to update.


The application updates the end dates of the member record(s)
and marks the status in the membership_sale record as "complete"

I use only the native Go
database interface "database/sql"
via a thin Database object that contains information like the type of the database in use.

As far as possible I keep the PostgreSQL and
SQLite queries the same.
PostgreSQL uses $1, $2 etc as placeholders in queries,
SQLite uses ?, ? etc.
I specify each query in PostgreSQL form
and the thin layer converts it to SQLite form
just before running it.

Each query is wrapped in a function that collects the input data,
runs the query and checks that it yields a result.
Where the logic is the same for PostgreSQL and SQLite,
either can be used to check that logic.

I use a PostgreSQL database for the production system, so there is an integration test to check that each query works with that database.
Other tests can be done using SQLite in-memory databases.

## Unit and Integration Testing

Automatic testing is currently adequate but not good,
so the application needs some manual testing to make up for that.
My intention is to improve that situation.

I try to separate out processing that can be tested using lightweight uint tests from stuff that can't.

For example, some of the configuration is taken from a JSON config file.
The file is read into memory and the contents consumed.
I separate the file reading from the consumption so most of the testing can be done on strings of text.

That minimises the need for tests that read files,
but some of that is still necessary, of course.
Rather than using a fixed file I do that by creating a temporary directory,
writing a config file to it,
reading it back,
checking it and then removing the directory. 
I use a different directory name each time so that such tests can be run in parallel.

Another example of this policy of separation is the validation of
the web form that collects the data for the sale.  The form has lots of options
requiring lots of validation
and some of that needs to access the database.
The validation runs in two stages.
Validation that doesn't need to access the database is done in the first stage.  That's most of the work and it can be tested thoroughly by lightweight unit tests.  The second stage of validation just checks that the given data matches one or two existing users.
That just needs a simple integration test connecting to a prepared test database.

Some of the processing of data from the database is separated out from the code that fetches the database record.
In those cases,
the separated out processing 
can be checked using lightweight unit tests
and it's then just necessary to write an integration test to check that the data from the database is presented correctly to the processing stage.

The test.sh script in the root of this repository
runs all of the tests.
This also produces a set of test coverage reports.
