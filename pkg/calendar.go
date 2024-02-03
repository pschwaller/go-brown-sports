package pkg

import (
	"google.golang.org/api/calendar/v3"
	"log"
	"strings"
	"time"
)

func GetSportLocation(sport string) string {
	// Build a map of the Sports to a Location string for the events
	sportToLocation := map[string]string{
		"Men's Tennis":                 "",
		"Women's Tennis":               "",
		"Men's Water Polo":             "",
		"Women's Water Polo":           "",
		"Men's Basketball":             "Pizzitola Sports Center, Providence, RI 02906",
		"Women's Basketball":           "Pizzitola Sports Center, Providence, RI 02906",
		"Wrestling":                    "",
		"Gymnastics":                   "",
		"Gymnastics (Rumble & Tumble)": "",
		"Wrestling (Rumble & Tumble)":  "",
		"Men's Ice Hockey":             "Meehan Auditorium, 225 Hope St, Providence, RI 02912",
		"Women's Ice Hockey":           "Meehan Auditorium, 225 Hope St, Providence, RI 02912",
		"Track and Field (OMAC)":       "Olney-Margolies Athletic Center (OMAC)",
		"Men's Lacrosse":               "",
		"Women's Lacrosse":             "",
		"Baseball":                     "Terrence Murray Baseball Stadium",
		"Softball":                     "",
		"Men's Crew":                   "",
	}
	location := sportToLocation[sport]
	return location
}

// GetCalendarMaps Get a map with key: datetime+sport and value: SportingEvent struct of all the future events
// on the calendar.
func GetCalendarMaps(calendarService *calendar.Service, currentTime time.Time) (map[string]SportingEvent, map[string]string) {
	// The Google APIs deal with times in specific string formats.
	currentTimeString := time.Now().Format(time.RFC3339)

	// Retrieve ALL of the future calendar events
	var err error
	events, err := calendarService.Events.List(CalendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(currentTimeString).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve the events from the calendar: %v", err)
	}

	// Create two maps.
	// The first is for the SportingEvent struct
	// The second is for the Calendar Event ID.  This allows us to delete or update the event.
	// These could be merged into a single map with a tuple of values, but this seems more
	// straightforward at the small expense of extra memory.
	calendarFutureEvents := make(map[string]SportingEvent)
	calendarFutureEventIds := make(map[string]string)
	for _, item := range events.Items {
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}

		if !strings.Contains(item.Description, AutomationMarker) {
			continue // We didn't create this entry... skip it.
		}

		sportingEvent := getSportingEventFromCalendarEvent(item)

		// Although we specify TimeMin when getting the list of calendar events,
		// this includes events with an _end time_ that if after the TimeMin.
		// Need to check against the start time of the event to ensure calendar
		// events aren't deleted when the event is in progress during a run.
		if sportingEvent.Datetime.After(currentTime) {
			calendarFutureEvents[sportingEvent.GetKey()] = sportingEvent
			calendarFutureEventIds[sportingEvent.GetKey()] = item.Id
		}
	}
	return calendarFutureEvents, calendarFutureEventIds
}

func getSportingEventFromCalendarEvent(calendarEvent *calendar.Event) SportingEvent {
	var sportingEvent SportingEvent

	sportingEvent.Sport = calendarEvent.Summary
	description := strings.ReplaceAll(calendarEvent.Description, AutomationMarker, "")
	for _, role := range strings.Split(description, "\n") {
		// Calendar entries can have an extra blank line.  Let's suppress it.
		if len(role) > 2 {
			sportingEvent.Roles = append(sportingEvent.Roles, role)
		}
	}
	eastern, _ := time.LoadLocation("America/New_York")
	datetime, _ := time.ParseInLocation("2006-01-02T15:04:05-05:00", calendarEvent.Start.DateTime, eastern)

	// Parser seems to set a seconds value from the timezone.  Truncate to the minute to match the spreadsheet.
	datetime = datetime.Truncate(time.Minute)
	sportingEvent.Datetime = datetime

	return sportingEvent
}

func CreateCalendarEvent(calendarService *calendar.Service, calendarId string, sportingEvent SportingEvent) error {
	event := createCalendarEntryObject(sportingEvent)
	var err error
	event, err = calendarService.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	return err
}

func UpdateCalendarEvent(calendarService *calendar.Service, calendarId string, eventId string, sportingEvent SportingEvent) error {
	event := createCalendarEntryObject(sportingEvent)
	var err error
	event, err = calendarService.Events.Update(calendarId, eventId, event).Do()
	if err != nil {
		log.Fatalf("Unable to update event. %v\n", err)
	}
	return err
}

func createCalendarEntryObject(sportingEvent SportingEvent) *calendar.Event {
	description := ""
	for _, role := range sportingEvent.Roles {
		description += role + "\n"
	}
	description += AutomationMarker

	startTime := sportingEvent.Datetime
	endTime := startTime.Add(time.Hour * time.Duration(2))
	event := &calendar.Event{
		Summary:     sportingEvent.Sport,
		Location:    GetSportLocation(sportingEvent.Sport),
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "America/New_York",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "America/New_York",
		},
		// TODO add code to fill in the emails of people who opt-in to invites.
		//Attendees: []*calendar.EventAttendee{
		//	&calendar.EventAttendee{Email:"lpage@example.com"},
		//	&calendar.EventAttendee{Email:"sbrin@example.com"},
		//},
	}
	return event
}
