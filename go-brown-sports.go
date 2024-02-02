// Copyright 2024 Peter Schwaller (peter@schwaller.org)

package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"log"
	"schwaller.org/go-brown-sports/pkg"
)

func main() {

	// Set up access to the Google APIs we're using
	ctx, client, err := pkg.AccessGoogleClient()
	if err != nil {
		log.Fatalf("Unable to create Google client: %v", err)
	}

	// We're going to create a series of maps using datetime+sport as the key, and
	// a SportingEvent struct as the value.

	// Access the calendar and generate the maps
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}
	calendarFutureEvents, calendarFutureEventIds := pkg.GetCalendarMaps(calendarService)

	// Access the spreadsheet and generate the map
	sheetService, err := pkg.AccessSpreadsheet(ctx, client)
	if err != nil {
		log.Fatalf("Unable to access spreadsheet: %v", err)
	}
	spreadsheetFutureEvents := pkg.GetSpreadsheetMap(sheetService)

	// Now that we have all the maps, let's sync the spreadsheet info into the calendar.
	synchronizeCalendar(spreadsheetFutureEvents, calendarFutureEvents, calendarFutureEventIds, calendarService)
}

func synchronizeCalendar(spreadsheetFutureEvents map[string]pkg.SportingEvent, calendarFutureEvents map[string]pkg.SportingEvent, calendarFutureEventIds map[string]string, calendarService *calendar.Service) {
	var err error

	// Create sets.  This makes determining what's missing and extra simple and straightforward.
	calendarSet := mapset.NewSetFromSlice(pkg.GetKeysFromSportingEventMap(calendarFutureEvents))
	spreadsheetSet := mapset.NewSetFromSlice(pkg.GetKeysFromSportingEventMap(spreadsheetFutureEvents))

	missingInCalendar := spreadsheetSet.Difference(calendarSet)
	for key := range missingInCalendar.Iter() {
		sportingEvent := spreadsheetFutureEvents[key.(string)]
		err = pkg.CreateCalendarEvent(calendarService, pkg.CalendarId, sportingEvent)
		if err != nil {
			fmt.Printf("Error creating calendar event for %s: %v", key, err)
		}
	}

	extraInCalendar := calendarSet.Difference(spreadsheetSet)
	for key := range extraInCalendar.Iter() {
		eventId := calendarFutureEventIds[key.(string)]
		err = calendarService.Events.Delete(pkg.CalendarId, eventId).Do()
		if err != nil {
			fmt.Printf("Error deleting %s: %v", key.(string), err)
		}
	}

	matchingKeys := calendarSet.Intersect(spreadsheetSet)
	updateCount := 0
	for key := range matchingKeys.Iter() {
		keyString := key.(string)
		if calendarFutureEvents[keyString].IsMostlyEqual(spreadsheetFutureEvents[keyString]) {
			continue
		}
		// If we're here, there must have been a change.  Update the calendar entry.
		eventId := calendarFutureEventIds[keyString]
		err = pkg.UpdateCalendarEvent(calendarService, pkg.CalendarId, eventId, spreadsheetFutureEvents[keyString])
		if err != nil {
			fmt.Printf("Error updating %s: %v", keyString, err)
		}
		updateCount += 1
	}

	fmt.Printf("Missing: %d, Extra: %d, Updated: %d\n", missingInCalendar.Cardinality(), extraInCalendar.Cardinality(), updateCount)
}
