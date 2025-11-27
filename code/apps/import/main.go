package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/goblimey/go-stripe-payments/code/apps/import/csvimport"
	"github.com/goblimey/go-stripe-payments/code/pkg/database"
)

var yearMembershipEnds int // The year that the membership ends.

func main() {
	usage := fmt.Sprintf("usage %s  CSV_file_name  year_membership_ends", os.Args[0])
	if len(os.Args) < 3 {
		slog.Error(usage)
		return
	}

	var yearError error
	yearMembershipEnds, yearError = strconv.Atoi(os.Args[2])

	if yearError != nil {
		slog.Error("The second argument (membership year) must the year that the membership ends, for example 2026")
		return
	}

	// Open the file.
	file, openError := os.Open(os.Args[1])

	// Checks for the error
	if openError != nil {
		m := "Error while reading " + os.Args[1] + "\n"
		log.Fatal(m, openError)
		os.Exit(-1)
	}

	records, importError := csvimport.Import(file, yearMembershipEnds)

	if importError != nil {
		slog.Error(importError.Error())
		os.Exit(-1)
	}

	dbConfig := database.GetDBConfigFromTheEnvironment()

	db := database.New(&dbConfig)

	connError := db.Connect()
	if connError != nil {
		slog.Error(connError.Error())
		os.Exit(-1)
	}

	txError := db.BeginTx()
	if txError != nil {
		slog.Error(txError.Error())
		os.Exit(-1)
	}
	defer db.Rollback()
	defer db.Close()

	// Traverse the imported data and produce DB records.
	for _, record := range records {
		line := fmt.Sprintf("%d %s %s: %s, %s",
			yearMembershipEnds,
			record.FirstName,
			record.Surname,
			record.UserName,
			record.Email)
		slog.Info(line)

		id, err := csvimport.ProcessRecord(db, &record)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if id <= 0 {
			em := fmt.Sprintf("user %s: ID not a positive integer - %d", record.UserName, id)
			slog.Error(em)
		}
	}

	// Success.
	log.Println("commit and close")
	db.Commit()
	db.Close()
}
