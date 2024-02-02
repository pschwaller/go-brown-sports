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

func AccessSpreadsheet(ctx context.Context, client *http.Client) (*sheets.Service, error) {

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return srv, err
}

// GetSpreadsheetMap Get a map with key: datetime+sport and value: SportingEvent struct of all the events
// on all months on the spreadsheet.
func GetSpreadsheetMap(sheetService *sheets.Service) map[string]SportingEvent {
	nameToEmailMap := loadNameToEmailMap(sheetService, SpreadsheetId)

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
	currentTime := time.Now()
	for _, month := range monthList {
		monthEvents := LoadMonthAssignments(sheetService, SpreadsheetId, month, nameToEmailMap)
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
	spreadsheetId string,
	month string,
	nameToEmailMap map[string]string) []SportingEvent {

	readRange := fmt.Sprintf("%s!A:ZZ", month)
	var rows [][]interface{} = loadSpreadsheetRows(srv, spreadsheetId, readRange)
	headers := rows[0]
	eventData := rows[1:]

	// The first columns are required.  If they don't exist in the correct order... panic!
	// The remaining columns are all treated as named Roles, and we don't care if more columns are added.
	requiredHeaders := []string{"Date", "Time", "Sport"}
	for index, expectedHeader := range requiredHeaders {
		if headers[index] != expectedHeader {
			panic(fmt.Sprintf("Unexpected headers in month %s.  Expected %s, found %s.", month, expectedHeader, headers[0]))
		}
	}
	missing := make(map[string]int)

	var sportingEvents []SportingEvent

	for _, event := range eventData {
		// Some rows are blank.  User should delete them, but let's be nice to users.
		if len(event) == 0 {
			continue
		}
		dateString := event[0].(string)
		timeString := event[1].(string)

		sportingEvent := SportingEvent{}

		sportingEvent.Datetime = resolveDatetime(dateString, timeString)
		sportingEvent.Sport = event[2].(string)

		if len(headers) != len(event) {
			fmt.Printf("No MATCH: %s\n", event)
		}

		for index, eventEntry := range event {
			if index < 3 {
				continue
			}
			if eventEntry == "x" || eventEntry == "" {
				continue
			}
			name := eventEntry.(string)
			if nameToEmailMap[name] == "" {
				missing[name] = missing[name] + 1
			} else {
				sportingEvent.Emails = append(sportingEvent.Emails, nameToEmailMap[name])
				sportingEvent.Roles = append(sportingEvent.Roles, fmt.Sprintf("%s: %s", headers[index], name))
			}
		}
		sportingEvents = append(sportingEvents, sportingEvent)

	}

	return sportingEvents
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
func loadNameToEmailMap(srv *sheets.Service, spreadsheetId string) map[string]string {
	var rows [][]interface{} = loadSpreadsheetRows(srv, spreadsheetId, "Worker Contact Info!A2:C")
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

func loadSpreadsheetRows(srv *sheets.Service, spreadsheetId string, readRange string) [][]interface{} {

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	var rows [][]interface{}
	for _, row := range resp.Values {
		rows = append(rows, row)
	}
	return rows
}
