package pkg

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"net/http"
	"strings"
	"time"
)

// Specify the column number of our required columns.
// If any changes are made to these, the requiredHeaders variable (see below) should also be updated.
const dateColumnNumber = 0
const timeColumnNumber = 1
const sportColumnNumber = 2

func AccessSpreadsheet(ctx context.Context, client *http.Client) (*sheets.Service, error) {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv, err
}

// GetSpreadsheetMap Get a map with key: datetime+sport and value: SportingEvent struct of all the events
// on all months on the spreadsheet.
func GetSpreadsheetMap(sheetService *sheets.Service, currentTime time.Time) map[string]SportingEvent {
	nameToEmailMap := loadNameToEmailMap(sheetService, GetSpreadsheetID())

	// This list corresponds to the sheet IDs (tabs) in the spreadsheet.
	monthList := []string{
		"September",
		"October",
		// TODO Need to ask Kelvin to fix the spreadsheet -- November does not have "Date" in cell A1.
		// "November",
		"December",
		"January",
		"February",
		"March",
		"April",
		"May",
	}

	spreadsheetFutureEvents := make(map[string]SportingEvent)

	for _, month := range monthList {
		monthEvents := LoadMonthAssignments(sheetService, GetSpreadsheetID(), month, nameToEmailMap)
		for _, event := range monthEvents {
			if event.Datetime.After(currentTime) {
				spreadsheetFutureEvents[event.GetKey()] = event
			}
		}
	}

	return spreadsheetFutureEvents
}

func LoadMonthAssignments(
	srv *sheets.Service,
	spreadsheetID string,
	month string,
	nameToEmailMap map[string]string) []SportingEvent {
	readRange := fmt.Sprintf("%s!A:ZZ", month)

	rows := loadSpreadsheetRows(srv, spreadsheetID, readRange)

	headers := rows[0]
	eventData := rows[1:]

	// The first columns are required.  If they don't exist in the correct order... panic!
	// The remaining columns are all treated as named Roles, and we don't care if more columns are added.
	requiredHeaders := []string{"Date", "Time", "Sport"}
	for index, expectedHeader := range requiredHeaders {
		if headers[index] != expectedHeader {
			log.Printf("Unexpected headers in month %s.  Expected %s, found %s.", month, expectedHeader, headers[0])
		}
	}

	missing := make(map[string]int)

	var sportingEvents []SportingEvent

	for _, event := range eventData {
		// Some rows are blank.  User should delete them, but let's be nice to users.
		if len(event) == 0 {
			continue
		}

		sportingEvent := buildSingleEvent(event, headers, requiredHeaders, nameToEmailMap, missing)
		sportingEvents = append(sportingEvents, sportingEvent)
	}

	return sportingEvents
}

func buildSingleEvent(
	event []interface{},
	headers []interface{},
	requiredHeaders []string,
	nameToEmailMap map[string]string,
	missing map[string]int) SportingEvent {
	dateString := event[dateColumnNumber].(string)
	timeString := event[timeColumnNumber].(string)

	sportingEvent := SportingEvent{}

	sportingEvent.Datetime = resolveDatetime(dateString, timeString)
	sportingEvent.Sport = event[sportColumnNumber].(string)

	if len(headers) != len(event) {
		fmt.Printf("No MATCH: %s\n", event)
	}

	for index, eventEntry := range event {
		if index < len(requiredHeaders) {
			continue
		}

		if eventEntry == "x" || eventEntry == "" {
			continue
		}

		name := eventEntry.(string)

		if nameToEmailMap[name] == "" {
			missing[name]++
		} else {
			sportingEvent.Emails = append(sportingEvent.Emails, nameToEmailMap[name])
			sportingEvent.Roles = append(sportingEvent.Roles, fmt.Sprintf("%s: %s", headers[index], name))
		}
	}

	return sportingEvent
}

func resolveDatetime(dateString string, timeString string) time.Time {
	// The times seem to be entered as "1 p.m.".  Let's normalize them a bit before proceeding.
	// TODO - Spreadsheet should store "(DH)" in the sport instead of in the Time column.
	replacer := strings.NewReplacer(".", "", " ", "", "TBA", "12:00am", "(DH)", "", "PM", "pm", "AM", "am")
	timeString = replacer.Replace(timeString)

	if !strings.Contains(timeString, ":") {
		// The timeString entries are not fully formed, because they are just typed text.  Let's fix the format.
		corrector := strings.NewReplacer("pm", ":00pm", "am", ":00am")
		timeString = corrector.Replace(timeString)
	}

	datetimeString := fmt.Sprintf("%s %s", dateString, timeString)
	// Hardcode the parsing to be for Eastern time zone.  America/New_York handles daylight savings time adjustments.
	eastern, _ := time.LoadLocation("America/New_York")
	datetime, err := time.ParseInLocation("Monday, January 2, 2006 3:04pm", datetimeString, eastern)

	if err != nil {
		fmt.Printf("Error formatting date %s: %v\n", datetimeString, err)
	}

	return datetime
}

// The spreadsheet has a separate tab for worker contact info.  Let's pull the emails for calendar invites.
func loadNameToEmailMap(srv *sheets.Service, spreadsheetID string) map[string]string {
	rows := loadSpreadsheetRows(srv, spreadsheetID, "Worker Contact Info!A2:C")

	nameToEmailMap := make(map[string]string)

	for _, row := range rows {
		if len(row) != 1 {
			name := row[0].(string)
			email := row[2].(string)

			if strings.Contains(email, "@brown.edu") {
				// Workers with Brown email addresses are often referred to on the schedule by first name only.
				// So we put them into the map by first name as well as by full name.
				firstName := strings.Split(name, " ")[0]
				if nameToEmailMap[firstName] != "" {
					fmt.Printf("Duplicate entry for %s\n", firstName)
				}

				nameToEmailMap[firstName] = email
			}

			nameToEmailMap[name] = email

			rows = append(rows, row)
		}
	}

	return nameToEmailMap
}

func loadSpreadsheetRows(srv *sheets.Service, spreadsheetID string, readRange string) [][]interface{} {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	return resp.Values
}
